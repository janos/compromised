// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"resenje.org/daemon"
	"resenje.org/x/application"
)

func stopCmd() error {
	err := application.StopDaemon(daemon.Daemon{
		PidFileName: options.PidFileName,
	})
	if err != nil {
		return err
	}
	fmt.Println("Stopped")
	return nil
}
