// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"

	filepasswords "resenje.org/compromised/pkg/passwords/file"
)

func indexPasswordsCmd() {
	cli := flag.NewFlagSet("index-passwords", flag.ExitOnError)

	minHashCount := cli.Uint64("min-hash-count", 1, "Skip hashes with counts lower than specified with this flag.")
	shardCount := cli.Int("shard-count", 32, "Split hashes into a several files. Possible values: 1, 2, 4, 8, 16, 32, 64, 128, 256.")
	hashCounting := cli.String("hash-counting", "exact", "Store approximate hash counts. Possible values: exact, approx, none.")

	help := cli.Bool("h", false, "Show program usage.")

	cli.Usage = func() {
		fmt.Fprintf(os.Stderr, `USAGE

  index-passwords [input filename] [output directory]

OPTIONS

`)
		cli.PrintDefaults()
	}

	if err := cli.Parse(os.Args[2:]); err != nil {
		cli.Usage()
		os.Exit(2)
	}

	if *help {
		cli.Usage()
		return
	}

	if cli.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "index-passwords command requires two arguments: input filename and output directory")
		cli.Usage()
		os.Exit(2)
	}

	if err := filepasswords.Index(cli.Arg(0), cli.Arg(1), &filepasswords.IndexOptions{
		MinHashCount: *minHashCount,
		ShardCount:   *shardCount,
		HashCounting: filepasswords.HashCounting(*hashCounting),
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
