// Copyright 2026 The go-qrl Authors
// This file is part of go-qrl.
//
// go-qrl is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-qrl is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// Command gen regenerates fixture files used by the go-qrl test suite.
//
// Usage:
//
//	go run ./hack/gen --target=all
//	go run ./hack/gen --target=addresses --out=.
//
// Each --target maps to a generator in this package. Generators must be
// deterministic so consecutive invocations produce byte-identical files; this
// keeps the CI diff-check honest.
package main

import (
	"flag"
	"fmt"
	"os"
)

type generator struct {
	name string
	fn   func(outDir string) error
}

var generators = []generator{
	{name: "addresses", fn: generateAddresses},
	{name: "blocks", fn: generateBlocks},
	{name: "transactions", fn: generateTransactions},
	{name: "receipts", fn: generateReceipts},
	{name: "abi", fn: generateABIFixtures},
}

func main() {
	target := flag.String("target", "all", "which generator to run (all, or a comma-separated subset)")
	out := flag.String("out", ".", "repository root where testdata/ lives")
	list := flag.Bool("list", false, "list available generators and exit")
	flag.Parse()

	if *list {
		fmt.Println("available generators:")
		for _, g := range generators {
			fmt.Printf("  %s\n", g.name)
		}
		return
	}

	selected := selectGenerators(*target)
	if len(selected) == 0 {
		fmt.Fprintf(os.Stderr, "no generators selected for target=%q\n", *target)
		os.Exit(2)
	}

	fmt.Printf("writing fixtures under %s\n", *out)
	for _, g := range selected {
		if err := g.fn(*out); err != nil {
			fmt.Fprintf(os.Stderr, "generator %s failed: %v\n", g.name, err)
			os.Exit(1)
		}
	}
}

func selectGenerators(target string) []generator {
	if target == "all" {
		return generators
	}
	wanted := map[string]bool{}
	start := 0
	for i := 0; i <= len(target); i++ {
		if i == len(target) || target[i] == ',' {
			if start < i {
				wanted[target[start:i]] = true
			}
			start = i + 1
		}
	}
	out := make([]generator, 0, len(wanted))
	for _, g := range generators {
		if wanted[g.name] {
			out = append(out, g)
		}
	}
	return out
}
