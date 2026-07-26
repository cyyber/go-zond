// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	if err != nil {
		t.Fatal(err)
	}
	client := &fakeKurtosis{
		Enclave:          kurtosis.EnclaveRef{UUID: strings.Repeat("a", 32)},
		ExecutionService: topologyFixture(),
	}
	manager := NewManager()
	manager.newClient = func() (kurtosisClient, error) { return client, nil }
	manager.probe = func(context.Context, string, string) error { return nil }
	manager.createRecoveryTimeout = 25 * time.Millisecond
	manager.createRecoveryPollPeriod = time.Millisecond
	return managerFixture{manager: manager, client: client, networkDir: networkDir}
}

func TestStartPersistsExactOwnershipAndRefusesStaleReuse(t *testing.T) {
	fixture := newManagerFixture(t)
	if err := fixture.manager.Start(context.Background(), fixture.networkDir, testExecutionImage); err != nil {
		t.Fatal(err)
	}
	ownership, err := loadOwnership(fixture.networkDir)
	if err != nil || ownership.UUID != fixture.client.Enclave.UUID {
		t.Fatalf("ownership=%+v err=%v", ownership, err)
	}
	if err := fixture.manager.Start(context.Background(), fixture.networkDir, testExecutionImage); err == nil ||
		!strings.Contains(err.Error(), "already exists") {
		t.Fatalf("stale-reuse error = %v", err)
	}
	environment, err := fixture.manager.Inspect(context.Background(), fixture.networkDir)
	if err != nil || environment.RPCURL != "http://127.0.0.1:18545" {
		t.Fatalf("environment=%+v err=%v", environment, err)
	}
	if countCalls(fixture.client.Calls, "create:") != 1 ||
		countCalls(fixture.client.Calls, "run:") != 1 {
		t.Fatalf("network was reprovisioned: %v", fixture.client.Calls)
	}
	if len(fixture.client.Runs) != 1 ||
		fixture.client.Runs[0].Locator != packageLocator ||
		!strings.Contains(fixture.client.Runs[0].SerializedParams, `"el_image":"local/go-qrl:test"`) {
		t.Fatalf("package runs = %+v", fixture.client.Runs)
	}
	if countCalls(fixture.client.Calls, "get:") != 0 {
		t.Fatalf("manager repeated adapter identity lookups: %v", fixture.client.Calls)
	}
}

func TestStatusFailsWhenNetworkIsNotRunning(t *testing.T) {
	fixture := newManagerFixture(t)
	if _, err := fixture.manager.Inspect(context.Background(), fixture.networkDir); err == nil ||
		!strings.Contains(err.Error(), "not running") {
		t.Fatalf("status error = %v", err)
	}
}

func TestLifecycleFailsClosedOnDanglingOwnership(t *testing.T) {
	fixture := newManagerFixture(t)
	if err := os.Symlink("missing-ownership", ownershipPath(fixture.networkDir)); err != nil {
		t.Fatal(err)
	}
	if err := fixture.manager.Start(context.Background(), fixture.networkDir, testExecutionImage); err == nil {
		t.Fatal("start accepted dangling ownership")
	}
	if err := fixture.manager.Stop(context.Background(), fixture.networkDir); err == nil {
		t.Fatal("stop accepted dangling ownership")
	}
	if len(fixture.client.Calls) != 0 {
		t.Fatalf("dangling ownership triggered external mutation: %v", fixture.client.Calls)
	}
}

