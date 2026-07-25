// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"os"
	"strings"
	"testing"
)

func TestEnsureWalletCreatesAndReusesPrivateSeed(t *testing.T) {
	networkDir, err := ensureNetworkDirectory(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	wallet, err := ensureWallet(networkDir)
	if err != nil {
		t.Fatal(err)
	}
	seedPath := walletSeedPath(networkDir)
	seed, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(seedPath)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("private seed metadata = %v, %v", info, err)
	}
	again, err := ensureWallet(networkDir)
	if err != nil || again != wallet {
		t.Fatalf("reused wallet = %+v, %v", again, err)
	}
	reused, _ := os.ReadFile(seedPath)
	if string(reused) != string(seed) {
		t.Fatal("wallet reuse replaced the seed")
	}
}

func TestEnsureWalletRejectsInvalidExistingSeed(t *testing.T) {
	networkDir, err := ensureNetworkDirectory(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(walletSeedPath(networkDir), []byte("invalid\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ensureWallet(networkDir); err == nil || !strings.Contains(err.Error(), "restore existing wallet") {
		t.Fatalf("invalid-seed error = %v", err)
	}
}
