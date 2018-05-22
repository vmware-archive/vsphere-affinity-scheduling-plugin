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
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestNodePodCache_Add(t *testing.T) {
	cache := newNodePodCache()
	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod1",
			UID:  types.UID("uid-pod1"),
		},
		Spec: v1.PodSpec{
			NodeName: "node1",
		},
	}

	pods := cache.GetPodsOnNode("node1")
	if len(pods) != 0 {
		t.Errorf("expect []; got %s", pods)
	}

	cache.Add(pod1)

	pods = cache.GetPodsOnNode("node1")
	if len(pods) != 1 || pods[0] != "uid-pod1" {
		t.Errorf("expect [uid-pod1]; got %s", pods)
	}
}

func TestNodePodCache_Update(t *testing.T) {
	cache := newNodePodCache()
	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod1",
			UID:  types.UID("uid-pod1"),
		},
		Spec: v1.PodSpec{
			NodeName: "node1",
		},
	}

	cache.Add(pod1)

	pod2 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod1",
			UID:  types.UID("uid-pod1"),
		},
		Spec: v1.PodSpec{
			NodeName: "node2",
		},
	}

	cache.Update(pod1, pod2)

	pods := cache.GetPodsOnNode("node1")
	if len(pods) != 0 {
		t.Errorf("expect []; got %s", pods)
	}

	pods = cache.GetPodsOnNode("node2")
	if len(pods) != 1 || pods[0] != "uid-pod1" {
		t.Errorf("expect [uid-pod1]; got %s", pods)
	}
}

func TestNodePodCache_Delete(t *testing.T) {
	cache := newNodePodCache()
	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod1",
			UID:  types.UID("uid-pod1"),
		},
		Spec: v1.PodSpec{
			NodeName: "node1",
		},
	}

	cache.Add(pod1)
	cache.Delete(pod1)

	pods := cache.GetPodsOnNode("node1")
	if len(pods) != 0 {
		t.Errorf("expect []; got %s", pods)
	}
}
