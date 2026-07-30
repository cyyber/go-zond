// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/rpc"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func (suite *liveSuite) assertRawDebug(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()
	fixture := suite.fixture
	blockNumber := rpc.BlockNumber(fixture.block.NumberU64())
	blockSelector := rpc.BlockNumberOrHashWithNumber(blockNumber)

	var rawHeader, rawBlock, rawTransaction hexutil.Bytes
	gomega.Expect(raw.CallContext(
		ctx,
		&rawHeader,
		"debug_getRawHeader",
		blockSelector,
	)).To(gomega.Succeed())
	gomega.Expect(rawHeader).NotTo(gomega.BeEmpty())
	gomega.Expect(raw.CallContext(
		ctx,
		&rawBlock,
		"debug_getRawBlock",
		blockSelector,
	)).To(gomega.Succeed())
	gomega.Expect(rawBlock).NotTo(gomega.BeEmpty())
	gomega.Expect(raw.CallContext(
		ctx,
		&rawTransaction,
		"debug_getRawTransaction",
		fixture.tx.Hash(),
	)).To(gomega.Succeed())
	gomega.Expect(rawTransaction).NotTo(gomega.BeEmpty())

	var rawReceipts []hexutil.Bytes
	gomega.Expect(raw.CallContext(
		ctx,
		&rawReceipts,
		"debug_getRawReceipts",
		blockSelector,
	)).To(gomega.Succeed())
	gomega.Expect(rawReceipts).NotTo(gomega.BeEmpty())

	var printed string
	gomega.Expect(raw.CallContext(
		ctx,
		&printed,
		"debug_printBlock",
		fixture.block.NumberU64(),
	)).To(gomega.Succeed())
	gomega.Expect(printed).NotTo(gomega.BeEmpty())

	var headHash hexutil.Bytes
	gomega.Expect(raw.CallContext(ctx, &headHash, "debug_dbGet", "LastBlock")).To(gomega.Succeed())
	gomega.Expect(headHash).To(gomega.HaveLen(common.HashLength))

	var ancientCount uint64
	gomega.Expect(raw.CallContext(ctx, &ancientCount, "debug_dbAncients")).To(gomega.Succeed())
	var ancientHash hexutil.Bytes
	if ancientCount > 0 {
		gomega.Expect(raw.CallContext(
			ctx,
			&ancientHash,
			"debug_dbAncient",
			"hashes",
			ancientCount-1,
		)).To(gomega.Succeed())
		gomega.Expect(ancientHash).To(gomega.HaveLen(common.HashLength))
	} else {
		expectRegisteredError(raw.CallContext(
			ctx,
			&ancientHash,
			"debug_dbAncient",
			"hashes",
			0,
		))
	}
}

func (suite *liveSuite) assertDebugState(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()
	fixture := suite.fixture
	blockNumber := rpc.BlockNumber(fixture.block.NumberU64())
	blockSelector := rpc.BlockNumberOrHashWithNumber(blockNumber)

	var dump json.RawMessage
	gomega.Expect(raw.CallContext(ctx, &dump, "debug_dumpBlock", blockNumber)).To(gomega.Succeed())
	gomega.Expect(json.Valid(dump)).To(gomega.BeTrue())

	var badBlocks []json.RawMessage
	gomega.Expect(raw.CallContext(ctx, &badBlocks, "debug_getBadBlocks")).To(gomega.Succeed())

	var accountRange json.RawMessage
	gomega.Expect(raw.CallContext(
		ctx,
		&accountRange,
		"debug_accountRange",
		blockSelector,
		hexutil.Bytes{},
		1,
		true,
		true,
		false,
	)).To(gomega.Succeed())
	gomega.Expect(json.Valid(accountRange)).To(gomega.BeTrue())

	callTx, err := suite.signTransaction(ctx, &fixture.address, new(big.Int), nil)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	callReceipt := suite.submitAndWait(ctx, callTx)
	gomega.Expect(callReceipt.Status).To(gomega.Equal(types.ReceiptStatusSuccessful))

	var storageRange struct {
		Storage map[common.Hash]struct {
			Key   *common.Hash          `json:"key"`
			Value common.StorageValue64 `json:"value"`
		} `json:"storage"`
	}
	gomega.Expect(raw.CallContext(
		ctx,
		&storageRange,
		"debug_storageRangeAt",
		rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(callReceipt.BlockNumber.Int64())),
		int(callReceipt.TransactionIndex),
		fixture.address,
		hexutil.Bytes{},
		10,
	)).To(gomega.Succeed())
	gomega.Expect(storageRange.Storage).NotTo(gomega.BeEmpty())

	var modifiedByNumber, modifiedByHash []common.Address
	expectRegisteredError(raw.CallContext(
		ctx,
		&modifiedByNumber,
		"debug_getModifiedAccountsByNumber",
		callReceipt.BlockNumber.Uint64(),
	))
	expectRegisteredError(raw.CallContext(
		ctx,
		&modifiedByHash,
		"debug_getModifiedAccountsByHash",
		callReceipt.BlockHash,
	))
}

