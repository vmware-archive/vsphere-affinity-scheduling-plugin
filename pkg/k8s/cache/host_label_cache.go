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

package k8scache

import (
	"log"
	"sync"

	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/constants"
	"k8s.io/api/core/v1"
)

type hostLabelCache struct {
	nodeToHost  map[string]string
	hostToNodes map[string]map[string]struct{}

	sync.Mutex
}

func newHostLabelCache() *hostLabelCache {
	return &hostLabelCache{
		nodeToHost:  make(map[string]string),
		hostToNodes: make(map[string]map[string]struct{}),
	}
}

func (c *hostLabelCache) Add(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		log.Printf("cannot convert to *v1.Node: %v", obj)
		return
	}

	c.Lock()
	defer c.Unlock()

	if host, ok := node.Labels[constants.HostLabel]; ok {
		c.addNodeHost(node, host)
	}
}

func (c *hostLabelCache) Update(oldObj, newObj interface{}) {
	oldnode, ok := oldObj.(*v1.Node)
	if !ok {
		log.Printf("cannot convert oldObj to *v1.Node: %v", oldObj)
		return
	}

	node, ok := newObj.(*v1.Node)
	if !ok {
		log.Printf("cannot convert newObj to *v1.Node: %v", newObj)
		return
	}

	if oldnode.Name != node.Name {
		log.Printf("cannot update node with different Name: %s != %s", oldnode.Name, node.Name)
		return
	}

	c.Lock()
	defer c.Unlock()

	log.Printf("hostLabelCache: updating node %s", node.Name)
	if oldnode.Labels[constants.HostLabel] != node.Labels[constants.HostLabel] {
		c.removeNodeHost(node, oldnode.Labels[constants.HostLabel])
		c.addNodeHost(node, node.Labels[constants.HostLabel])
	}
}

func (c *hostLabelCache) Delete(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		log.Printf("cannot convert to *v1.Node: %v", obj)
		return
	}

	c.Lock()
	defer c.Unlock()

	if host, ok := node.Labels[constants.HostLabel]; ok {
		c.removeNodeHost(node, host)
	}
}

func (c *hostLabelCache) addNodeHost(node *v1.Node, host string) {
	if host == "" {
		return
	}

	log.Printf("hostLabelCache: adding node to host mapping %s=>%s", node.Name, host)
	c.nodeToHost[node.Name] = host
	if _, ok := c.hostToNodes[host]; !ok {
		c.hostToNodes[host] = make(map[string]struct{})
	}
	c.hostToNodes[host][node.Name] = struct{}{}
}

func (c *hostLabelCache) removeNodeHost(node *v1.Node, host string) {
	if host == "" {
		return
	}

	log.Printf("hostLabelCache: removing node to host mapping %s=>%s", node.Name, host)
	delete(c.nodeToHost, node.Name)
	delete(c.hostToNodes[host], node.Name)
}

// GetHost returns the hostname of a given node
func (c *hostLabelCache) GetHost(node string) string {
	return c.nodeToHost[node]
}

// GetNodes returns all the nodes running on a given host
func (c *hostLabelCache) GetNodes(host string) []string {
	result := []string{}

	for k := range c.hostToNodes[host] {
		result = append(result, k)
	}

	return result
}
