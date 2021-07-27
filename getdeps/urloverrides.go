// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
)

// NewURLOverrides creates a new URLOverrides object after parsing the
// specified configuration file.
func NewURLOverrides(data []byte) (*URLOverrides, error) {
	var urlOverrides URLOverrides
	if err := json.Unmarshal(data, &urlOverrides); err != nil {
		return nil, fmt.Errorf("failed to unmarshal URL overrides JSON: %v", err)
	}
	return &urlOverrides, nil
}

// URLOverrides maps URLs to be overridden with custom ones. This is
// useful for example if you want to use your own mirrors of certain
// repositories.
type URLOverrides map[string]string

// Override applies the override, if any, to the provided URL. If no
// override for that URL exists, the original URL is returned unchanged.
func (uo URLOverrides) Override(origURL string) string {
	// Entire URL
	if override, ok := uo[origURL]; ok {
		return override
	}

	// Just base name
	u, err := url.Parse(origURL)
	if err != nil {
		return origURL
	}
	if override, ok := uo[path.Base(u.Path)]; ok {
		return override
	}

	return origURL
}
