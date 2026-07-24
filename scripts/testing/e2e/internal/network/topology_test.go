// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"testing"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

func topologyFixture() []kurtosis.Service {
	return []kurtosis.Service{
		{Name: executionServiceName, PublicIP: "127.0.0.1", PublicPorts: map[string]kurtosis.Port{rpcPortID: {Number: 18545}, webSocketPortID: {Number: 18546}}},
		{Name: consensusServiceName},
		{Name: validatorServiceName},
	}
}

func TestDiscoverTopology(t *testing.T) {
	got, err := discoverTopology(topologyFixture())
	if err != nil {
		t.Fatal(err)
	}
	if got.RPCURL != "http://127.0.0.1:18545" ||
		got.WebSocketURL != "ws://127.0.0.1:18546" ||
		got.GraphQLURL != "http://127.0.0.1:18545/graphql" {
		t.Fatalf("topology = %+v", got)
	}
}

func TestDiscoverTopologyRequiresServicesAndEndpoints(t *testing.T) {
	for name, mutate := range map[string]func([]kurtosis.Service){
		"missing": func(services []kurtosis.Service) { services[1].Name = "helper" },
		"port":    func(services []kurtosis.Service) { delete(services[0].PublicPorts, rpcPortID) },
	} {
		t.Run(name, func(t *testing.T) {
			services := topologyFixture()
			mutate(services)
			if _, err := discoverTopology(services); err == nil {
				t.Fatal("invalid topology was accepted")
			}
		})
	}
}
