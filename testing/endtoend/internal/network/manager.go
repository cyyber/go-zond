// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package network owns separately managed E2E network slots.
package network

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v7"
	"github.com/theQRL/go-qrl/testing/endtoend/internal/kurtosis"
)

type kurtosisClient interface {
	CreateEnclave(context.Context, string) (kurtosis.EnclaveRef, error)
	LookupEnclave(context.Context, string) (kurtosis.EnclaveRef, bool, error)
	RunRemotePackage(context.Context, kurtosis.EnclaveRef, string, string) error
	Service(context.Context, kurtosis.EnclaveRef, string) (kurtosis.Service, error)
	DestroyEnclave(context.Context, kurtosis.EnclaveRef) error
}

type Environment struct {
	RPCURL       string
	GraphQLURL   string
	WebSocketURL string
	SeedFile     string
}

type Manager struct {
	newClient             func() (kurtosisClient, error)
	probe                 func(context.Context, string, string) error
	createRecoveryTimeout time.Duration
	createRecoveryInitial time.Duration
}

func NewManager() *Manager {
	return &Manager{
		newClient:             func() (kurtosisClient, error) { return kurtosis.NewSDKClient() },
		probe:                 probeNetwork,
		createRecoveryTimeout: 15 * time.Second,
		createRecoveryInitial: 250 * time.Millisecond,
	}
}

func (manager *Manager) Start(ctx context.Context, requestedDir, executionImage string) error {
	networkDir, err := ensureNetworkDirectory(requestedDir)
	if err != nil {
		return err
	}
	mutation, err := AcquireMutationLease(networkDir)
	if err != nil {
		return err
	}
	defer mutation.Close()

	name := enclaveName(networkDir)
	client, err := manager.newClient()
	if err != nil {
		return fmt.Errorf("connect to Kurtosis engine: %w", err)
	}
	if _, found, err := client.LookupEnclave(ctx, name); err != nil {
		return fmt.Errorf("resolve deterministic network slot: %w", err)
	} else if found {
		return errors.New("network already exists or provisioning is incomplete; use status or network-stop")
	}

	walletAddress, err := ensureWallet(networkDir)
	if err != nil {
		return fmt.Errorf("prepare private E2E wallet: %w", err)
	}
	parameters, err := effectiveParameters(walletAddress, executionImage)
	if err != nil {
		return fmt.Errorf("prepare qrl-package parameters: %w", err)
	}
	enclave, createErr := client.CreateEnclave(ctx, name)
	if createErr != nil {
		enclave, err = manager.recoverAmbiguousCreation(client, name, createErr)
		if err != nil {
			return err
		}
	}
	if enclave.Name != name || enclave.UUID == "" {
		return errors.Join(
			errors.New("Kurtosis returned an unexpected enclave identity"),
			cleanupCreatedEnclave(client, enclave),
		)
	}
	if err := ctx.Err(); err != nil {
		if createErr != nil {
			return errors.Join(
				fmt.Errorf(
					"create Kurtosis enclave %q returned an error; recovered its deterministic slot for network-stop: %w",
					name,
					createErr,
				),
				fmt.Errorf("caller context is no longer usable: %w", err),
			)
		}
		return fmt.Errorf(
			"Kurtosis enclave was created; its deterministic slot remains for network-stop: %w",
			err,
		)
	}

	if err := client.RunRemotePackage(ctx, enclave, packageLocator, parameters); err != nil {
		return fmt.Errorf("run pinned qrl-package; network slot remains for network-stop: %w", err)
	}

	if err := retryUntil(ctx, 500*time.Millisecond, 2*time.Second, func(attempt context.Context) error {
		_, err := manager.inspectEnclave(attempt, client, enclave, networkDir)
		return err
	}); err != nil {
		return fmt.Errorf("wait for network readiness; network slot remains for network-stop: %w", err)
	}
	return nil
}

func (manager *Manager) recoverAmbiguousCreation(
	client kurtosisClient,
	name string,
	createErr error,
) (kurtosis.EnclaveRef, error) {
	recoveryCtx, cancel := context.WithTimeout(context.Background(), manager.createRecoveryTimeout)
	defer cancel()
	var enclave kurtosis.EnclaveRef
	if lookupErr := retryUntil(recoveryCtx, manager.createRecoveryInitial, time.Second, func(attempt context.Context) error {
		var (
			found bool
			err   error
		)
		enclave, found, err = client.LookupEnclave(attempt, name)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("Kurtosis enclave %q is not visible yet", name)
		}
		return nil
	}); lookupErr != nil {
		return kurtosis.EnclaveRef{}, errors.Join(
			fmt.Errorf("create Kurtosis enclave %q: %w", name, createErr),
			fmt.Errorf("recover ambiguous creation by name: %w", lookupErr),
		)
	}
	return enclave, nil
}

func (manager *Manager) Inspect(ctx context.Context, requestedDir string) (Environment, error) {
	networkDir, err := canonicalNetworkDirectory(requestedDir)
	if err != nil {
		return Environment{}, err
	}
	client, err := manager.newClient()
	if err != nil {
		return Environment{}, err
	}
	enclave, found, err := client.LookupEnclave(ctx, enclaveName(networkDir))
	if err != nil {
		return Environment{}, fmt.Errorf("resolve deterministic network slot: %w", err)
	}
	if !found {
		return Environment{}, errors.New("network is not running")
	}
	return manager.inspectEnclave(ctx, client, enclave, networkDir)
}

func (manager *Manager) inspectEnclave(
	ctx context.Context,
	client kurtosisClient,
	enclave kurtosis.EnclaveRef,
	networkDir string,
) (Environment, error) {
	execution, err := client.Service(ctx, enclave, executionServiceName)
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
	seedFile := walletSeedPath(networkDir)
	environment := Environment{
		RPCURL:       rpcURL,
		GraphQLURL:   rpcURL + graphQLPath,
		WebSocketURL: webSocketURL,
		SeedFile:     seedFile,
	}
	walletAddress, err := validateWalletSeed(seedFile)
	if err != nil {
		return Environment{}, fmt.Errorf("validate private wallet: %w", err)
	}
	if err = manager.probe(ctx, environment.RPCURL, walletAddress); err != nil {
		return Environment{}, err
	}
	return environment, nil
}

func (manager *Manager) Stop(ctx context.Context, requestedDir string) error {
	networkDir, err := canonicalNetworkDirectory(requestedDir)
	if err != nil {
		return err
	}
	mutation, err := AcquireMutationLease(networkDir)
	if err != nil {
		return err
	}
	defer mutation.Close()

	client, err := manager.newClient()
	if err != nil {
		return err
	}
	enclave, found, err := client.LookupEnclave(ctx, enclaveName(networkDir))
	if err != nil {
		return fmt.Errorf("resolve deterministic network slot: %w", err)
	}
	if !found {
		return nil
	}
	return client.DestroyEnclave(ctx, enclave)
}

func cleanupCreatedEnclave(client kurtosisClient, enclave kurtosis.EnclaveRef) error {
	if enclave.UUID == "" {
		return errors.New("cannot clean up an invalid returned enclave identity")
	}
	cleanupCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	return client.DestroyEnclave(cleanupCtx, enclave)
}

func enclaveName(canonicalNetworkDir string) string {
	digest := sha256.Sum256([]byte(canonicalNetworkDir))
	return fmt.Sprintf("go-qrl-e2e-%x", digest[:24])
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
