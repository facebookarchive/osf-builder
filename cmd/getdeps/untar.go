// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"log"
	"os"

	"github.com/ulikunitz/xz"
)

// Untar represents a tarball
type Untar struct {
	Label string `json:"label"`
	URL   string `json:"url"`
	Hash  string `json:"hash,omitempty"`
}

// CompressionType is the type that defines compression types.
type CompressionType int

// compression types.
const (
	CompressionTypeUnsupported = iota
	CompressionTypeGzip
	CompressionTypeXz
)

var (
	magicBytesGzip = []byte{0x1f, 0x8b}
	magicBytesXz   = []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}
)

func detectCompressionType(data []byte) CompressionType {
	switch {
	case len(data) >= len(magicBytesGzip) && bytes.Equal(data[:len(magicBytesGzip)], magicBytesGzip):
		return CompressionTypeGzip
	case len(data) >= len(magicBytesXz) && bytes.Equal(data[:len(magicBytesXz)], magicBytesXz):
		return CompressionTypeXz
	default:
		return CompressionTypeUnsupported
	}
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

	// uncompress. We support gzip and xz.
	reader := bytes.NewReader(data)

	var archive io.Reader
	// gzip can be detected with http.DetectContentType, but xz is not
	// supported. So we match the magic bytes in the xz header, as specified in
	// the XZ file format Section 2.1.1.1, see
	// https://tukaani.org/xz/xz-file-format.txt .
	//
	// For gzip, the magic bytes are 1F 8B (starting at 0)
	// for xz the magic bytes are FD 37 7A 58 5A 00 (starting at 0)
	compressionType := detectCompressionType(data)
	switch compressionType {
	case CompressionTypeGzip:
		archive, err = gzip.NewReader(reader)
		// Close only required for gzip package
		defer archive.(io.ReadCloser).Close()
	case CompressionTypeXz:
		archive, err = xz.NewReader(reader)
	case CompressionTypeUnsupported:
		fallthrough
	default:
		return errors.New("unsupported compression type")
	}
	if err != nil {
		return err
	}

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
