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

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestAddressesDeterministic runs the generator twice into separate tmp dirs
// and asserts the output is byte-identical. Protects the CI contract that
// "regenerate then diff" stays empty when no source changed.
func TestAddressesDeterministic(t *testing.T) {
	a := t.TempDir()
	b := t.TempDir()
	if err := generateAddresses(a); err != nil {
		t.Fatalf("first run: %v", err)
	}
	if err := generateAddresses(b); err != nil {
		t.Fatalf("second run: %v", err)
	}
	read := func(dir string) []byte {
		data, err := os.ReadFile(filepath.Join(dir, "testdata", "addresses.json"))
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	got := read(a)
	want := read(b)
	if !bytes.Equal(got, want) {
		t.Fatalf("generator is non-deterministic:\nsha256(a)=%x\nsha256(b)=%x",
			sha256.Sum256(got), sha256.Sum256(want))
	}

	var accounts []testAccount
	if err := json.Unmarshal(got, &accounts); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	if len(accounts) != len(addressLabels) {
		t.Fatalf("expected %d accounts, got %d", len(addressLabels), len(accounts))
	}
	// Every address must be a well-formed 48-byte Q-prefixed hex string.
	for _, acc := range accounts {
		if len(acc.Address) != 1+2*48 {
			t.Errorf("%s: address wrong length %d, want %d", acc.Label, len(acc.Address), 1+2*48)
		}
		if acc.Address[0] != 'Q' {
			t.Errorf("%s: address missing Q prefix: %q", acc.Label, acc.Address)
		}
		if _, err := hex.DecodeString(acc.Address[1:]); err != nil {
			t.Errorf("%s: address not hex: %v", acc.Label, err)
		}
	}
}
