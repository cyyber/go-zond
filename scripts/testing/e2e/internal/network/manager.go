// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

type Manager struct {
	NewClient func() (kurtosis.Client, error)
	Commands  commandRunner
	Stdout    io.Writer
	Stderr    io.Writer
	Prepare   func(context.Context, commandRunner, StartRequest, string, io.Writer, io.Writer) (preparedNetwork, error)
	Probe     func(context.Context, probeRequest) error
}

func NewManager() *Manager {
	return &Manager{
		NewClient: func() (kurtosis.Client, error) { return kurtosis.NewSDKClient() },
		Commands:  execRunner{},
		Stdout:    io.Discard,
		Stderr:    io.Discard,
		Prepare:   prepareNetwork,
		Probe:     probeNetwork,
	}
}

func (manager *Manager) Start(ctx context.Context, request StartRequest) (Result, error) {
	repoRoot, err := canonicalExistingDirectory(request.RepoRoot, "repository root")
	if err != nil {
		return Result{}, err
	}
	request.RepoRoot = repoRoot
	networkDir, err := ensureNetworkDirectory(request.NetworkDir)
	if err != nil {
		return Result{}, err
	}
	request.NetworkDir = networkDir
	if relative, err := filepath.Rel(repoRoot, networkDir); err != nil {
		return Result{}, err
	} else if relative == "." || relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return Result{}, errors.New("network directory must not be the repository or one of its descendants")
	}
	mutation, err := acquireMutationLease(networkDir)
	if err != nil {
		return Result{}, err
	}
	defer mutation.Close()
	if request.StartTimeout <= 0 {
		request.StartTimeout = 150 * time.Minute
	}
	startCtx, cancel := context.WithTimeout(ctx, request.StartTimeout)
	defer cancel()

	readyExists, err := pathExists(statePath(networkDir))
	if err != nil {
		return Result{}, err
	}
	ownershipExists, err := pathExists(ownershipPath(networkDir))
	if err != nil {
		return Result{}, err
	}
	switch {
	case readyExists && !ownershipExists:
		return Result{}, errors.New("ready network state has no exact-UUID ownership record")
	case ownershipExists && !readyExists:
		ownership, loadErr := loadOwnership(networkDir)
		if loadErr != nil {
			return Result{}, loadErr
		}
		if _, ownershipErr := ownership.OwnedEnclave(); ownershipErr != nil {
			return Result{}, ownershipErr
		}
		return Result{}, errors.New("network provisioning is incomplete; run network-stop before starting again")
	case readyExists:
		return Result{}, errors.New("network is already running; use status or stop first")
	}

	walletAddress, err := ensureWallet(privatePath(networkDir))
	if err != nil {
		return Result{}, fmt.Errorf("prepare private E2E wallet: %w", err)
	}
	client, err := manager.NewClient()
	if err != nil {
		return Result{}, fmt.Errorf("connect to Kurtosis engine: %w", err)
	}
	prepared, err := manager.Prepare(startCtx, manager.Commands, request, walletAddress, manager.Stdout, manager.Stderr)
	if err != nil {
		return Result{}, err
	}
	name := defaultEnclaveName(networkDir)
	intent := OwnershipRecord{
		NetworkDir: networkDir,
		Name:       name,
	}
	if err := createOwnership(intent); err != nil {
		return Result{}, fmt.Errorf("persist enclave creation intent: %w", err)
	}
	enclave, err := client.CreateEnclave(startCtx, name)
	if err != nil {
		return Result{}, fmt.Errorf(
			"create Kurtosis enclave %q returned an ambiguous result; creation intent was retained and create will not be replayed: %w",
			name,
			err,
		)
	}
	if enclave.Name != name || !enclave.Owned || enclave.Validate() != nil {
		cleanupErr := cleanupCreatedEnclave(client, enclave)
		if cleanupErr == nil {
			cleanupErr = removeOwnership(intent)
		}
		return Result{}, errors.Join(
			errors.New("Kurtosis returned an unexpected or unowned enclave identity"),
			wrapCleanupError(cleanupErr),
		)
	}
	ownership := intent
	ownership.UUID = enclave.UUID
	if err := captureOwnership(ownership); err != nil {
		cleanupErr := cleanupCreatedEnclave(client, enclave)
		if cleanupErr == nil {
			cleanupErr = removeOwnership(intent)
		} else if recoveryErr := captureOwnership(ownership); recoveryErr != nil {
			cleanupErr = errors.Join(
				cleanupErr,
				fmt.Errorf(
					"persist recovery ownership for enclave %s/%s: %w",
					enclave.Name,
					enclave.UUID,
					recoveryErr,
				),
			)
		}
		return Result{}, errors.Join(
			fmt.Errorf("persist exact enclave ownership: %w", err),
			wrapCleanupError(cleanupErr),
		)
	}

	if err := client.RunRemotePackage(startCtx, enclave, kurtosis.PackageRun{
		Locator:          packageLocator,
		SerializedParams: prepared.Params,
	}); err != nil {
		return Result{}, fmt.Errorf("run pinned qrl-package; network ownership was retained for network-stop: %w", err)
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
		return Result{}, fmt.Errorf("discover qrl-package topology; network ownership was retained for network-stop: %w", err)
	}
	if err := waitUntil(startCtx, 2*time.Second, func(attempt context.Context) error {
		return manager.Probe(attempt, probeRequest{
			RPCURL:  runtime.RPCURL,
			Address: walletAddress, ExpectedChainID: expectedChainID,
		})
	}); err != nil {
		return Result{}, fmt.Errorf("wait for network readiness; network ownership was retained for network-stop: %w", err)
	}
	state := State{Ready: true}
	if err := writeState(networkDir, state); err != nil {
		return Result{}, fmt.Errorf("publish ready network state; exact ownership was retained for network-stop: %w", err)
	}
	return Result{topology: runtime, Ready: true}, nil
}

