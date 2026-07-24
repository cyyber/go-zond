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
	Services(context.Context, kurtosis.EnclaveRef) ([]kurtosis.Service, error)
	DestroyEnclave(context.Context, kurtosis.EnclaveRef) error
}

type Manager struct {
	newClient func() (kurtosisClient, error)
	probe     func(context.Context, probeRequest) error
}

func NewManager() *Manager {
	return &Manager{
		newClient: func() (kurtosisClient, error) { return kurtosis.NewSDKClient() },
		probe:     probeNetwork,
	}
}

func (manager *Manager) Start(ctx context.Context, request StartRequest) error {
	networkDir, err := ensureNetworkDirectory(request.NetworkDir)
	if err != nil {
		return err
	}
	request.NetworkDir = networkDir
	mutation, err := acquireMutationLease(networkDir)
	if err != nil {
		return err
	}
	defer mutation.Close()
	if request.StartTimeout <= 0 {
		request.StartTimeout = 150 * time.Minute
	}
	startCtx, cancel := context.WithTimeout(ctx, request.StartTimeout)
	defer cancel()

	ownershipExists, err := pathExists(ownershipPath(networkDir))
	if err != nil {
		return err
	}
	if ownershipExists {
		ownership, loadErr := loadOwnership(networkDir)
		if loadErr != nil {
			return loadErr
		}
		if _, ownershipErr := ownership.Enclave(); ownershipErr != nil {
			return ownershipErr
		}
		return errors.New("network already exists or provisioning is incomplete; use status or network-stop")
	}

	walletAddress, err := ensureWallet(privatePath(networkDir))
	if err != nil {
		return fmt.Errorf("prepare private E2E wallet: %w", err)
	}
	client, err := manager.newClient()
	if err != nil {
		return fmt.Errorf("connect to Kurtosis engine: %w", err)
	}
	if request.ExecutionImage == "" {
		request.ExecutionImage = DefaultExecutionImage
	}
	parameters, err := effectiveParameters(walletAddress, request.ExecutionImage)
	if err != nil {
		return fmt.Errorf("prepare qrl-package parameters: %w", err)
	}
	name, err := newEnclaveName(networkDir)
	if err != nil {
		return fmt.Errorf("create unique Kurtosis enclave name: %w", err)
	}
	enclave, err := client.CreateEnclave(startCtx, name)
	if err != nil {
		return recoverAmbiguousCreation(startCtx, client, networkDir, name, err)
	}
	if enclave.Name != name || enclave.Validate() != nil {
		cleanupErr := cleanupCreatedEnclave(client, enclave)
		return errors.Join(
			errors.New("Kurtosis returned an unexpected enclave identity"),
			wrapCleanupError(cleanupErr),
		)
	}
	ownership := OwnershipRecord{Name: enclave.Name, UUID: enclave.UUID}
	if err := createOwnership(networkDir, ownership); err != nil {
		cleanupErr := cleanupCreatedEnclave(client, enclave)
		return errors.Join(
			fmt.Errorf("persist exact enclave ownership: %w", err),
			wrapCleanupError(cleanupErr),
		)
	}

	if err := client.RunRemotePackage(startCtx, enclave, kurtosis.PackageRun{
		Locator:          packageLocator,
		SerializedParams: parameters,
	}); err != nil {
		return fmt.Errorf("run pinned qrl-package; network ownership was retained for network-stop: %w", err)
	}

	var runtime runtimeTopology
	if err := waitUntil(startCtx, 2*time.Second, func(attempt context.Context) error {
		services, err := client.Services(attempt, enclave)
		if err != nil {
			return err
		}
		discovered, err := discoverTopology(services)
		if err != nil {
			return err
		}
		runtime = discovered
		return nil
	}); err != nil {
		return fmt.Errorf("discover qrl-package topology; network ownership was retained for network-stop: %w", err)
	}
	if err := waitUntil(startCtx, 2*time.Second, func(attempt context.Context) error {
		return manager.probe(attempt, probeRequest{
			RPCURL:  runtime.RPCURL,
			Address: walletAddress, ExpectedChainID: expectedChainID,
		})
	}); err != nil {
		return fmt.Errorf("wait for network readiness; network ownership was retained for network-stop: %w", err)
	}
	return nil
}

