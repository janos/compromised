// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package approxcount

import (
	"fmt"
	"math"
)

// Encoder encodes and decodes integer values to and from 8 bytes with precision
// degrading logarithmically with higher values. Encoder must be aware of
// maximal possible value and starts always from zero.
//
// Minimal value that can be encoded is 1.
type Encoder struct {
	max uint64
	c   float64
}

// NewEncoder creates a new Encoder with up to max value able to encode.
func NewEncoder(max uint64) (*Encoder, error) {
	if max < 1 {
		return nil, fmt.Errorf("invalid max value %v", max)
	}
	return &Encoder{
		max: max,
		c:   255 / math.Log(float64(max)),
	}, nil
}

// Encode returns an approximation of integer in 8 bytes format.
func (e *Encoder) Encode(value uint64) uint8 {
	if value > e.max || value < 1 {
		panic("overflow")
	}

	return uint8(math.Round(math.Log(float64(value)) * e.c))
}

// Decode returns an approximated integer from 8 bytes.
func (e *Encoder) Decode(encoded uint8) uint64 {
	return uint64(math.Round(math.Exp(float64(encoded) / e.c)))
}
