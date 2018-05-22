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
	"sync"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// kubeNodeCache keeps the mapping between node name and hostname of the node
type kubeNodeCache struct {
	hostnameToKubeNode map[string]string
	kubeNodeToHostname map[string]string
	sync.Mutex
}

// newKubeNodeCache creates a kubeNodeCache instance
func newKubeNodeCache(nodeInformer cache.SharedIndexInformer) *kubeNodeCache {
	c := &kubeNodeCache{
		hostnameToKubeNode: make(map[string]string),
		kubeNodeToHostname: make(map[string]string),
	}

	nodeInformer.AddEventHandler(c)

	return c
}

// GetHostnameFromNodeName returns hostname that matches the node name
// in cache.
func (c *kubeNodeCache) GetHostnameFromNodeName(nodename string) string {
	c.Lock()
	defer c.Unlock()

	return c.kubeNodeToHostname[nodename]
}

// GetNodeNameFromHostname returns node name that matches the hostname
// in cache.
func (c *kubeNodeCache) GetNodeNameFromHostname(hostname string) string {
	c.Lock()
	defer c.Unlock()

	return c.hostnameToKubeNode[hostname]
}

// OnAdd is the callback that gets triggered when informer informs the
// event of object addition.
func (c *kubeNodeCache) OnAdd(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		return
	}

	log.Printf("kubeNodeCache: OnAdd(%s)", node.Name)

	hostname := getHostname(node)
	if hostname == "" {
		return
	}

	c.Lock()
	defer c.Unlock()

	c.add(node, hostname)
}

// add adds a Node in cache. Caller needs to own the lock.
func (c *kubeNodeCache) add(node *v1.Node, hostname string) {
	c.hostnameToKubeNode[hostname] = node.Name
	c.kubeNodeToHostname[node.Name] = hostname
}

// OnDelete is the callback that gets triggered when informer informs the
// event of object deletion.
func (c *kubeNodeCache) OnDelete(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		return
	}

	log.Printf("kubeNodeCache: OnDelete(%s)", node.Name)

	c.Lock()
	defer c.Unlock()

	c.delete(node)
}

// delete deletes a Node from cache. Caller needs to own the lock.
func (c *kubeNodeCache) delete(node *v1.Node) {
	hostname := c.kubeNodeToHostname[node.Name]
	delete(c.kubeNodeToHostname, node.Name)
	delete(c.hostnameToKubeNode, hostname)
}

// OnDelete is the callback that gets triggered when informer informs the
// event of object update.
func (c *kubeNodeCache) OnUpdate(old, new interface{}) {
	oldNode, ok := old.(*v1.Node)
	if !ok {
		return
	}

	newNode, ok := new.(*v1.Node)
	if !ok {
		return
	}

	log.Printf("kubeNodeCache: OnUpdate(%s)", oldNode.Name)

	hostname := getHostname(newNode)

	c.Lock()
	defer c.Unlock()

	c.delete(oldNode)
	c.add(newNode, hostname)
}

// getHostname retrieves the hostname from Node status.
func getHostname(node *v1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type == v1.NodeHostName {
			return address.Address
		}
	}
	return ""
}
