// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureWalletCreatesAndReusesPrivateSeed(t *testing.T) {
	networkDir, err := ensureNetworkDirectory(filepath.Join(t.TempDir(), "network"))
	require.NoError(t, err)
	walletAddress, err := ensureWallet(networkDir)
	require.NoError(t, err)
	reusedAddress, err := ensureWallet(networkDir)
	require.NoError(t, err)
	require.Equal(t, walletAddress, reusedAddress)
}
