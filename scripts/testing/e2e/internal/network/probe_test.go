// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/theQRL/go-qrl/rpc"
)

type probeService struct {
	mu           sync.Mutex
	chainID      string
	balance      string
	freezeBlocks bool
	blockCalls   int
	address      string
}

func (service *probeService) ChainId() string { return service.chainID }

func (service *probeService) BlockNumber() string {
	service.mu.Lock()
	defer service.mu.Unlock()
	service.blockCalls++
	if service.freezeBlocks || service.blockCalls == 1 {
		return "0x1"
	}
	return "0x2"
}

func (service *probeService) GetBalance(address, _ string) string {
	service.mu.Lock()
	defer service.mu.Unlock()
	service.address = address
	return service.balance
}

func newProbeServer(t *testing.T, service *probeService) *httptest.Server {
	t.Helper()
	server := rpc.NewServer()
	if err := server.RegisterName("qrl", service); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(server.Stop)
	return httptest.NewServer(server)
}

func TestProbeNetworkRequiresAdvancingFundedChain(t *testing.T) {
	service := &probeService{chainID: "0x539", balance: "0x1"}
	server := newProbeServer(t, service)
	defer server.Close()
	address := "Q" + strings.Repeat("b", 128)
	if err := probeNetwork(context.Background(), probeRequest{
		RPCURL: server.URL, Address: address, ExpectedChainID: "0x539",
	}); err != nil {
		t.Fatal(err)
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	if service.blockCalls < 2 || !strings.EqualFold(service.address, address) {
		t.Fatalf("block calls=%d address=%q", service.blockCalls, service.address)
	}
}

func TestProbeNetworkRejectsWrongChainAndEmptyWallet(t *testing.T) {
	for name, service := range map[string]*probeService{
		"chain":   {chainID: "0x540", balance: "0x1"},
		"balance": {chainID: "0x539", balance: "0x0"},
	} {
		t.Run(name, func(t *testing.T) {
			server := newProbeServer(t, service)
			defer server.Close()
			err := probeNetwork(context.Background(), probeRequest{
				RPCURL: server.URL, Address: "Q" + strings.Repeat("c", 128),
				ExpectedChainID: "0x539",
			})
			if err == nil {
				t.Fatal("invalid network was accepted")
			}
		})
	}
}

func TestProbeNetworkHonorsCallerDeadline(t *testing.T) {
	service := &probeService{chainID: "0x539", balance: "0x1", freezeBlocks: true}
	server := newProbeServer(t, service)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err := probeNetwork(ctx, probeRequest{
		RPCURL: server.URL, Address: "Q" + strings.Repeat("d", 128),
		ExpectedChainID: "0x539",
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline error = %v", err)
	}
}
