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
	"github.com/vmware/govmomi"
)

// Vsphere is the vSphere client to talk to vSphere
type Vsphere interface {
	// Apply affinity rule to a list of vms so that DRS will schedule them on the
	// same host
	ApplyAffinityRule(name string, vms ...string) error

	// Apply anti-affinity rule to a list of vms so that DRS will schedule them on
	// different hosts
	ApplyAntiAffinityRule(name string, vms ...string) error

	// DeleteAffinityRule deletes an affinity rule
	DeleteAffinityRule(name string) error

	// DeleteAntiAffinityRule deletes an anti-affinity rule
	DeleteAntiAffinityRule(name string) error

	// Rules returns the applied VM-to-VM affinity and anti-affinity rules
	Rules() map[string]Rule

	// Logout signs off the session
	Logout()

	// Client returns the govmomi client
	Client() *govmomi.Client

	Querier
}

// Querier is a read-only interface for client to query vSphere.
//
// VMID: unique reference id inside vSphere to represent VM as a managed object.
//       It's a string that looks like this: "VirtualMachine:vm-1".
type Querier interface {
	// GetHostFromVMID gets ESX server's hostname from VMID.
	GetHostFromVMID(vmid string) (string, error)

	// GetHostnameFromVMID gets the hostname of a virtual machine identified by
	// VMID. Empty string will be returned if it isn't found.
	GetHostnameFromVMID(hostname string) string

	// GetVMIDFromHostname gets the vSphere VMID from the virtual machine's
	// hostname. Empty string will be returned if it isn't found.
	GetVMIDFromHostname(vmid string) string
}
