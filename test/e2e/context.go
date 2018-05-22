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

package e2e

import (
	"log"
	"os/exec"

	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

type context struct {
	// cmd is the Cmd for plugin process
	cmd *exec.Cmd

	// k8sClient is the kubernetes client instance
	k8sClient kubernetes.Interface
}

func (ctx *context) GetNodes() []v1.Node {
	nodeList, err := ctx.k8sClient.Core().Nodes().List(meta_v1.ListOptions{})
	if err != nil {
		log.Fatalf("failed to get node list: %s", err)
	}

	return nodeList.Items
}

func (ctx *context) GetNodeNames() []string {
	nodes := ctx.GetNodes()
	nodeNames := []string{}
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.Name)
	}
	return nodeNames
}

/*
func (ctx *context) SetNodeToHost(nodeName, hostName string) {
	ctx.Lock()
	defer ctx.Unlock()

	if ctx.nodeToHost == nil {
		ctx.nodeToHost = make(map[string]string)
	}
	if ctx.hostToNodes == nil {
		ctx.hostToNodes = make(map[string][]string)
	}

	ctx.nodeToHost[nodeName] = hostName
	ctx.hostToNodes[hostName] = append(ctx.hostToNodes[hostName], nodeName)
}

func (ctx *context) GetNodeToHost() map[string]string {
	ctx.Lock()
	defer ctx.Unlock()

	return ctx.nodeToHost
}

func (ctx *context) GetHostToNodes() map[string][]string {
	ctx.Lock()
	defer ctx.Unlock()

	return ctx.hostToNodes
}
*/
