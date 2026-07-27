// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theQRL/go-qrl/devnet/internal/kurtosis"
)

const testExecutionImage = "local/go-qrl:test"

func TestNetworkLifecycle(t *testing.T) {
	const enclaveName = "go-qrl-devnet-test"
	client := &fakeKurtosis{
		ExecutionService: kurtosis.Service{
			PublicIP:    "127.0.0.1",
			PublicPorts: map[string]uint16{rpcPortID: 18545, webSocketPortID: 18546},
		},
	}
	manager := NewManager()
	manager.newClient = func() (kurtosisClient, error) { return client, nil }
	manager.probe = func(context.Context, string, string) error { return nil }

	require.NoError(t, manager.Start(t.Context(), StartOptions{
		EnclaveName:    enclaveName,
		ExecutionImage: testExecutionImage,
	}))
	require.Equal(t, enclaveName, client.Name)
	require.Equal(t, packageLocator, client.RunLocator)
	require.Contains(t, client.RunParameters, `"el_image":"local/go-qrl:test"`)

	environment, err := manager.Inspect(t.Context(), enclaveName)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:18545", environment.RPCURL)
	require.Equal(t, "http://127.0.0.1:18545/graphql", environment.GraphQLURL)
	require.Equal(t, "ws://127.0.0.1:18546", environment.WebSocketURL)
	require.Equal(t, enclaveName, client.ServiceName)

	require.NoError(t, manager.Stop(t.Context(), enclaveName))
	require.Equal(t, enclaveName, client.DestroyName)
	_, err = manager.Inspect(t.Context(), enclaveName)
	require.ErrorContains(t, err, "not running")
}

type fakeKurtosis struct {
	Name                      string
	ExecutionService          kurtosis.Service
	Exists                    bool
	ServiceName, DestroyName  string
	RunLocator, RunParameters string
}

func (fake *fakeKurtosis) CreateAndRunRemotePackage(_ context.Context, name, locator, parameters string) error {
	fake.Name = name
	fake.Exists = true
	fake.RunLocator, fake.RunParameters = locator, parameters
	return nil
}

func (fake *fakeKurtosis) EnclaveExists(_ context.Context, name string) (bool, error) {
	return fake.Exists && name == fake.Name, nil
}

func (fake *fakeKurtosis) Service(_ context.Context, enclaveName, serviceName string) (kurtosis.Service, error) {
	fake.ServiceName = enclaveName
	if serviceName != executionServiceName {
		return kurtosis.Service{}, fmt.Errorf("service %q not found", serviceName)
	}
	return fake.ExecutionService, nil
}

func (fake *fakeKurtosis) DestroyEnclave(_ context.Context, name string) error {
	fake.DestroyName = name
	fake.Exists = false
	return nil
}
