// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"math/big"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/internal/qrlapi"
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

	var byNumber, byHash, byTransactionHash qrlapi.RPCTransaction
	gomega.Expect(raw.CallContext(
		ctx,
		&byNumber,
		"qrl_getTransactionByBlockNumberAndIndex",
		rpc.BlockNumber(blockNumber.Int64()),
		index,
	)).To(gomega.Succeed())
	gomega.Expect(raw.CallContext(
		ctx,
		&byHash,
		"qrl_getTransactionByBlockHashAndIndex",
		fixture.block.Hash(),
		index,
	)).To(gomega.Succeed())
	gomega.Expect(raw.CallContext(
		ctx,
		&byTransactionHash,
		"qrl_getTransactionByHash",
		fixture.tx.Hash(),
	)).To(gomega.Succeed())
	for _, transaction := range []qrlapi.RPCTransaction{byNumber, byHash, byTransactionHash} {
		expectRPCTransaction(transaction, fixture.tx, fixture.receipt, fixture.block, suite.chainID)
	}

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

	var receiptJSON struct {
		BlockHash         common.Hash     `json:"blockHash"`
		BlockNumber       hexutil.Uint64  `json:"blockNumber"`
		TransactionHash   common.Hash     `json:"transactionHash"`
		TransactionIndex  hexutil.Uint64  `json:"transactionIndex"`
		From              common.Address  `json:"from"`
		To                *common.Address `json:"to"`
		GasUsed           hexutil.Uint64  `json:"gasUsed"`
		CumulativeGasUsed hexutil.Uint64  `json:"cumulativeGasUsed"`
		ContractAddress   *common.Address `json:"contractAddress"`
		Logs              []*types.Log    `json:"logs"`
		LogsBloom         types.Bloom     `json:"logsBloom"`
		Type              hexutil.Uint    `json:"type"`
		EffectiveGasPrice *hexutil.Big    `json:"effectiveGasPrice"`
		Status            hexutil.Uint    `json:"status"`
	}
	gomega.Expect(raw.CallContext(
		ctx,
		&receiptJSON,
		"qrl_getTransactionReceipt",
		fixture.tx.Hash(),
	)).To(gomega.Succeed())
	sender, err := types.Sender(types.LatestSignerForChainID(suite.chainID), fixture.tx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(receiptJSON.BlockHash).To(gomega.Equal(fixture.receipt.BlockHash))
	gomega.Expect(uint64(receiptJSON.BlockNumber)).To(gomega.Equal(fixture.receipt.BlockNumber.Uint64()))
	gomega.Expect(receiptJSON.TransactionHash).To(gomega.Equal(fixture.tx.Hash()))
	gomega.Expect(uint64(receiptJSON.TransactionIndex)).To(
		gomega.Equal(uint64(fixture.receipt.TransactionIndex)),
	)
	gomega.Expect(receiptJSON.From).To(gomega.Equal(sender))
	gomega.Expect(receiptJSON.To).To(gomega.BeNil())
	gomega.Expect(uint64(receiptJSON.GasUsed)).To(gomega.Equal(fixture.receipt.GasUsed))
	gomega.Expect(uint64(receiptJSON.CumulativeGasUsed)).To(
		gomega.Equal(fixture.receipt.CumulativeGasUsed),
	)
	gomega.Expect(receiptJSON.ContractAddress).NotTo(gomega.BeNil())
	gomega.Expect(*receiptJSON.ContractAddress).To(gomega.Equal(fixture.receipt.ContractAddress))
	gomega.Expect(receiptJSON.Logs).To(gomega.Equal(fixture.receipt.Logs))
	gomega.Expect(receiptJSON.LogsBloom).To(gomega.Equal(fixture.receipt.Bloom))
	gomega.Expect(uint64(receiptJSON.Type)).To(gomega.Equal(uint64(fixture.tx.Type())))
	gomega.Expect(receiptJSON.EffectiveGasPrice).NotTo(gomega.BeNil())
	gomega.Expect(receiptJSON.EffectiveGasPrice.ToInt()).To(
		gomega.Equal(fixture.receipt.EffectiveGasPrice),
	)
	gomega.Expect(uint64(receiptJSON.Status)).To(gomega.Equal(fixture.receipt.Status))

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

func expectRPCTransaction(
	got qrlapi.RPCTransaction,
	want *types.Transaction,
	receipt *types.Receipt,
	block *types.Block,
	chainID *big.Int,
) {
	ginkgo.GinkgoHelper()

	sender, err := types.Sender(types.LatestSignerForChainID(chainID), want)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(got.BlockHash).NotTo(gomega.BeNil())
	gomega.Expect(*got.BlockHash).To(gomega.Equal(block.Hash()))
	gomega.Expect(got.BlockNumber).NotTo(gomega.BeNil())
	gomega.Expect(got.BlockNumber.ToInt()).To(gomega.Equal(block.Number()))
	gomega.Expect(got.From).To(gomega.Equal(sender))
	gomega.Expect(uint64(got.Gas)).To(gomega.Equal(want.Gas()))
	gomega.Expect(got.GasPrice).NotTo(gomega.BeNil())
	gomega.Expect(got.GasPrice.ToInt()).To(gomega.Equal(receipt.EffectiveGasPrice))
	gomega.Expect(got.GasFeeCap).NotTo(gomega.BeNil())
	gomega.Expect(got.GasFeeCap.ToInt()).To(gomega.Equal(want.GasFeeCap()))
	gomega.Expect(got.GasTipCap).NotTo(gomega.BeNil())
	gomega.Expect(got.GasTipCap.ToInt()).To(gomega.Equal(want.GasTipCap()))
	gomega.Expect(got.Hash).To(gomega.Equal(want.Hash()))
	gomega.Expect(got.Input).To(gomega.Equal(hexutil.Bytes(want.Data())))
	gomega.Expect(uint64(got.Nonce)).To(gomega.Equal(want.Nonce()))
	gomega.Expect(got.To).To(gomega.BeNil())
	gomega.Expect(got.TransactionIndex).NotTo(gomega.BeNil())
	gomega.Expect(uint64(*got.TransactionIndex)).To(gomega.Equal(uint64(receipt.TransactionIndex)))
	gomega.Expect(got.Value).NotTo(gomega.BeNil())
	gomega.Expect(got.Value.ToInt().Cmp(want.Value())).To(gomega.BeZero())
	gomega.Expect(uint64(got.Type)).To(gomega.Equal(uint64(want.Type())))
	gomega.Expect(got.Accesses).NotTo(gomega.BeNil())
	gomega.Expect(*got.Accesses).To(gomega.Equal(want.AccessList()))
	gomega.Expect(got.ChainID).NotTo(gomega.BeNil())
	gomega.Expect(got.ChainID.ToInt()).To(gomega.Equal(want.ChainId()))
	gomega.Expect(got.Descriptor).To(gomega.Equal(hexutil.Bytes(want.Descriptor())))
	gomega.Expect(got.ExtraParams).To(gomega.Equal(hexutil.Bytes(want.ExtraParams())))
	gomega.Expect(got.PublicKey).To(gomega.Equal(hexutil.Bytes(want.RawPublicKeyValue())))
	gomega.Expect(got.Signature).To(gomega.Equal(hexutil.Bytes(want.RawSignatureValue())))
}
