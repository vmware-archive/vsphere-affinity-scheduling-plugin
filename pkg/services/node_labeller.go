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

package services

import (
	"log"
	"time"

	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/bridgecache"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/k8s/cache"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/k8s/nodeupdater"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/vsphere"
)

// NodeLabeller updates node's host information based on watching on vSphere.
type NodeLabeller struct {
	Interval time.Duration

	nodeUpdater nodeupdater.NodeUpdater
	nodeLister  k8scache.NodeLister
	vsclient    vsphere.Vsphere
	bridgecache bridgecache.Cache
}

// NewNodeLabeller creates a NodeLabeller
func NewNodeLabeller(nodeLister k8scache.NodeLister, nodeUpdater nodeupdater.NodeUpdater,
	vsclient vsphere.Vsphere, bridgecache bridgecache.Cache) *NodeLabeller {
	return &NodeLabeller{
		nodeLister:  nodeLister,
		nodeUpdater: nodeUpdater,
		vsclient:    vsclient,
		bridgecache: bridgecache,
		Interval:    10 * time.Second,
	}
}

// Run starts a service until stopCh is closed
func (n *NodeLabeller) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(n.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			n.doLabel()
		case <-stopCh:
			log.Println("service exits: NodeLabeller")
			return
		}
	}
}

func (n *NodeLabeller) doLabel() {
	log.Println("NodeLabeller start labelling")
	defer log.Println("NodeLabeller finish labelling")

	nodes, err := n.nodeLister.ListNode()
	if err != nil {
		log.Printf("[ERROR] list node from k8scache, %s", err)
		return
	}

	for _, node := range nodes {
		vmid := n.bridgecache.GetVMIDFromNode(node.Name)
		log.Printf("node,vmid: %s, %s", node.Name, vmid)

		host, err := n.vsclient.GetHostFromVMID(vmid)
		if err != nil {
			log.Printf("[ERROR] failed to get vsphere host from vmid: %s", err)
			continue
		}

		err = n.nodeUpdater.Update(node.Name, host)
		if err != nil {
			log.Printf("[ERROR] failed to update k8s node %s with host %s",
				node.Name, host)
		}
	}
}
