// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"fmt"
	"strings"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

func discoverEnvironment(
	services map[string]kurtosis.Service,
	seedFile string,
) (Environment, error) {
	// Kurtosis service contexts do not expose lifecycle status. The network
	// probe that follows discovery proves readiness by observing chain progress.
	for role, name := range requiredServices {
		_, ok := services[name]
		if !ok {
			return Environment{}, fmt.Errorf("required %s service %q is missing", role, name)
		}
	}
	execution := services[executionServiceName]
	rpcURL, ok := execution.PublicEndpoint(rpcPortID, "http")
	if !ok {
		return Environment{}, fmt.Errorf(
			"execution service %q lacks public RPC port %q",
			executionServiceName,
			rpcPortID,
		)
	}
	webSocketURL, ok := execution.PublicEndpoint(webSocketPortID, "ws")
	if !ok {
		return Environment{}, fmt.Errorf(
			"execution service %q lacks public WebSocket port %q",
			executionServiceName,
			webSocketPortID,
		)
	}
	return Environment{
		RPCURL:       rpcURL,
		GraphQLURL:   strings.TrimRight(rpcURL, "/") + graphQLPath,
		WebSocketURL: webSocketURL,
		SeedFile:     seedFile,
	}, nil
}
