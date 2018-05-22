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

// NodeCache keeps node to host mapping
type NodeCache map[string]string

// GetHost returns the hostname of a given node
func (c NodeCache) GetHost(node string) string {
	return c[node]
}

// GetNodes returns all the nodes running on a given host
func (c NodeCache) GetNodes(host string) []string {
	result := []string{}

	for k, v := range c {
		if v == host {
			result = append(result, k)
		}
	}

	return result
}
