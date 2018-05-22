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
	"reflect"
	"sort"
	"testing"

	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/constants"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHostLabelCache_Add(t *testing.T) {
	cache := newHostLabelCache()
	node1 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node1",
			Labels: map[string]string{constants.HostLabel: "host1"},
		},
	}

	cache.Add(node1)

	if host := cache.GetHost("node1"); host != "host1" {
		t.Errorf("expect cache.GetHost return %s; got %s", "host1", host)
	}
	if nodes := cache.GetNodes("host1"); !reflect.DeepEqual(nodes, []string{"node1"}) {
		t.Errorf("expect cache.GetNodes return %s; got %s", []string{"node1"}, nodes)
	}

	node2 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node2",
			Labels: map[string]string{constants.HostLabel: "host1"},
		},
	}

	cache.Add(node2)

	if host := cache.GetHost("node2"); host != "host1" {
		t.Errorf("expect cache.GetHost return %s; got %s", "host1", host)
	}

	nodes := cache.GetNodes("host1")
	sort.Strings(nodes)
	if !reflect.DeepEqual(nodes, []string{"node1", "node2"}) {
		t.Errorf("expect cache.GetNodes return %s; got %s", []string{"node1", "node2"}, nodes)
	}
}

func TestHostLabelCache_Update(t *testing.T) {
	cache := newHostLabelCache()
	node1 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node1",
			Labels: map[string]string{constants.HostLabel: "host1"},
		},
	}

	cache.Add(node1)

	node2 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node1",
			Labels: map[string]string{constants.HostLabel: "host2"},
		},
	}

	cache.Update(node1, node2)

	if host := cache.GetHost("node1"); host != "host2" {
		t.Errorf("expect cache.GetHost return %s; got %s", "host2", host)
	}

	nodes := cache.GetNodes("host1")
	if len(nodes) != 0 {
		t.Errorf("expect cache.GetNodes return empty list for old node; got %s", nodes)
	}

	nodes = cache.GetNodes("host2")
	if !reflect.DeepEqual(nodes, []string{"node1"}) {
		t.Errorf("expect cache.GetNodes return %s; got %s", []string{"node1"}, nodes)
	}
}

func TestHostLabelCache_Delete(t *testing.T) {
	cache := newHostLabelCache()
	node1 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node1",
			Labels: map[string]string{constants.HostLabel: "host1"},
		},
	}
	node2 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node2",
			Labels: map[string]string{constants.HostLabel: "host1"},
		},
	}

	cache.Add(node1)
	cache.Add(node2)
	cache.Delete(node2)

	if host := cache.GetHost("node1"); host != "host1" {
		t.Errorf("expect cache.GetHost return %s; got %s", "host1", host)
	}
	if nodes := cache.GetNodes("host1"); !reflect.DeepEqual(nodes, []string{"node1"}) {
		t.Errorf("expect cache.GetNodes return %s; got %s", []string{"node1"}, nodes)
	}
}
