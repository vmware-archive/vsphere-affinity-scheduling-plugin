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

package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/vmware/vsphere-affinity-scheduling-plugin/pkg/algorithm"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api/v1"
)

// SchedExtenderHandler implements http.Handler as a scheduler extender
type SchedExtenderHandler struct {
	Filter algorithm.Filter
}

// ServeHTTP implements http.Handler
func (s *SchedExtenderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("processing %s %q", r.Method, r.URL.Path)

	if strings.Contains(r.URL.Path, "filter") { // /scheduler/filter
		s.processFilter(w, r)
	} else if strings.Contains(r.URL.Path, "prioritize") { // /scheduler/prioritize
		s.processPrioritize(w, r)
	} else if strings.Contains(r.URL.Path, "bind") { // /scheduler/bind
		s.processBind(w, r)
	} else {
		http.Error(w, "Unsupported request", http.StatusNotFound)
	}
}

func (s *SchedExtenderHandler) processFilter(w http.ResponseWriter, r *http.Request) {
	log.Printf("process filter %s", r.URL.Path)

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var args schedulerapi.ExtenderArgs
	if err := decoder.Decode(&args); err != nil {
		log.Printf("[ERROR] decode error: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if args.NodeNames == nil {
		http.Error(w, "NodeNames is nil", http.StatusBadRequest)
		return
	}

	nodes, err := s.Filter.Filter(&args.Pod, *args.NodeNames)
	if err != nil {
		log.Printf("[ERROR] filter error: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := schedulerapi.ExtenderFilterResult{
		NodeNames: &nodes,
	}

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(&resp); err != nil {
		log.Printf("[ERROR] encode response %s", err)
	}
}

func (s *SchedExtenderHandler) processPrioritize(w http.ResponseWriter, r *http.Request) {
	log.Printf("process prioritize %s", r.URL.Path)

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var args schedulerapi.ExtenderArgs
	if err := decoder.Decode(&args); err != nil {
		log.Printf("[ERROR] decode error: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Implement prioritize

	var resp schedulerapi.HostPriorityList

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(&resp); err != nil {
		log.Printf("[ERROR] encode response %s", err)
	}
}

func (s *SchedExtenderHandler) processBind(w http.ResponseWriter, r *http.Request) {
	log.Printf("process binding %s", r.URL.Path)

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var args schedulerapi.ExtenderBindingArgs
	if err := decoder.Decode(&args); err != nil {
		log.Printf("[ERROR] decode error: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := schedulerapi.ExtenderBindingResult{}

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(&resp); err != nil {
		log.Printf("[ERROR] encode response %s", err)
	}
}
