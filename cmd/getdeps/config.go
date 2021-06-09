// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// Config contains the sources which need to be fetched
type Config struct {
	// Identification of the repo the build is being run from.
	// Either `git describe` or `hg id` output.
	BuildID string `json:"build_id"`
	// Additional config files to include. Their order matters: subsequent ones
	// may override values from previous ones.
	Includes  []string  `json:"includes,omitempty"`
	Initramfs Initramfs `json:"initramfs"`
	Kernel    Kernel    `json:"kernel"`
	Coreboot  Coreboot  `json:"coreboot"`
}

// NewConfig creates a new config object by parsing the specified file,
// without loading the includes. See NewConfigWithIncludes to fully load
// a file with its includes.
func NewConfig(data []byte) (*Config, error) {
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %v", err)
	}
	return &config, nil
}

// NewConfigWithIncludes parses a configuration blob and recursively
// any specified `include` directive.
// Any relative include path is considered to be relative to `basedir`.
// If `basedir` is empty, relative paths will be resolved from the current
// working directory.
//
// Note that the order of the includes is meaningful: the latest includes will
// override the earliest. The inclusion tree is traversed depth-first
// pre-order, so the inner includes have always more priority than the outer
// (top-level) ones. Recursion alert!
// The recursion depth has a hard limit of 512, which should be enough to avoid
// loops.
func NewConfigWithIncludes(data []byte, basedir string) (*Config, error) {
	maxDepth, currentDepth := uint(512), uint(0)
	return newConfigWithIncludes(data, basedir, maxDepth, currentDepth)
}

func newConfigWithIncludes(data []byte, basedir string, maxDepth, currentDepth uint) (*Config, error) {
	currentDepth++
	if currentDepth > maxDepth {
		return nil, fmt.Errorf("maximum recursion depth of %d reached", maxDepth)
	}
	topConfig, err := NewConfig(data)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	for _, include := range topConfig.Includes {
		if !filepath.IsAbs(include) {
			include = filepath.Join(basedir, include)
		}
		includeData, err := ioutil.ReadFile(include)
		if err != nil {
			return nil, fmt.Errorf("failed to read file '%s': %v", include, err)
		}
		other, err := newConfigWithIncludes(includeData, basedir, maxDepth, currentDepth)
		if err != nil {
			return nil, err
		}
		config, err = mergeConfigs(config, other)
		if err != nil {
			return nil, fmt.Errorf("failed to merge file '%s' into the configuration: %v", include, err)
		}
	}
	config, err = mergeConfigs(config, topConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to merge top config into the configuration: %v", err)
	}
	return config, nil
}

func mergeConfigs(config1 *Config, config2 *Config) (*Config, error) {
	var (
		newConfig Config
		err       error
	)
	if config1 == nil || config2 == nil {
		return nil, fmt.Errorf("config objects to merge must be non-nil")
	}

	newConfig.Initramfs, err = mergeInitramfsConfigs(config1.Initramfs, config2.Initramfs)
	if err != nil {
		return &newConfig, fmt.Errorf("error merging initramfs config: %w", err)
	}
	newConfig.Kernel, err = mergeKernelConfigs(config1.Kernel, config2.Kernel)
	if err != nil {
		return &newConfig, fmt.Errorf("error merging kernel config: %w", err)
	}
	newConfig.Coreboot, err = mergeCorebootConfigs(config1.Coreboot, config2.Coreboot)
	if err != nil {
		return &newConfig, fmt.Errorf("error merging coreboot config: %w", err)
	}

	return &newConfig, nil
}
