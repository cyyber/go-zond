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
	require.NotContains(t, networkDir, linkRoot)
	info, err := os.Stat(networkDir)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o700), info.Mode().Perm())
}
