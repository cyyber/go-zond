// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/rpc"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func (suite *liveSuite) assertDebugErrorPaths(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()
	blockNumber := rpc.BlockNumber(suite.fixture.block.NumberU64())

	var property string
	if err := raw.CallContext(ctx, &property, "debug_chaindbProperty", ""); err != nil {
		expectRegisteredError(err)
	}
	var accessible hexutil.Uint64
	if err := raw.CallContext(
		ctx,
		&accessible,
		"debug_getAccessibleState",
		blockNumber,
		blockNumber-1,
	); err != nil {
		expectRegisteredError(err)
	}
	var flushInterval string
	if err := raw.CallContext(ctx, &flushInterval, "debug_getTrieFlushInterval"); err != nil {
		expectRegisteredError(err)
	}

	for _, test := range []struct {
		method     string
		parameters []any
	}{
		{"debug_preimage", []any{common.Hash{}}},
		{"debug_traceBadBlock", []any{common.Hash{}, map[string]any{}}},
		{"debug_traceBlockFromFile", []any{"/path/that/does/not/exist", map[string]any{}}},
		{"debug_standardTraceBadBlockToFile", []any{common.Hash{}, map[string]any{}}},
		{"debug_setTrieFlushInterval", []any{"not-a-duration"}},
		{"admin_addPeer", []any{"not-a-qnode"}},
		{"admin_removePeer", []any{"not-a-qnode"}},
		{"admin_addTrustedPeer", []any{"not-a-qnode"}},
		{"admin_removeTrustedPeer", []any{"not-a-qnode"}},
		{"admin_exportChain", []any{"/"}},
		{"admin_importChain", []any{"/path/that/does/not/exist"}},
	} {
		var result json.RawMessage
		expectRegisteredError(raw.CallContext(ctx, &result, test.method, test.parameters...))
	}
}

func expectRegisteredError(err error) {
	ginkgo.GinkgoHelper()

	gomega.Expect(err).To(gomega.HaveOccurred())
	var rpcError rpc.Error
	if errors.As(err, &rpcError) {
		gomega.Expect(rpcError.ErrorCode()).NotTo(
			gomega.Equal(-32601),
			fmt.Sprintf("RPC method was not registered: %v", err),
		)
	}
}
