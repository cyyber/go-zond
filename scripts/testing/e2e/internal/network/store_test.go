// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

func TestNetworkDirectory(t *testing.T) {
	parent, linkRoot := t.TempDir(), t.TempDir()
	link := filepath.Join(linkRoot, "parent")
	require.NoError(t, os.Symlink(parent, link))
	networkDir, err := ensureNetworkDirectory(filepath.Join(link, "network"))
	require.NoError(t, err)
	require.True(t, filepath.IsAbs(networkDir))
	require.NotContains(t, networkDir, linkRoot)
	info, err := os.Stat(networkDir)
	require.NoError(t, err)
	require.True(t, info.IsDir())
	require.Equal(t, os.FileMode(0o700), info.Mode().Perm())
}

func TestOwnershipRequiresExactUUIDForDestruction(t *testing.T) {
	networkDir := t.TempDir()
	enclave := kurtosis.EnclaveRef{Name: enclaveNamePrefix(networkDir) + "attempt"}
	require.ErrorContains(t, validateOwnershipDirectory(networkDir, enclave), "UUID")
	enclave.UUID = strings.Repeat("a", 32)
	require.NoError(t, validateOwnershipDirectory(networkDir, enclave))
}

func TestOwnershipPersistsOneExclusiveExactIdentity(t *testing.T) {
	networkDir := t.TempDir()
	record := kurtosis.EnclaveRef{
		Name: enclaveNamePrefix(networkDir) + "0123456789abcdef0123456789abcdef",
		UUID: strings.Repeat("a", 32),
	}
	require.NoError(t, createOwnership(networkDir, record))
	require.ErrorIs(t, createOwnership(networkDir, record), os.ErrExist)
	loaded, err := loadOwnership(networkDir)
	require.NoError(t, err)
	require.Equal(t, record, loaded)
}

func TestOwnershipCannotBeCopiedToAnotherNetworkDirectory(t *testing.T) {
	source, destination := t.TempDir(), t.TempDir()
	record := kurtosis.EnclaveRef{
		Name: enclaveNamePrefix(source) + "0123456789abcdef0123456789abcdef",
		UUID: strings.Repeat("a", 32),
	}
	require.NoError(t, createOwnership(source, record))
	payload, err := os.ReadFile(ownershipPath(source))
	require.NoError(t, err)
	require.NoError(t, writeExclusive(ownershipPath(destination), payload))
	_, err = loadOwnership(destination)
	require.ErrorContains(t, err, "does not belong")
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
