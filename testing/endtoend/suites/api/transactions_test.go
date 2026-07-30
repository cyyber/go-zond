// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"encoding/json"

	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/rpc"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func (suite *liveSuite) assertTransactions(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()
	fixture := suite.fixture
	blockNumber := fixture.receipt.BlockNumber

	index := hexutil.Uint64(fixture.receipt.TransactionIndex)
	var countByNumber, countByHash hexutil.Uint
	gomega.Expect(raw.CallContext(
		ctx,
		&countByNumber,
		"qrl_getBlockTransactionCountByNumber",
		rpc.BlockNumber(blockNumber.Int64()),
	)).To(gomega.Succeed())
	gomega.Expect(raw.CallContext(
		ctx,
		&countByHash,
		"qrl_getBlockTransactionCountByHash",
		fixture.block.Hash(),
	)).To(gomega.Succeed())
	gomega.Expect(uint64(countByNumber)).To(gomega.BeNumerically(">", uint64(index)))
	gomega.Expect(countByHash).To(gomega.Equal(countByNumber))

	var transaction map[string]json.RawMessage
	gomega.Expect(raw.CallContext(
		ctx,
		&transaction,
		"qrl_getTransactionByBlockNumberAndIndex",
		rpc.BlockNumber(blockNumber.Int64()),
		index,
	)).To(gomega.Succeed())
	gomega.Expect(transaction).To(gomega.HaveKey("hash"))
	gomega.Expect(raw.CallContext(
		ctx,
		&transaction,
		"qrl_getTransactionByBlockHashAndIndex",
		fixture.block.Hash(),
		index,
	)).To(gomega.Succeed())

	var rawByNumber, rawByHash, rawByTransactionHash hexutil.Bytes
	gomega.Expect(raw.CallContext(
		ctx,
		&rawByNumber,
		"qrl_getRawTransactionByBlockNumberAndIndex",
		rpc.BlockNumber(blockNumber.Int64()),
		index,
	)).To(gomega.Succeed())
	gomega.Expect(raw.CallContext(
		ctx,
		&rawByHash,
		"qrl_getRawTransactionByBlockHashAndIndex",
		fixture.block.Hash(),
		index,
	)).To(gomega.Succeed())
	gomega.Expect(raw.CallContext(
		ctx,
		&rawByTransactionHash,
		"qrl_getRawTransactionByHash",
		fixture.tx.Hash(),
	)).To(gomega.Succeed())
	gomega.Expect(rawByNumber).To(gomega.Equal(rawByHash))
	gomega.Expect(rawByHash).To(gomega.Equal(rawByTransactionHash))

	wantRaw, err := fixture.tx.MarshalBinary()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(rawByTransactionHash).To(gomega.Equal(hexutil.Bytes(wantRaw)))

	found, pending, err := suite.client.TransactionByHash(ctx, fixture.tx.Hash())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(pending).To(gomega.BeFalse())
	gomega.Expect(found.Hash()).To(gomega.Equal(fixture.tx.Hash()))

	inBlock, err := suite.client.TransactionInBlock(
		ctx,
		fixture.block.Hash(),
		uint(fixture.receipt.TransactionIndex),
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(inBlock.Hash()).To(gomega.Equal(fixture.tx.Hash()))

	receipt, err := suite.client.TransactionReceipt(ctx, fixture.tx.Hash())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(receipt.TxHash).To(gomega.Equal(fixture.tx.Hash()))

	var receiptJSON map[string]json.RawMessage
	gomega.Expect(raw.CallContext(
		ctx,
		&receiptJSON,
		"qrl_getTransactionReceipt",
		fixture.tx.Hash(),
	)).To(gomega.Succeed())
	gomega.Expect(receiptJSON).To(gomega.HaveKey("logs"))

	var filled struct {
		Raw hexutil.Bytes      `json:"raw"`
		Tx  *types.Transaction `json:"tx"`
	}
	gomega.Expect(raw.CallContext(ctx, &filled, "qrl_fillTransaction", map[string]any{
		"from":  suite.from,
		"to":    suite.from,
		"value": "0x0",
	})).To(gomega.Succeed())
	gomega.Expect(filled.Raw).NotTo(gomega.BeEmpty())
	gomega.Expect(filled.Tx).NotTo(gomega.BeNil())

	var decoded types.Transaction
	gomega.Expect(decoded.UnmarshalBinary(filled.Raw)).To(gomega.Succeed())
	gomega.Expect(decoded.Hash()).To(gomega.Equal(filled.Tx.Hash()))
	gomega.Expect(decoded.To()).NotTo(gomega.BeNil())
	gomega.Expect(*decoded.To()).To(gomega.Equal(suite.from))
	gomega.Expect(decoded.Value().Sign()).To(gomega.Equal(0))
	gomega.Expect(decoded.ChainId()).To(gomega.Equal(suite.chainID))
}
