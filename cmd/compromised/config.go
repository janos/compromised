// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	xconfig "resenje.org/x/config"

	"resenje.org/compromised/cmd/compromised/config"
)

var (
	// Initialize configurations with default values.
	options = config.NewCompromisedOptions()
)

func init() {
	// Register options in config.
	cfg.Register(config.Name, options)
}

var cfg = xconfig.New(config.Name)

func configCmd() error {
	// Print loaded configuration.
	fmt.Print(cfg.String())
	return nil
}

func updateConfig() error {
	if *configDir == "" {
		*configDir = os.Getenv(strings.ToUpper(config.Name) + "_CONFIGDIR")
	}
	if *configDir == "" {
		*configDir = config.Dir
	}

	cfg.Dirs = []string{
		*configDir,
	}
	if d, err := os.UserConfigDir(); err == nil {
		cfg.Dirs = append(cfg.Dirs, filepath.Join(d, config.Name))
	}
	return cfg.Load()
}

func verifyAndPrepareConfig() error {
	return cfg.VerifyAndPrepare()
}
