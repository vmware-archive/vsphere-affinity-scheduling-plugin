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
	"testing"

	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/k8s/client"
)

func exampleUpdater(t *testing.T) {
	client, err := client.New()
	if err != nil {
		t.Fatal(err)
	}

	updater := New(client)
	err = updater.Update("ip-10-0-14-19.us-west-1.compute.internal", "host1")
	fmt.Println(err)
}