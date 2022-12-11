// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package file

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"resenje.org/compromised/pkg/approxcount"
)

// IndexOptions holds oprional parameter for indexing pwned passwords data.
type IndexOptions struct {
	// MinHashCount filters out hashes with lower compromised counts from indexing.
	MinHashCount uint64
	// ShardCount specifies the number of files into which hashes should be stored.
	// Allowed values are 1, 2, 4, 8, 16, 32, 64, 128 and 256.
	ShardCount int
	// HashCounting specifies if hashes compromised count should be exact,
	// approximate or none in order to have more compact database.
	HashCounting HashCounting
	// LogFunc can be specified as a custom receiver of log messages.
	LogFunc func(string, ...interface{})
}

// HashCounting enumerates hash counting types.
type HashCounting string

var (
	// HashCountingExact stores counts as they are.
	HashCountingExact HashCounting = "exact"
	// HashCountingApprox stores counts with approximation of around 5%.
	HashCountingApprox HashCounting = "approx"
	// HashCountingNone does not store counts.
	HashCountingNone HashCounting = "none"
)

// Index creates an indexed database of pwned passwords by reading hashes and
// their counts from a textual file where hashes are ordered by their values
// provided by https://haveibeenpwned.com/Passwords. It returns the number of
// saved hashes.
func Index(inputFilename, outputDir string, o *IndexOptions) (uint64, error) {
	if o == nil {
		o = new(IndexOptions)
	}
	if o.MinHashCount < 1 {
		o.MinHashCount = 1
	}
	if o.ShardCount == 0 {
		o.ShardCount = defaultShardCount
	}
	if !isShardCountValid(o.ShardCount) {
		return 0, errors.New("invalid shard count")
	}
	if o.HashCounting == "" {
		o.HashCounting = HashCountingExact
	}
	if o.LogFunc == nil {
		o.LogFunc = func(format string, a ...interface{}) {
			fmt.Printf(format+"\n", a...)
		}
	}
	logFunc := o.LogFunc

	if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
		return 0, fmt.Errorf("database directory %s already exists", outputDir)
	}

	inputFile, err := os.Open(inputFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("input file %s does not exist", inputFilename)
		}
		return 0, fmt.Errorf("open input file: %w", err)
	}
	defer inputFile.Close()

	logFunc("analyzing input file %s", inputFilename)

	progressTicker := time.NewTicker(10 * time.Second)
	defer progressTicker.Stop()

	start := time.Now()

	stat, err := inputFile.Stat()
	if err != nil {
		return 0, fmt.Errorf("input file stat: %w", err)
	}

	inputFileSize := stat.Size()

	var i uint64
	var fileCursor uint64
	var maxHashCount uint64
	scanner := bufio.NewScanner(inputFile)
	var prevLine string
	for scanner.Scan() {
		s := scanner.Text()

		fileCursor += uint64(len(s)) + 1

		if s[:40] < prevLine {
			return 0, errors.New("input file is not sorted by hashes")
		}

		c, err := strconv.ParseUint(s[41:], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("convert count to integer: line %v: %v", i, err)
		}

		if c < o.MinHashCount {
			continue
		}

		i++

		if c > maxHashCount {
			maxHashCount = c
		}

		select {
		case <-progressTicker.C:
			p := float64(fileCursor) / float64(inputFileSize) * 100
			d := time.Since(start)
			logFunc("line: %v\tprogress: %.2f%%\teta: %v", i, p, time.Duration(float64(d)*100/p)-d)
		default:
		}
	}

	count := i

	var countDecoder string
	switch o.HashCounting {
	case HashCountingExact:
		countDecoder = "big32"
	case HashCountingApprox:
		countDecoder = "approx8"
	case HashCountingNone:
		countDecoder = "none"
	default:
		return 0, fmt.Errorf("unsupported hash counter %s", o.HashCounting)
	}

	dbSize, err := getDBSize(count, o.ShardCount, supportedHash, o.HashCounting)
	if err != nil {
		return 0, fmt.Errorf("get db size: %w", err)
	}

	logFunc("total hashes: %v", count)
	logFunc("estimated db size: %v", formatBytes(dbSize))
	logFunc("max hash count: %v", maxHashCount)

	if err := os.MkdirAll(outputDir, 0777); err != nil {
		return 0, fmt.Errorf("create output dir %s: %w", outputDir, err)
	}

	metaFile, err := os.Create(filepath.Join(outputDir, "db.json"))
	if err != nil {
		return 0, fmt.Errorf("create db.json file: %w", err)
	}
	defer metaFile.Close()

	b, err := json.MarshalIndent(meta{
		Version:      version,
		Hash:         supportedHash,
		Count:        count,
		MaxHashCount: maxHashCount,
		MinHashCount: o.MinHashCount,
		ShardCount:   o.ShardCount,
		CountDecoder: countDecoder,
	}, "", "    ")
	if err != nil {
		return 0, fmt.Errorf("encode db.json: %w", err)
	}

	if err := os.WriteFile(filepath.Join(outputDir, "db.json"), b, 0666); err != nil {
		return 0, fmt.Errorf("write db.json: %w", err)
	}

	hashFileWriters := make(map[int]*bufio.Writer, o.ShardCount)
	for i := 0; i < o.ShardCount; i++ {
		f, err := os.Create(filepath.Join(
			outputDir,
			getShardFilename(i, o.ShardCount),
		))
		if err != nil {
			return 0, fmt.Errorf("create hashes file %v: %w", i, err)
		}
		defer f.Close()

		w := bufio.NewWriterSize(f, 64*1024)
		hashFileWriters[i] = w
		defer w.Flush()
	}

	logFunc("saving to: %v", outputDir)

	start = time.Now()

	if _, err := inputFile.Seek(0, io.SeekStart); err != nil {
		return 0, fmt.Errorf("seek to the beginning of input file: %w", err)
	}

	var hashCountEncoder func(uint64) []byte
	switch o.HashCounting {
	case HashCountingExact:
		b := make([]byte, 4)
		hashCountEncoder = func(v uint64) []byte {
			binary.BigEndian.PutUint32(b, uint32(v))
			return b
		}
	case HashCountingApprox:
		e, err := approxcount.NewEncoder(uint64(maxHashCount))
		if err != nil {
			return 0, fmt.Errorf("new approxcount %v: %w", maxHashCount, err)
		}
		hashCountEncoder = func(v uint64) []byte {
			return []byte{e.Encode(v)}
		}
	case HashCountingNone:
		hashCountEncoder = func(v uint64) []byte {
			return nil
		}
	default:
		return 0, fmt.Errorf("unsupported hash counter %s", o.HashCounting)
	}

	indexFile, err := os.Create(filepath.Join(outputDir, "index.db"))
	if err != nil {
		return 0, fmt.Errorf("create index file: %w", err)
	}
	defer indexFile.Close()

	indexFileWriter := bufio.NewWriterSize(indexFile, 64*1024)
	defer indexFileWriter.Flush()

	i = 0
	fileCursor = 0
	var hashFileIndex uint32
	var partition, previousPartition uint64
	var previousShard int
	scanner = bufio.NewScanner(inputFile)
	if _, err := indexFileWriter.Write(make([]byte, 4)); err != nil {
		return 0, fmt.Errorf("write index zero entry: %w", err)
	}
	buf := make([]byte, 4)
	for scanner.Scan() {
		s := scanner.Text()

		fileCursor += uint64(len(s)) + 1

		count, err := strconv.ParseUint(s[41:], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("convert count to integer: line %v: %w", i, err)
		}

		prefix, err := hex.DecodeString(s[:6])
		if err != nil {
			return 0, fmt.Errorf("decode prefix: line %v: %w", i, err)
		}

		partition = uint24(prefix)
		if partition != previousPartition {
			if partition < previousPartition {
				return 0, fmt.Errorf("partition %v not after partition %v", partition, previousPartition)
			}
			for i := previousPartition; i < partition; i++ {
				binary.BigEndian.PutUint32(buf, hashFileIndex)
				if _, err := indexFileWriter.Write(buf); err != nil {
					return 0, fmt.Errorf("write index cursor: line %v: %w", i, err)
				}
			}
			previousPartition = partition
		}

		shard := getShard(int(prefix[0]), o.ShardCount)
		if shard != previousShard {
			hashFileIndex = 0
			previousShard = shard
			if _, err := indexFileWriter.Write(make([]byte, 4)); err != nil {
				return 0, fmt.Errorf("write index zero entry: %w", err)
			}
		}

		if count >= o.MinHashCount {
			hash, err := hex.DecodeString(s[6:40])
			if err != nil {
				return 0, fmt.Errorf("decode hash: line %v: %w", i, err)
			}

			if _, err := hashFileWriters[shard].Write(hash); err != nil {
				return 0, fmt.Errorf("write hash: line %v: %w", i, err)
			}
			if _, err := hashFileWriters[shard].Write(hashCountEncoder(uint64(count))); err != nil {
				return 0, fmt.Errorf("write hash count: line %v: %w", i, err)
			}

			i++
			hashFileIndex++
		}

		select {
		case <-progressTicker.C:
			p := float64(fileCursor) / float64(inputFileSize) * 100
			d := time.Since(start)
			logFunc("line: %v\tprogress: %.2f%%\teta: %v", i, p, time.Duration(float64(d)*100/p)-d)
		default:
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("read input file: %w", err)
	}

	binary.BigEndian.PutUint32(buf, uint32(hashFileIndex))
	for i := partition; i <= maxUint24; i++ {
		if _, err := indexFileWriter.Write(buf); err != nil {
			return 0, fmt.Errorf("write index end %v: %w", i, err)
		}
	}

	logFunc("saved %v hashes", i)

	return i, nil
}

