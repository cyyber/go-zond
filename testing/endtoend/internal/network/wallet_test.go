// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureWalletCreatesAndReusesPrivateSeed(t *testing.T) {
	networkDir, err := ensureNetworkDirectory(t.TempDir())
	require.NoError(t, err)
	wallet, err := ensureWallet(networkDir)
	require.NoError(t, err)
	seedPath := walletSeedPath(networkDir)
	seed, err := os.ReadFile(seedPath)
	require.NoError(t, err)
	info, err := os.Lstat(seedPath)
	require.NoError(t, err)
	require.True(t, info.Mode().IsRegular())
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	require.Zero(t, info.Mode()&os.ModeSymlink)
	again, err := ensureWallet(networkDir)
	require.NoError(t, err)
	require.Equal(t, wallet, again)
	reused, err := os.ReadFile(seedPath)
	require.NoError(t, err)
	require.Equal(t, seed, reused, "wallet reuse replaced the seed")
}

func TestEnsureWalletRejectsInvalidExistingSeed(t *testing.T) {
	networkDir, err := ensureNetworkDirectory(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(walletSeedPath(networkDir), []byte("invalid\n"), 0o600))
	_, err = ensureWallet(networkDir)
	require.ErrorContains(t, err, "restore existing wallet")
}

func TestExclusiveWriteCreatesOnePrivateFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret")
	require.NoError(t, writeExclusive(path, []byte("first\n")))
	require.ErrorIs(t, writeExclusive(path, []byte("second\n")), os.ErrExist)
	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "first\n", string(contents))
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}
