// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/theQRL/go-qrl/testing/endtoend/internal/kurtosis"
)

const testExecutionImage = "local/go-qrl:test"

type managerFixture struct {
	manager    *Manager
	client     *fakeKurtosis
	networkDir string
}

func newManagerFixture(t *testing.T) managerFixture {
	t.Helper()
	networkDir, err := ensureNetworkDirectory(t.TempDir())
	require.NoError(t, err)
	client := &fakeKurtosis{
		Enclave: kurtosis.EnclaveRef{UUID: strings.Repeat("a", 32)},
		ExecutionService: kurtosis.Service{
			PublicIP:    "127.0.0.1",
			PublicPorts: map[string]uint16{rpcPortID: 18545, webSocketPortID: 18546},
		},
	}
	manager := NewManager()
	manager.newClient = func() (kurtosisClient, error) { return client, nil }
	manager.probe = func(context.Context, string, string) error { return nil }
	manager.createRecoveryTimeout, manager.createRecoveryInitial = 25*time.Millisecond, time.Millisecond
	return managerFixture{manager, client, networkDir}
}

func TestEnclaveNameIsStable192BitSlot(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	name := enclaveName(first)
	require.Equal(t, name, enclaveName(first))
	require.NotEqual(t, name, enclaveName(second))
	digest, err := hex.DecodeString(strings.TrimPrefix(name, "go-qrl-e2e-"))
	require.NoError(t, err)
	require.Len(t, digest, 24)
}

func TestStartOwnsOneDeterministicSlot(t *testing.T) {
	fixture := newManagerFixture(t)
	require.NoError(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage))
	require.Equal(t, enclaveName(fixture.networkDir), fixture.client.Enclave.Name)
	require.Equal(t, 1, fixture.client.Creates)
	require.Equal(t, 1, fixture.client.Runs)
	require.Equal(t, packageLocator, fixture.client.RunLocator)
	require.Contains(t, fixture.client.RunParameters, `"el_image":"local/go-qrl:test"`)

	environment, err := fixture.manager.Inspect(t.Context(), fixture.networkDir)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:18545", environment.RPCURL)
	require.Equal(t, "http://127.0.0.1:18545/graphql", environment.GraphQLURL)
	require.Equal(t, "ws://127.0.0.1:18546", environment.WebSocketURL)
	require.Equal(t, walletSeedPath(fixture.networkDir), environment.SeedFile)

	require.ErrorContains(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage), "already exists")
	require.Equal(t, 1, fixture.client.Creates)
	require.Equal(t, 1, fixture.client.Runs)
}

func TestLifecycleDoesNotTreatLookupFailureAsAbsence(t *testing.T) {
	operations := []struct {
		name string
		run  func(context.Context, *managerFixture) error
	}{
		{"start", func(ctx context.Context, f *managerFixture) error {
			return f.manager.Start(ctx, f.networkDir, testExecutionImage)
		}},
		{"status", func(ctx context.Context, f *managerFixture) error {
			_, err := f.manager.Inspect(ctx, f.networkDir)
			return err
		}},
		{"stop", func(ctx context.Context, f *managerFixture) error {
			return f.manager.Stop(ctx, f.networkDir)
		}},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			fixture := newManagerFixture(t)
			lookupErr := errors.New("engine listing failed")
			fixture.client.LookupError = lookupErr
			err := operation.run(t.Context(), &fixture)
			require.ErrorIs(t, err, lookupErr)
			require.ErrorContains(t, err, "resolve deterministic network slot")
			require.Zero(t, fixture.client.Creates)
			require.Zero(t, fixture.client.Runs)
			require.Zero(t, fixture.client.Destroys)
		})
	}
}

func TestInspectAndStopAbsentSlot(t *testing.T) {
	fixture := newManagerFixture(t)
	_, err := fixture.manager.Inspect(t.Context(), fixture.networkDir)
	require.ErrorContains(t, err, "not running")
	require.NoError(t, fixture.manager.Stop(t.Context(), fixture.networkDir))
	require.Zero(t, fixture.client.Destroys)
}

