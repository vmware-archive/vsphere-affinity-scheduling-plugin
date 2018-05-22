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

package client

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// New tries to creates a client in cluster as first choice, if it failed,
// it tries it as kubectl does on client side.
func New() (kubernetes.Interface, error) {
	if client, err := NewInCluster(); err == nil {
		return client, err
	}

	return NewClient()
}

// NewClient creates a clientset to talk to apiserver
func NewClient() (kubernetes.Interface, error) {
	path := os.Getenv("KUBECONFIG")
	if path == "" {
		path = filepath.Join(os.Getenv("HOME"), ".kube/config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

// NewInCluster creates a clientset if it runs inside a cluster. It will get
// serviceaccount token to construct a clientset.
func NewInCluster() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}