func getDBSize(count uint64, shardCount int, hash string, hashCounting HashCounting) (uint64, error) {
	indexMaxSeek := (maxUint24 + shardCount) * indexLocationEncodedSize

	indexFileSize := indexMaxSeek + indexReadSize

	if hash != supportedHash {
		return 0, errors.New("unsupported hashing algorithm")
	}

	var countEncodedSize uint64
	switch hashCounting {
	case HashCountingExact:
		countEncodedSize = 4
	case HashCountingApprox:
		countEncodedSize = 1
	case HashCountingNone:
		countEncodedSize = 0
	default:
		return 0, fmt.Errorf("unsupported hash counter %s", hashCounting)
	}

	hashesFileSize := (hashRemainderSize + countEncodedSize) * count

	return uint64(indexFileSize) + hashesFileSize, nil
}

func formatBytes(v uint64) string {
	if v < 1024 {
		return fmt.Sprintf("%d bytes", v)
	}
	var d uint64 = 1024
	if v < d*1024 {
		return fmt.Sprintf("%.2f KiB", float64(v)/float64(d))
	}
	d *= 1024
	if v < d*1024 {
		return fmt.Sprintf("%.2f MiB", float64(v)/float64(d))
	}
	d *= 1024
	return fmt.Sprintf("%.2f GiB", float64(v)/float64(d))
}
