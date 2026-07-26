// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMutationLeaseContendsAndReopens(t *testing.T) {
	networkDir, err := ensureNetworkDirectory(t.TempDir())
	require.NoError(t, err)
	first, err := AcquireMutationLease(networkDir)
	require.NoError(t, err)
	if second, err := AcquireMutationLease(networkDir); err == nil {
		_ = second.Close()
		t.Fatal("concurrent lease was acquired")
	} else {
		require.ErrorContains(t, err, "already in progress")
	}
	require.NoError(t, first.Close())
	reopened, err := AcquireMutationLease(networkDir)
	require.NoError(t, err, "reopen lease")
	require.NoError(t, reopened.Close())
}
