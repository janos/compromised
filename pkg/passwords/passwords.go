// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package passwords

import "context"

// Service specifies operations agains compromised passwords database.
type Service interface {
	IsPasswordCompromised(ctx context.Context, sha1Sum [20]byte) (count uint64, err error)
}
