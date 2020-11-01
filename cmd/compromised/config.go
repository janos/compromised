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

	"resenje.org/x/application"

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

var cfg = application.NewConfig(config.Name)

func configCmd() {
	// Print loaded configuration.
	fmt.Print(cfg.String())
}

func updateConfig() {
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
	if err := cfg.Load(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
		os.Exit(2)
	}
}

func verifyAndPrepareConfig() {
	if err := cfg.VerifyAndPrepare(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		if e, ok := err.(*application.HelpError); ok {
			fmt.Println()
			fmt.Println(e.Help)
		}
		os.Exit(2)
	}
}
