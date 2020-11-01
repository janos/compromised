// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package compromised creates executable for the compromised program.
//
// Configuration loading, validation and initialization of all required
// services for server to function is done in this package. It integrates
// all server dependencies into a form of a single executable.

package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	cli = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	configDir = cli.String("config-dir", "", "Directory that contains configuration files.")
	help      = cli.Bool("h", false, "Show program usage.")
)

func main() {
	cli.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, os.Args[0])
		cli.PrintDefaults()
	}

	if err := cli.Parse(os.Args[1:]); err != nil {
		helpCmd()
		return
	}

	cmd := cli.Arg(0)

	if *help && cmd == "" {
		helpCmd()
		return
	}

	switch cmd {
	case "index-passwords":
		indexPasswordsCmd()
		return

	case "version":
		versionCmd()
		return

	}

	updateConfig()

	switch cmd {
	case "", "daemon":
		verifyAndPrepareConfig()
		startCmd(cmd == "daemon")

	case "stop":
		stopCmd()

	case "status":
		statusCmd()

	case "debug-dump":
		debugDumpCmd()

	case "config":
		configCmd()

	case "index-passwords":
		indexPasswordsCmd()

	default:
		helpUnknownCmd(cmd)
	}
}
