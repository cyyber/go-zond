// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

type kurtosisClient interface {
	CreateEnclave(context.Context, string) (kurtosis.EnclaveRef, error)
	GetEnclave(context.Context, string) (kurtosis.EnclaveRef, error)
	EnclaveExists(context.Context, string) (bool, error)
	RunRemotePackage(context.Context, kurtosis.EnclaveRef, kurtosis.PackageRun) error
	Services(context.Context, kurtosis.EnclaveRef) (map[string]kurtosis.Service, error)
	DestroyEnclave(context.Context, kurtosis.EnclaveRef) error
}

type Manager struct {
	newClient                func() (kurtosisClient, error)
	probe                    func(context.Context, probeRequest) error
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

func (manager *Manager) Start(ctx context.Context, request StartRequest) error {
	networkDir, err := ensureNetworkDirectory(request.NetworkDir)
	if err != nil {
		return err
	}
	mutation, err := acquireMutationLease(networkDir)
	if err != nil {
		return err
	}
	defer mutation.Close()

	ownershipExists, err := pathExists(ownershipPath(networkDir))
	if err != nil {
		return err
	}
	if ownershipExists {
		_, loadErr := loadOwnership(networkDir)
		if loadErr != nil {
			return loadErr
		}
		return errors.New("network already exists or provisioning is incomplete; use status or network-stop")
	}

	walletAddress, err := ensureWallet(networkDir)
	if err != nil {
		return fmt.Errorf("prepare private E2E wallet: %w", err)
	}
	client, err := manager.newClient()
	if err != nil {
		return fmt.Errorf("connect to Kurtosis engine: %w", err)
	}
	parameters, err := effectiveParameters(walletAddress, request.ExecutionImage)
	if err != nil {
		return fmt.Errorf("prepare qrl-package parameters: %w", err)
	}
	name, err := newEnclaveName(networkDir)
	if err != nil {
		return fmt.Errorf("create unique Kurtosis enclave name: %w", err)
	}
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

	if err := client.RunRemotePackage(ctx, enclave, kurtosis.PackageRun{
		Locator:          packageLocator,
		SerializedParams: parameters,
	}); err != nil {
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
		cleanupErr := cleanupCreatedEnclave(client, enclave)
		return errors.Join(
			errors.New("Kurtosis returned an unexpected enclave identity"),
			wrapCleanupError(cleanupErr),
		)
	}
	if err := createOwnership(networkDir, enclave); err != nil {
		cleanupErr := cleanupCreatedEnclave(client, enclave)
		return errors.Join(
			fmt.Errorf("persist exact enclave ownership: %w", err),
			wrapCleanupError(cleanupErr),
		)
	}
	return nil
}

func (manager *Manager) Status(ctx context.Context, requestedDir string) error {
	_, err := manager.inspect(ctx, requestedDir)
	return err
}

func (manager *Manager) inspect(ctx context.Context, requestedDir string) (Environment, error) {
	networkDir, err := canonicalExistingDirectory(requestedDir, "network directory")
	if err != nil {
		return Environment{}, err
	}
	ownershipExists, err := pathExists(ownershipPath(networkDir))
	if err != nil {
		return Environment{}, err
	}
	if !ownershipExists {
		return Environment{}, errors.New("network is not running")
	}
	enclave, err := loadOwnership(networkDir)
	if err != nil {
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
	services, err := client.Services(ctx, enclave)
	if err != nil {
		return Environment{}, err
	}
	seedFile := walletSeedPath(networkDir)
	environment, err := discoverEnvironment(services, seedFile)
	if err != nil {
		return Environment{}, err
	}
	walletAddress, err := validateWalletSeed(seedFile)
	if err != nil {
		return Environment{}, fmt.Errorf("validate private wallet: %w", err)
	}
	if err = manager.probe(ctx, probeRequest{
		RPCURL: environment.RPCURL, Address: walletAddress,
	}); err != nil {
		return Environment{}, err
	}
	return environment, nil
}

func (manager *Manager) Authenticate(ctx context.Context, networkDir string) (Environment, error) {
	return manager.inspect(ctx, networkDir)
}

func (manager *Manager) Stop(ctx context.Context, requestedDir string) error {
	networkDir, err := canonicalExistingDirectory(requestedDir, "network directory")
	if err != nil {
		return err
	}
	mutation, err := acquireMutationLease(networkDir)
	if err != nil {
		return err
	}
	defer mutation.Close()

	ownershipExists, err := pathExists(ownershipPath(networkDir))
	if err != nil {
		return err
	}
	if !ownershipExists {
		return nil
	}
	enclave, err := loadOwnership(networkDir)
	if err != nil {
		return err
	}
	client, err := manager.newClient()
	if err != nil {
		return err
	}
	if err := destroyExactEnclave(ctx, client, enclave); err != nil {
		return err
	}
	if err := removeOwnership(networkDir); err != nil {
		return err
	}
	return nil
}

func destroyExactEnclave(ctx context.Context, client kurtosisClient, enclave kurtosis.EnclaveRef) error {
	destroyErr := client.DestroyEnclave(ctx, enclave)
	exists, inspectErr := client.EnclaveExists(ctx, enclave.UUID)
	if inspectErr != nil {
		return errors.Join(destroyErr, fmt.Errorf("confirm owned enclave destruction: %w", inspectErr))
	}
	if exists {
		return errors.Join(destroyErr, errors.New("owned enclave still exists after destruction"))
	}
	return nil
}

func cleanupCreatedEnclave(client kurtosisClient, enclave kurtosis.EnclaveRef) error {
	if err := enclave.Validate(); err != nil {
		return errors.New("cannot clean up an invalid returned enclave identity")
	}
	cleanupCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	return destroyExactEnclave(cleanupCtx, client, enclave)
}

func wrapCleanupError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("clean up returned enclave identity: %w", err)
}

func pathExists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func enclaveNamePrefix(canonicalNetworkDir string) string {
	return "go-qrl-e2e-" + networkID(canonicalNetworkDir) + "-"
}

func newEnclaveName(canonicalNetworkDir string) (string, error) {
	var suffix [16]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%x", enclaveNamePrefix(canonicalNetworkDir), suffix), nil
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
