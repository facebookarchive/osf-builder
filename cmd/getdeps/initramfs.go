// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"fmt"
	"reflect"
)

// Initramfs contains the sources that are needed to build the initramfs
type Initramfs struct {
	Goget []Gopkg `json:"goget"`
	Untar []Untar `json:"untar"`
}

// Get downloads the sources to build the initramfs
func (ir *Initramfs) Get(projectDir string, urlOverrides *URLOverrides, hashMode HashMode) error {
	for i, g := range ir.Goget {
		if err := g.Get(projectDir, urlOverrides, hashMode); err != nil {
			return err
		}
		ir.Goget[i] = g
	}
	for i, u := range ir.Untar {
		if err := u.Get(projectDir, urlOverrides, hashMode); err != nil {
			return err
		}
		ir.Untar[i] = u
	}
	return nil
}

func mergeInitramfsConfigs(base Initramfs, patch Initramfs) (Initramfs, error) {
	var ret Initramfs = base

	for i := range patch.Goget {
		if patch.Goget[i].Label == "" {
			return ret, fmt.Errorf("label for %s cannot be empty", patch.Goget[i].Pkg)
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
			return ret, fmt.Errorf("label for %s cannot be empty", patch.Untar[i].URL)
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

	return ret, nil
}
