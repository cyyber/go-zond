// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNetworkDirectoryIsPrivateCanonicalDirectory(t *testing.T) {
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

func TestEnsureNetworkDirectoryRejectsWithoutChangingExistingPermissions(t *testing.T) {
	networkDir := t.TempDir()
	require.NoError(t, os.Chmod(networkDir, 0o755))
	_, err := ensureNetworkDirectory(networkDir)
	require.ErrorContains(t, err, "0700 permissions")
	info, statErr := os.Stat(networkDir)
	require.NoError(t, statErr)
	require.Equal(t, os.FileMode(0o755), info.Mode().Perm())
}
