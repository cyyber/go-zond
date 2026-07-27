// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/rpc"
)

type probeService struct {
	balance    string
	blockCalls int
	address    string
}

func (service *probeService) BlockNumber() string {
	service.blockCalls++
	if service.blockCalls == 1 {
		return "0x1"
	}
	return "0x2"
}

func (service *probeService) GetBalance(address, _ string) string {
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
	service := &probeService{balance: "0x1"}
	server := newProbeServer(t, service)
	defer server.Close()
	address := common.MustParseAddress("Q" + strings.Repeat("b", 128))
	require.NoError(t, probeNetwork(context.Background(), server.URL, address))
	require.True(t, strings.EqualFold(service.address, address.Hex()))
}

func TestProbeNetworkRejectsEmptyWallet(t *testing.T) {
	server := newProbeServer(t, &probeService{balance: "0x0"})
	defer server.Close()
	err := probeNetwork(
		context.Background(),
		server.URL,
		common.MustParseAddress("Q"+strings.Repeat("c", 128)),
	)
	require.ErrorContains(t, err, "has no balance")
}
