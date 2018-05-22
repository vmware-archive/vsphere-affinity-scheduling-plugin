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

package services

import (
	"reflect"
	"sort"
	"testing"

	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/algorithm/fake"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/constants"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/test"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/vsphere"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestDRSRulerDesiredRules(t *testing.T) {
	ruler := &DRSRuler{
		affinityPods:     make(map[string]*v1.Pod),
		antiAffinityPods: make(map[string]*v1.Pod),
	}

	anchorPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"type": "anchor"},
		},
		Spec: v1.PodSpec{
			NodeName: "node0",
		},
	}

	affinityPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID: types.UID("pod-uid1"),
		},
		Spec: v1.PodSpec{
			NodeName: "node1",
			Affinity: &v1.Affinity{
				PodAffinity: &v1.PodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
						{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"type": "anchor"},
							},
							Namespaces:  []string{},
							TopologyKey: constants.HostLabel,
						},
					},
				},
			},
		},
	}

	affinityPod2 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID: types.UID("pod-uid2"),
		},
		Spec: v1.PodSpec{
			NodeName: "node2",
			Affinity: &v1.Affinity{
				PodAntiAffinity: &v1.PodAntiAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
						{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"type": "anchor"},
							},
							Namespaces:  []string{},
							TopologyKey: constants.HostLabel,
						},
					},
				},
			},
		},
	}

	ruler.OnAdd(anchorPod)
	ruler.OnAdd(affinityPod)
	ruler.OnAdd(affinityPod2)

	ruler.podLister = fake.NewPodLister([]*v1.Pod{
		anchorPod,
		affinityPod,
		affinityPod2,
	})
	ruler.bcache = test.FakeBCache(map[string]string{
		"node0": "vm0",
		"node1": "vm1",
		"node2": "vm2",
	})

	rules := ruler.desiredRules()

	expected := map[string]vsphere.Rule{
		"affi-pod-uid1": vsphere.Rule{
			Name:     "affi-pod-uid1",
			VMs:      []string{"vm0", "vm1"},
			Affinity: true,
		},
		"anti-pod-uid2": vsphere.Rule{
			Name:     "anti-pod-uid2",
			VMs:      []string{"vm0", "vm2"},
			Affinity: false,
		},
	}

	for _, rule := range rules {
		sort.Strings(rule.VMs)
	}
	if !reflect.DeepEqual(expected, rules) {
		t.Errorf("expected desiredRules=%+v; got %+v", expected, rules)
	}
}

func TestDRSRulerHandler(t *testing.T) {
	ruler := &DRSRuler{
		affinityPods:     make(map[string]*v1.Pod),
		antiAffinityPods: make(map[string]*v1.Pod),
	}

	affinityPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID: types.UID("pod-uid1"),
		},
		Spec: v1.PodSpec{
			NodeName: "node1",
			Affinity: &v1.Affinity{
				PodAffinity: &v1.PodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{},
				},
			},
		},
	}

	// OnAdd affinityPod
	ruler.OnAdd(affinityPod)

	if len(ruler.affinityPods) != 1 {
		t.Errorf("expect len(affinityPods)==1; got %d", len(ruler.affinityPods))
	}
	if ruler.affinityPods["pod-uid1"] != affinityPod {
		t.Errorf("expect affinityPod=%+v; got %+v", affinityPod, ruler.affinityPods["pod-uid1"])
	}

	antiAffinityPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID: types.UID("pod-uid2"),
		},
		Spec: v1.PodSpec{
			NodeName: "node1",
			Affinity: &v1.Affinity{
				PodAntiAffinity: &v1.PodAntiAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{},
				},
			},
		},
	}

	// OnAdd anti-affinity pod
	ruler.OnAdd(antiAffinityPod)

	if len(ruler.antiAffinityPods) != 1 {
		t.Errorf("expect len(affinityPods)==1; got %d", len(ruler.antiAffinityPods))
	}
	if ruler.antiAffinityPods["pod-uid2"] != antiAffinityPod {
		t.Errorf("expect affinityPod=%+v; got %+v", antiAffinityPod, ruler.antiAffinityPods["pod-uid2"])
	}

	ruler.OnDelete(affinityPod)
	ruler.OnDelete(antiAffinityPod)

	if len(ruler.affinityPods) != 0 {
		t.Errorf("expect len(affinityPods)==0; got %d", len(ruler.affinityPods))
	}
	if len(ruler.antiAffinityPods) != 0 {
		t.Errorf("expect len(affinityPods)==0; got %d", len(ruler.antiAffinityPods))
	}
}
