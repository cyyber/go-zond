// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package network owns the lifecycle and immutable identity of separately
// managed E2E networks.
package network

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

type kurtosisClient interface {
	CreateEnclave(context.Context, string) (kurtosis.EnclaveRef, error)
	GetEnclave(context.Context, string) (kurtosis.EnclaveRef, error)
	RunRemotePackage(context.Context, kurtosis.EnclaveRef, string, string) error
	Service(context.Context, kurtosis.EnclaveRef, string) (kurtosis.Service, error)
	DestroyEnclave(context.Context, kurtosis.EnclaveRef) error
}

type Manager struct {
	newClient                func() (kurtosisClient, error)
	probe                    func(context.Context, string, string) error
	createRecoveryTimeout    time.Duration
	createRecoveryPollPeriod time.Duration
}

func NewManager() *Manager {
	return &Manager{
		newClient:                func() (kurtosisClient, error) { return kurtosis.NewSDKClient() },
		probe:                    probeNetwork,
		createRecoveryTimeout:    15 * time.Second,
		createRecoveryPollPeriod: 250 * time.Millisecond,
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

	if _, err := loadOwnership(networkDir); err == nil {
		return errors.New("network already exists or provisioning is incomplete; use status or network-stop")
	} else if !errors.Is(err, errOwnershipAbsent) {
		return err
	}

	walletAddress, err := ensureWallet(networkDir)
	if err != nil {
		return fmt.Errorf("prepare private E2E wallet: %w", err)
	}
	client, err := manager.newClient()
	if err != nil {
		return fmt.Errorf("connect to Kurtosis engine: %w", err)
	}
	parameters, err := effectiveParameters(walletAddress, executionImage)
	if err != nil {
		return fmt.Errorf("prepare qrl-package parameters: %w", err)
	}
	name := newEnclaveName(networkDir)
	enclave, createErr := client.CreateEnclave(ctx, name)
	if createErr != nil {
		enclave, err = manager.recoverAmbiguousCreation(client, name, createErr)
		if err != nil {
			return err
		}
	}
	if err := retainCreatedEnclave(client, networkDir, name, enclave); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		if createErr != nil {
			return errors.Join(
				fmt.Errorf(
					"create Kurtosis enclave %q returned an error; recovered exact ownership for network-stop: %w",
					name,
					createErr,
				),
				fmt.Errorf("caller context is no longer usable: %w", err),
			)
		}
		return fmt.Errorf(
			"Kurtosis enclave was created; exact ownership was retained for network-stop: %w",
			err,
		)
	}

	if err := client.RunRemotePackage(ctx, enclave, packageLocator, parameters); err != nil {
		return fmt.Errorf("run pinned qrl-package; network ownership was retained for network-stop: %w", err)
	}

	if err := waitUntil(ctx, 2*time.Second, func(attempt context.Context) error {
		_, err := manager.inspectEnclave(attempt, client, enclave, networkDir)
		return err
	}); err != nil {
		return fmt.Errorf("wait for network readiness; network ownership was retained for network-stop: %w", err)
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
	if lookupErr := waitUntil(recoveryCtx, manager.createRecoveryPollPeriod, func(attempt context.Context) error {
		var err error
		enclave, err = client.GetEnclave(attempt, name)
		return err
	}); lookupErr != nil {
		return kurtosis.EnclaveRef{}, errors.Join(
			fmt.Errorf("create Kurtosis enclave %q: %w", name, createErr),
			fmt.Errorf("recover ambiguous creation by name: %w", lookupErr),
		)
	}
	return enclave, nil
}

func retainCreatedEnclave(
	client kurtosisClient,
	networkDir,
	expectedName string,
	enclave kurtosis.EnclaveRef,
) error {
	if enclave.Name != expectedName || enclave.Validate() != nil {
		return errors.Join(
			errors.New("Kurtosis returned an unexpected enclave identity"),
			cleanupCreatedEnclave(client, enclave),
		)
	}
	if err := createOwnership(networkDir, enclave); err != nil {
		return errors.Join(
			fmt.Errorf("persist exact enclave ownership: %w", err),
			cleanupCreatedEnclave(client, enclave),
		)
	}
	return nil
}

func (manager *Manager) Inspect(ctx context.Context, requestedDir string) (Environment, error) {
	networkDir, err := canonicalNetworkDirectory(requestedDir)
	if err != nil {
		return Environment{}, err
	}
	enclave, err := loadOwnership(networkDir)
	if err != nil {
		if errors.Is(err, errOwnershipAbsent) {
			return Environment{}, errors.New("network is not running")
		}
		return Environment{}, err
	}
	client, err := manager.newClient()
	if err != nil {
		return Environment{}, err
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
	seedFile := walletSeedPath(networkDir)
	environment, err := discoverEnvironment(execution, seedFile)
	if err != nil {
		return Environment{}, err
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

	enclave, err := loadOwnership(networkDir)
	if err != nil {
		if errors.Is(err, errOwnershipAbsent) {
			return nil
		}
		return err
	}
	client, err := manager.newClient()
	if err != nil {
		return err
	}
	if err := client.DestroyEnclave(ctx, enclave); err != nil {
		return err
	}
	return removeOwnership(networkDir)
}

func cleanupCreatedEnclave(client kurtosisClient, enclave kurtosis.EnclaveRef) error {
	if err := enclave.Validate(); err != nil {
		return errors.New("cannot clean up an invalid returned enclave identity")
	}
	cleanupCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	return client.DestroyEnclave(cleanupCtx, enclave)
}

func enclaveNamePrefix(canonicalNetworkDir string) string {
	digest := sha256.Sum256([]byte(canonicalNetworkDir))
	return fmt.Sprintf("go-qrl-e2e-%x-", digest[:6])
}

func newEnclaveName(canonicalNetworkDir string) string {
	var suffix [16]byte
	rand.Read(suffix[:])
	return fmt.Sprintf("%s%x", enclaveNamePrefix(canonicalNetworkDir), suffix)
}

func waitUntil(ctx context.Context, interval time.Duration, operation func(context.Context) error) error {
	var last error
	for {
		if err := operation(ctx); err == nil {
			return nil
		} else {
			last = err
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("last attempt: %v: %w", last, ctx.Err())
		case <-timer.C:
		}
	}
}