func recoverAmbiguousCreation(
	ctx context.Context,
	client kurtosisClient,
	networkDir,
	name string,
	createErr error,
) error {
	enclave, lookupErr := client.GetEnclave(ctx, name)
	if lookupErr != nil {
		return errors.Join(
			fmt.Errorf("create Kurtosis enclave %q: %w", name, createErr),
			fmt.Errorf("recover ambiguous creation by name: %w", lookupErr),
		)
	}
	if enclave.Name != name || enclave.Validate() != nil {
		return errors.Join(
			fmt.Errorf("create Kurtosis enclave %q: %w", name, createErr),
			errors.New("Kurtosis returned an unexpected recovered enclave identity"),
		)
	}
	if err := createOwnership(networkDir, OwnershipRecord{
		Name: enclave.Name,
		UUID: enclave.UUID,
	}); err != nil {
		return errors.Join(
			fmt.Errorf("create Kurtosis enclave %q: %w", name, createErr),
			fmt.Errorf("persist recovered exact enclave ownership: %w", err),
		)
	}
	return fmt.Errorf(
		"create Kurtosis enclave %q returned an error; recovered exact ownership for network-stop: %w",
		name,
		createErr,
	)
}

func (manager *Manager) Status(ctx context.Context, requestedDir string) error {
	_, err := manager.status(ctx, requestedDir)
	return err
}

func (manager *Manager) status(ctx context.Context, requestedDir string) (runtimeTopology, error) {
	networkDir, err := canonicalExistingDirectory(requestedDir, "network directory")
	if err != nil {
		return runtimeTopology{}, err
	}
	ownershipExists, err := pathExists(ownershipPath(networkDir))
	if err != nil {
		return runtimeTopology{}, err
	}
	if !ownershipExists {
		return runtimeTopology{}, errors.New("network is not running")
	}
	ownership, err := loadOwnership(networkDir)
	if err != nil {
		return runtimeTopology{}, err
	}
	enclave, err := ownership.Enclave()
	if err != nil {
		return runtimeTopology{}, err
	}
	client, err := manager.newClient()
	if err != nil {
		return runtimeTopology{}, err
	}
	services, err := client.Services(ctx, enclave)
	if err != nil {
		return runtimeTopology{}, err
	}
	runtime, err := discoverTopology(services)
	if err != nil {
		return runtimeTopology{}, err
	}
	walletAddress, err := validateWalletSeed(walletSeedPath(networkDir))
	if err != nil {
		return runtimeTopology{}, fmt.Errorf("validate private wallet: %w", err)
	}
	if err = manager.probe(ctx, probeRequest{
		RPCURL: runtime.RPCURL, Address: walletAddress, ExpectedChainID: expectedChainID,
	}); err != nil {
		return runtimeTopology{}, err
	}
	return runtime, nil
}

func (manager *Manager) Authenticate(ctx context.Context, networkDir string) (Environment, error) {
	canonicalNetworkDir, err := canonicalExistingDirectory(networkDir, "network directory")
	if err != nil {
		return Environment{}, err
	}
	runtime, err := manager.status(ctx, canonicalNetworkDir)
	if err != nil {
		return Environment{}, err
	}
	environment := Environment{
		RPCURL:       runtime.RPCURL,
		GraphQLURL:   runtime.GraphQLURL,
		WebSocketURL: runtime.WebSocketURL,
		SeedFile:     walletSeedPath(canonicalNetworkDir),
	}
	return environment, nil
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
	ownership, err := loadOwnership(networkDir)
	if err != nil {
		return err
	}
	client, err := manager.newClient()
	if err != nil {
		return err
	}
	enclave, err := ownership.Enclave()
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
			return fmt.Errorf("readiness deadline reached after: %v: %w", last, ctx.Err())
		case <-timer.C:
		}
	}
}

var _ Authenticator = (*Manager)(nil)
var _ Controller = (*Manager)(nil)
