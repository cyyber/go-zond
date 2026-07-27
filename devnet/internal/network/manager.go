// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package network owns separately managed development network slots.
package network

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cenkalti/backoff/v7"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/devnet"
	"github.com/theQRL/go-qrl/devnet/internal/kurtosis"
)

type kurtosisClient interface {
	EnclaveExists(context.Context, string) (bool, error)
	CreateAndRunRemotePackage(context.Context, string, string, string) error
	Service(context.Context, string, string) (kurtosis.Service, error)
	DestroyEnclave(context.Context, string) error
}

const DefaultEnclaveName = "go-qrl-devnet"

type Environment struct {
	RPCURL       string
	GraphQLURL   string
	WebSocketURL string
}

type StartOptions struct {
	EnclaveName    string
	ExecutionImage string
	Parameters     []byte
}

type Manager struct {
	newClient func() (kurtosisClient, error)
	probe     func(context.Context, string, string) error
}

func NewManager() *Manager {
	return &Manager{
		newClient: func() (kurtosisClient, error) { return kurtosis.NewSDKClient() },
		probe:     probeNetwork,
	}
}

// Inspect validates and inspects the separately started development network.
func Inspect(ctx context.Context) (Environment, error) {
	name := os.Getenv("DEVNET_ENCLAVE_NAME")
	if name == "" {
		name = DefaultEnclaveName
	}
	return NewManager().Inspect(ctx, name)
}

func (manager *Manager) Start(ctx context.Context, options StartOptions) error {
	client, err := manager.newClient()
	if err != nil {
		return fmt.Errorf("connect to Kurtosis engine: %w", err)
	}
	if found, err := client.EnclaveExists(ctx, options.EnclaveName); err != nil {
		return fmt.Errorf("resolve development network: %w", err)
	} else if found {
		return errors.New("network already exists or provisioning is incomplete; use status or network-stop")
	}

	wallet, err := devnet.UnsafeDevelopmentWallet()
	if err != nil {
		return fmt.Errorf("restore public development wallet: %w", err)
	}
	walletAddress := common.Address(wallet.GetAddress()).Hex()
	parameters, err := effectiveParameters(walletAddress, options.ExecutionImage, options.Parameters)
	if err != nil {
		return fmt.Errorf("prepare qrl-package parameters: %w", err)
	}
	if err := client.CreateAndRunRemotePackage(
		ctx,
		options.EnclaveName,
		packageLocator,
		parameters,
	); err != nil {
		return fmt.Errorf(
			"create enclave or run pinned qrl-package; enclave may remain for network-stop: %w",
			err,
		)
	}

	if err := retryUntil(ctx, 500*time.Millisecond, 2*time.Second, func(attempt context.Context) error {
		_, err := manager.inspectEnclave(attempt, client, options.EnclaveName, walletAddress)
		return err
	}); err != nil {
		return fmt.Errorf("wait for network readiness; enclave remains for network-stop: %w", err)
	}
	return nil
}

func (manager *Manager) Inspect(ctx context.Context, name string) (Environment, error) {
	client, err := manager.newClient()
	if err != nil {
		return Environment{}, err
	}
	found, err := client.EnclaveExists(ctx, name)
	if err != nil {
		return Environment{}, fmt.Errorf("resolve development network: %w", err)
	}
	if !found {
		return Environment{}, errors.New("network is not running")
	}
	wallet, err := devnet.UnsafeDevelopmentWallet()
	if err != nil {
		return Environment{}, fmt.Errorf("restore public development wallet: %w", err)
	}
	walletAddress := common.Address(wallet.GetAddress()).Hex()
	return manager.inspectEnclave(ctx, client, name, walletAddress)
}

func (manager *Manager) inspectEnclave(
	ctx context.Context,
	client kurtosisClient,
	name,
	walletAddress string,
) (Environment, error) {
	execution, err := client.Service(ctx, name, executionServiceName)
	if err != nil {
		return Environment{}, err
	}
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
	environment := Environment{
		RPCURL:       rpcURL,
		GraphQLURL:   rpcURL + graphQLPath,
		WebSocketURL: webSocketURL,
	}
	if err = manager.probe(ctx, environment.RPCURL, walletAddress); err != nil {
		return Environment{}, err
	}
	return environment, nil
}

func (manager *Manager) Stop(ctx context.Context, name string) error {
	client, err := manager.newClient()
	if err != nil {
		return err
	}
	found, err := client.EnclaveExists(ctx, name)
	if err != nil {
		return fmt.Errorf("resolve development network: %w", err)
	}
	if !found {
		return nil
	}
	return client.DestroyEnclave(ctx, name)
}

func retryUntil(ctx context.Context, initial, maximum time.Duration, operation func(context.Context) error) error {
	policy := backoff.NewExponentialBackOff()
	policy.InitialInterval = initial
	policy.MaxInterval = maximum
	_, err := backoff.Retry(
		ctx,
		func() (struct{}, error) { return struct{}{}, operation(ctx) },
		backoff.WithBackOff(policy),
		backoff.WithMaxElapsedTime(0),
	)
	return err
}
