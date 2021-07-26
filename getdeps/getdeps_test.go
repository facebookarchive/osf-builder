// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBaseDir(t *testing.T) {
	var (
		bd  string
		err error
	)
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("Failed to get current working directory: %v", err))
	}
	// input basedir is not empty. Expect the same basedir, ignoring configfile.
	bd, err = getBaseDir("configs", "")
	require.NoError(t, err)
	assert.Equal(t, "configs", bd)

	bd, err = getBaseDir("configs", "blah")
	require.NoError(t, err)
	assert.Equal(t, "configs", bd)

	// input basedir is empty, configfile is relative.
	bd, err = getBaseDir("", "config.json")
	require.NoError(t, err)
	assert.Equal(t, cwd, bd)

	bd, err = getBaseDir("", "configs/config.json")
	require.NoError(t, err)
	assert.Equal(t, path.Join(cwd, "configs"), bd)

	bd, err = getBaseDir("", "configs/")
	require.NoError(t, err)
	assert.Equal(t, path.Join(cwd, "configs"), bd)

	// input basedir is empty, configfile is absolute.
	bd, err = getBaseDir("", "/config.json")
	require.NoError(t, err)
	assert.Equal(t, "/", bd)

	bd, err = getBaseDir("", "/configs/config.json")
	require.NoError(t, err)
	assert.Equal(t, "/configs", bd)

	bd, err = getBaseDir("", "/configs/")
	require.NoError(t, err)
	assert.Equal(t, "/configs", bd)

	// input basedir is empty, configfile is empty.
	bd, err = getBaseDir("", "")
	require.NoError(t, err)
	assert.Equal(t, cwd, bd)
}
