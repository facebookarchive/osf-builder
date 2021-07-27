// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// getdeps is a tool to fetch the dependencies necessary to build OSF (Open
// System System Firmware). It works by parsing a JSON configuration containing
// the definitions of all the components that should be fetched.
// This tool also generates a JSON file containing all the components' versions,
// suitable for using in the `internal_versions` VPD variable for OSF.
//
// The supported components currently are:
// - coreboot
// - kernel (linux)
// - initramfs (u-root)
//
// For each component it is possible to specify the source (e.g. the git repo
// or package URL), its version (e.g. the branch and git commit hash, or the
// package version and hash).
//
// It is also possible to specify an URL overrides file, which will replace the
// corresponding component's URL with the override. This is useful if you want,
// for example, use alternative mirrors and repositories for a specific
// component.
//
// The hash mode allows you to be strict or permissive in the hash validation,
// and, when used in update mode, it lets you use the latest commit hashes.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	flag "github.com/spf13/pflag"
)

// Flags.
var (
	flagHashMode = flag.StringP("hashmode", "H", string(hashModeStrict),
		"Hash verification mode: "+
			"strict - require hashes for repos and blobs, check out for repos and verify for blobs; "+
			"permissive - use hashes that are present but don't require; "+
			"update - zero out all the hashes in the beginning, update to whatever is found.")
	flagComponents = flag.StringP("components", "C", "",
		fmt.Sprintf("Comma-separated list of components to fetch. If empty or not specified, fetch all of them Supported components: %s", strings.Join(supportedComponents, ", ")))
	flagConfigFile       = flag.StringP("config", "c", "config.json", "Configuration file")
	flagURLOverridesFile = flag.StringP("url-overrides", "u", "", "URL overrides file")
	flagFinalConfigFile  = flag.StringP("output", "o", "", "Path to the output config file after all expansions, suitable for storing in the `internal_versions` VPD variable")
	flagBaseDir          = flag.StringP("basedir", "d", "", "Base directory for relative includes. If unspecified, the current working directory is used for relative includes")
)

// HashMode represents the hash mode to use. See constants below.
type HashMode string

const (
	hashModeStrict     HashMode = "strict"
	hashModePermissive HashMode = "permissive"
	hashModeUpdate     HashMode = "update"
)

var (
	defaultBranch       = "master"
	supportedComponents = []string{"initramfs", "kernel", "coreboot"}
	supportedHashModes  = []HashMode{hashModeStrict, hashModePermissive, hashModeUpdate}
)

// Component defines an interface for the different components
type Component interface {
	Get(projectDir string, overrides *URLOverrides, hashMode HashMode) error
}

// Merge fields from src object into dst.
// - If source field is zero value, skip.
//   For non-pointer fields this means they cannot be cleared.
// - If a field is a pointer field and the source points to a zero value,
//   clear the corresponding dst field (set ot nil).
// - For anything else, copy the src to dst.
func mergeFields(dst reflect.Value, src reflect.Value) {
	for i := 0; i < dst.NumField(); i++ {
		sf, df := src.Field(i), dst.Field(i)
		if sf.IsZero() {
			continue
		}
		if sf.Kind() == reflect.Ptr {
			if sf.Elem().IsZero() {
				// Create a nil pointer of field's type.
				nilValue := reflect.New(sf.Type()).Elem()
				df.Set(nilValue)
			} else {
				df.Set(sf)
			}
		} else {
			df.Set(sf)
		}
	}
}

// identifyRepo returns an identifier for the repository checked out at the
// specified directory, if any.
func identifyRepo(dir string) (string, error) {
	// Is it a Git repo?
	cmd := exec.Command("git", "describe", "--dirty", "--tags", "--always")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	// is it a Hg repo?
	cmd = exec.Command("hg", "identify")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	return "", fmt.Errorf("not a Git or Hg repo")
}

func getBuildID(configFile string, cwd string) string {
	// If running under Buck, use its working dir at the time of build invocation.
	if val, ok := os.LookupEnv("BUCK_CLIENT_PWD"); ok {
		if id, err := identifyRepo(val); err != nil {
			return id
		}
	}
	// Next, try directory of the config.
	if absConfigFilePath, err := filepath.Abs(configFile); err == nil {
		if id, err := identifyRepo(filepath.Dir(absConfigFilePath)); err == nil {
			return id
		}
	}
	// Finally, try CWD.
	if id, err := identifyRepo(cwd); err == nil {
		return id
	}
	// Give up.
	return "???"
}

// expandComponent parses a comma-separated list of components, validates their
// names, and removes duplicates.
func expandComponents(componentString string) ([]string, error) {
	// if no component is specified, assume all supported components.
	if componentString == "" {
		return supportedComponents, nil
	}
	components := strings.Split(componentString, ",")
	if len(components) == 0 {
		return supportedComponents, nil
	}
	cMap := make(map[string]struct{}, 0)
	for _, c := range components {
		found := false
		c = strings.ToLower(c)
		for _, sc := range supportedComponents {
			if c == sc {
				found = true
			}
		}
		if found == false {
			return nil, fmt.Errorf("unsupported component '%s'", c)
		}
		cMap[c] = struct{}{}
	}
	ret := make([]string, 0, len(cMap))
	for k := range cMap {
		ret = append(ret, k)
	}
	return ret, nil
}

