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
)

func runCommand(bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	log.Printf("Running %v", cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running %v: %w", cmd, err)
	}
	return nil
}
