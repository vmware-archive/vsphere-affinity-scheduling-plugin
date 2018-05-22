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

package bridgecache

import (
	"log"

	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/vsphere"
	"k8s.io/client-go/tools/cache"
)

// Cache provides bridging information between vSphere and Kubernetes in cache.
// Specifically, when client has a Kubernetes node name, he wants to know the
// Corresponding VM ID in vSphere, this Cache will provide this information.
type Cache interface {
	// GetVMIDFromNode returns virtual machine from vSphere
	GetVMIDFromNode(name string) string
}

// cacheStore is an implementation of Cache.
//
// cacheStore keeps watching vSphere and Kubernetes to collect updates and save
// them internally.
type cacheStore struct {
	*kubeNodeCache
	vsphere.Querier
}

// NewCache returns a Cache instance
func NewCache(nodeInformer cache.SharedIndexInformer, querier vsphere.Querier) Cache {
	return &cacheStore{
		Querier:       querier,
		kubeNodeCache: newKubeNodeCache(nodeInformer),
	}
}

// GetVMIDFromNode returns virtual machine from vSphere
func (c *cacheStore) GetVMIDFromNode(name string) string {
	hostname := c.GetHostnameFromNodeName(name)
	log.Printf("node,hostname: %s,%s", name, hostname)
	if hostname == "" {
		return ""
	}

	return c.GetVMIDFromHostname(hostname)
}
