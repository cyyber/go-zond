// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package suitekit

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/network"
)

func TestOpenLiveNetwork(t *testing.T) {
	const rpcURL = "http://127.0.0.1:18545"
	networkDir, seedFile := liveNetworkDirectory(t)
	inspect := func(
		_ context.Context,
		requestedNetwork string,
	) (network.Environment, error) {
		require.Equal(t, networkDir, requestedNetwork)
		if competing, err := network.AcquireMutationLease(networkDir); err == nil {
			_ = competing.Close()
			t.Fatal("network lease was not held during inspection")
		}
		return network.Environment{
			RPCURL:       rpcURL,
			GraphQLURL:   rpcURL + "/graphql",
			WebSocketURL: "ws://127.0.0.1/ws",
			SeedFile:     seedFile,
		}, nil
	}

	live, err := openLiveNetwork(t.Context(), networkDir, inspect)
	require.NoError(t, err)
	require.Equal(t, rpcURL, live.RPCURL)
	require.Equal(t, rpcURL+"/graphql", live.GraphQLURL)
	require.Equal(t, "ws://127.0.0.1/ws", live.WebSocketURL)
	require.Equal(t, seedFile, live.SeedFile)
	require.NoError(t, live.Close())
	require.NoError(t, live.Close())
	reopened, err := network.AcquireMutationLease(networkDir)
	require.NoError(t, err, "session leaked network lease")
	_ = reopened.Close()
}

func TestOpenLiveNetworkReleasesLeaseOnFailure(t *testing.T) {
	networkDir, _ := liveNetworkDirectory(t)
	inspect := func(context.Context, string) (network.Environment, error) {
		return network.Environment{}, errors.New("network unavailable")
	}

	_, err := openLiveNetwork(t.Context(), networkDir, inspect)
	require.ErrorContains(t, err, "inspect live network")
	reopened, err := network.AcquireMutationLease(networkDir)
	require.NoError(t, err, "failure leaked network lease")
	_ = reopened.Close()
}

func liveNetworkDirectory(t *testing.T) (string, string) {
	t.Helper()
	networkDir := t.TempDir()
	require.NoError(t, os.Chmod(networkDir, 0o700))
	return networkDir, filepath.Join(networkDir, "wallet.seed")
}
