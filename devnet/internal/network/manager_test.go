// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theQRL/go-qrl/devnet/internal/kurtosis"
)

const testExecutionImage = "local/go-qrl:test"

func TestNetworkLifecycle(t *testing.T) {
	networkDir, err := ensureNetworkDirectory(filepath.Join(t.TempDir(), "network"))
	require.NoError(t, err)
	client := &fakeKurtosis{
		Enclave: kurtosis.EnclaveRef{UUID: strings.Repeat("a", 32)},
		ExecutionService: kurtosis.Service{
			PublicIP:    "127.0.0.1",
			PublicPorts: map[string]uint16{rpcPortID: 18545, webSocketPortID: 18546},
		},
	}
	manager := NewManager()
	manager.newClient = func() (kurtosisClient, error) { return client, nil }
	manager.probe = func(context.Context, string, string) error { return nil }

	require.NoError(t, manager.Start(t.Context(), StartOptions{
		Directory:      networkDir,
		ExecutionImage: testExecutionImage,
	}))
	require.Equal(t, enclaveName(networkDir), client.Enclave.Name)
	require.Equal(t, packageLocator, client.RunLocator)
	require.Contains(t, client.RunParameters, `"el_image":"local/go-qrl:test"`)
	require.Equal(t, client.Enclave, client.RunRef)
	require.Regexp(t, `^go-qrl-e2e-[0-9a-f]{48}$`, client.Enclave.Name)
	require.Len(t, client.Enclave.Name, 59)

	environment, err := manager.Inspect(t.Context(), networkDir)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:18545", environment.RPCURL)
	require.Equal(t, "http://127.0.0.1:18545/graphql", environment.GraphQLURL)
	require.Equal(t, "ws://127.0.0.1:18546", environment.WebSocketURL)
	require.Equal(t, filepath.Join(networkDir, "wallet.seed"), environment.SeedFile)
	require.Equal(t, client.Enclave, client.ServiceRef)

	require.NoError(t, manager.Stop(t.Context(), networkDir))
	require.False(t, client.Exists)
	require.Equal(t, client.Enclave, client.DestroyRef)
	_, err = manager.Inspect(t.Context(), networkDir)
	require.ErrorContains(t, err, "not running")
}
