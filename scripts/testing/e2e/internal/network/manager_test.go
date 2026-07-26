// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

type managerFixture struct {
	manager    *Manager
	client     *fakeKurtosis
	networkDir string
}

const testExecutionImage = "local/go-qrl:test"

func newManagerFixture(t *testing.T) managerFixture {
	t.Helper()
	networkDir, err := ensureNetworkDirectory(t.TempDir())
	require.NoError(t, err)
	client := &fakeKurtosis{
		Enclave:          kurtosis.EnclaveRef{UUID: strings.Repeat("a", 32)},
		ExecutionService: executionServiceFixture(),
	}
	manager := NewManager()
	manager.newClient = func() (kurtosisClient, error) { return client, nil }
	manager.probe = func(context.Context, string, string) error { return nil }
	manager.createRecoveryTimeout = 25 * time.Millisecond
	manager.createRecoveryInitial = time.Millisecond
	return managerFixture{manager: manager, client: client, networkDir: networkDir}
}

func executionServiceFixture() kurtosis.Service {
	return kurtosis.Service{
		PublicIP:    "127.0.0.1",
		PublicPorts: map[string]uint16{rpcPortID: 18545, webSocketPortID: 18546},
	}
}

func TestStartPersistsExactOwnershipAndRefusesStaleReuse(t *testing.T) {
	fixture := newManagerFixture(t)
	require.NoError(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage))
	ownership, err := loadOwnership(fixture.networkDir)
	require.NoError(t, err)
	require.Equal(t, fixture.client.Enclave.UUID, ownership.UUID)
	require.ErrorContains(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage), "already exists")
	environment, err := fixture.manager.Inspect(t.Context(), fixture.networkDir)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:18545", environment.RPCURL)
	require.Equal(t, "http://127.0.0.1:18545/graphql", environment.GraphQLURL)
	require.Equal(t, "ws://127.0.0.1:18546", environment.WebSocketURL)
	require.Equal(t, walletSeedPath(fixture.networkDir), environment.SeedFile)
	require.Equal(t, 1, countCalls(fixture.client.Calls, "create:"))
	require.Equal(t, 1, countCalls(fixture.client.Calls, "run:"))
	require.Len(t, fixture.client.Runs, 1)
	require.Equal(t, packageLocator, fixture.client.Runs[0].Locator)
	require.Contains(t, fixture.client.Runs[0].SerializedParams, `"el_image":"local/go-qrl:test"`)
	require.Zero(t, countCalls(fixture.client.Calls, "get:"))
}

func TestStatusFailsWhenNetworkIsNotRunning(t *testing.T) {
	fixture := newManagerFixture(t)
	_, err := fixture.manager.Inspect(t.Context(), fixture.networkDir)
	require.ErrorContains(t, err, "not running")
}

func TestInspectRequiresExecutionEndpoints(t *testing.T) {
	for _, portID := range []string{rpcPortID, webSocketPortID} {
		t.Run(portID, func(t *testing.T) {
			fixture := newManagerFixture(t)
			_, err := ensureWallet(fixture.networkDir)
			require.NoError(t, err)
			delete(fixture.client.ExecutionService.PublicPorts, portID)

			_, err = fixture.manager.inspectEnclave(
				t.Context(),
				fixture.client,
				fixture.client.Enclave,
				fixture.networkDir,
			)
			require.ErrorContains(t, err, "lacks public")
			require.ErrorContains(t, err, portID)
		})
	}
}

func TestLifecycleFailsClosedOnDanglingOwnership(t *testing.T) {
	fixture := newManagerFixture(t)
	require.NoError(t, os.Symlink("missing-ownership", ownershipPath(fixture.networkDir)))
	require.Error(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage))
	require.Error(t, fixture.manager.Stop(t.Context(), fixture.networkDir))
	require.Empty(t, fixture.client.Calls, "dangling ownership triggered external mutation")
}

func TestFailedProvisioningRetainsExactOwnershipUntilStop(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.RunError = errors.New("package failed")
	require.Error(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage))
	ownership, err := loadOwnership(fixture.networkDir)
	require.NoError(t, err)
	require.Equal(t, fixture.client.Enclave.UUID, ownership.UUID)
	fixture.client.RunError = nil
	require.ErrorContains(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage), "network-stop")
	require.NoError(t, fixture.manager.Stop(t.Context(), fixture.networkDir))
	require.NoError(t, fixture.manager.Stop(t.Context(), fixture.networkDir))
	require.True(t, fixture.client.Destroyed)
	require.Equal(t, 1, countCalls(fixture.client.Calls, "destroy:"))
}

