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

package test

import (
	"sync"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

// FakeNodeListWatch implements an in memory ListWatch for node, mainly for
// testing purpose
type FakeNodeListWatch struct {
	m        map[string]*v1.Node
	watchers []*watch.FakeWatcher
	sync.Mutex
}

// NewFakeNodeListWatch creates an instance of FakeNodeListWatch
func NewFakeNodeListWatch() *FakeNodeListWatch {
	return &FakeNodeListWatch{
		m:        make(map[string]*v1.Node),
		watchers: make([]*watch.FakeWatcher, 0),
	}
}

// List returns a list representation of objects. The ListOptions is ignored.
func (lw *FakeNodeListWatch) List(_ metav1.ListOptions) (runtime.Object, error) {
	lw.Lock()
	defer lw.Unlock()

	list := &v1.NodeList{}
	for _, node := range lw.m {
		list.Items = append(list.Items, *node)
	}
	return list, nil
}

// Watch returns a watcher. The ListOptions is ignored.
func (lw *FakeNodeListWatch) Watch(_ metav1.ListOptions) (watch.Interface, error) {
	lw.Lock()
	defer lw.Unlock()

	newWatch := watch.NewFake()
	for _, node := range lw.m {
		newWatch.Add(node)
	}

	lw.watchers = append(lw.watchers, newWatch)

	return newWatch, nil
}

// Add adds a node to FakeNodeListWatch
func (lw *FakeNodeListWatch) Add(obj *v1.Node) {
	lw.Lock()
	defer lw.Unlock()

	lw.m[string(obj.GetUID())] = obj
	for _, w := range lw.watchers {
		w.Add(obj)
	}
}

// Update updates a node in FakeNodeListWatch
func (lw *FakeNodeListWatch) Update(obj *v1.Node) {
	lw.Lock()
	defer lw.Unlock()

	lw.m[string(obj.GetUID())] = obj
	for _, w := range lw.watchers {
		w.Modify(obj)
	}
}

// Delete deletes a node from FakeNodeListWatch
func (lw *FakeNodeListWatch) Delete(obj *v1.Node) {
	lw.Lock()
	defer lw.Unlock()

	delete(lw.m, string(obj.GetUID()))
	for _, w := range lw.watchers {
		w.Delete(obj)
	}
}
