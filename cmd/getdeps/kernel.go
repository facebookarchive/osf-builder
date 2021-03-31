// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"fmt"
	"reflect"
)

// Kernel contains the sources that are needed to build the kernel
type Kernel struct {
	Untar []Untar `json:"untar"`
}

// Get downloads the sources to build the kernel
func (k *Kernel) Get(projectDir string, urlOverrides *URLOverrides, hashMode HashMode) error {
	for i, u := range k.Untar {
		if err := u.Get(projectDir, urlOverrides, hashMode); err != nil {
			return err
		}
		k.Untar[i] = u
	}
	return nil
}

func mergeKernelConfigs(base Kernel, patch Kernel) (Kernel, error) {
	var ret Kernel = base

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
