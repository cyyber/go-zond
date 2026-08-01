// Copyright 2021 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package qrlstats

import (
	"context"
	"math/big"
	"strconv"
	"testing"

	qrl "github.com/theQRL/go-qrl"
	"github.com/theQRL/go-qrl/consensus/beacon"
	"github.com/theQRL/go-qrl/core"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/event"
	qrlbackend "github.com/theQRL/go-qrl/qrl"
	"github.com/theQRL/go-qrl/rpc"
)

type blockReportingBackend struct {
	head      *types.Header
	block     *types.Block
	requested rpc.BlockNumber
}

var _ fullNodeBackend = (*blockReportingBackend)(nil)
var _ fullNodeBackend = (*qrlbackend.QRLAPIBackend)(nil)

func (*blockReportingBackend) SubscribeChainHeadEvent(chan<- core.ChainHeadEvent) event.Subscription {
	return nil
}

func (*blockReportingBackend) SubscribeNewTxsEvent(chan<- core.NewTxsEvent) event.Subscription {
	return nil
}

func (b *blockReportingBackend) CurrentHeader() *types.Header {
	return b.head
}

func (b *blockReportingBackend) HeaderByNumber(context.Context, rpc.BlockNumber) (*types.Header, error) {
	return b.head, nil
}

func (*blockReportingBackend) Stats() (int, int) {
	return 0, 0
}

func (*blockReportingBackend) SyncProgress() qrl.SyncProgress {
	return qrl.SyncProgress{}
}

func (b *blockReportingBackend) BlockByNumber(_ context.Context, number rpc.BlockNumber) (*types.Block, error) {
	b.requested = number
	return b.block, nil
}

func (b *blockReportingBackend) CurrentBlock() *types.Header {
	return b.head
}

func (*blockReportingBackend) SuggestGasTipCap(context.Context) (*big.Int, error) {
	return new(big.Int), nil
}

func TestParseQRLstatsURL(t *testing.T) {
	cases := []struct {
		url              string
		node, pass, host string
	}{
		{
			url:  `"debug meowsbits":mypass@ws://mordor.dash.fault.dev:3000`,
			node: "debug meowsbits", pass: "mypass", host: "ws://mordor.dash.fault.dev:3000",
		},
		{
			url:  `"debug @meowsbits":mypass@ws://mordor.dash.fault.dev:3000`,
			node: "debug @meowsbits", pass: "mypass", host: "ws://mordor.dash.fault.dev:3000",
		},
		{
			url:  `"debug: @meowsbits":mypass@ws://mordor.dash.fault.dev:3000`,
			node: "debug: @meowsbits", pass: "mypass", host: "ws://mordor.dash.fault.dev:3000",
		},
		{
			url:  `name:@ws://mordor.dash.fault.dev:3000`,
			node: "name", pass: "", host: "ws://mordor.dash.fault.dev:3000",
		},
		{
			url:  `name@ws://mordor.dash.fault.dev:3000`,
			node: "name", pass: "", host: "ws://mordor.dash.fault.dev:3000",
		},
		{
			url:  `:mypass@ws://mordor.dash.fault.dev:3000`,
			node: "", pass: "mypass", host: "ws://mordor.dash.fault.dev:3000",
		},
		{
			url:  `:@ws://mordor.dash.fault.dev:3000`,
			node: "", pass: "", host: "ws://mordor.dash.fault.dev:3000",
		},
	}

	for i, c := range cases {
		parts, err := parseQRLstatsURL(c.url)
		if err != nil {
			t.Fatal(err)
		}
		node, pass, host := parts[0], parts[1], parts[2]

		// unquote because the value provided will be used as a CLI flag value, so unescaped quotes will be removed
		nodeUnquote, err := strconv.Unquote(node)
		if err == nil {
			node = nodeUnquote
		}

		if node != c.node {
			t.Errorf("case=%d mismatch node value, got: %v ,want: %v", i, node, c.node)
		}
		if pass != c.pass {
			t.Errorf("case=%d mismatch pass value, got: %v ,want: %v", i, pass, c.pass)
		}
		if host != c.host {
			t.Errorf("case=%d mismatch host value, got: %v ,want: %v", i, host, c.host)
		}
	}
}

func TestAssembleBlockStatsFetchesFullBlock(t *testing.T) {
	header := &types.Header{Number: big.NewInt(7), GasLimit: 30_000_000, Time: 1}
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(1),
		GasTipCap: big.NewInt(1),
		GasFeeCap: big.NewInt(1),
		Value:     new(big.Int),
	})
	block := types.NewBlockWithHeader(header).WithBody(types.Body{Transactions: types.Transactions{tx}})
	backend := &blockReportingBackend{head: header, block: block}
	service := &Service{backend: backend, engine: beacon.NewFaker()}

	stats := service.assembleBlockStats(nil)
	if stats == nil {
		t.Fatal("block stats are nil")
	}
	if backend.requested != rpc.BlockNumber(7) {
		t.Fatalf("requested block mismatch: have %d, want 7", backend.requested)
	}
	if len(stats.Txs) != 1 || stats.Txs[0].Hash != tx.Hash() {
		t.Fatalf("transaction stats mismatch: have %v, want %s", stats.Txs, tx.Hash())
	}
}

func TestAssembleBlockStatsMissingHeadBlock(t *testing.T) {
	header := &types.Header{Number: big.NewInt(7)}
	backend := &blockReportingBackend{head: header}
	service := &Service{backend: backend, engine: beacon.NewFaker()}

	if stats := service.assembleBlockStats(nil); stats != nil {
		t.Fatalf("unexpected block stats: %#v", stats)
	}
}
