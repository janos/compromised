// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package config holds project and service related data and structures
// that define optional parameters for different parts of the service.
package config

import (
	"os"
	"path/filepath"
)

var (
	// Name is the name of the service.
	Name = "compromised"

	// BaseDir is the directory where the service's executable is located.
	BaseDir = func() string {
		path, err := os.Executable()
		if err != nil {
			panic(err)
		}
		path, err = filepath.EvalSymlinks(path)
		if err != nil {
			panic(err)
		}
		return filepath.Dir(path)
	}()

	// Dir is default directory where configuration files are located.
	Dir = "/etc/compromised"
)
