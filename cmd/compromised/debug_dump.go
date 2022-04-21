// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !windows
// +build !windows

package main

import (
	"syscall"

	"resenje.org/daemon"
)

func debugDumpCmd() error {
	// Send SIGUSR1 signal to a daemonized process.
	// Service is able to receive the signal and dump debugging
	// information to files or stderr.
	return (&daemon.Daemon{
		PidFileName: options.PidFileName,
	}).Signal(syscall.SIGUSR1)
}
