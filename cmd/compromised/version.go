// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"resenje.org/compromised"
)

func versionCmd() {
	fmt.Println(compromised.Version)
}
