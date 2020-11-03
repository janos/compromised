// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package file

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"resenje.org/compromised/pkg/approxcount"
	"resenje.org/compromised/pkg/passwords"
)

var _ passwords.Service = (*Service)(nil)

// Service implements passwords service by reading the passwords hash data
// directly from files stored on the filesystem.
type Service struct {
	index            *os.File
	shards           map[int]*os.File
	shardCount       int
	countDecoder     func([]byte) uint64
	countEncodedSize int64
	metrics          metrics
}

// New creates a new instance of Service by opening database files in a provided
// directory location on the filesystem.
func New(dir string) (*Service, error) {
	b, err := ioutil.ReadFile(filepath.Join(dir, "db.json"))
	if err != nil {
		return nil, err
	}
	var m meta
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if m.Version > version {
		return nil, errors.New("unsupported data version")
	}
	if m.Hash != supportedHash {
		return nil, errors.New("unsupported hashing algorithm")
	}
	if !isShardCountValid(m.ShardCount) {
		return nil, errors.New("invalid shard count")
	}

	var countDecoder func([]byte) uint64
	var countEncodedSize int64
	switch m.CountDecoder {
	case "big32":
		countDecoder = func(b []byte) uint64 {
			return uint64(binary.BigEndian.Uint32(b))
		}
		countEncodedSize = 4
	case "approx8":
		e, err := approxcount.NewEncoder(uint64(m.MaxHashCount))
		if err != nil {
			return nil, err
		}
		countDecoder = func(b []byte) uint64 {
			return e.Decode(b[0])
		}
		countEncodedSize = 1
	case "none":
		countDecoder = func(b []byte) uint64 {
			return 1
		}
		countEncodedSize = 0
	default:
		return nil, errors.New("invalid count decoder")
	}

	index, err := os.Open(filepath.Join(dir, "index.db"))
	if err != nil {
		return nil, err
	}
	shards := make(map[int]*os.File, m.ShardCount)
	for i := 0; i < m.ShardCount; i++ {
		shards[i], err = os.Open(filepath.Join(
			dir,
			getShardFilename(i, m.ShardCount),
		))
		if err != nil {
			return nil, fmt.Errorf("open hashes file %v: %w", i, err)
		}
	}
	return &Service{
		index:            index,
		shards:           shards,
		shardCount:       m.ShardCount,
		countDecoder:     countDecoder,
		countEncodedSize: countEncodedSize,
		metrics:          newMetrics(),
	}, nil
}

const maxReaderBufferSize = 4096

// IsPasswordCompromised provides information if the password is compromised by
// reading the index and hashes files.
func (s *Service) IsPasswordCompromised(_ context.Context, sum [20]byte) (count uint64, err error) {
	shard := getShard(int(sum[0]), s.shardCount)

	partition := uint24(sum[:partitionSize])

	indexLocation := (int64(partition) + int64(shard)) * indexLocationEncodedSize // add the shard count as it starts with a zero value step
	indexCursor, err := s.index.Seek(indexLocation, io.SeekStart)
	if err != nil {
		return 0, fmt.Errorf("index seek to %v: %w", indexLocation, err)
	}
	if indexLocation != indexCursor {
		return 0, fmt.Errorf("index out of range: %v instead %v", indexCursor, indexLocation)
	}

	buf := make([]byte, indexReadSize)
	n, err := s.index.Read(buf)
	if err != nil {
		return 0, fmt.Errorf("index read %v at %v: %w", len(buf), indexCursor, err)
	}
	if n != len(buf) {
		return 0, fmt.Errorf("index short read at %v: %v instead %v", indexCursor, n, len(buf))
	}

	hashRemainderStep := hashRemainderSize + s.countEncodedSize

	hashRemaindersStart := int64(binary.BigEndian.Uint32(buf[:indexLocationEncodedSize])) * hashRemainderStep
	hashRemaindersEnd := int64(binary.BigEndian.Uint32(buf[indexLocationEncodedSize:indexLocationEncodedSize*2])) * hashRemainderStep

	hashFile := s.shards[shard]

	hashRemaindersCursor, err := hashFile.Seek(hashRemaindersStart, io.SeekStart)
	if err != nil {
		return 0, fmt.Errorf("hashes %v seek to %v: %w", shard, hashRemaindersStart, err)
	}
	if hashRemaindersStart != hashRemaindersCursor {
		return 0, fmt.Errorf("hashes out of range: %v instead %v", hashRemaindersCursor, hashRemaindersStart)
	}

	readerBufferSize := hashRemaindersEnd - hashRemaindersStart
	if readerBufferSize > maxReaderBufferSize || readerBufferSize < hashRemainderStep {
		readerBufferSize = maxReaderBufferSize
	}

	buf = make([]byte, hashRemainderStep)
	passwordHashRemainder := sum[partitionSize:]
	hashRemaindersCursor = hashRemaindersStart
	hashFileReader := bufio.NewReaderSize(hashFile, int(readerBufferSize))
	for hashRemaindersCursor < hashRemaindersEnd {
		n, err := hashFileReader.Read(buf)
		if err != nil {
			return 0, fmt.Errorf("hashes %v read %v at %v: %w", shard, len(buf), hashRemaindersCursor, err)
		}
		if n != len(buf) {
			return 0, fmt.Errorf("hashes short read at cursor %v: %v instead %v", hashRemaindersCursor, n, len(buf))
		}
		if bytes.Equal(passwordHashRemainder, buf[:hashRemainderSize]) {
			return s.countDecoder(buf[hashRemainderSize:]), nil
		}
		hashRemaindersCursor += int64(n)
	}

	return 0, nil
}

// Close closes all open files.
func (s *Service) Close() error {
	for v, f := range s.shards {
		if err := f.Close(); err != nil {
			return fmt.Errorf("close hashes file %v: %w", v, err)
		}
	}
	return s.index.Close()
}
