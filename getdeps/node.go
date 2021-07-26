// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"fmt"
	"reflect"
)

type Node struct {
	Git   []Git   `json:"git,omitempty"`
	Goget []Gopkg `json:"goget,omitempty"`
	Untar []Untar `json:"untar,omitempty"`
	Files *Files  `json:"files,omitempty"`
}

// Get performs the specified actions.
func (n *Node) Get(projectDir string, urlOverrides *URLOverrides, hashMode HashMode) error {
	for i, g := range n.Git {
		if err := g.Get(projectDir, urlOverrides, hashMode); err != nil {
			return fmt.Errorf("error processing %s entry %d: %w", "git", i, err)
		}
		n.Git[i] = g
	}
	for i, gg := range n.Goget {
		if err := gg.Get(projectDir, urlOverrides, hashMode); err != nil {
			return fmt.Errorf("error processing %s entry %d: %w", "goget", i, err)
		}
		n.Goget[i] = gg
	}
	for i, u := range n.Untar {
		if err := u.Get(projectDir, urlOverrides, hashMode); err != nil {
			return fmt.Errorf("error processing %s entry %d: %w", "untar", i, err)
		}
		n.Untar[i] = u
	}
	if n.Files != nil {
		if err := n.Files.Get(projectDir, urlOverrides, hashMode); err != nil {
			return fmt.Errorf("error processing %s entry: %w", "files", err)
		}
	}
	return nil
}

func mergeNodes(base, patch *Node) (*Node, error) {
	var ret Node
	if base != nil {
		ret = *base
	} else {
		if patch == nil {
			return nil, nil
		}
	}
	if patch == nil {
		return &ret, nil
	}

	for i := range patch.Git {
		if patch.Git[i].Label == "" {
			return nil, fmt.Errorf("label for %s cannot be empty", patch.Git[i].URL)
		}
		matchFound := false
		for j := range ret.Git {
			if patch.Git[i].Label != ret.Git[j].Label {
				continue
			}
			dst := reflect.ValueOf(&ret.Git[j]).Elem()
			src := reflect.ValueOf(&patch.Git[i]).Elem()
			mergeFields(dst, src)
			matchFound = true
			break
		}
		if !matchFound {
			ret.Git = append(ret.Git, patch.Git[i])
		}
	}

	for i := range patch.Goget {
		if patch.Goget[i].Label == "" {
			return nil, fmt.Errorf("label for %s cannot be empty", patch.Goget[i].Pkg)
		}
		matchFound := false
		for j := range ret.Goget {
			if patch.Goget[i].Label != ret.Goget[j].Label {
				continue
			}
			dst := reflect.ValueOf(&ret.Goget[j]).Elem()
			src := reflect.ValueOf(&patch.Goget[i]).Elem()
			mergeFields(dst, src)
			matchFound = true
			break
		}
		if !matchFound {
			ret.Goget = append(ret.Goget, patch.Goget[i])
		}
	}

	for i := range patch.Untar {
		if patch.Untar[i].Label == "" {
			return nil, fmt.Errorf("label for %s cannot be empty", patch.Untar[i].URL)
		}
		matchFound := false
		for j := range ret.Untar {
			if patch.Untar[i].Label != ret.Untar[j].Label {
				continue
			}
			dst := reflect.ValueOf(&ret.Untar[j]).Elem()
			src := reflect.ValueOf(&patch.Untar[i]).Elem()
			mergeFields(dst, src)
			matchFound = true
			break
		}
		if !matchFound {
			ret.Untar = append(ret.Untar, patch.Untar[i])
		}
	}

	if patch.Files != nil {
		ret.Files = patch.Files
	}

	return &ret, nil
}
