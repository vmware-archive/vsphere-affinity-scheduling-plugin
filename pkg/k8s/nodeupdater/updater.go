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

package nodeupdater

import (
	"fmt"
	"strings"

	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/constants"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
)

const patchTemplate = `
[
  {
    "op": "add",
    "path": "/metadata/labels/%s",
    "value": "%s"
  }
]
`

const addTemplate = `
[
  {
    "op": "add",
    "path": "/metadata/labels/%s",
    "value": "%s"
  }
]
`

const removeTemplate = `
[
  {
    "op": "remove",
    "path": "/metadata/labels/%s"
  }
]
`

// NodeUpdater updates a node with a label to indicate the physical host it is
// running on
type NodeUpdater interface {
	// Update labels the Kubernetes node with the pre-defined label to indicate
	// the physical host of the node.
	Update(node, host string) error

	// DeleteLabel removes the host label from the node
	DeleteLabel(node string) error
}

type nodeUpdater struct {
	nodeIfc v1core.NodeInterface
}

// New creates a NodeUpdater instance
func New(client kubernetes.Interface) NodeUpdater {
	return &nodeUpdater{
		nodeIfc: client.Core().Nodes(),
	}
}

// Update labels the Kubernetes node with the pre-defined label to indicate
// the physical host of the node.
func (u *nodeUpdater) Update(nodeName, hostName string) error {
	add := []byte(fmt.Sprintf(addTemplate, hostLabel, hostName))
	_, err := u.nodeIfc.Patch(nodeName, types.JSONPatchType, add)
	if err == nil {
		return nil
	}

	patch := []byte(fmt.Sprintf(patchTemplate, hostLabel, hostName))
	_, err = u.nodeIfc.Patch(nodeName, types.JSONPatchType, patch)

	return err
}

func (u *nodeUpdater) DeleteLabel(nodeName string) error {
	remove := []byte(fmt.Sprintf(removeTemplate, hostLabel))
	_, _ = u.nodeIfc.Patch(nodeName, types.JSONPatchType, remove)

	return nil
}

var hostLabel = escape(constants.HostLabel)

// escape is not a complete escape function. It simply replace `/` to `~1`.
func escape(path string) string {
	return strings.Replace(path, "/", "~1", -1)
}
