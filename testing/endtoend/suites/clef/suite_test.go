// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package clef

import (
	"math/big"
	"slices"
	"testing"
)

func TestClefServerArgsUseChainID(t *testing.T) {
	chainID := big.NewInt(424_242)
	args := clefServerArgs(t.TempDir(), 12345, chainID)
	index := slices.Index(args, "--chainid")
	if index == -1 || index+1 == len(args) {
		t.Fatal("Clef arguments have no --chainid value")
	}
	if got := args[index+1]; got != chainID.String() {
		t.Fatalf("Clef --chainid = %s, want %s", got, chainID)
	}
}
