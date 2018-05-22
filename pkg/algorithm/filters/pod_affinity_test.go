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

package filters

import (
	"reflect"
	"testing"

	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/algorithm/fake"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodAffinity(t *testing.T) {
	tests := []struct {
		desc       string
		pod        *v1.Pod
		pods       []*v1.Pod
		nodeToHost map[string]string
		nodes      []string
		expect     []string
	}{
		{
			desc: "pod affinity; pod matchLabels; nodes on different hosts.",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					Affinity: &v1.Affinity{
						PodAffinity: &v1.PodAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{"key": "value"},
										// MatchExpressions
									},
									Namespaces:  []string{},
									TopologyKey: HostTopologyKey,
								},
							},
						},
					},
				},
			},
			pods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"key": "value"},
					},
					Spec: v1.PodSpec{
						NodeName: "node1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"key": "value2"},
					},
					Spec: v1.PodSpec{
						NodeName: "node2",
					},
				},
			},
			nodeToHost: map[string]string{
				"node1": "host1",
				"node2": "host2",
			},
			nodes:  []string{"node1", "node2"},
			expect: []string{"node1"},
		},
		{
			desc: "pod affinity; pod matchLabels; nodes on same hosts.",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					Affinity: &v1.Affinity{
						PodAffinity: &v1.PodAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{"key": "value"},
										// MatchExpressions
									},
									Namespaces:  []string{},
									TopologyKey: HostTopologyKey,
								},
							},
						},
					},
				},
			},
			pods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"key": "value"},
					},
					Spec: v1.PodSpec{
						NodeName: "node1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"key": "value2"},
					},
					Spec: v1.PodSpec{
						NodeName: "node3",
					},
				},
			},
			nodeToHost: map[string]string{
				"node1": "host1",
				"node2": "host1",
				"node3": "host2",
			},
			nodes:  []string{"node1", "node2", "node3"},
			expect: []string{"node1", "node2"},
		},
	}

	for _, test := range tests {
		filter := &podAffinity{
			podLister: fake.NewPodLister(test.pods),
			hostCache: fake.NodeCache(test.nodeToHost),
		}
		result, err := filter.Filter(test.pod, test.nodes)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(result, test.expect) {
			t.Errorf("[%s] expect %s; got %s", test.desc, test.expect, result)
		}
	}
}
