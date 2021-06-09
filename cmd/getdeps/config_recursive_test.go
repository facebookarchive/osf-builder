// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testRecursiveConfig         = filepath.Join("testdata", "recursive_config.json")
	testInfiniteRecursiveConfig = filepath.Join("testdata", "infinite_recursive_config.json")
)

func TestNewConfigWithIncludesRecursive(t *testing.T) {
	data, err := ioutil.ReadFile(testRecursiveConfig)
	require.NoError(t, err)
	config, err := NewConfigWithIncludes(data, "testdata")
	require.NoError(t, err)
	require.Equal(t, 1, len(config.Coreboot.Git))
	require.NotNil(t, config.Coreboot.Git[0].Branch)
	require.Equal(t, "master", *config.Coreboot.Git[0].Branch)
	// Hash gets overridden in recursive_config.json.
	require.Nil(t, config.Coreboot.Git[0].Hash)
}

func TestNewConfigWithIncludesInfiniteRecursion(t *testing.T) {
	data, err := ioutil.ReadFile(testInfiniteRecursiveConfig)
	require.NoError(t, err)
	config, err := NewConfigWithIncludes(data, "testdata")
	require.Error(t, err)
	require.Nil(t, config)
}
