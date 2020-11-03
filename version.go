// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compromised

var (
	version = "0.0.0" // automatically set on release
	commit  string    // automatically set git commit hash

	Version = func() string {
		if commit != "" {
			return version + "-" + commit
		}
		return version
	}()
)
