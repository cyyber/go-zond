// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"testing"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

func topologyFixture() kurtosis.Service {
	return kurtosis.Service{
		PublicIP:    "127.0.0.1",
		PublicPorts: map[string]uint16{rpcPortID: 18545, webSocketPortID: 18546},
	}
}

func TestDiscoverEnvironment(t *testing.T) {
	got, err := discoverEnvironment(topologyFixture(), "/wallet.seed")
	if err != nil {
		t.Fatal(err)
	}
	if got.RPCURL != "http://127.0.0.1:18545" ||
		got.WebSocketURL != "ws://127.0.0.1:18546" ||
		got.GraphQLURL != "http://127.0.0.1:18545/graphql" ||
		got.SeedFile != "/wallet.seed" {
		t.Fatalf("environment = %+v", got)
	}
}

func TestDiscoverEnvironmentRequiresEndpoints(t *testing.T) {
	for _, portID := range []string{rpcPortID, webSocketPortID} {
		t.Run(portID, func(t *testing.T) {
			execution := topologyFixture()
			delete(execution.PublicPorts, portID)
			if _, err := discoverEnvironment(execution, "/wallet.seed"); err == nil {
				t.Fatal("missing endpoint was accepted")
			}
		})
	}
}