// getBaseDir returns the absolute path of the base directory for the config files,
// given an input base dir. If the input base dir is empty, use the directory where
// the specified config file is located. This default enables config files to
// include other config files relatively to their own directory.
//
// Examples (where <PWD> means "replace with the present working directory")
//
// `basedir` is `configs`
//   then `configfile` is ignored, and the returned basedir is "<PWD>/configs".
//
// `basedir` is empty, `configfile` is "configs/myconfig.json"
//   then the returned basedir is <PWD>"/configs" (relative).
//
// `basedir` is empty, `configfile` is "/configs/myconfig.json"
//   then the returned basedir is "/configs" (absolute).
//
// `basedir` is empty, `configfile` is `configs/`
//   then the returned basedir is <PWD>"/configs".
//
// `basedir` is empty, `configfile` is empty
//   then the returned basedir is <PWD>.
func getBaseDir(basedir, configfile string) (string, error) {
	if basedir != "" {
		return basedir, nil
	}
	basedir, err := filepath.Abs(filepath.Dir(configfile))
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for '%s': %w", filepath.Dir(configfile), err)
	}
	return basedir, nil
}

func main() {
	flag.Parse()

	baseDir, err := getBaseDir(*flagBaseDir, *flagConfigFile)
	if err != nil {
		log.Fatalf("Failed to get base dir: %v", err)
	}

	// expand requested components
	components, err := expandComponents(*flagComponents)
	if err != nil {
		log.Fatalf("Invalid components: %v", err)
	}

	found := false
	for _, hm := range supportedHashModes {
		if *flagHashMode == string(hm) {
			found = true
			break
		}
	}
	if !found {
		log.Fatalf("unsupported hash mode %q", *flagHashMode)
	}

	configData, err := ioutil.ReadFile(*flagConfigFile)
	if err != nil {
		log.Fatalf("Failed to read configuration file '%s': %v", *flagConfigFile, err)
	}
	config, err := NewConfigWithIncludes(configData, baseDir)
	if err != nil {
		log.Fatalln(err)
	}

	// load URL overrides file
	var urlOverrides *URLOverrides
	if *flagURLOverridesFile != "" {
		urloverridesData, err := ioutil.ReadFile(*flagURLOverridesFile)
		if err != nil {
			log.Fatalf("Failed to open URL overrides file '%s': %v", *flagURLOverridesFile, err)
		}
		urlOverrides, err = NewURLOverrides(urloverridesData)
		if err != nil {
			log.Fatalln(err)
		}
	}

	projectDir, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	buildID := getBuildID(*flagConfigFile, projectDir)
	for _, componentName := range components {
		var component Component
		switch componentName {
		case "coreboot":
			component = config.Coreboot
		case "kernel":
			component = config.Kernel
		case "initramfs":
			component = config.Initramfs
		default:
			// this should not happen, unless the switch cases are not kept in
			// sync with the `supportedComponents` variable.
			log.Fatalf("Unsupported component '%s'. This could be a bug, please report it to the maintainers", componentName)
		}
		workingDir := path.Join(projectDir, componentName)

		log.Printf("Build ID: %s", buildID)

		// clean up previous working directory
		if err = os.RemoveAll(workingDir); err != nil {
			log.Fatalln(err)
		}

		// create new working directory
		if err = os.Mkdir(workingDir, os.ModePerm); err != nil {
			log.Fatalln(err)
		}

		// change working directory
		if err := os.Chdir(workingDir); err != nil {
			log.Fatalln(err)
		}

		// get the sources
		if err := component.Get(baseDir, urlOverrides, HashMode(*flagHashMode)); err != nil {
			log.Fatalln(err)
		}
	}

	// To ensure consistent formatting when the config is fed into vpd,
	// write out a final versions file whether or not the base config was
	// patched. This will also expose fields that were not explicitly set
	// in hand-written config files.
	// If the file already exists, override only the portion that was processed.
	finalConfigFile := *flagFinalConfigFile
	if finalConfigFile != "" {
		if !filepath.IsAbs(finalConfigFile) {
			finalConfigFile = filepath.Join(projectDir, finalConfigFile)
		}
		act := "Wrote"
		var finalConfig *Config
		finalConfigData, err := ioutil.ReadFile(finalConfigFile)
		if err == nil {
			if fc, err := NewConfig(finalConfigData); err == nil {
				finalConfig = fc
				act = "Updated"
			}
		}
		if finalConfig == nil {
			finalConfig = &Config{}
		}
		finalConfig.BuildID = buildID
		for _, componentName := range components {
			switch componentName {
			case "coreboot":
				finalConfig.Coreboot = config.Coreboot
			case "kernel":
				finalConfig.Kernel = config.Kernel
			case "initramfs":
				finalConfig.Initramfs = config.Initramfs
			}
		}
		indentedConfig, err := json.MarshalIndent(finalConfig, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal configuration: %v", err)
		}
		if err := ioutil.WriteFile(finalConfigFile, indentedConfig, 0644); err != nil {
			log.Fatalf("Failed to write generated versions to file '%s': %v", finalConfigFile, err)
		}
		log.Printf("%s %s", act, finalConfigFile)
	}
}
