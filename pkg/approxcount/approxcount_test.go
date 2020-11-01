// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package approxcount_test

import (
	"math"
	"strconv"
	"testing"

	"resenje.org/compromised/pkg/approxcount"
)

func TestEncoder(t *testing.T) {
	for _, tc := range []struct {
		max    uint64
		values []uint64
		want   []uint64
	}{
		{
			max:    1,
			values: []uint64{1},
			want:   []uint64{1},
		},
		{
			max:    2,
			values: []uint64{1, 2},
			want:   []uint64{1, 2},
		},
		{
			max:    10,
			values: []uint64{1, 2, 3, 9, 10},
			want:   []uint64{1, 2, 3, 9, 10},
		},
		{
			max:    254,
			values: []uint64{1, 2, 3, 253, 254},
			want:   []uint64{1, 2, 3, 254, 254},
		},
		{
			max:    255,
			values: []uint64{1, 2, 254, 255},
			want:   []uint64{1, 2, 255, 255},
		},
		{
			max:    256,
			values: []uint64{1, 53, 54, 55, 56, 57, 253, 254, 255, 256},
			want:   []uint64{1, 53, 53, 55, 56, 57, 250, 256, 256, 256},
		},
		{
			max:    23597311,
			values: []uint64{1, 14, 15, 16, 17, 18, 1000, 23597310, 23597311},
			want:   []uint64{1, 14, 15, 16, 18, 18, 1016, 23597311, 23597311},
		},
		{
			max:    math.MaxUint32,
			values: []uint64{1, 7, 8, 9, 10, 11, 1000000, math.MaxUint32 - 1, math.MaxUint32},
			want:   []uint64{1, 7, 8, 9, 10, 11, 1014925, math.MaxUint32, math.MaxUint32},
		},
		{
			max:    math.MaxUint64,
			values: []uint64{1, 7, 8, 9, 10, 11, 1000000, math.MaxUint64 - 1, math.MaxUint64},
			want:   []uint64{1, 7, 8, 10, 10, 11, 930374, 18446744073709524992, 18446744073709524992},
		},
	} {
		t.Run(strconv.FormatUint(tc.max, 10), func(t *testing.T) {
			e, err := approxcount.NewEncoder(tc.max)
			if err != nil {
				t.Fatal(err)
			}
			for i, value := range tc.values {
				t.Run(strconv.FormatUint(value, 10), func(t *testing.T) {
					want := tc.want[i]
					encoded := e.Encode(value)
					got := e.Decode(encoded)
					if got != want {
						t.Errorf("got %v, want %v for %v (encoded %v)", got, want, value, encoded)
					}
				})
			}
		})
	}
}

func TestEncoderPanic(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("got no panic")
		}
		want := "overflow"
		if err != want {
			t.Fatalf("got panic error message %q, want %q", err, want)
		}
	}()

	e, err := approxcount.NewEncoder(100)
	if err != nil {
		t.Fatal(err)
	}
	e.Encode(101)
}
