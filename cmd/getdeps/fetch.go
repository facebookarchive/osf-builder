// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

func fetch(label, urlStr string) ([]byte, error) {
	log.Printf("%s: Downloading %s...", label, urlStr)

	// Get the data
	var (
		resp *http.Response
		err  error
	)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// some servers will behave differently upon redirects if a Referer
			// header is found, and this may cause the download to fail. So here
			// we remove the Referer header.
			req.Header.Del("Referer")
			return nil
		},
	}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create new http.Request: %w", err)
	}
	for attempts := 0; attempts < 3; attempts++ {
		resp, err = client.Do(req)
		if err != nil {
			if uErr, ok := err.(*url.Error); ok {
				if uErr.Temporary() || uErr.Timeout() {
					// retryable error
					log.Printf("Failed to get file, trying again. Error was: %v", err)
					continue
				}
			}
			// non-retryable error
			return nil, fmt.Errorf("%s: error while downloading %s: %w", label, urlStr, err)
		}
		defer resp.Body.Close()
		log.Printf("Status code is %s", resp.Status)
		break
	}
	// At this point either the last attempt succeeded, or it failed with
	// a retryable error, but we are out of retrie.
	if err != nil {
		return nil, fmt.Errorf("every download attempt has failed. Last error: %v", err)
	}

	var data []byte
	for attempts := 0; attempts < 3; attempts++ {
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			switch err {
			case io.ErrUnexpectedEOF, io.ErrClosedPipe:
				// retryable error
				log.Printf("Failed to retrieve file, trying again. Error was: %v", err)
				continue
			default:
				// non-retryable error
				return nil, fmt.Errorf("%s: error while downloading %s: %w", label, urlStr, err)
			}
		} else {
			break
		}
	}
	return data, nil
}

func fetchAndVerify(label, projectDir, urlStr string, hashMode HashMode, hash *string, urlOverrides *URLOverrides) ([]byte, os.FileInfo, error) {
	if urlOverrides != nil {
		urlStr = urlOverrides.Override(urlStr)
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: invalid URL %q: %w", label, urlStr, err)
	}

	if strings.ToLower(u.Scheme) == "file" {
		filePath := path.Join(projectDir, u.Host, u.Path)
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return nil, nil, err
		}
		data, err := ioutil.ReadFile(filePath)
		return data, fileInfo, err
	}

	switch hashMode {
	case hashModeStrict:
		if hash == nil || *hash == "" {
			return nil, nil, fmt.Errorf("%s: %s: hash mode is strict and no hash supplied", label, urlStr)
		}
	case hashModeUpdate:
		if hash != nil {
			*hash = ""
		}
	case hashModePermissive:
		// Proceed
	}

	var data []byte
	if hash != nil {
		var actualHash string
		// blindly retry to downloading the file when hash check fails. This is
		// to work around an odd behaviour of the GNU mirrors where the files
		// are updated but their content is wrong for a few seconds (e.g. the
		// tar.gz file with tar'ed but not gzip'ed content, like it's being
		// compressed in prod).
		for attempts := 0; attempts < 3; attempts++ {
			data, err = fetch(label, urlStr)
			if err != nil {
				return nil, nil, err
			}
			actualHash, err = verifyHash(data, *hash)
			if err != nil {
				log.Printf("Hash validation for %s failed, will try downloading the file again. Error is: %v", label, err)
				continue
			}
			if *hash == "" {
				*hash = actualHash
				log.Printf("%s: Hash %s", label, actualHash)
			} else {
				log.Printf("%s: Hash %s (verified)", label, actualHash)
			}
			return data, nil, nil
		}

		// at this point err is `nil` if the last attempt was successful,
		// and not `nil` otherwise.
		return data, nil, err
	}

	data, err = fetch(label, urlStr)
	return data, nil, err
}

func verifyHash(data []byte, expectedHash string) (string, error) {
	var ct string
	expectedHash = strings.ToLower(expectedHash)
	if expectedHash == "" {
		// Hash update mode
		ct = "sha256"
	} else {
		parts := strings.Split(expectedHash, ":")
		if len(parts) != 2 {
			return "", fmt.Errorf("unsupported hash format %q", expectedHash)
		}
		expectedHashType := parts[0]
		switch expectedHashType {
		case "sha256":
			ct = "sha256"
		default:
			return "", fmt.Errorf("unsupported hash type %q", expectedHashType)
		}
	}

	var csHex string
	switch ct {
	case "sha256":
		cs := sha256.Sum256(data)
		csHex = strings.ToLower(hex.EncodeToString(cs[:]))
	}

	actualHash := fmt.Sprintf("%s:%s", ct, csHex)

	var err error

	if expectedHash != "" && actualHash != expectedHash {
		return actualHash, fmt.Errorf("hash mismatch: expected %q, got %q", expectedHash, actualHash)
	}

	return actualHash, err
}
