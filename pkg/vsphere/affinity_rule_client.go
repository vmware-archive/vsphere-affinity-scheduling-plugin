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
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// FIXME: there is a small window of race condition in this implementation
// where the cache of rules is not yet update-to-date, where a delete request
// is coming in, the rule won't be deleted. But retry will work.

var (
	// ErrAffinityRuleDupKey is raised when the name of the affinity rule
	// conflicts with another one in system that has already been enabled.
	ErrAffinityRuleDupKey = errors.New("name of affinity rule is duplicated")
)

// affinityClient helps set affinity/anti-affinity rule to VMs. Every rule
// should be assigned with a unique name, if name is duplicated, client Will
// not accept the new rule.
type affinityClient struct {
	client  *govmomi.Client
	ctx     context.Context
	cluster types.ManagedObjectReference

	rules     map[int32]*types.ClusterRuleInfo
	ruleKey   map[string]int32
	rrules    map[string]Rule
	rulesLock sync.RWMutex
}

// Rule represents a VM-to-VM affinity/anti-affinity rule
type Rule struct {
	Name     string
	VMs      []string
	Affinity bool
}

func newAffinityClient(clusterName string) *affinityClient {
	ctx := context.Background()
	vsclient, err := NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	clt := &affinityClient{
		client: vsclient,
		ctx:    ctx,
	}

	// Init cluster object
	m := view.NewManager(vsclient.Client)

	v, err := m.CreateContainerView(ctx, vsclient.ServiceContent.RootFolder, []string{"ClusterComputeResource"}, true)
	if err != nil {
		log.Fatal(err)
	}

	defer v.Destroy(ctx)

	var clusters []mo.ClusterComputeResource
	err = v.Retrieve(ctx, []string{"ClusterComputeResource"}, []string{"name"}, &clusters)
	if err != nil {
		log.Fatal(err)
	}

	for i := range clusters {
		if clusters[i].Name == clusterName {
			clt.cluster = clusters[i].Reference()
			break
		}
	}

	if clt.cluster.Type == "" {
		log.Printf("[WARNING] cannot find cluster named %s", clusterName)
	}

	return clt
}

func (c *affinityClient) Rules() map[string]Rule {
	return c.rrules
}

// Run runs in background keep the key to rules in sync
func (c *affinityClient) Run(stopCh <-chan struct{}) error {
	ctx := context.Background()
	filter := new(property.WaitFilter)
	filter.Add(c.cluster.Reference(), "ClusterComputeResource", []string{"configurationEx"})

	err := property.WaitForUpdates(ctx, c.client.PropertyCollector(), filter, func(updates []types.ObjectUpdate) bool {
		for _, update := range updates {
			switch update.Kind {
			case types.ObjectUpdateKindModify:
				fallthrough
			case types.ObjectUpdateKindEnter:
				rules := make(map[int32]*types.ClusterRuleInfo)
				ruleKey := make(map[string]int32)
				rrules := make(map[string]Rule)

				for _, cs := range update.ChangeSet {
					config := cs.Val.(types.ClusterConfigInfoEx)
					for _, rule := range config.Rule {

						info := rule.GetClusterRuleInfo()
						rules[info.Key] = info
						ruleKey[info.Name] = info.Key

						theRule := Rule{
							Name: info.Name,
						}
						switch rule.(type) {
						case *types.ClusterAffinityRuleSpec:
							theRule.Affinity = true
							for _, vm := range rule.(*types.ClusterAffinityRuleSpec).Vm {
								theRule.VMs = append(theRule.VMs, vm.String())
							}
						case *types.ClusterAntiAffinityRuleSpec:
							theRule.Affinity = false
							for _, vm := range rule.(*types.ClusterAntiAffinityRuleSpec).Vm {
								theRule.VMs = append(theRule.VMs, vm.String())
							}
						}
						rrules[info.Name] = theRule
					}
				}
				c.rulesLock.Lock()
				c.rules = rules
				c.ruleKey = ruleKey
				c.rrules = rrules
				c.rulesLock.Unlock()
			case types.ObjectUpdateKindLeave:
			}
		}

		log.Println("vsphere-affinity-client: updated rrules", c.rrules)

		select {
		case <-stopCh:
			return true
		default:
		}

		return false
	})

	return err
}

