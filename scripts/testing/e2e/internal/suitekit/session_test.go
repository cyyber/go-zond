// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package suitekit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/network"
)

const testSeed = "010000f29f58aff0b00de2844f7e20bd9eeaacc379150043beeb328335817512b29fbb7184da84a092f842b2a06d72a24a5d28"

func TestOpenLiveSession(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	networkDir, seedFile := liveNetworkDirectory(t)
	values := map[string]string{networkDirVariable: networkDir}
	authenticate := func(
		_ context.Context,
		requestedNetwork string,
	) (network.Environment, error) {
		if requestedNetwork != networkDir {
			t.Fatalf("authenticate path = %q", requestedNetwork)
		}
		if competing, err := network.AcquireMutationLease(networkDir); err == nil {
			_ = competing.Close()
			t.Fatal("network lease was not held during authentication")
		}
		return network.Environment{
			RPCURL:       server.URL,
			GraphQLURL:   server.URL + "/graphql",
			WebSocketURL: "ws://127.0.0.1/ws",
			SeedFile:     seedFile,
		}, nil
	}

	session, err := openLiveSession(t.Context(), mapGetenv(values), authenticate)
	if err != nil {
		t.Fatal(err)
	}
	expectedWallet, err := wallet.RestoreFromSeedHex(testSeed)
	if err != nil {
		t.Fatal(err)
	}
	if session.Client == nil || session.Wallet == nil ||
		session.Sender != common.Address(expectedWallet.GetAddress()) {
		t.Fatalf("incomplete signing session: %+v", session)
	}
	if session.RPCURL != server.URL ||
		session.GraphQLURL != server.URL+"/graphql" ||
		session.WebSocketURL != "ws://127.0.0.1/ws" {
		t.Fatalf("session endpoints = %+v", session)
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := network.AcquireMutationLease(networkDir)
	if err != nil {
		t.Fatalf("session leaked network lease: %v", err)
	}
	_ = reopened.Close()
}

func TestOpenLiveSessionReleasesLeaseOnFailure(t *testing.T) {
	networkDir, _ := liveNetworkDirectory(t)
	values := map[string]string{networkDirVariable: networkDir}
	authenticate := func(context.Context, string) (network.Environment, error) {
		return network.Environment{}, errors.New("network unavailable")
	}

	_, err := openLiveSession(t.Context(), mapGetenv(values), authenticate)
	if err == nil || !strings.Contains(err.Error(), "authenticate live network") {
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
	if err := os.Mkdir(filepath.Join(networkDir, "private"), 0o700); err != nil {
		t.Fatal(err)
	}
	seedFile := filepath.Join(networkDir, "private", "wallet.seed")
	if err := os.WriteFile(seedFile, []byte(testSeed+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	canonical, err := filepath.EvalSymlinks(networkDir)
	if err != nil {
		t.Fatal(err)
	}
	return canonical, filepath.Join(canonical, "private", "wallet.seed")
}

func mapGetenv(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}
