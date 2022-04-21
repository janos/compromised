// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
)

var (
	usage = `USAGE

  %s [options...] [command]

  Executing the program without specifying a command will start a process in
  the foreground and log all messages to stderr.

COMMANDS

  daemon
    Start program in the background.

  stop
    Stop program that runs in the background.

  status
    Display status of a running process.

  config
    Print configuration that program will load on start. This command is
    dependent of -config-dir option value.

  debug-dump
    Send to a running process USR1 signal to log debug information in the log.

  index-passwords
    Generate passwords database from pwned passwords sha1 file.

  version
    Print version to Stdout.

OPTIONS

`
)

func helpCmd() {
	cli.Usage()
}

func helpUnknownCmd(cmd string) error {
	return fmt.Errorf("unknown command: %s", cmd)
}