func (suite *liveSuite) assertDebugTracing(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()
	fixture := suite.fixture
	blockNumber := rpc.BlockNumber(fixture.block.NumberU64())
	blockSelector := rpc.BlockNumberOrHashWithNumber(blockNumber)

	var trace json.RawMessage
	gomega.Expect(raw.CallContext(
		ctx,
		&trace,
		"debug_traceTransaction",
		fixture.tx.Hash(),
		map[string]any{},
	)).To(gomega.Succeed())
	gomega.Expect(json.Valid(trace)).To(gomega.BeTrue())

	var blockTrace []json.RawMessage
	gomega.Expect(raw.CallContext(
		ctx,
		&blockTrace,
		"debug_traceBlockByNumber",
		blockNumber,
		map[string]any{},
	)).To(gomega.Succeed())
	gomega.Expect(blockTrace).NotTo(gomega.BeEmpty())

	var rawBlock hexutil.Bytes
	gomega.Expect(raw.CallContext(
		ctx,
		&rawBlock,
		"debug_getRawBlock",
		blockSelector,
	)).To(gomega.Succeed())
	gomega.Expect(raw.CallContext(
		ctx,
		&blockTrace,
		"debug_traceBlockByHash",
		fixture.block.Hash(),
		map[string]any{},
	)).To(gomega.Succeed())
	gomega.Expect(raw.CallContext(
		ctx,
		&blockTrace,
		"debug_traceBlock",
		rawBlock,
		map[string]any{},
	)).To(gomega.Succeed())

	gomega.Expect(blockNumber).To(gomega.BeNumerically(">", 0))
	traceResults := make(chan json.RawMessage, 1)
	traceSub, err := suite.wsClient.Client().Subscribe(
		ctx,
		"debug",
		traceResults,
		"traceChain",
		blockNumber-1,
		blockNumber,
		map[string]any{},
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer traceSub.Unsubscribe()

	select {
	case result := <-traceResults:
		gomega.Expect(json.Valid(result)).To(gomega.BeTrue())
	case err := <-traceSub.Err():
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	case <-time.After(30 * time.Second):
		ginkgo.Fail("timed out waiting for debug_traceChain result")
	case <-ctx.Done():
		ginkgo.Fail("trace context ended: " + ctx.Err().Error())
	}

	gomega.Expect(raw.CallContext(ctx, &trace, "debug_traceCall", map[string]any{
		"from": suite.from,
		"to":   fixture.address,
		"data": "0x",
	}, blockSelector, map[string]any{})).To(gomega.Succeed())
	gomega.Expect(json.Valid(trace)).To(gomega.BeTrue())

	var roots []common.Hash
	gomega.Expect(raw.CallContext(
		ctx,
		&roots,
		"debug_intermediateRoots",
		fixture.block.Hash(),
		map[string]any{},
	)).To(gomega.Succeed())
	gomega.Expect(roots).To(gomega.HaveLen(len(fixture.block.Transactions())))
}

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
