// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows
// +build windows

package main

import (
	"errors"
)

func debugDumpCmd() error {
	return errors.New("debug dump is not supported on Windows")
}