func TestInspectRequiresExecutionEndpoints(t *testing.T) {
	for _, portID := range []string{rpcPortID, webSocketPortID} {
		t.Run(portID, func(t *testing.T) {
			fixture := newManagerFixture(t)
			_, err := ensureWallet(fixture.networkDir)
			require.NoError(t, err)
			delete(fixture.client.ExecutionService.PublicPorts, portID)
			_, err = fixture.manager.inspectEnclave(t.Context(), fixture.client, fixture.client.Enclave, fixture.networkDir)
			require.ErrorContains(t, err, portID)
		})
	}
}

func TestProvisioningFailureLeavesStoppableSlot(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.RunError = errors.New("package failed")
	require.ErrorContains(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage), "slot remains")
	require.True(t, fixture.client.Exists)
	require.ErrorContains(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage), "network-stop")

	fixture.client.RunError = nil
	require.NoError(t, fixture.manager.Stop(t.Context(), fixture.networkDir))
	require.NoError(t, fixture.manager.Stop(t.Context(), fixture.networkDir))
	require.False(t, fixture.client.Exists)
	require.Equal(t, 1, fixture.client.Destroys)
}

func TestAmbiguousCreateRecovery(t *testing.T) {
	t.Run("published slot", func(t *testing.T) {
		fixture := newManagerFixture(t)
		fixture.client.CreateError = errors.New("create response lost")
		fixture.client.CreateAfterError = true
		fixture.client.CreateCallback = func() { fixture.client.RecoveryFailures = 1 }
		require.NoError(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage))
		require.Equal(t, 3, fixture.client.Lookups) // preflight plus two recovery attempts
		require.Equal(t, 1, fixture.client.Runs)
	})
	t.Run("independent recovery context", func(t *testing.T) {
		fixture := newManagerFixture(t)
		fixture.client.CreateError = errors.New("create response lost")
		fixture.client.CreateAfterError = true
		ctx, cancel := context.WithCancel(t.Context())
		fixture.client.CreateCallback = cancel
		err := fixture.manager.Start(ctx, fixture.networkDir, testExecutionImage)
		require.ErrorIs(t, err, context.Canceled)
		require.ErrorContains(t, err, "recovered its deterministic slot")
		require.True(t, fixture.client.Exists)
		require.Equal(t, 2, fixture.client.Lookups)
		require.Zero(t, fixture.client.Runs)
	})
	t.Run("unresolved", func(t *testing.T) {
		fixture := newManagerFixture(t)
		fixture.client.CreateError = errors.New("create failed")
		err := fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage)
		require.ErrorContains(t, err, "recover ambiguous creation")
		require.False(t, fixture.client.Exists)
		require.Zero(t, fixture.client.Runs)
	})
}

func TestUnexpectedCreatedIdentityIsCleanedUp(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.Enclave.Name = "unexpected-enclave"
	err := fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage)
	require.ErrorContains(t, err, "unexpected enclave identity")
	require.False(t, fixture.client.Exists)
	require.Equal(t, 1, fixture.client.Destroys)
	require.Zero(t, fixture.client.Runs)
}

func TestRetryUntilPreservesCancellationAndLastError(t *testing.T) {
	lastErr := errors.New("last operation failed")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	attempts := 0
	err := retryUntil(ctx, time.Hour, time.Hour, func(context.Context) error {
		attempts++
		return lastErr
	})
	require.Equal(t, 1, attempts)
	require.ErrorIs(t, err, context.Canceled)
	require.ErrorIs(t, err, lastErr)
}

func TestStopReconcilesDestroyOutcomes(t *testing.T) {
	t.Run("rejected then retried", func(t *testing.T) {
		fixture := newManagerFixture(t)
		require.NoError(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage))
		fixture.client.DestroyError = errors.New("destroy rejected")
		require.ErrorContains(t, fixture.manager.Stop(t.Context(), fixture.networkDir), "destroy rejected")
		require.True(t, fixture.client.Exists)
		fixture.client.DestroyError = nil
		require.NoError(t, fixture.manager.Stop(t.Context(), fixture.networkDir))
		require.Equal(t, 2, fixture.client.Destroys)
	})
	t.Run("lost response", func(t *testing.T) {
		fixture := newManagerFixture(t)
		require.NoError(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage))
		fixture.client.DestroyError = errors.New("destroy response lost")
		fixture.client.DestroyAfterError = true
		require.NoError(t, fixture.manager.Stop(t.Context(), fixture.networkDir))
		require.False(t, fixture.client.Exists)
		require.Equal(t, 1, fixture.client.Destroys)
	})
}
