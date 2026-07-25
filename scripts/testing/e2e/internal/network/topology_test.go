// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"testing"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

func topologyFixture() map[string]kurtosis.Service {
	return map[string]kurtosis.Service{
		executionServiceName: {
			PublicIP:    "127.0.0.1",
			PublicPorts: map[string]uint16{rpcPortID: 18545, webSocketPortID: 18546},
		},
		consensusServiceName: {},
		validatorServiceName: {},
	}
}

func TestDiscoverEnvironment(t *testing.T) {
	got, err := discoverEnvironment(topologyFixture(), "/private/wallet.seed")
	if err != nil {
		t.Fatal(err)
	}
	if got.RPCURL != "http://127.0.0.1:18545" ||
		got.WebSocketURL != "ws://127.0.0.1:18546" ||
		got.GraphQLURL != "http://127.0.0.1:18545/graphql" ||
		got.SeedFile != "/private/wallet.seed" {
		t.Fatalf("environment = %+v", got)
	}
}

func TestDiscoverEnvironmentRequiresServicesAndEndpoints(t *testing.T) {
	for name, mutate := range map[string]func(map[string]kurtosis.Service){
		"missing": func(services map[string]kurtosis.Service) {
			delete(services, consensusServiceName)
		},
		"port": func(services map[string]kurtosis.Service) {
			execution := services[executionServiceName]
			delete(execution.PublicPorts, rpcPortID)
			services[executionServiceName] = execution
		},
	} {
		t.Run(name, func(t *testing.T) {
			services := topologyFixture()
			mutate(services)
			if _, err := discoverEnvironment(services, "/private/wallet.seed"); err == nil {
				t.Fatal("invalid topology was accepted")
			}
		})
	}
}
