// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package file_test

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"resenje.org/compromised/pkg/passwords"
	"resenje.org/compromised/pkg/passwords/file"
)

func TestService(t *testing.T) {
	t.Run("default index options", newServiceTest(nil))

	t.Run("min hash count", newServiceTest(&file.IndexOptions{
		MinHashCount: 10,
	}))

	for _, shardCount := range []int{1, 2, 4, 8, 16, 32, 64, 128, 256} {
		t.Run(fmt.Sprintf("shard count %v", shardCount), newServiceTest(&file.IndexOptions{
			ShardCount: shardCount,
		}))
	}

	t.Run("approximate hash count", newServiceTest(&file.IndexOptions{
		HashCounting: file.HashCountingApprox,
	}))

	t.Run("no hash count", newServiceTest(&file.IndexOptions{
		HashCounting: file.HashCountingNone,
	}))

	t.Run("all custom index options", newServiceTest(&file.IndexOptions{
		MinHashCount: 5,
		HashCounting: file.HashCountingApprox,
		ShardCount:   8,
	}))
}

func newServiceTest(o *file.IndexOptions) func(t *testing.T) {
	return func(t *testing.T) {
		if o == nil {
			o = new(file.IndexOptions)
		}

		dir := t.TempDir()

		inputFilename := "testdata/pwned-passwords-sha1-ordered-by-hash.txt"
		dbDir := filepath.Join(dir, "db")

		count, err := file.Index(inputFilename, dbDir, o)
		if err != nil {
			t.Fatal(err)
		}

		s, err := file.New(dbDir)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()

		t.Run("hit", func(t *testing.T) {
			inputFile, err := os.Open(inputFilename)
			if err != nil {
				t.Fatal(err)
			}

			scanner := bufio.NewScanner(inputFile)
			var i uint64
			for scanner.Scan() {
				line := scanner.Text()
				i++

				hash := line[:40]

				want, err := strconv.ParseUint(line[41:], 10, 64)
				if err != nil {
					t.Fatal(err)
				}

				if o.HashCounting == file.HashCountingNone {
					want = 1
				}

				var tolerance uint64
				if o.HashCounting == file.HashCountingApprox {
					tolerance = uint64(math.Round(float64(want) / 25))
				}

				if want < o.MinHashCount {
					want = 0
				}

				isPasswordCompromised(t, s, hash, want, tolerance)
			}
			if i < count {
				t.Errorf("validated %v hashes, expected at least %v", i, count)
			}
		})

		t.Run("miss", func(t *testing.T) {
			for _, hash := range []string{
				"0000000000000000000000000000000000000000",
				"7890abcdef0123456789abcdef0123456789abcd",
				"ffffffffffffffffffffffffffffffffffffffff",
			} {
				isPasswordCompromised(t, s, hash, 0, 0)
			}
		})

		t.Run("miss edges", func(t *testing.T) {
			inputFile, err := os.Open(inputFilename)
			if err != nil {
				t.Fatal(err)
			}

			var prevLine string

			scanner := bufio.NewScanner(inputFile)
			var i uint64
			for scanner.Scan() {
				line := scanner.Text()
				i++

				if prevLine == "" {
					prevLine = line
					continue
				}

				partition := line[:6]
				remainder := line[6:40]
				prevPartition := prevLine[:6]
				prevRemainder := prevLine[6:40]

				if h := prevPartition + remainder; h != line[:40] {
					isPasswordCompromised(t, s, h, 0, 0)
				}
				if h := partition + prevRemainder; h != prevLine[:40] {
					isPasswordCompromised(t, s, h, 0, 0)
				}

				prevLine = line
			}
			if i < count {
				t.Errorf("validated %v hashes, expected at least %v", i, count)
			}
		})
	}
}

func isPasswordCompromised(t *testing.T, s passwords.Service, hash string, want, tolerance uint64) {
	t.Helper()

	got, err := s.IsPasswordCompromised(context.Background(), hexDecodeSHA1Sum(t, hash))
	if err != nil {
		t.Fatal(err)
	}
	if tolerance > 0 {
		if got < want-tolerance || got > want+tolerance {
			t.Errorf("hash %s: got count %v, want %v with tolerance %v", hash, got, want, tolerance)
		}
	} else {
		if got != want {
			t.Errorf("hash %s: got count %v, want %v", hash, got, want)
		}
	}
}

func hexDecodeSHA1Sum(t *testing.T, s string) (sum [20]byte) {
	t.Helper()

	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}

	copy(sum[:], b)
	return sum
}
