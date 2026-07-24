// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

type managerFixture struct {
	manager    *Manager
	client     *kurtosis.Fake
	request    StartRequest
	networkDir string
}

func newManagerFixture(t *testing.T) managerFixture {
	t.Helper()
	networkDir, err := ensureNetworkDirectory(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	client := &kurtosis.Fake{
		Enclave: kurtosis.EnclaveRef{
			Name: defaultEnclaveName(networkDir), UUID: strings.Repeat("a", 32), Owned: true,
		},
		ServiceList: topologyFixture(),
	}
	manager := NewManager()
	manager.NewClient = func() (kurtosis.Client, error) { return client, nil }
	manager.Stdout, manager.Stderr = io.Discard, io.Discard
	manager.Probe = func(context.Context, probeRequest) error { return nil }
	manager.Prepare = func(
		context.Context, commandRunner, StartRequest, string, io.Writer, io.Writer,
	) (preparedNetwork, error) {
		return preparedNetwork{Params: `{"network":"test"}`}, nil
	}
	return managerFixture{
		manager: manager, client: client, networkDir: networkDir,
		request: StartRequest{
			RepoRoot: t.TempDir(), NetworkDir: networkDir,
			StartTimeout: time.Minute,
		},
	}
}

func TestStartPublishesReadyStateAndRefusesStaleReuse(t *testing.T) {
	fixture := newManagerFixture(t)
	first, err := fixture.manager.Start(context.Background(), fixture.request)
	if err != nil || !first.Ready {
		t.Fatalf("start=%+v err=%v", first, err)
	}
	ownership, err := loadOwnership(fixture.networkDir)
	if err != nil || ownership.UUID != fixture.client.Enclave.UUID {
		t.Fatalf("ownership=%+v err=%v", ownership, err)
	}
	state, err := loadState(fixture.networkDir)
	if err != nil || !state.Ready {
		t.Fatalf("state=%+v err=%v", state, err)
	}
	if _, err := fixture.manager.Start(context.Background(), fixture.request); err == nil ||
		!strings.Contains(err.Error(), "already running") {
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
}

func TestFailedProvisioningRetainsExactOwnershipUntilStop(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.RunError = errors.New("package failed")
	if _, err := fixture.manager.Start(context.Background(), fixture.request); err == nil {
		t.Fatal("failed provisioning succeeded")
	}
	ownership, err := loadOwnership(fixture.networkDir)
	if err != nil || ownership.UUID != fixture.client.Enclave.UUID {
		t.Fatalf("retained ownership=%+v err=%v", ownership, err)
	}
	fixture.client.RunError = nil
	if _, err := fixture.manager.Start(context.Background(), fixture.request); err == nil ||
		!strings.Contains(err.Error(), "network-stop") {
		t.Fatalf("incomplete restart error = %v", err)
	}
	if _, err := fixture.manager.Stop(context.Background(), fixture.networkDir); err != nil {
		t.Fatal(err)
	}
	if !fixture.client.Destroyed {
		t.Fatal("owned enclave was not destroyed")
	}
}

func TestAmbiguousCreateIsNeverReplayedOrDestroyedByName(t *testing.T) {
	fixture := newManagerFixture(t)
	fixture.client.CreateError = errors.New("create response lost")
	fixture.client.CreateAfterError = true
	if _, err := fixture.manager.Start(context.Background(), fixture.request); err == nil {
		t.Fatal("ambiguous create succeeded")
	}
	record, err := loadOwnership(fixture.networkDir)
	if err != nil || record.UUID != "" {
		t.Fatalf("creation intent=%+v err=%v", record, err)
	}
	fixture.client.CreateError = nil
	if _, err := fixture.manager.Start(context.Background(), fixture.request); err == nil ||
		!strings.Contains(err.Error(), "exact UUID") {
		t.Fatalf("ambiguous restart error = %v", err)
	}
	if _, err := fixture.manager.Stop(context.Background(), fixture.networkDir); err == nil ||
		!strings.Contains(err.Error(), "exact UUID") {
		t.Fatalf("ambiguous stop error = %v", err)
	}
	if countCalls(fixture.client.Calls, "create:") != 1 || fixture.client.Destroyed {
		t.Fatalf("unsafe ambiguous handling: %v", fixture.client.Calls)
	}
}

type enclaveDriftClient struct{ *kurtosis.Fake }

func (client enclaveDriftClient) GetEnclave(context.Context, string) (kurtosis.EnclaveRef, error) {
	enclave := client.Enclave
	enclave.Name += "-other"
	return enclave, nil
}

func TestStopRefusesEnclaveIdentityMismatch(t *testing.T) {
	fixture := newManagerFixture(t)
	if _, err := fixture.manager.Start(context.Background(), fixture.request); err != nil {
		t.Fatal(err)
	}
	fixture.manager.NewClient = func() (kurtosis.Client, error) {
		return enclaveDriftClient{fixture.client}, nil
	}
	if _, err := fixture.manager.Stop(context.Background(), fixture.networkDir); err == nil ||
		!strings.Contains(err.Error(), "name/UUID changed") {
		t.Fatalf("mismatch stop error = %v", err)
	}
	if fixture.client.Destroyed {
		t.Fatal("mismatched enclave was destroyed")
	}
}

func TestStopReconcilesLostDestroyResponseAndIgnoresBadReadyMarker(t *testing.T) {
	fixture := newManagerFixture(t)
	if _, err := fixture.manager.Start(context.Background(), fixture.request); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statePath(fixture.networkDir), []byte("{bad"), 0o600); err != nil {
		t.Fatal(err)
	}
	fixture.client.DestroyError = errors.New("destroy response lost")
	fixture.client.DestroyAfterError = true
	if _, err := fixture.manager.Stop(context.Background(), fixture.networkDir); err != nil {
		t.Fatal(err)
	}
	if !fixture.client.Destroyed || countCalls(fixture.client.Calls, "destroy:") != 1 {
		t.Fatalf("destroyed=%t calls=%v", fixture.client.Destroyed, fixture.client.Calls)
	}
	for _, path := range []string{statePath(fixture.networkDir), ownershipPath(fixture.networkDir)} {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("%s remains: %v", path, err)
		}
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
