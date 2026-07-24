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
	request    StartRequest
	networkDir string
}

func newManagerFixture(t *testing.T) managerFixture {
	t.Helper()
	networkDir, err := ensureNetworkDirectory(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	client := &fakeKurtosis{
		Enclave:     kurtosis.EnclaveRef{UUID: strings.Repeat("a", 32)},
		ServiceList: topologyFixture(),
	}
	manager := NewManager()
	manager.newClient = func() (kurtosisClient, error) { return client, nil }
	manager.probe = func(context.Context, probeRequest) error { return nil }
	return managerFixture{
		manager: manager, client: client, networkDir: networkDir,
		request: StartRequest{
			NetworkDir: networkDir, ExecutionImage: "local/go-qrl:test",
			StartTimeout: time.Minute,
		},
	}
}

func TestStartPersistsExactOwnershipAndRefusesStaleReuse(t *testing.T) {
	fixture := newManagerFixture(t)
	if err := fixture.manager.Start(context.Background(), fixture.request); err != nil {
		t.Fatal(err)
	}
	ownership, err := loadOwnership(fixture.networkDir)
	if err != nil || ownership.UUID != fixture.client.Enclave.UUID {
		t.Fatalf("ownership=%+v err=%v", ownership, err)
	}
	if _, err := os.Stat(filepath.Join(fixture.networkDir, "network.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("public network state exists: %v", err)
	}
	if err := fixture.manager.Start(context.Background(), fixture.request); err == nil ||
		!strings.Contains(err.Error(), "already exists") {
		t.Fatalf("stale-reuse error = %v", err)
	}
	environment, err := fixture.manager.Authenticate(context.Background(), fixture.networkDir)
	if err != nil || environment.RPCURL != "http://127.0.0.1:18545" {
		t.Fatalf("environment=%+v err=%v", environment, err)
	}
	if countCalls(fixture.client.Calls, "create:") != 1 ||
		countCalls(fixture.client.Calls, "run:") != 1 {
		t.Fatalf("network was reprovisioned: %v", fixture.client.Calls)
	}
	if len(fixture.client.Runs) != 1 ||
		!strings.Contains(fixture.client.Runs[0].SerializedParams, `"el_image":"local/go-qrl:test"`) {
		t.Fatalf("package runs = %+v", fixture.client.Runs)
	}
	if countCalls(fixture.client.Calls, "get:") != 0 {
		t.Fatalf("manager repeated adapter identity lookups: %v", fixture.client.Calls)
	}
}

func TestStatusFailsWhenNetworkIsNotRunning(t *testing.T) {
	fixture := newManagerFixture(t)
	if err := fixture.manager.Status(context.Background(), fixture.networkDir); err == nil ||
		!strings.Contains(err.Error(), "not running") {
		t.Fatalf("status error = %v", err)
	}
}

func TestFailedProvisioningRetainsExactOwnershipUntilStop(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.RunError = errors.New("package failed")
	if err := fixture.manager.Start(context.Background(), fixture.request); err == nil {
		t.Fatal("failed provisioning succeeded")
	}
	ownership, err := loadOwnership(fixture.networkDir)
	if err != nil || ownership.UUID != fixture.client.Enclave.UUID {
		t.Fatalf("retained ownership=%+v err=%v", ownership, err)
	}
	fixture.client.RunError = nil
	if err := fixture.manager.Start(context.Background(), fixture.request); err == nil ||
		!strings.Contains(err.Error(), "network-stop") {
		t.Fatalf("incomplete restart error = %v", err)
	}
	if err := fixture.manager.Stop(context.Background(), fixture.networkDir); err != nil {
		t.Fatal(err)
	}
	if !fixture.client.Destroyed {
		t.Fatal("owned enclave was not destroyed")
	}
}

func TestAmbiguousCreateRecoversAndPersistsExactOwnership(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.CreateError = errors.New("create response lost")
	fixture.client.CreateAfterError = true
	if err := fixture.manager.Start(context.Background(), fixture.request); err == nil ||
		!strings.Contains(err.Error(), "recovered exact ownership") {
		t.Fatalf("ambiguous create error = %v", err)
	}
	record, err := loadOwnership(fixture.networkDir)
	if err != nil || record.UUID != fixture.client.Enclave.UUID {
		t.Fatalf("recovered ownership=%+v err=%v", record, err)
	}
	fixture.client.CreateError = nil
	if err := fixture.manager.Start(context.Background(), fixture.request); err == nil ||
		!strings.Contains(err.Error(), "already exists") {
		t.Fatalf("ambiguous restart error = %v", err)
	}
	if err := fixture.manager.Stop(context.Background(), fixture.networkDir); err != nil {
		t.Fatal(err)
	}
	if countCalls(fixture.client.Calls, "create:") != 1 || !fixture.client.Destroyed {
		t.Fatalf("unsafe ambiguous handling: %v", fixture.client.Calls)
	}
}

func TestUnresolvedCreateErrorLeavesNoPartialOwnership(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.CreateError = errors.New("create failed")
	if err := fixture.manager.Start(context.Background(), fixture.request); err == nil ||
		!strings.Contains(err.Error(), "recover ambiguous creation") {
		t.Fatalf("create error = %v", err)
	}
	if _, err := os.Stat(ownershipPath(fixture.networkDir)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("partial ownership remains: %v", err)
	}
}

type identityRefusingClient struct{ *fakeKurtosis }

func (client identityRefusingClient) DestroyEnclave(context.Context, kurtosis.EnclaveRef) error {
	return errors.New("enclave identity changed")
}

func TestStopRetainsOwnershipWhenAdapterRejectsIdentity(t *testing.T) {
	fixture := newManagerFixture(t)
	if err := fixture.manager.Start(context.Background(), fixture.request); err != nil {
		t.Fatal(err)
	}
	fixture.manager.newClient = func() (kurtosisClient, error) {
		return identityRefusingClient{fixture.client}, nil
	}
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

func TestStopReconcilesLostDestroyResponseAndRemovesOwnership(t *testing.T) {
	fixture := newManagerFixture(t)
	if err := fixture.manager.Start(context.Background(), fixture.request); err != nil {
		t.Fatal(err)
	}
	fixture.client.DestroyError = errors.New("destroy response lost")
	fixture.client.DestroyAfterError = true
	if err := fixture.manager.Stop(context.Background(), fixture.networkDir); err != nil {
		t.Fatal(err)
	}
	if !fixture.client.Destroyed || countCalls(fixture.client.Calls, "destroy:") != 1 {
		t.Fatalf("destroyed=%t calls=%v", fixture.client.Destroyed, fixture.client.Calls)
	}
	if _, err := os.Stat(ownershipPath(fixture.networkDir)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("ownership remains: %v", err)
	}
}

type alreadyAbsentClient struct{ *fakeKurtosis }

func (client alreadyAbsentClient) DestroyEnclave(_ context.Context, ref kurtosis.EnclaveRef) error {
	client.Calls = append(client.Calls, "destroy:"+ref.UUID)
	return errors.New("enclave not found")
}

func TestStopRemovesOwnershipWhenEnclaveIsAlreadyAbsent(t *testing.T) {
	fixture := newManagerFixture(t)
	if err := fixture.manager.Start(context.Background(), fixture.request); err != nil {
		t.Fatal(err)
	}
	fixture.client.Destroyed = true
	fixture.manager.newClient = func() (kurtosisClient, error) {
		return alreadyAbsentClient{fixture.client}, nil
	}
	if err := fixture.manager.Stop(context.Background(), fixture.networkDir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ownershipPath(fixture.networkDir)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("ownership remains: %v", err)
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
