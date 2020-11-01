// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !windows

package main

import (
	"fmt"
	"os"
	"syscall"

	"resenje.org/daemon"
)

func debugDumpCmd() {
	// Send SIGUSR1 signal to a daemonized process.
	// Service is able to receive the signal and dump debugging
	// information to files or stderr.
	err := (&daemon.Daemon{
		PidFileName: options.PidFileName,
	}).Signal(syscall.SIGUSR1)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(2)
	}
}
