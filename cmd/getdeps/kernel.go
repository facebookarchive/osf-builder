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
	Git   []Git `json:"git"`
}

// Get downloads the sources to build the kernel
func (k *Kernel) Get(projectDir string, urlOverrides *URLOverrides, hashMode HashMode) error {
	for i, u := range k.Git {
		if err := u.Get(projectDir, urlOverrides, hashMode); err != nil {
			return err
		}
		k.Git[i] = u
	}
	return nil
}

func mergeKernelConfigs(base Kernel, patch Kernel) (Kernel, error) {
	var ret Kernel = base

	for i := range patch.Git {
		if patch.Git[i].Label == "" {
			return ret, fmt.Errorf("label for %s cannot be empty", patch.Git[i].URL)
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

	return ret, nil
}
