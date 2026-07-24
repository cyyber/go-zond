// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureWalletCreatesAndReusesPrivateSeed(t *testing.T) {
	dir := t.TempDir()
	wallet, err := ensureWallet(dir)
	if err != nil {
		t.Fatal(err)
	}
	seedPath := filepath.Join(dir, seedName)
	seed, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(seedPath)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("private seed metadata = %v, %v", info, err)
	}
	again, err := ensureWallet(dir)
	if err != nil || again != wallet {
		t.Fatalf("reused wallet = %+v, %v", again, err)
	}
	reused, _ := os.ReadFile(seedPath)
	if string(reused) != string(seed) {
		t.Fatal("wallet reuse replaced the seed")
	}
}

func TestEnsureWalletRejectsInvalidExistingSeed(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, seedName), []byte("invalid\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ensureWallet(dir); err == nil || !strings.Contains(err.Error(), "restore existing wallet") {
		t.Fatalf("invalid-seed error = %v", err)
	}
}
