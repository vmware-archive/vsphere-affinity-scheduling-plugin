/*
Copyright (c) 201８ VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vsphere

import (
	"context"
	"log"
	"sync"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type cachedQuerier struct {
	client *govmomi.Client

	sync.Mutex
	hostnameToVMID map[string]string
	vmidToHostname map[string]string
	vmidToHost     map[string]*mo.HostSystem

	// internal cache
	hostCache map[types.ManagedObjectReference]*mo.HostSystem
}

// newCachedQuerier creates a cached querier
func newCachedQuerier(client *govmomi.Client, stopCh <-chan struct{}) Querier {
	c := &cachedQuerier{
		client:         client,
		hostnameToVMID: make(map[string]string),
		vmidToHostname: make(map[string]string),
		vmidToHost:     make(map[string]*mo.HostSystem),
		hostCache:      make(map[types.ManagedObjectReference]*mo.HostSystem),
	}

	go c.Run(stopCh)

	return c
}

func (c *cachedQuerier) GetHostnameFromVMID(vmid string) string {
	c.Lock()
	defer c.Unlock()
	return c.vmidToHostname[vmid]
}

func (c *cachedQuerier) GetVMIDFromHostname(hostname string) string {
	c.Lock()
	defer c.Unlock()
	return c.hostnameToVMID[hostname]
}

func (c *cachedQuerier) GetHostFromVMID(vmid string) (string, error) {
	c.Lock()
	defer c.Unlock()

	if host, ok := c.vmidToHost[vmid]; ok {
		return host.Name, nil
	}
	return "", nil
}

func (c *cachedQuerier) Run(stopCh <-chan struct{}) {
	// Create view of VirtualMachine objects
	m := view.NewManager(c.client.Client)
	ctx := context.Background()

	v, err := m.CreateContainerView(ctx, c.client.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		log.Fatal(err)
	}

	defer v.Destroy(ctx)

	filter := new(property.WaitFilter)
	filter.Add(v.Reference(), "VirtualMachine", []string{"runtime.host", "summary.guest.hostName"}, v.TraversalSpec())

	property.WaitForUpdates(ctx, c.client.PropertyCollector(), filter, func(updates []types.ObjectUpdate) bool {
		c.Lock()
		defer c.Unlock()

		for _, update := range updates {
			// FIXME: Can I assume PropertyChangeOp is always assign?

			switch update.Kind {
			case types.ObjectUpdateKindModify:
				fallthrough
			case types.ObjectUpdateKindEnter:
				for _, cs := range update.ChangeSet {
					log.Printf("vsphere: update %s %s", update.Obj.String(), cs.Name)
					if cs.Name == "summary.guest.hostName" && cs.Val != nil {
						hostname := cs.Val.(string)
						log.Printf("vsphere: cache update, vmid<=>hostname, %s<=>%s", update.Obj.String(), hostname)
						c.vmidToHostname[update.Obj.String()] = hostname
						c.hostnameToVMID[hostname] = update.Obj.String()
					} else if cs.Name == "runtime.host" && cs.Val != nil {
						moref := cs.Val.(types.ManagedObjectReference)
						c.vmidToHost[update.Obj.String()] = c.getHost(moref)
					}
				}
			case types.ObjectUpdateKindLeave:
				log.Printf("vsphere: delete %s", update.Obj.String())
				if hostname, ok := c.vmidToHostname[update.Obj.String()]; ok {
					delete(c.vmidToHostname, update.Obj.String())
					delete(c.hostnameToVMID, hostname)
				}
				delete(c.vmidToHost, update.Obj.String())
			}
		}

		select {
		case <-stopCh:
			return true
		default:
		}

		return false // keep waiting
	})
}

func (c *cachedQuerier) getHost(ref types.ManagedObjectReference) *mo.HostSystem {
	if host, ok := c.hostCache[ref]; ok {
		return host
	}

	pc := property.DefaultCollector(c.client.Client)
	dst := &mo.HostSystem{}
	err := pc.RetrieveOne(context.Background(), ref, []string{"name"}, dst)
	if err != nil {
		log.Printf("vsphere: failed to retrieve hostsystem info")
		return dst
	}

	c.hostCache[ref] = dst
	return dst
}
