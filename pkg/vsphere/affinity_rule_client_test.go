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

package vsphere

import (
	"fmt"
	"testing"
	"time"
)

func exampleAffinityRuleClient(t *testing.T) {
	client := newAffinityClient("cluster1")
	stopCh := make(chan struct{})
	var err error

	go func() {
		_ = client.Run(stopCh)
	}()

	go func() {
		time.Sleep(5 * time.Second)
		close(stopCh)
	}()

	err = client.ApplyAffinityRule("affinity-1",
		"VirtualMachine:vm-29", "VirtualMachine:vm-26", "VirtualMachine:vm-30")
	fmt.Println(err)

	<-stopCh

	err = client.DeleteAffinityRule("affinity-1")
	fmt.Println(err)
}
