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

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
)

type client struct {
	client *govmomi.Client
	ctx    context.Context
	stopCh chan struct{}

	*affinityClient // affinity rules
	Querier         // cached or nocache
}

// NewCachedClient creates a cached vsphere client.
func NewCachedClient(clusterName string) Vsphere {
	ctx := context.Background()
	vsclient, err := NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	stopCh := make(chan struct{})

	clt := &client{
		client:         vsclient,
		ctx:            ctx,
		stopCh:         stopCh,
		affinityClient: newAffinityClient(clusterName),
		Querier:        newCachedQuerier(vsclient, stopCh),
	}

	// Init cluster object
	m := view.NewManager(vsclient.Client)

	v, err := m.CreateContainerView(ctx, vsclient.ServiceContent.RootFolder, []string{"ClusterComputeResource"}, true)
	if err != nil {
		log.Fatal(err)
	}

	defer v.Destroy(ctx)

	var clusters []mo.ClusterComputeResource
	err = v.Retrieve(ctx, []string{"ClusterComputeResource"}, []string{"name"}, &clusters)
	if err != nil {
		log.Fatal(err)
	}

	for i := range clusters {
		if clusters[i].Name == clusterName {
			clt.cluster = clusters[i].Reference()
			break
		}
	}

	if clt.cluster.Type == "" {
		log.Printf("[WARNING] cannot find cluster named %s", clusterName)
	}

	go clt.affinityClient.Run(stopCh)

	return clt
}

func (c *client) Client() *govmomi.Client {
	return c.client
}

func (c *client) Logout() {
	close(c.stopCh)
	c.client.Logout(c.ctx)
}
