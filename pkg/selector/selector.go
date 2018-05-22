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

package selector

import (
	"k8s.io/apimachinery/pkg/labels"
)

// Selector selects labels
type Selector interface {
	Matches(labels.Labels) bool
}

// And is a group of selectors which matches if all of them match
type And []Selector

// Matches returns true if all selectors return true, otherwise false
func (and And) Matches(labels labels.Labels) bool {
	if len(and) == 0 {
		return true
	}

	for _, s := range and {
		if !s.Matches(labels) {
			return false
		}
	}

	return true
}

// Or is a group of selectors which matches if any one of them matches
type Or []Selector

// Matches returns true if any selector returns true, otherwise false
func (or Or) Matches(labels labels.Labels) bool {
	if len(or) == 0 {
		return true
	}

	for _, s := range or {
		if s.Matches(labels) {
			return true
		}
	}

	return false
}

type not struct {
	s Selector
}

func (n *not) Matches(labels labels.Labels) bool {
	return !n.s.Matches(labels)
}

// Not returns a reverse of selector
func Not(s Selector) Selector {
	if n, ok := s.(*not); ok {
		return n
	}
	return &not{s}
}