func (manager *Manager) Status(ctx context.Context, requestedDir string) (Result, error) {
	return manager.status(ctx, requestedDir)
}

func (manager *Manager) status(ctx context.Context, requestedDir string) (Result, error) {
	networkDir, err := canonicalExistingDirectory(requestedDir, "network directory")
	if err != nil {
		return Result{}, err
	}
	readyExists, err := pathExists(statePath(networkDir))
	if err != nil {
		return Result{}, err
	}
	ownershipExists, err := pathExists(ownershipPath(networkDir))
	if err != nil {
		return Result{}, err
	}
	if !ownershipExists {
		if readyExists {
			return Result{}, errors.New("ready network state has no exact-UUID ownership record")
		}
		return Result{Message: "network is not running"}, nil
	}
	ownership, err := loadOwnership(networkDir)
	if err != nil {
		return Result{}, err
	}
	enclave, err := ownership.OwnedEnclave()
	if err != nil {
		return Result{Message: err.Error()}, nil
	}
	client, err := manager.NewClient()
	if err != nil {
		return Result{}, err
	}
	if err := authenticateOwnedEnclave(ctx, client, enclave); err != nil {
		return Result{}, err
	}
	if !readyExists {
		return Result{Message: "network provisioning is incomplete; run network-stop before starting again"}, nil
	}
	if _, err := loadState(networkDir); err != nil {
		return Result{}, err
	}
	services, err := client.Services(ctx, enclave)
	if err != nil {
		return Result{}, err
	}
	runtime, err := discoverTopology(services)
	if err != nil {
		return Result{}, err
	}
	walletAddress, err := validateWalletSeed(walletSeedPath(networkDir))
	if err != nil {
		return Result{}, fmt.Errorf("validate private wallet: %w", err)
	}
	if err = manager.Probe(ctx, probeRequest{
		RPCURL: runtime.RPCURL, Address: walletAddress, ExpectedChainID: expectedChainID,
	}); err != nil {
		return Result{}, err
	}
	return Result{topology: runtime, Ready: true}, nil
}

