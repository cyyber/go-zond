// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package suitekit inspects the shared live network for an E2E suite.
package suitekit

import (
	"context"
	"fmt"
	"os"

	"github.com/theQRL/go-qrl/testing/endtoend/internal/network"
)

const networkDirVariable = "E2E_NETWORK_DIR"

// InspectLiveNetwork validates and inspects the separately started network.
func InspectLiveNetwork(ctx context.Context) (network.Environment, error) {
	return inspectLiveNetwork(ctx, os.Getenv(networkDirVariable), network.NewManager().Inspect)
}

func inspectLiveNetwork(
	ctx context.Context,
	networkDir string,
	inspect func(context.Context, string) (network.Environment, error),
) (network.Environment, error) {
	environment, err := inspect(ctx, networkDir)
	if err != nil {
		return network.Environment{}, fmt.Errorf("inspect live network: %w", err)
	}
	return environment, nil
}