func TestLifecycleFailsClosedOnLegacyOwnership(t *testing.T) {
	fixture := newManagerFixture(t)
	current := fixture.client.Enclave
	current.Name = enclaveNamePrefix(fixture.networkDir) + "current"
	if err := createOwnership(fixture.networkDir, current); err != nil {
		t.Fatal(err)
	}
	legacy := ownershipPath(filepath.Join(fixture.networkDir, "private"))
	if err := os.Mkdir(filepath.Dir(legacy), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := writeExclusive(legacy, []byte("{}\n")); err != nil {
		t.Fatal(err)
	}
	for name, operation := range map[string]func() error{
		"start": func() error {
			return fixture.manager.Start(context.Background(), fixture.networkDir, testExecutionImage)
		},
		"stop": func() error {
			return fixture.manager.Stop(context.Background(), fixture.networkDir)
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := operation(); err == nil || !strings.Contains(err.Error(), "legacy network ownership") {
				t.Fatalf("%s error = %v", name, err)
			}
		})
	}
	if len(fixture.client.Calls) != 0 {
		t.Fatalf("legacy ownership triggered external mutation: %v", fixture.client.Calls)
	}
}

func TestFailedProvisioningRetainsExactOwnershipUntilStop(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.RunError = errors.New("package failed")
	if err := fixture.manager.Start(context.Background(), fixture.networkDir, testExecutionImage); err == nil {
		t.Fatal("failed provisioning succeeded")
	}
	ownership, err := loadOwnership(fixture.networkDir)
	if err != nil || ownership.UUID != fixture.client.Enclave.UUID {
		t.Fatalf("retained ownership=%+v err=%v", ownership, err)
	}
	fixture.client.RunError = nil
	if err := fixture.manager.Start(context.Background(), fixture.networkDir, testExecutionImage); err == nil ||
		!strings.Contains(err.Error(), "network-stop") {
		t.Fatalf("incomplete restart error = %v", err)
	}
	if err := fixture.manager.Stop(context.Background(), fixture.networkDir); err != nil {
		t.Fatal(err)
	}
	if err := fixture.manager.Stop(context.Background(), fixture.networkDir); err != nil {
		t.Fatalf("second stop: %v", err)
	}
	if !fixture.client.Destroyed {
		t.Fatal("owned enclave was not destroyed")
	}
	if calls := countCalls(fixture.client.Calls, "destroy:"); calls != 1 {
		t.Fatalf("destroy calls after second stop = %d", calls)
	}
}

func TestAmbiguousCreateRecoversAndPersistsExactOwnership(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.CreateError = errors.New("create response lost")
	fixture.client.CreateAfterError = true
	if err := fixture.manager.Start(context.Background(), fixture.networkDir, testExecutionImage); err != nil {
		t.Fatalf("recover ambiguous create: %v", err)
	}
	record, err := loadOwnership(fixture.networkDir)
	if err != nil || record.UUID != fixture.client.Enclave.UUID {
		t.Fatalf("recovered ownership=%+v err=%v", record, err)
	}
	if err := fixture.manager.Start(context.Background(), fixture.networkDir, testExecutionImage); err == nil ||
		!strings.Contains(err.Error(), "already exists") {
		t.Fatalf("ambiguous restart error = %v", err)
	}
	if err := fixture.manager.Stop(context.Background(), fixture.networkDir); err != nil {
		t.Fatal(err)
	}
	if countCalls(fixture.client.Calls, "create:") != 1 ||
		countCalls(fixture.client.Calls, "run:") != 1 ||
		!fixture.client.Destroyed {
		t.Fatalf("unsafe ambiguous handling: %v", fixture.client.Calls)
	}
}

func TestAmbiguousCreateWaitsForEnclavePublication(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.CreateError = errors.New("create response lost")
	fixture.client.CreateAfterError = true
	fixture.client.GetFailures = 1

	if err := fixture.manager.Start(context.Background(), fixture.networkDir, testExecutionImage); err != nil {
		t.Fatalf("recover delayed enclave publication: %v", err)
	}
	if countCalls(fixture.client.Calls, "get:") != 2 ||
		countCalls(fixture.client.Calls, "run:") != 1 {
		t.Fatalf("ambiguous creation was not retried safely: %v", fixture.client.Calls)
	}
	record, err := loadOwnership(fixture.networkDir)
	if err != nil || record.UUID != fixture.client.Enclave.UUID {
		t.Fatalf("recovered ownership=%+v err=%v", record, err)
	}
}

