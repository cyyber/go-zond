// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/rpc"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

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
	gomega.Expect(badBlocks).To(gomega.BeEmpty())

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
