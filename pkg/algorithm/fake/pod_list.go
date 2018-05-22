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

package fake

import (
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/algorithm"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/selector"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type podLister struct {
	pods []*v1.Pod
}

// NewPodLister creates a fake PodLister
func NewPodLister(pods []*v1.Pod) algorithm.PodLister {
	return &podLister{pods}
}

func (l *podLister) ListPod(selector selector.Selector) ([]*v1.Pod, error) {
	result := []*v1.Pod{}
	for _, pod := range l.pods {
		if selector.Matches(labels.Set(pod.GetLabels())) {
			result = append(result, pod)
		}
	}
	return result, nil
}