func (c *affinityClient) ApplyAffinityRule(name string, vms ...string) error {
	log.Printf("vsphere: apply affinity rule %s on vms %s", name, vms)

	c.rulesLock.RLock()
	if _, ok := c.ruleKey[name]; ok {
		c.rulesLock.RUnlock()
		return ErrAffinityRuleDupKey
	}
	c.rulesLock.RUnlock()

	morefs := make([]types.ManagedObjectReference, len(vms))
	for i := range vms {
		morefs[i].FromString(vms[i])
	}

	cluster := object.NewClusterComputeResource(c.client.Client, c.cluster.Reference())

	spec := &types.ClusterConfigSpecEx{
		RulesSpec: []types.ClusterRuleSpec{
			types.ClusterRuleSpec{
				ArrayUpdateSpec: types.ArrayUpdateSpec{
					Operation: types.ArrayUpdateOperationAdd,
				},
				Info: &types.ClusterAffinityRuleSpec{
					ClusterRuleInfo: types.ClusterRuleInfo{
						Name:    name,
						Enabled: addressOfBool(true),
					},
					Vm: morefs,
				},
			},
		},
	}

	task, err := cluster.Reconfigure(c.ctx, spec, true)
	if err != nil {
		return err
	}

	return task.Wait(c.ctx)
}

func (c *client) ApplyAntiAffinityRule(name string, vms ...string) error {
	log.Printf("vsphere: apply anti-affinity rule %s on vms %s", name, vms)

	c.rulesLock.RLock()
	if _, ok := c.ruleKey[name]; ok {
		c.rulesLock.RUnlock()
		return ErrAffinityRuleDupKey
	}
	c.rulesLock.RUnlock()

	morefs := make([]types.ManagedObjectReference, len(vms))
	for i := range vms {
		morefs[i].FromString(vms[i])
	}

	cluster := object.NewClusterComputeResource(c.client.Client, c.cluster.Reference())

	spec := &types.ClusterConfigSpecEx{
		RulesSpec: []types.ClusterRuleSpec{
			types.ClusterRuleSpec{
				ArrayUpdateSpec: types.ArrayUpdateSpec{
					Operation: types.ArrayUpdateOperationAdd,
				},
				Info: &types.ClusterAntiAffinityRuleSpec{
					ClusterRuleInfo: types.ClusterRuleInfo{
						Name:    name,
						Enabled: addressOfBool(true),
					},
					Vm: morefs,
				},
			},
		},
	}

	task, err := cluster.Reconfigure(c.ctx, spec, true)
	if err != nil {
		return err
	}

	return task.Wait(c.ctx)
}

func (c *affinityClient) DeleteAffinityRule(name string) error {
	return c.deleteRule(name)
}

func (c *affinityClient) DeleteAntiAffinityRule(name string) error {
	return c.deleteRule(name)
}

func (c *affinityClient) deleteRule(name string) error {
	log.Printf("vsphere: delete affinity rule %s", name)

	c.rulesLock.RLock()
	key, ok := c.ruleKey[name]
	c.rulesLock.RUnlock()

	if !ok {
		return fmt.Errorf("affinity: affinity rule %s not found", name)
	}

	// sync with all the rules and the key mapping to the rules
	log.Printf("vsphere: delete affinity rule with key %d", key)

	cluster := object.NewClusterComputeResource(c.client.Client, c.cluster.Reference())

	spec := &types.ClusterConfigSpecEx{
		RulesSpec: []types.ClusterRuleSpec{
			types.ClusterRuleSpec{
				ArrayUpdateSpec: types.ArrayUpdateSpec{
					Operation: types.ArrayUpdateOperationRemove,
					RemoveKey: key,
				},
			},
		},
	}

	task, err := cluster.Reconfigure(c.ctx, spec, true)
	if err != nil {
		return err
	}

	return task.Wait(c.ctx)
}

func addressOfBool(v bool) *bool {
	return &v
}