func TestAmbiguousCreateRecoversAndPersistsExactOwnership(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.CreateError = errors.New("create response lost")
	fixture.client.CreateAfterError = true
	require.NoError(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage))
	record, err := loadOwnership(fixture.networkDir)
	require.NoError(t, err)
	require.Equal(t, fixture.client.Enclave.UUID, record.UUID)
	require.ErrorContains(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage), "already exists")
	require.NoError(t, fixture.manager.Stop(t.Context(), fixture.networkDir))
	require.Equal(t, 1, countCalls(fixture.client.Calls, "create:"))
	require.Equal(t, 1, countCalls(fixture.client.Calls, "run:"))
	require.True(t, fixture.client.Destroyed)
}

func TestAmbiguousCreateWaitsForEnclavePublication(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.CreateError = errors.New("create response lost")
	fixture.client.CreateAfterError = true
	fixture.client.GetFailures = 1

	require.NoError(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage))
	require.Equal(t, 2, countCalls(fixture.client.Calls, "get:"))
	require.Equal(t, 1, countCalls(fixture.client.Calls, "run:"))
	record, err := loadOwnership(fixture.networkDir)
	require.NoError(t, err)
	require.Equal(t, fixture.client.Enclave.UUID, record.UUID)
}

func TestAmbiguousCreateUsesIndependentRecoveryContext(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.CreateError = errors.New("create response lost")
	fixture.client.CreateAfterError = true
	ctx, cancel := context.WithCancel(t.Context())
	fixture.client.CreateCallback = cancel

	err := fixture.manager.Start(ctx, fixture.networkDir, testExecutionImage)
	require.ErrorIs(t, err, context.Canceled)
	require.ErrorContains(t, err, "recovered exact ownership")
	record, loadErr := loadOwnership(fixture.networkDir)
	require.NoError(t, loadErr)
	require.Equal(t, fixture.client.Enclave.UUID, record.UUID)
	require.Equal(t, 1, countCalls(fixture.client.Calls, "get:"))
	require.Zero(t, countCalls(fixture.client.Calls, "run:"))
}

func TestAmbiguousCreateCleansUpWhenOwnershipCannotBePersisted(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.CreateError = errors.New("create response lost")
	fixture.client.CreateAfterError = true
	fixture.client.CreateCallback = func() {
		require.NoError(t, writeExclusive(ownershipPath(fixture.networkDir), []byte("{}\n")))
	}

	err := fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage)
	require.ErrorContains(t, err, "persist exact enclave ownership")
	require.True(t, fixture.client.Destroyed)
	require.Equal(t, 1, countCalls(fixture.client.Calls, "destroy:"))
	require.Zero(t, countCalls(fixture.client.Calls, "run:"))
}

func TestUnresolvedCreateErrorLeavesNoPartialOwnership(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.CreateError = errors.New("create failed")
	require.ErrorContains(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage), "recover ambiguous creation")
	_, err := os.Stat(ownershipPath(fixture.networkDir))
	require.ErrorIs(t, err, os.ErrNotExist)
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

func TestStopRetainsOwnershipWhenAdapterRejectsIdentity(t *testing.T) {
	fixture := newManagerFixture(t)
	require.NoError(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage))
	fixture.client.DestroyError = errors.New("enclave identity changed")
	require.ErrorContains(t, fixture.manager.Stop(t.Context(), fixture.networkDir), "identity changed")
	require.False(t, fixture.client.Destroyed)
	_, err := loadOwnership(fixture.networkDir)
	require.NoError(t, err)
}

func TestStopRemovesOwnershipWhenDestroyReportsAnAbsentEnclave(t *testing.T) {
	for name, arrange := range map[string]func(*fakeKurtosis){
		"lost response": func(client *fakeKurtosis) {
			client.DestroyError = errors.New("destroy response lost")
			client.DestroyAfterError = true
		},
		"already absent": func(client *fakeKurtosis) {
			client.Destroyed = true
			client.DestroyError = errors.New("enclave not found")
		},
	} {
		t.Run(name, func(t *testing.T) {
			fixture := newManagerFixture(t)
			require.NoError(t, fixture.manager.Start(t.Context(), fixture.networkDir, testExecutionImage))
			arrange(fixture.client)
			require.NoError(t, fixture.manager.Stop(t.Context(), fixture.networkDir))
			require.True(t, fixture.client.Destroyed)
			require.Equal(t, 1, countCalls(fixture.client.Calls, "destroy:"))
			_, err := os.Stat(ownershipPath(fixture.networkDir))
			require.ErrorIs(t, err, os.ErrNotExist)
		})
	}
}

func countCalls(calls []string, prefix string) int {
	count := 0
	for _, call := range calls {
		if strings.HasPrefix(call, prefix) {
			count++
		}
	}
	return count
}
