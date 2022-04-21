// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"resenje.org/daemon"
)

func statusCmd() error {
	// Use daemon.Daemon to obtain status information and print it.
	pid, err := (&daemon.Daemon{
		PidFileName: options.PidFileName,
	}).Status()
	if err != nil {
		return fmt.Errorf("not running: %w", err)
	}
	fmt.Println("Running: PID", pid)
	return nil
}
