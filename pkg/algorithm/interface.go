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

package algorithm

import (
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/selector"
	"k8s.io/api/core/v1"
)

// Filter filters nodes based on pod's spec, it returns valid nodes that the
// pod is okay to run on.
type Filter interface {
	Filter(pod *v1.Pod, nodes []string) ([]string, error)
}

// PodLister list pods
type PodLister interface {
	ListPod(selector.Selector) ([]*v1.Pod, error)
}

// HostCache keeps the node to host relationship, and supports host and node
// query.
type HostCache interface {
	// GetHost returns the hostname of a given node
	GetHost(node string) string

	// GetNodes returns all the nodes running on a given host
	GetNodes(host string) []string
}

// Filters is an ordered list of Filters
type Filters []Filter

// Filter implements interface Filter
func (filters Filters) Filter(pod *v1.Pod, nodes []string) ([]string, error) {
	var err error
	for _, filter := range filters {
		nodes, err = filter.Filter(pod, nodes)
		if err != nil {
			return nodes, err
		}
	}

	return nodes, err
}