func TestAmbiguousCreateUsesIndependentRecoveryContext(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.CreateError = errors.New("create response lost")
	fixture.client.CreateAfterError = true
	ctx, cancel := context.WithCancel(context.Background())
	fixture.client.CreateCallback = cancel

	err := fixture.manager.Start(ctx, fixture.networkDir, testExecutionImage)
	if !errors.Is(err, context.Canceled) ||
		!strings.Contains(err.Error(), "recovered exact ownership") {
		t.Fatalf("ambiguous canceled create error = %v", err)
	}
	record, loadErr := loadOwnership(fixture.networkDir)
	if loadErr != nil || record.UUID != fixture.client.Enclave.UUID {
		t.Fatalf("recovered ownership=%+v err=%v", record, loadErr)
	}
	if countCalls(fixture.client.Calls, "get:") != 1 ||
		countCalls(fixture.client.Calls, "run:") != 0 {
		t.Fatalf("recovery did not isolate cancellation: %v", fixture.client.Calls)
	}
}

func TestAmbiguousCreateCleansUpWhenOwnershipCannotBePersisted(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.CreateError = errors.New("create response lost")
	fixture.client.CreateAfterError = true
	fixture.client.CreateCallback = func() {
		if err := writeExclusive(ownershipPath(fixture.networkDir), []byte("{}\n")); err != nil {
			t.Errorf("create conflicting ownership: %v", err)
		}
	}

	err := fixture.manager.Start(context.Background(), fixture.networkDir, testExecutionImage)
	if err == nil || !strings.Contains(err.Error(), "persist exact enclave ownership") {
		t.Fatalf("persistence error = %v", err)
	}
	if !fixture.client.Destroyed ||
		countCalls(fixture.client.Calls, "destroy:") != 1 ||
		countCalls(fixture.client.Calls, "run:") != 0 {
		t.Fatalf("recovered enclave was orphaned: %v", fixture.client.Calls)
	}
}

func TestUnresolvedCreateErrorLeavesNoPartialOwnership(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.CreateError = errors.New("create failed")
	if err := fixture.manager.Start(context.Background(), fixture.networkDir, testExecutionImage); err == nil ||
		!strings.Contains(err.Error(), "recover ambiguous creation") {
		t.Fatalf("create error = %v", err)
	}
	if _, err := os.Stat(ownershipPath(fixture.networkDir)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("partial ownership remains: %v", err)
	}
}

func TestStopRetainsOwnershipWhenAdapterRejectsIdentity(t *testing.T) {
	fixture := newManagerFixture(t)
	if err := fixture.manager.Start(context.Background(), fixture.networkDir, testExecutionImage); err != nil {
		t.Fatal(err)
	}
	fixture.client.DestroyError = errors.New("enclave identity changed")
	if err := fixture.manager.Stop(context.Background(), fixture.networkDir); err == nil ||
		!strings.Contains(err.Error(), "identity changed") {
		t.Fatalf("mismatch stop error = %v", err)
	}
	if fixture.client.Destroyed {
		t.Fatal("mismatched enclave was destroyed")
	}
	if _, err := loadOwnership(fixture.networkDir); err != nil {
		t.Fatalf("ownership was not retained: %v", err)
	}
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
			if err := fixture.manager.Start(context.Background(), fixture.networkDir, testExecutionImage); err != nil {
				t.Fatal(err)
			}
			arrange(fixture.client)
			if err := fixture.manager.Stop(context.Background(), fixture.networkDir); err != nil {
				t.Fatal(err)
			}
			if !fixture.client.Destroyed || countCalls(fixture.client.Calls, "destroy:") != 1 {
				t.Fatalf("destroyed=%t calls=%v", fixture.client.Destroyed, fixture.client.Calls)
			}
			if _, err := os.Stat(ownershipPath(fixture.networkDir)); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("ownership remains: %v", err)
			}
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
