// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"fmt"
	"strings"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

type runtimeTopology struct {
	RPCURL, GraphQLURL, WebSocketURL string
}

func discoverTopology(services []kurtosis.Service) (runtimeTopology, error) {
	byName := make(map[string]kurtosis.Service, len(services))
	for _, service := range services {
		if _, exists := byName[service.Name]; exists {
			return runtimeTopology{}, fmt.Errorf("Kurtosis returned duplicate service name %q", service.Name)
		}
		byName[service.Name] = service
	}
	// Kurtosis service contexts do not expose lifecycle status. The network
	// probe that follows discovery proves readiness by observing chain progress.
	for role, name := range requiredServices {
		_, ok := byName[name]
		if !ok {
			return runtimeTopology{}, fmt.Errorf("required %s service %q is missing", role, name)
		}
	}
	execution := byName[executionServiceName]
	rpcURL, ok := execution.PublicEndpoint(rpcPortID, "http")
	if !ok {
		return runtimeTopology{}, fmt.Errorf("execution service %q lacks public RPC port %q", execution.Name, rpcPortID)
	}
	webSocketURL, ok := execution.PublicEndpoint(webSocketPortID, "ws")
	if !ok {
		return runtimeTopology{}, fmt.Errorf("execution service %q lacks public WebSocket port %q", execution.Name, webSocketPortID)
	}
	return runtimeTopology{
		RPCURL:       rpcURL,
		GraphQLURL:   strings.TrimRight(rpcURL, "/") + graphQLPath,
		WebSocketURL: webSocketURL,
	}, nil
}
