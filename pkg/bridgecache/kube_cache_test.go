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
	"testing"
	"time"

	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/test"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
)

func TestKubeCache(t *testing.T) {
	lw := test.NewFakeNodeListWatch()

	nodeInformer := cache.NewSharedIndexInformer(
		lw,
		&v1.Node{},
		0,
		cache.Indexers{},
	)

	cache := newKubeNodeCache(nodeInformer)

	stopCh := make(chan struct{})
	defer close(stopCh)
	go nodeInformer.Run(stopCh)

	err := wait.Poll(time.Millisecond, 5*time.Second, func() (bool, error) {
		return nodeInformer.HasSynced() == true, nil
	})
	if err != nil {
		t.Fatal("timeout waiting nodeInformer.HasSynced()")
	}

	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
			UID:  types.UID("uid-node1"),
		},
		Status: v1.NodeStatus{
			Addresses: []v1.NodeAddress{
				v1.NodeAddress{
					Type:    v1.NodeHostName,
					Address: "hostname1",
				},
			},
		},
	}

	lw.Add(node)

	err = wait.Poll(time.Millisecond, 5*time.Second, func() (bool, error) {
		return nodeInformer.HasSynced() == true, nil
	})
	if err != nil {
		t.Fatal("timeout waiting nodeInformer.HasSynced()")
	}

	if hostname := cache.GetHostnameFromNodeName("node1"); hostname != "hostname1" {
		t.Errorf("expect hostname == '%s'; got '%s'", "hostname1", hostname)
	}

	if nodename := cache.GetNodeNameFromHostname("hostname1"); nodename != "node1" {
		t.Errorf("expect nodename == '%s'; got '%s'", "node1", nodename)
	}

	// Test cache after updating node
	updatedHostname := "hostname-updated"
	node.Status.Addresses[0].Address = "hostname-updated"
	lw.Update(node)

	err = wait.Poll(time.Millisecond, 5*time.Second, func() (bool, error) {
		return nodeInformer.HasSynced() == true, nil
	})
	if err != nil {
		t.Fatal("timeout waiting nodeInformer.HasSynced()")
	}

	if hostname := cache.GetHostnameFromNodeName("node1"); hostname != updatedHostname {
		t.Errorf("expect hostname == '%s'; got '%s'", updatedHostname, hostname)
	}

	if nodename := cache.GetNodeNameFromHostname(updatedHostname); nodename != "node1" {
		t.Errorf("expect nodename == '%s'; got '%s'", "node1", nodename)
	}

	if nodename := cache.GetNodeNameFromHostname("hostname1"); nodename != "" {
		t.Errorf("expect nodename == '%s'; got '%s'", "", nodename)
	}

	// Test after deleting node
	lw.Delete(node)

	err = wait.Poll(time.Millisecond, 5*time.Second, func() (bool, error) {
		return nodeInformer.HasSynced() == true, nil
	})
	if err != nil {
		t.Fatal("timeout waiting nodeInformer.HasSynced()")
	}

	if hostname := cache.GetHostnameFromNodeName("node1"); hostname != "" {
		t.Errorf("expect hostname == '%s'; got '%s'", "", hostname)
	}

	if nodename := cache.GetNodeNameFromHostname("hostname1"); nodename != "" {
		t.Errorf("expect nodename == '%s'; got '%s'", "", nodename)
	}
}
