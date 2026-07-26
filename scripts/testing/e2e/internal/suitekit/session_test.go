// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package suitekit

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/network"
)

func TestOpenLiveNetwork(t *testing.T) {
	const rpcURL = "http://127.0.0.1:18545"
	networkDir, seedFile := liveNetworkDirectory(t)
	inspect := func(
		_ context.Context,
		requestedNetwork string,
	) (network.Environment, error) {
		if requestedNetwork != networkDir {
			t.Fatalf("inspect path = %q", requestedNetwork)
		}
		if competing, err := network.AcquireMutationLease(networkDir); err == nil {
			_ = competing.Close()
			t.Fatal("network lease was not held during inspection")
		}
		return network.Environment{
			RPCURL:       rpcURL,
			GraphQLURL:   rpcURL + "/graphql",
			WebSocketURL: "ws://127.0.0.1/ws",
			SeedFile:     seedFile,
		}, nil
	}

	live, err := openLiveNetwork(t.Context(), networkDir, inspect)
	if err != nil {
		t.Fatal(err)
	}
	if live.RPCURL != rpcURL ||
		live.GraphQLURL != rpcURL+"/graphql" ||
		live.WebSocketURL != "ws://127.0.0.1/ws" ||
		live.SeedFile != seedFile {
		t.Fatalf("live network = %+v", live)
	}
	if err := live.Close(); err != nil {
		t.Fatal(err)
	}
	if err := live.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := network.AcquireMutationLease(networkDir)
	if err != nil {
		t.Fatalf("session leaked network lease: %v", err)
	}
	_ = reopened.Close()
}

func TestOpenLiveNetworkReleasesLeaseOnFailure(t *testing.T) {
	networkDir, _ := liveNetworkDirectory(t)
	inspect := func(context.Context, string) (network.Environment, error) {
		return network.Environment{}, errors.New("network unavailable")
	}

	_, err := openLiveNetwork(t.Context(), networkDir, inspect)
	if err == nil || !strings.Contains(err.Error(), "inspect live network") {
		t.Fatalf("authentication error = %v", err)
	}
	reopened, err := network.AcquireMutationLease(networkDir)
	if err != nil {
		t.Fatalf("failure leaked network lease: %v", err)
	}
	_ = reopened.Close()
}

func liveNetworkDirectory(t *testing.T) (string, string) {
	t.Helper()
	networkDir := t.TempDir()
	if err := os.Chmod(networkDir, 0o700); err != nil {
		t.Fatal(err)
	}
	return networkDir, filepath.Join(networkDir, "wallet.seed")
}
