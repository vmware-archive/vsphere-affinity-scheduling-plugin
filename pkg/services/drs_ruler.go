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

package services

import (
	"fmt"
	"log"
	"reflect"
	"sort"
	"time"

	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/algorithm"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/bridgecache"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/constants"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/selector"
	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/vsphere"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

// DRSRuler watches Kubernetes pods, whenever a pod is assigned to a node with
// pod-to-pod affinity or anti-affinity rule, it sets the VM-to-VM Affinity
// or anti-affinity rule to vSphere so that DRS (vSphere scheduler) doesn't
// do scheduling against Kubernetes. For example, DRS doesn't migrate VM to a
// host that breaks Kubernetes' anti-affinity rule.
type DRSRuler struct {
	bcache    bridgecache.Cache
	podLister algorithm.PodLister
	vsclient  vsphere.Vsphere

	// kubernetes pods with affinity rules
	affinityPods     map[string]*v1.Pod
	antiAffinityPods map[string]*v1.Pod
}

// NewDRSRuler creates an DRSRuler instance
func NewDRSRuler(
	podInformer cache.SharedIndexInformer,
	bcache bridgecache.Cache,
	podLister algorithm.PodLister,
	vsclient vsphere.Vsphere) *DRSRuler {
	drs := &DRSRuler{
		podLister:        podLister,
		bcache:           bcache,
		vsclient:         vsclient,
		affinityPods:     make(map[string]*v1.Pod),
		antiAffinityPods: make(map[string]*v1.Pod),
	}

	podInformer.AddEventHandler(drs)

	return drs
}

// Run starts the service
func (r *DRSRuler) Run(stopCh <-chan struct{}) {
	log.Println("Start service DRSRuler...")
	for {
		time.Sleep(15 * time.Second)

		r.sync()

		select {
		case <-stopCh:
			return
		default:
		}
	}
}

func (r *DRSRuler) sync() {
	actualRules := r.vsclient.Rules()
	desiredRules := r.desiredRules()

	log.Printf("actual rules: %v", actualRules)
	log.Printf("desired rules: %v", desiredRules)

	// Delete not-needed rules
	for uid, rule := range actualRules {
		if _, ok := desiredRules[uid]; !ok {
			log.Printf("delete rule: %v", rule)
			if rule.Affinity {
				r.vsclient.DeleteAffinityRule(rule.Name)
			} else {
				r.vsclient.DeleteAntiAffinityRule(rule.Name)
			}
		}
	}

	// Apply missing rules
	for uid, rule := range desiredRules {
		if _, ok := actualRules[uid]; !ok {
			log.Printf("apply rule: %s(%v)", uid, rule)
			if rule.Affinity {
				r.vsclient.ApplyAffinityRule(rule.Name, rule.VMs...)
			} else {
				r.vsclient.ApplyAntiAffinityRule(rule.Name, rule.VMs...)
			}
		}
	}

	// Modify changed rules
	for uid, desiredRule := range desiredRules {
		if actualRule, ok := actualRules[uid]; ok {
			sort.Strings(desiredRule.VMs)
			sort.Strings(actualRule.VMs)
			if !reflect.DeepEqual(desiredRule.VMs, actualRule.VMs) {
				log.Printf("modify rule: %v", actualRule)
				if actualRule.Affinity {
					r.vsclient.DeleteAffinityRule(actualRule.Name)
					r.vsclient.ApplyAffinityRule(actualRule.Name, actualRule.VMs...)
				} else {
					r.vsclient.DeleteAffinityRule(actualRule.Name)
					r.vsclient.ApplyAntiAffinityRule(actualRule.Name, actualRule.VMs...)
				}
			}

		}
	}
}

func (r *DRSRuler) desiredRules() map[string]vsphere.Rule {
	rules := make(map[string]vsphere.Rule)

	r.calculateRules(r.affinityPods, true, rules)
	r.calculateRules(r.antiAffinityPods, false, rules)

	return rules
}

func (r *DRSRuler) calculateRules(podsWithTerm map[string]*v1.Pod, affinity bool, rules map[string]vsphere.Rule) {
	for _, pod := range podsWithTerm {
		var terms []v1.PodAffinityTerm
		if affinity {
			terms = pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		} else {
			terms = pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		}

		// Get pod selector
		var selectors selector.And
		for _, af := range terms {
			if af.TopologyKey != constants.HostLabel {
				continue
			}

			selector, err := metav1.LabelSelectorAsSelector(af.LabelSelector)
			if err != nil {
				log.Printf("[WARNING] failed to get selector: %s", err)
				continue
			}

			selectors = append(selectors, selector)
		}

		pods, err := r.podLister.ListPod(selectors)
		if err != nil {
			log.Printf("[ERROR] failed to list pods: %s", err)
		}
		if len(pods) == 0 {
			continue
		}

		rule := vsphere.Rule{
			Name:     r.ruleName(pod, affinity),
			Affinity: affinity,
		}

		vmids := make(map[string]interface{})
		for _, matchedPod := range pods {
			nodename := matchedPod.Spec.NodeName
			if nodename == "" {
				continue
			}
			vmid := r.bcache.GetVMIDFromNode(nodename)
			vmids[vmid] = struct{}{}
		}
		vmid := r.bcache.GetVMIDFromNode(pod.Spec.NodeName)
		vmids[vmid] = struct{}{}

		for vmid := range vmids {
			rule.VMs = append(rule.VMs, vmid)
		}

		rules[rule.Name] = rule
	}
}

func (r *DRSRuler) ruleName(pod *v1.Pod, affinity bool) string {
	if affinity {
		return fmt.Sprintf("affi-%s", pod.UID)
	}
	return fmt.Sprintf("anti-%s", pod.UID)
}

// OnAdd is handler for adding an pod object
func (r *DRSRuler) OnAdd(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return
	}

	if pod.Spec.NodeName != "" {
		if rule := pod.Spec.Affinity; rule != nil {
			if affinity := rule.PodAffinity; affinity != nil {
				r.affinityPods[string(pod.UID)] = pod
			}
			if anti := rule.PodAntiAffinity; anti != nil {
				r.antiAffinityPods[string(pod.UID)] = pod
			}
		}
	}
}

// OnUpdate is handler for updating an pod object
func (r *DRSRuler) OnUpdate(old, new interface{}) {
	oldPod, ok := old.(*v1.Pod)
	if !ok {
		return
	}
	newPod, ok := new.(*v1.Pod)
	if !ok {
		return
	}

	if oldPod.Spec.NodeName != newPod.Spec.NodeName && oldPod.Spec.NodeName == "" {
		// newly assigned nodeName
		r.OnAdd(newPod)
	}
}

// OnDelete is handler for adding an pod object
func (r *DRSRuler) OnDelete(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return
	}

	if pod.Spec.NodeName != "" {
		if rule := pod.Spec.Affinity; rule != nil {
			if affinity := rule.PodAffinity; affinity != nil {
				delete(r.affinityPods, string(pod.UID))
			}
			if anti := rule.PodAntiAffinity; anti != nil {
				delete(r.antiAffinityPods, string(pod.UID))
			}
		}
	}
}

func getSelector(terms []v1.PodAffinityTerm) selector.Selector {
	var selectors selector.And

	for _, term := range terms {
		if term.TopologyKey != constants.HostLabel {
			continue
		}

		selector, err := metav1.LabelSelectorAsSelector(term.LabelSelector)
		if err != nil {
			log.Printf("[WARNING] invalid selector: %s", err)
			continue
		}

		selectors = append(selectors, selector)
	}

	return selectors
}
