// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/rlp"
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

	var rawHeader hexutil.Bytes
	gomega.Expect(raw.CallContext(
		ctx,
		&rawHeader,
		"debug_getRawHeader",
		blockSelector,
	)).To(gomega.Succeed())
	wantHeader, err := rlp.EncodeToBytes(fixture.block.Header())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(rawHeader).To(gomega.Equal(hexutil.Bytes(wantHeader)))

	var rawBlock hexutil.Bytes
	gomega.Expect(raw.CallContext(
		ctx,
		&rawBlock,
		"debug_getRawBlock",
		blockSelector,
	)).To(gomega.Succeed())
	wantBlock, err := rlp.EncodeToBytes(fixture.block)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(rawBlock).To(gomega.Equal(hexutil.Bytes(wantBlock)))

	var rawTransaction hexutil.Bytes
	gomega.Expect(raw.CallContext(
		ctx,
		&rawTransaction,
		"debug_getRawTransaction",
		fixture.tx.Hash(),
	)).To(gomega.Succeed())
	wantTransaction, err := fixture.tx.MarshalBinary()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(rawTransaction).To(gomega.Equal(hexutil.Bytes(wantTransaction)))

	var rawReceipts []hexutil.Bytes
	gomega.Expect(raw.CallContext(
		ctx,
		&rawReceipts,
		"debug_getRawReceipts",
		blockSelector,
	)).To(gomega.Succeed())
	gomega.Expect(rawReceipts).To(gomega.HaveLen(len(fixture.block.Transactions())))
	wantReceipt, err := fixture.receipt.MarshalBinary()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(rawReceipts[int(fixture.receipt.TransactionIndex)]).To(
		gomega.Equal(hexutil.Bytes(wantReceipt)),
	)

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
