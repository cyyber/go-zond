// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, server.RegisterName("qrl", service))
	t.Cleanup(server.Stop)
	return httptest.NewServer(server)
}

func TestProbeNetworkRequiresAdvancingFundedChain(t *testing.T) {
	service := &probeService{chainID: "0x539", balance: "0x1"}
	server := newProbeServer(t, service)
	defer server.Close()
	address := "Q" + strings.Repeat("b", 128)
	require.NoError(t, probeNetwork(context.Background(), server.URL, address))
	service.mu.Lock()
	defer service.mu.Unlock()
	require.GreaterOrEqual(t, service.blockCalls, 2)
	require.True(t, strings.EqualFold(service.address, address))
}

func TestProbeNetworkRejectsWrongChainAndEmptyWallet(t *testing.T) {
	for name, service := range map[string]*probeService{
		"chain":   {chainID: "0x540", balance: "0x1"},
		"balance": {chainID: "0x539", balance: "0x0"},
	} {
		t.Run(name, func(t *testing.T) {
			server := newProbeServer(t, service)
			defer server.Close()
			err := probeNetwork(
				context.Background(),
				server.URL,
				"Q"+strings.Repeat("c", 128),
			)
			require.Error(t, err)
		})
	}
}

func TestProbeNetworkHonorsCallerDeadline(t *testing.T) {
	service := &probeService{chainID: "0x539", balance: "0x1", freezeBlocks: true}
	server := newProbeServer(t, service)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err := probeNetwork(ctx, server.URL, "Q"+strings.Repeat("d", 128))
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.ErrorContains(t, err, "chain advancement probe interrupted")
	require.ErrorContains(t, err, "block number remains at 1")
}
