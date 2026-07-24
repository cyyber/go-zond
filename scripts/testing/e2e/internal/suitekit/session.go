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

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
	"github.com/theQRL/go-qrl/qrlclient"
	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/network"
)

const networkDirVariable = "E2E_NETWORK_DIR"

// LiveSession owns the authenticated network lease, RPC connection, and
// restored wallet used by one live suite.
type LiveSession struct {
	RPCURL       string
	GraphQLURL   string
	WebSocketURL string
	Client       *qrlclient.Client
	Wallet       wallet.Wallet
	Sender       common.Address

	lease    *network.MutationLease
	closeErr error
	close    sync.Once
}

// OpenLiveSession authenticates the separately started network and opens its
// funded signing session. Close must be called before another suite or network
// lifecycle command can use the network.
func OpenLiveSession(ctx context.Context) (*LiveSession, error) {
	return openLiveSession(ctx, os.Getenv, network.NewManager().Authenticate)
}

func openLiveSession(
	ctx context.Context,
	getenv func(string) string,
	authenticate func(context.Context, string) (network.Environment, error),
) (*LiveSession, error) {
	if ctx == nil {
		return nil, errors.New("open live session: context is nil")
	}
	lease, err := network.AcquireMutationLease(getenv(networkDirVariable))
	if err != nil {
		return nil, err
	}
	fail := func(cause error) (*LiveSession, error) {
		return nil, errors.Join(cause, lease.Close())
	}

	environment, err := authenticate(ctx, lease.NetworkDir())
	if err != nil {
		return fail(fmt.Errorf("authenticate live network: %w", err))
	}
	restored, err := wallet.RestoreFromFile(environment.SeedFile)
	if err != nil {
		return fail(fmt.Errorf("restore live wallet: %w", err))
	}
	client, err := qrlclient.DialContext(ctx, environment.RPCURL)
	if err != nil {
		return fail(fmt.Errorf("dial live RPC: %w", err))
	}
	return &LiveSession{
		RPCURL:       environment.RPCURL,
		GraphQLURL:   environment.GraphQLURL,
		WebSocketURL: environment.WebSocketURL,
		Client:       client,
		Wallet:       restored,
		Sender:       common.Address(restored.GetAddress()),
		lease:        lease,
	}, nil
}

// Close closes the RPC connection and releases the network mutation lease.
// It is nil-safe and idempotent.
func (session *LiveSession) Close() error {
	if session == nil {
		return nil
	}
	session.close.Do(func() {
		if session.Client != nil {
			session.Client.Close()
		}
		session.closeErr = session.lease.Close()
	})
	return session.closeErr
}
