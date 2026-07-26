// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package suitekit opens the shared live network for an E2E suite.
package suitekit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/theQRL/go-qrl/testing/endtoend/internal/network"
)

const networkDirVariable = "E2E_NETWORK_DIR"

// LiveNetwork owns the authenticated network lease used by one live suite.
// Suites opt into their own RPC clients, wallets, consoles, or command
// processes instead of paying for suite-specific setup here.
type LiveNetwork struct {
	network.Environment
	lease io.Closer
}

// OpenLiveNetwork authenticates the separately started network and holds its
// mutation lease. Close must be called before another suite or lifecycle
// command can use the network.
func OpenLiveNetwork(ctx context.Context) (*LiveNetwork, error) {
	return openLiveNetwork(ctx, os.Getenv(networkDirVariable), network.NewManager().Inspect)
}

func openLiveNetwork(
	ctx context.Context,
	networkDir string,
	inspect func(context.Context, string) (network.Environment, error),
) (*LiveNetwork, error) {
	lease, err := network.AcquireMutationLease(networkDir)
	if err != nil {
		return nil, err
	}
	environment, err := inspect(ctx, networkDir)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("inspect live network: %w", err), lease.Close())
	}
	return &LiveNetwork{
		Environment: environment,
		lease:       lease,
	}, nil
}

// Close releases the network mutation lease. It is nil-safe and idempotent.
func (live *LiveNetwork) Close() error {
	if live == nil {
		return nil
	}
	return live.lease.Close()
}
