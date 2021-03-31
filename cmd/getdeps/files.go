// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
)

// File represents a single file to be fetched
type File struct {
	URL  string `json:"url"`
	Hash string `json:"hash,omitempty"`
}

// Files represents a list of files that need to be fetched
type Files struct {
	Label    string `json:"label"`
	Dest     string `json:"dest,omitempty"`
	Filelist []File `json:"filelist,omitempty"`
}

// Get download the list of files
func (ff *Files) Get(projectDir string, urlOverrides *URLOverrides, hashMode HashMode) error {
	for i, f := range ff.Filelist {
		u, err := url.Parse(f.URL)
		if err != nil {
			return fmt.Errorf("%s: Invalid URL %q", ff.Label, f.URL)
		}

		name := path.Base(u.Path)

		bytes, err := fetchAndVerify(ff.Label, projectDir, f.URL, hashMode, &f.Hash, urlOverrides)
		if err != nil {
			return fmt.Errorf("%s: %s: %w", ff.Label, name, err)
		}

		if err = os.MkdirAll(ff.Dest, os.ModePerm); err != nil {
			return err
		}

		path := path.Join(ff.Dest, name)
		if err = ioutil.WriteFile(path, bytes, 0644); err != nil {
			return err
		}

		ff.Filelist[i] = f
	}

	return nil
}
