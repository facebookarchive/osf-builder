// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Git represents a Git repository
type Git struct {
	Label  string  `json:"label"`
	URL    string  `json:"url"`
	Dest   string  `json:"dest,omitempty"`
	Branch *string `json:"branch,omitempty"`
	Hash   *string `json:"hash,omitempty"`
}

// Get downloads a Git repository
func (g *Git) Get(projectDir string, urlOverrides *URLOverrides, hashMode HashMode) error {
	branch := defaultBranch
	if g.Branch != nil && *g.Branch != "" {
		branch = *g.Branch
	} else {
		g.Branch = &branch
	}

	dest := "."
	if g.Dest != "" {
		dest = g.Dest
	}

	hash := ""
	if g.Hash != nil {
		hash = *g.Hash
	}
	currentHash, err := gitClone(g.Label, g.URL, dest, branch, hashMode, hash, urlOverrides)
	if err != nil {
		return err
	}
	g.Hash = &currentHash

	return nil
}

func gitCloneShallow(label, repo, ref, dest string) (err error) {
	if err = os.MkdirAll(dest, 0o755); err != nil {
		return fmt.Errorf("%s: error creating %q: %w", label, dest, err)
	}
	defer func() {
		if err != nil {
			if dest != "." {
				os.RemoveAll(dest)
			} else {
				os.RemoveAll(".git")
			}
		}
	}()
	if err = runCommand("git", "-C", dest, "init", "-q"); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	if err = runCommand("git", "-C", dest, "remote", "add", "origin", repo); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	if err = runCommand("git", "-C", dest, "fetch", "-q", "--depth=1", "origin", ref); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	if err = runCommand("git", "-C", dest, "checkout", "-q", ref); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	return nil
}

func gitClone(label, repo, dest, branch string, hashMode HashMode, hash string, urlOverrides *URLOverrides) (string, error) {
	if urlOverrides != nil {
		repo = urlOverrides.Override(repo)
	}

	if branch == "" {
		return "", fmt.Errorf("%s: branch not specified", label)
	}

	if hashMode == hashModeUpdate {
		hash = ""
	}

	log.Printf("%s: Cloning %s (%s %s)...", label, repo, branch, hash)

	// Try the shallow clone first. This is much faster but requires
	// `uploadpack.allowReachableSHA1InWant` to be enabled on the server,
	// which is not the default.
	ok := false

	ref := hash
	if ref == "" {
		ref = branch
	}
	if err := gitCloneShallow(label, repo, ref, dest); err == nil {
		ok = true
	} else {
		log.Printf(
			"%s: shallow clone failed: %s,\n"+
				"== NOTE: This is likely because uploadpack.allowReachableSHA1InWant is not enabled on the server.",
			label, err)
		// Fall back to full clone
	}

	if !ok {
		if err := runCommand("git", "clone", "-q", "-b", branch, repo, dest); err != nil {
			return "", fmt.Errorf("%s: %w", label, err)
		}
		if hash != "" {
			if err := runCommand("git", "-C", dest, "checkout", "-q", hash); err != nil {
				return "", fmt.Errorf("%s: %w", label, err)
			}
		}
	}

	cmd := exec.Command("git", "-C", dest, "rev-parse", "HEAD")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: error running %v: %w", label, cmd, err)
	}
	currentHash := strings.TrimSpace(string(out))
	log.Printf("%s: Current hash is %s", label, currentHash)

	if hashMode == hashModeStrict && hash == "" {
		return currentHash, fmt.Errorf("%s: %s: hash mode is strict and no hash supplied (current is %s)", label, repo, currentHash)
	}

	return currentHash, nil
}
