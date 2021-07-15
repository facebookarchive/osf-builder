// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"os"
)

// Untar represents a tarball
type Untar struct {
	Label string `json:"label"`
	URL   string `json:"url"`
	Hash  string `json:"hash,omitempty"`
}

// Get downloads a tar.gz file and uncompresses it
func (pkg *Untar) Get(projectDir string, urlOverrides *URLOverrides, hashMode HashMode) error {
	// ignore file info, will use permissions from the tar metadata
	data, _, err := fetchAndVerify(pkg.Label, projectDir, pkg.URL, hashMode, &pkg.Hash, urlOverrides)
	if err != nil {
		return err
	}

	dir, _ := os.Getwd()
	log.Printf("%s: Uncompressing into %s...", pkg.Label, dir)

	// ungzip
	reader := bytes.NewReader(data)
	archive, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer archive.Close()

	// untar
	tarReader := tar.NewReader(archive)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(header.Name, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(header.Name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		if _, err = io.Copy(file, tarReader); err != nil {
			return err
		}
		if err = file.Close(); err != nil {
			return err
		}
	}

	return err
}
