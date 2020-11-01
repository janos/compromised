// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package main

import (
	"fmt"
	"os"
)

func debugDumpCmd() {
	fmt.Fprintln(os.Stderr, "Debug dump is not supported on Windows")
	os.Exit(2)
}
