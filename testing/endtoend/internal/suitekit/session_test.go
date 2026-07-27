// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package suitekit

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theQRL/go-qrl/testing/endtoend/internal/network"
)

func TestInspectLiveNetwork(t *testing.T) {
	const rpcURL = "http://127.0.0.1:18545"
	networkDir := t.TempDir()
	seedFile := filepath.Join(networkDir, "wallet.seed")
	inspect := func(
		_ context.Context,
		requestedNetwork string,
	) (network.Environment, error) {
		require.Equal(t, networkDir, requestedNetwork)
		return network.Environment{
			RPCURL:       rpcURL,
			GraphQLURL:   rpcURL + "/graphql",
			WebSocketURL: "ws://127.0.0.1/ws",
			SeedFile:     seedFile,
		}, nil
	}

	live, err := inspectLiveNetwork(t.Context(), networkDir, inspect)
	require.NoError(t, err)
	require.Equal(t, rpcURL, live.RPCURL)
	require.Equal(t, rpcURL+"/graphql", live.GraphQLURL)
	require.Equal(t, "ws://127.0.0.1/ws", live.WebSocketURL)
	require.Equal(t, seedFile, live.SeedFile)
}

func TestInspectLiveNetworkWrapsFailure(t *testing.T) {
	inspectError := errors.New("network unavailable")
	inspect := func(context.Context, string) (network.Environment, error) {
		return network.Environment{}, inspectError
	}

	_, err := inspectLiveNetwork(t.Context(), t.TempDir(), inspect)
	require.ErrorContains(t, err, "inspect live network")
	require.ErrorIs(t, err, inspectError)
}
