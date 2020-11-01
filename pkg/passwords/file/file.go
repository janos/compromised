// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package file

import (
	"crypto/sha1"
	"fmt"
	"strconv"
)

const (
	version           = 1
	supportedHash     = "sha1"
	defaultShardCount = 32
	maxShardCount     = 256

	partitionSize            = 3 // uint24 size in bytes
	indexLocationEncodedSize = 4 // uint32 size in bytes
	indexReadSize            = indexLocationEncodedSize * 2
	hashRemainderSize        = sha1.Size - partitionSize
	maxUint24                = 1<<24 - 1
)

var validShardCounts = []int{1, 2, 4, 8, 16, 32, 64, 128, 256}

type meta struct {
	Version      int    `json:"version"`
	Hash         string `json:"hash"`
	Count        uint64 `json:"count"`
	MinHashCount uint64 `json:"min_hash_count"`
	MaxHashCount uint64 `json:"max_hash_count"`
	ShardCount   int    `json:"shard_count"`
	CountDecoder string `json:"count_decoder"`
}

func isShardCountValid(v int) bool {
	for _, c := range validShardCounts {
		if v == c {
			return true
		}
	}
	return false
}

func getShard(b, shardCount int) int {
	d := maxShardCount / shardCount
	return b / d
}

func getShardFilename(shard, shardCount int) string {
	if shardCount == 1 {
		return "hashes.db"
	}
	if shardCount > 36 {
		n := strconv.FormatUint(uint64(shard), 36)
		if len(n) == 1 {
			n = "0" + n
		}
		return fmt.Sprintf("hashes-%s.db", n)
	}
	return fmt.Sprintf("hashes-%s.db", strconv.FormatUint(uint64(shard), 36))
}

func uint24(b []byte) uint64 {
	_ = b[2] // bounds check hint to compiler
	return uint64(b[2]) | uint64(b[1])<<8 | uint64(b[0])<<16
}
