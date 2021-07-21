// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"fmt"
	"reflect"
)

// Coreboot contains the sources that are needed to build coreboot
type Coreboot struct {
	Git   []Git `json:"git"`
	Files Files `json:"files"`
}

// Get downloads the sources to build coreboot
func (c *Coreboot) Get(projectDir string, urlOverrides *URLOverrides, hashMode HashMode) error {
	for i, g := range c.Git {
		if err := g.Get(projectDir, urlOverrides, hashMode); err != nil {
			return err
		}
		c.Git[i] = g
	}
	return c.Files.Get(projectDir, urlOverrides, hashMode)
}

func mergeCorebootConfigs(base Coreboot, patch Coreboot) (Coreboot, error) {
	var ret Coreboot = base

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

	// Files are a bit special. For now we assume that only Coreboot config
	// objects will have URLs for files, and there is only a single Files
	// member unlike most others which are arrays.
	if len(patch.Files.Filelist) != 0 {
		ret.Files = patch.Files
	}

	return ret, nil
}