func (manager *Manager) Authenticate(ctx context.Context, networkDir string) (Environment, error) {
	canonicalNetworkDir, err := canonicalExistingDirectory(networkDir, "network directory")
	if err != nil {
		return Environment{}, err
	}
	result, err := manager.status(ctx, canonicalNetworkDir)
	if err != nil {
		return Environment{}, err
	}
	if !result.Ready {
		return Environment{}, errors.New("E2E network is not ready")
	}
	environment := Environment{
		RPCURL:       result.topology.RPCURL,
		GraphQLURL:   result.topology.GraphQLURL,
		WebSocketURL: result.topology.WebSocketURL,
		SeedFile:     walletSeedPath(canonicalNetworkDir),
	}
	return environment, nil
}

func (manager *Manager) Stop(ctx context.Context, requestedDir string) (Result, error) {
	networkDir, err := canonicalExistingDirectory(requestedDir, "network directory")
	if err != nil {
		return Result{}, err
	}
	mutation, err := acquireMutationLease(networkDir)
	if err != nil {
		return Result{}, err
	}
	defer mutation.Close()

	readyExists, err := pathExists(statePath(networkDir))
	if err != nil {
		return Result{}, err
	}
	ownershipExists, err := pathExists(ownershipPath(networkDir))
	if err != nil {
		return Result{}, err
	}
	if !ownershipExists {
		if readyExists {
			return Result{}, errors.New("refusing stop: ready network state has no exact-UUID ownership record")
		}
		return Result{Message: "network is not running"}, nil
	}
	ownership, err := loadOwnership(networkDir)
	if err != nil {
		return Result{}, err
	}
	client, err := manager.NewClient()
	if err != nil {
		return Result{}, err
	}
	enclave, err := ownership.OwnedEnclave()
	if err != nil {
		return Result{}, err
	}
	if err := destroyExactEnclave(ctx, client, enclave); err != nil {
		return Result{}, err
	}
	// Remove the public ready state first. If either removal is interrupted,
	// the private exact-UUID record remains available for an idempotent stop.
	if readyExists {
		if err := removeState(networkDir); err != nil {
			return Result{}, err
		}
	}
	if err := removeOwnership(ownership); err != nil {
		return Result{}, err
	}
	return Result{Message: "network stopped"}, nil
}

func destroyExactEnclave(ctx context.Context, client kurtosis.Client, enclave kurtosis.EnclaveRef) error {
	exists, err := client.EnclaveExists(ctx, enclave.UUID)
	if err != nil {
		return fmt.Errorf("inspect owned enclave existence by exact UUID: %w", err)
	}
	if !exists {
		return nil
	}
	if err := authenticateOwnedEnclave(ctx, client, enclave); err != nil {
		return err
	}
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

func cleanupCreatedEnclave(client kurtosis.Client, enclave kurtosis.EnclaveRef) error {
	if err := enclave.Validate(); err != nil || !enclave.Owned {
		return errors.New("cannot clean up an invalid or unowned returned enclave identity")
	}
	cleanupCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	return destroyExactEnclave(cleanupCtx, client, enclave)
}

func authenticateOwnedEnclave(ctx context.Context, client kurtosis.Client, enclave kurtosis.EnclaveRef) error {
	current, err := client.GetEnclave(ctx, enclave.UUID)
	if err != nil {
		return fmt.Errorf("inspect owned enclave by exact UUID: %w", err)
	}
	if current.Name != enclave.Name || current.UUID != enclave.UUID {
		return errors.New("refusing operation: owned enclave name/UUID changed")
	}
	return nil
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

func defaultEnclaveName(canonicalNetworkDir string) string {
	return "go-qrl-e2e-" + networkID(canonicalNetworkDir)
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
