// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"net/url"
	"os"
	"path"
	"strings"
)

// Gopkg represents a Go package
type Gopkg struct {
	Label  string  `json:"label"`
	Pkg    string  `json:"pkg"`
	Branch *string `json:"branch,omitempty"`
	Hash   *string `json:"hash,omitempty"`
}

// Get downloads a Go package
func (pkg *Gopkg) Get(projectDir string, urlOverrides *URLOverrides, hashMode HashMode) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	goDir := path.Join(dir, "gopath/src")
	if err = os.MkdirAll(goDir, os.ModePerm); err != nil {
		return err
	}

	url, err := url.Parse(pkg.Pkg)
	if err != nil {
		return err
	}
	dirs := strings.Split(url.Host+url.Path, "/")
	for i := 0; i < len(dirs); i++ {
		goDir = path.Join(goDir, dirs[i])
		if err = os.MkdirAll(goDir, os.ModePerm); err != nil {
			return err
		}
	}

	repo := strings.Replace(pkg.Pkg, "golang.org/x", "go.googlesource.com", 1)

	branch := defaultBranch
	if pkg.Branch != nil && *pkg.Branch != "" {
		branch = *pkg.Branch
	} else {
		pkg.Branch = &branch
	}
	hash := ""
	if pkg.Hash != nil {
		hash = *pkg.Hash
	}
	currentHash, err := gitClone(pkg.Label, repo, goDir, branch, hashMode, hash, urlOverrides)
	if err != nil {
		return err
	}
	pkg.Hash = &currentHash
	return nil
}
