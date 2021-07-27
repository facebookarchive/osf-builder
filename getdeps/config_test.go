// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testFlatConfig = filepath.Join("testdata", "flat_config.json")
)

func TestNewConfig(t *testing.T) {
	data, err := ioutil.ReadFile(testFlatConfig)
	require.NoError(t, err)
	c, err := NewConfig(data)
	require.NoError(t, err)

	assert.Nil(t, c.Includes)

	require.NotNil(t, c.Coreboot)
	assert.NotNil(t, c.Coreboot.Git)
	require.Len(t, c.Coreboot.Git, 1)
	assert.Equal(t, c.Coreboot.Git[0].Label, "coreboot")
	assert.Equal(t, c.Coreboot.Git[0].URL, "https://review.coreboot.org/coreboot")
	require.NotNil(t, c.Coreboot.Git[0].Branch)
	assert.Equal(t, *c.Coreboot.Git[0].Branch, "master")
	require.NotNil(t, c.Coreboot.Git[0].Hash)
	assert.Equal(t, *c.Coreboot.Git[0].Hash, "HEAD")
	require.NotNil(t, c.Coreboot.Files)
	assert.Equal(t, c.Coreboot.Files.Label, "crossgcc_tarballs")
	assert.Equal(t, c.Coreboot.Files.Dest, "util/crossgcc/tarballs")
	assert.NotNil(t, c.Coreboot.Files.Filelist)
	require.Len(t, c.Coreboot.Files.Filelist, 1)
	assert.Equal(t, c.Coreboot.Files.Filelist[0].URL, "https://ftpmirror.gnu.org/gmp/gmp-6.1.2.tar.xz")
	assert.Equal(t, c.Coreboot.Files.Filelist[0].Hash, "sha256:87b565e89a9a684fe4ebeeddb8399dce2599f9c9049854ca8c0dfbdea0e21912")

	require.NotNil(t, c.Kernel)
	require.NotNil(t, c.Kernel.Untar)
	require.Len(t, c.Kernel.Untar, 1)
	assert.Equal(t, c.Kernel.Untar[0].Label, "kernel")
	assert.Equal(t, c.Kernel.Untar[0].URL, "https://cdn.kernel.org/pub/linux/kernel/v5.x/linux-5.9.12.tar.xz")
	assert.Equal(t, c.Kernel.Untar[0].Hash, "sha256:d97f56192e3474c9c8a44ca39957d51800a26497c9a13c9c5e8cc0f1f5b0d9bd")

	require.NotNil(t, c.Initramfs)
	require.NotNil(t, c.Initramfs.Untar)
	require.Len(t, c.Initramfs.Untar, 1)
	assert.Equal(t, c.Initramfs.Untar[0].Label, "go")
	assert.Equal(t, c.Initramfs.Untar[0].URL, "https://golang.org/dl/go1.15.linux-amd64.tar.gz")
	require.NotNil(t, c.Initramfs.Untar)
	require.Len(t, c.Initramfs.Goget, 1)
	assert.Equal(t, c.Initramfs.Goget[0].Label, "uroot")
	assert.Equal(t, c.Initramfs.Goget[0].Pkg, "https://github.com/u-root/u-root")
	require.NotNil(t, c.Initramfs.Goget[0].Branch)
	assert.Equal(t, *c.Initramfs.Goget[0].Branch, "master")
	require.NotNil(t, c.Initramfs.Goget[0].Hash)
	assert.Equal(t, *c.Initramfs.Goget[0].Hash, "60aeb0ab57dfac6e19f057de0b3e25793ede1616")
}

func TestNewConfigBrokenJSON(t *testing.T) {
	_, err := NewConfig([]byte("broken JSON"))
	assert.Error(t, err)
}

func TestMergeConfigs(t *testing.T) {
	_, err := mergeConfigs(nil, nil)
	assert.Error(t, err)

	_, err = mergeConfigs(nil, &Config{})
	assert.Error(t, err)

	_, err = mergeConfigs(&Config{}, nil)
	assert.Error(t, err)
}

func TestMergeConfigsInitramfs(t *testing.T) {
	leftBranch, leftHash := "leftbranch", "lefthash"
	left := Config{
		Initramfs: &Node{
			Goget: []Gopkg{
				{Label: "override", Pkg: "pkg_thisshouldbeoverridden", Branch: &leftBranch, Hash: &leftHash},
				{Label: "nooverride", Pkg: "pkg_thisshouldremain", Branch: &leftBranch, Hash: &leftHash},
			},
			Untar: []Untar{
				{Label: "override", URL: "url_thisshouldbeoverridden", Hash: "hash_override"},
				{Label: "nooverride", URL: "url_thisshouldremain", Hash: "hash_nooverride"},
			},
		},
	}
	right := Config{
		Initramfs: &Node{
			Goget: []Gopkg{
				{Label: "override", Pkg: "pkg_thisshouldbehere", Branch: &leftBranch, Hash: &leftHash},
			},
			Untar: []Untar{
				{Label: "override", URL: "url_thisshouldbehere", Hash: "hash_overridden"},
			},
		},
	}
	merged, err := mergeConfigs(&left, &right)
	require.NoError(t, err)
	assert.Equal(t, merged.Initramfs.Goget[0].Label, "override")
	assert.Equal(t, merged.Initramfs.Goget[0].Pkg, "pkg_thisshouldbehere")
	assert.Equal(t, merged.Initramfs.Goget[1].Label, "nooverride")
	assert.Equal(t, merged.Initramfs.Goget[1].Pkg, "pkg_thisshouldremain")

	// errors are returned when any label of the right-hand-side is empty
	rightWithoutLabel := Config(right)
	rightWithoutLabel.Initramfs.Goget[0].Label = ""
	_, err = mergeConfigs(&left, &rightWithoutLabel)
	require.Error(t, err)
}

// TODO test mergeKernel and mergeCoreboot via mergeConfigs as done above for
//      initramfs
