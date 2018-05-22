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
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/selector"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// CacheHandler is the interface for cache updates from external source.
type CacheHandler interface {
	Add(obj interface{})
	Update(old, new interface{})
	Delete(obj interface{})
}

// CreateHandler creates a cache.ResourceEventHandler from CacheHandler
func CreateHandler(h CacheHandler) cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    h.Add,
		UpdateFunc: h.Update,
		DeleteFunc: h.Delete,
	}
}

// SchedCache is the cache for scheduler
type SchedCache struct {
	*nodePodCache
	*hostLabelCache

	podInformer  cache.SharedIndexInformer
	nodeInformer cache.SharedIndexInformer
}

// New creates a SchedCache instance
func New(client kubernetes.Interface) *SchedCache {
	c := &SchedCache{
		nodePodCache:   newNodePodCache(),
		hostLabelCache: newHostLabelCache(),
	}

	lw := cache.NewListWatchFromClient(
		client.CoreV1().RESTClient(),
		"pods",
		meta_v1.NamespaceAll,
		fields.Everything())

	c.podInformer = cache.NewSharedIndexInformer(
		lw,
		&v1.Pod{},
		0, // skip resync
		cache.Indexers{},
	)

	c.podInformer.AddEventHandler(CreateHandler(c.nodePodCache))

	nodeLw := cache.NewListWatchFromClient(
		client.CoreV1().RESTClient(),
		"nodes",
		meta_v1.NamespaceAll,
		fields.Everything())

	c.nodeInformer = cache.NewSharedIndexInformer(
		nodeLw,
		&v1.Node{},
		0, // skip resync
		cache.Indexers{},
	)

	c.nodeInformer.AddEventHandler(CreateHandler(c.hostLabelCache))

	return c
}

// Run starts the shared informer in cache, which will be stopped when stopCh
// is closed.
func (c *SchedCache) Run(stopCh <-chan struct{}) {
	go c.podInformer.Run(stopCh)
	go c.nodeInformer.Run(stopCh)
}

// ListNode lists all nodes cached in SchedCached
func (c *SchedCache) ListNode() ([]*v1.Node, error) {
	result := []*v1.Node{}
	list := c.nodeInformer.GetStore().List()

	for _, node := range list {
		result = append(result, node.(*v1.Node))
	}

	return result, nil
}

// PodInformer returns the SharedIndexInformer for pods
func (c *SchedCache) PodInformer() cache.SharedIndexInformer {
	return c.podInformer
}

// NodeInformer returns the SharedIndexInformer for nodes
func (c *SchedCache) NodeInformer() cache.SharedIndexInformer {
	return c.nodeInformer
}

// ListPod list pods that matched the selector cached in SchedCache
func (c *SchedCache) ListPod(selector selector.Selector) ([]*v1.Pod, error) {
	result := []*v1.Pod{}
	list := c.podInformer.GetStore().List()

	for _, obj := range list {
		pod := obj.(*v1.Pod)
		if selector.Matches(labels.Set(pod.GetLabels())) {
			result = append(result, pod)
		}
	}

	return result, nil
}
