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

	"k8s.io/api/core/v1"
)

// nodePodCache keeps node-to-pod assignment based on node name
type nodePodCache struct {
	// podsOnNode save the set of pod UID per node
	podsOnNode map[string]map[string]struct{}

	sync.Mutex
}

func newNodePodCache() *nodePodCache {
	return &nodePodCache{
		podsOnNode: make(map[string]map[string]struct{}),
	}
}

// Add adds a new pod into cache
func (c *nodePodCache) Add(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		log.Printf("cannot convert to *v1.Pod: %v", obj)
		return
	}

	c.Lock()
	defer c.Unlock()

	if pod.Spec.NodeName != "" {
		c.doAdd(pod)
	}
}

// Update updates old pod with new pod in cache
func (c *nodePodCache) Update(oldObj, newObj interface{}) {
	oldPod, ok := oldObj.(*v1.Pod)
	if !ok {
		log.Printf("cannot convert oldObj to *v1.Pod: %v", oldObj)
		return
	}

	pod, ok := newObj.(*v1.Pod)
	if !ok {
		log.Printf("cannot convert newObj to *v1.Pod: %v", newObj)
		return
	}

	if oldPod.UID != pod.UID {
		log.Printf("cannot update pod with different UID: %s != %s", oldPod.UID, pod.UID)
		return
	}

	c.Lock()
	defer c.Unlock()

	if oldPod.Spec.NodeName != "" && pod.Spec.NodeName == "" {
		c.doDelete(oldPod)
	} else if oldPod.Spec.NodeName == "" && pod.Spec.NodeName != "" {
		c.doAdd(pod)
	} else if oldPod.Spec.NodeName != pod.Spec.NodeName {
		c.doDelete(oldPod)
		c.doAdd(pod)
	}
}

// Delete delete a pod from cache
func (c *nodePodCache) Delete(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		log.Printf("cannot convert to *v1.Pod: %v", obj)
		return
	}

	c.Lock()
	defer c.Unlock()

	if pod.Spec.NodeName != "" {
		c.doDelete(pod)
	}
}

func (c *nodePodCache) doAdd(pod *v1.Pod) {
	nodename := pod.Spec.NodeName
	log.Printf("adding pod to node cache (%s => %s)", pod.UID, nodename)

	if set, ok := c.podsOnNode[nodename]; ok {
		set[string(pod.UID)] = struct{}{}
	} else {
		c.podsOnNode[nodename] = map[string]struct{}{
			string(pod.UID): struct{}{},
		}
	}
}

func (c *nodePodCache) doDelete(pod *v1.Pod) {
	nodename := pod.Spec.NodeName
	log.Printf("deleting pod to node cache (%s => %s)", pod.UID, nodename)

	delete(c.podsOnNode[nodename], string(pod.UID))
	if len(c.podsOnNode[nodename]) == 0 {
		delete(c.podsOnNode, nodename)
	}
}

func (c *nodePodCache) GetPodsOnNode(nodename string) []string {
	c.Lock()
	c.Unlock()

	if m, ok := c.podsOnNode[nodename]; ok {
		result := make([]string, 0, len(m))
		for p := range m {
			result = append(result, p)
		}
		return result
	}

	return []string{}
}
