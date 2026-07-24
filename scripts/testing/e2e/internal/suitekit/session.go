// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package suitekit opens the shared live network for an E2E suite.
package suitekit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/network"
)

const networkDirVariable = "E2E_NETWORK_DIR"

// LiveNetwork owns the authenticated network lease used by one live suite.
// Suites opt into their own RPC clients, wallets, consoles, or command
// processes instead of paying for suite-specific setup here.
type LiveNetwork struct {
	RPCURL       string
	GraphQLURL   string
	WebSocketURL string
	SeedFile     string

	lease    *network.MutationLease
	closeErr error
	close    sync.Once
}

// OpenLiveNetwork authenticates the separately started network and holds its
// mutation lease. Close must be called before another suite or lifecycle
// command can use the network.
func OpenLiveNetwork(ctx context.Context) (*LiveNetwork, error) {
	return openLiveNetwork(ctx, os.Getenv, network.NewManager().Authenticate)
}

func openLiveNetwork(
	ctx context.Context,
	getenv func(string) string,
	authenticate func(context.Context, string) (network.Environment, error),
) (*LiveNetwork, error) {
	if ctx == nil {
		return nil, errors.New("open live network: context is nil")
	}
	lease, err := network.AcquireMutationLease(getenv(networkDirVariable))
	if err != nil {
		return nil, err
	}
	fail := func(cause error) (*LiveNetwork, error) {
		return nil, errors.Join(cause, lease.Close())
	}

	environment, err := authenticate(ctx, lease.NetworkDir())
	if err != nil {
		return fail(fmt.Errorf("authenticate live network: %w", err))
	}
	return &LiveNetwork{
		RPCURL:       environment.RPCURL,
		GraphQLURL:   environment.GraphQLURL,
		WebSocketURL: environment.WebSocketURL,
		SeedFile:     environment.SeedFile,
		lease:        lease,
	}, nil
}

// Close releases the network mutation lease. It is nil-safe and idempotent.
func (live *LiveNetwork) Close() error {
	if live == nil {
		return nil
	}
	live.close.Do(func() {
		live.closeErr = live.lease.Close()
	})
	return live.closeErr
}
