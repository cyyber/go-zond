// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"encoding/json"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/consensus/misc/eip1559"
	"github.com/theQRL/go-qrl/params"
	"github.com/theQRL/go-qrl/rlp"
	"github.com/theQRL/go-qrl/rpc"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func (suite *liveSuite) assertGraphQLQueries(ctx context.Context) {
	ginkgo.GinkgoHelper()

	fixture := suite.fixture
	block := hexutil.EncodeBig(fixture.receipt.BlockNumber)
	index := hexutil.EncodeUint64(uint64(fixture.receipt.TransactionIndex))
	slot := (common.Hash{}).Hex()

	data := suite.queryGraphQL(ctx, apiGraphQLQuery, map[string]any{
		"block":   block,
		"hash":    fixture.block.Hash().Hex(),
		"txHash":  fixture.tx.Hash().Hex(),
		"address": fixture.address.Hex(),
		"sender":  suite.from.Hex(),
		"slot":    slot,
		"topic":   fixture.topic.Hex(),
		"index":   index,
	})
	type callResult struct {
		Data    string `json:"data"`
		GasUsed string `json:"gasUsed"`
		Status  string `json:"status"`
	}
	type account struct {
		Address          string `json:"address"`
		Balance          string `json:"balance"`
		TransactionCount string `json:"transactionCount"`
		Code             string `json:"code"`
		Storage          string `json:"storage"`
	}
	type withdrawal struct {
		Index     string `json:"index"`
		Validator string `json:"validator"`
		Address   string `json:"address"`
		Amount    string `json:"amount"`
	}
	var root struct {
		Block struct {
			Number string `json:"number"`
			Hash   string `json:"hash"`
			Parent struct {
				Number string `json:"number"`
				Hash   string `json:"hash"`
			} `json:"parent"`
			TransactionsRoot string               `json:"transactionsRoot"`
			TransactionCount string               `json:"transactionCount"`
			StateRoot        string               `json:"stateRoot"`
			ReceiptsRoot     string               `json:"receiptsRoot"`
			Miner            account              `json:"miner"`
			ExtraData        string               `json:"extraData"`
			GasLimit         string               `json:"gasLimit"`
			GasUsed          string               `json:"gasUsed"`
			BaseFeePerGas    string               `json:"baseFeePerGas"`
			NextBaseFee      string               `json:"nextBaseFeePerGas"`
			Timestamp        string               `json:"timestamp"`
			LogsBloom        string               `json:"logsBloom"`
			Random           string               `json:"random"`
			Transactions     []graphQLTransaction `json:"transactions"`
			TransactionAt    graphQLTransaction   `json:"transactionAt"`
			Logs             []graphQLLog         `json:"logs"`
			Withdrawals      []withdrawal         `json:"withdrawals"`
			Account          account              `json:"account"`
			Call             callResult           `json:"call"`
			EstimateGas      string               `json:"estimateGas"`
			RawHeader        string               `json:"rawHeader"`
			Raw              string               `json:"raw"`
			WithdrawalsRoot  *string              `json:"withdrawalsRoot"`
		} `json:"block"`
		BlockByHash struct {
			Number string `json:"number"`
			Hash   string `json:"hash"`
		} `json:"blockByHash"`
		Blocks []struct {
			Number string `json:"number"`
			Hash   string `json:"hash"`
		} `json:"blocks"`
		Pending struct {
			TransactionCount string               `json:"transactionCount"`
			Transactions     []graphQLTransaction `json:"transactions"`
			Account          account              `json:"account"`
			Call             callResult           `json:"call"`
			EstimateGas      string               `json:"estimateGas"`
		} `json:"pending"`
		Transaction graphQLTransaction `json:"transaction"`
		Logs        []graphQLLog       `json:"logs"`
		GasPrice    string             `json:"gasPrice"`
		PriorityFee string             `json:"maxPriorityFeePerGas"`
		Syncing     *json.RawMessage   `json:"syncing"`
		ChainID     string             `json:"chainID"`
	}
	gomega.Expect(json.Unmarshal(data, &root)).To(gomega.Succeed())

	header := fixture.block.Header()
	rawHeader, err := rlp.EncodeToBytes(header)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	rawBlock, err := rlp.EncodeToBytes(fixture.block)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	balance, err := suite.client.BalanceAt(ctx, fixture.address, fixture.receipt.BlockNumber)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	nonce, err := suite.client.NonceAt(ctx, fixture.address, fixture.receipt.BlockNumber)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	code, err := suite.client.CodeAt(ctx, fixture.address, fixture.receipt.BlockNumber)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	minerBalance, err := suite.client.BalanceAt(
		ctx,
		header.Coinbase,
		fixture.receipt.BlockNumber,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	minerNonce, err := suite.client.NonceAt(ctx, header.Coinbase, fixture.receipt.BlockNumber)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	minerCode, err := suite.client.CodeAt(ctx, header.Coinbase, fixture.receipt.BlockNumber)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	minerStorage, err := suite.client.StorageAt(
		ctx,
		header.Coinbase,
		common.Hash{},
		fixture.receipt.BlockNumber,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	pendingBalance, err := suite.client.PendingBalanceAt(ctx, suite.from)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	pendingNonce, err := suite.client.PendingNonceAt(ctx, suite.from)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	pendingCode, err := suite.client.PendingCodeAt(ctx, suite.from)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	nextBaseFee := eip1559.CalcBaseFee(params.AllBeaconProtocolChanges, header)

	callArgs := map[string]any{
		"from": suite.from,
		"to":   fixture.address,
		"data": "0x",
	}
	var blockEstimate hexutil.Uint64
	gomega.Expect(suite.client.Client().CallContext(
		ctx,
		&blockEstimate,
		"qrl_estimateGas",
		callArgs,
		rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(fixture.block.NumberU64())),
	)).To(gomega.Succeed())
	var pendingEstimate hexutil.Uint64
	gomega.Expect(suite.client.Client().CallContext(
		ctx,
		&pendingEstimate,
		"qrl_estimateGas",
		callArgs,
		rpc.BlockNumberOrHashWithNumber(rpc.PendingBlockNumber),
	)).To(gomega.Succeed())

	gomega.Expect(root.Block.Number).To(gomega.Equal(block))
	gomega.Expect(root.Block.Hash).To(gomega.Equal(fixture.block.Hash().Hex()))
	gomega.Expect(root.Block.Parent.Number).To(
		gomega.Equal(hexutil.EncodeUint64(fixture.block.NumberU64() - 1)),
	)
	gomega.Expect(root.Block.Parent.Hash).To(gomega.Equal(header.ParentHash.Hex()))
	gomega.Expect(root.Block.TransactionsRoot).To(gomega.Equal(header.TxHash.Hex()))
	gomega.Expect(root.Block.TransactionCount).To(
		gomega.Equal(hexutil.EncodeUint64(uint64(len(fixture.block.Transactions())))),
	)
	gomega.Expect(root.Block.StateRoot).To(gomega.Equal(header.Root.Hex()))
	gomega.Expect(root.Block.ReceiptsRoot).To(gomega.Equal(header.ReceiptHash.Hex()))
	gomega.Expect(root.Block.Miner.Address).To(gomega.Equal(header.Coinbase.Hex()))
	gomega.Expect(root.Block.Miner.Balance).To(
		gomega.Equal(hexutil.EncodeBig(minerBalance)),
		"GraphQL block miner balance",
	)
	gomega.Expect(root.Block.Miner.TransactionCount).To(
		gomega.Equal(hexutil.EncodeUint64(minerNonce)),
	)
	gomega.Expect(root.Block.Miner.Code).To(gomega.Equal(hexutil.Encode(minerCode)))
	gomega.Expect(root.Block.Miner.Storage).To(
		gomega.Equal(common.BytesToStorageValue64(minerStorage).Hex()),
	)
	gomega.Expect(root.Block.ExtraData).To(gomega.Equal(hexutil.Encode(header.Extra)))
	gomega.Expect(root.Block.GasLimit).To(gomega.Equal(hexutil.EncodeUint64(header.GasLimit)))
	gomega.Expect(root.Block.GasUsed).To(gomega.Equal(hexutil.EncodeUint64(header.GasUsed)))
	gomega.Expect(root.Block.BaseFeePerGas).To(gomega.Equal(hexutil.EncodeBig(header.BaseFee)))
	gomega.Expect(root.Block.NextBaseFee).To(
		gomega.Equal(hexutil.EncodeBig(nextBaseFee)),
		"GraphQL next base fee",
	)
	gomega.Expect(root.Block.Timestamp).To(gomega.Equal(hexutil.EncodeUint64(header.Time)))
	gomega.Expect(root.Block.LogsBloom).To(gomega.Equal(hexutil.Encode(header.Bloom.Bytes())))
	gomega.Expect(root.Block.Random).To(gomega.Equal(header.Random.Hex()))
	gomega.Expect(root.Block.TransactionAt.Hash).To(gomega.Equal(fixture.tx.Hash().Hex()))
	gomega.Expect(root.Block.RawHeader).To(gomega.Equal(hexutil.Encode(rawHeader)))
	gomega.Expect(root.Block.Raw).To(gomega.Equal(hexutil.Encode(rawBlock)))
	gomega.Expect(root.BlockByHash.Hash).To(gomega.Equal(fixture.block.Hash().Hex()))
	gomega.Expect(root.BlockByHash.Number).To(gomega.Equal(block))
	gomega.Expect(root.Blocks).To(gomega.Equal([]struct {
		Number string `json:"number"`
		Hash   string `json:"hash"`
	}{{Number: block, Hash: fixture.block.Hash().Hex()}}))
	gomega.Expect(root.Pending.TransactionCount).To(
		gomega.Equal(hexutil.EncodeUint64(uint64(len(root.Pending.Transactions)))),
	)
	gomega.Expect(root.Pending.Account.Address).To(gomega.Equal(suite.from.Hex()))
	gomega.Expect(root.Pending.Account.Balance).To(gomega.Equal(hexutil.EncodeBig(pendingBalance)))
	gomega.Expect(root.Pending.Account.TransactionCount).To(
		gomega.Equal(hexutil.EncodeUint64(pendingNonce)),
	)
	gomega.Expect(root.Pending.Account.Code).To(gomega.Equal(hexutil.Encode(pendingCode)))
	gomega.Expect(root.Block.Account.Address).To(gomega.Equal(fixture.address.Hex()))
	gomega.Expect(root.Block.Account.Balance).To(gomega.Equal(hexutil.EncodeBig(balance)))
	gomega.Expect(root.Block.Account.TransactionCount).To(gomega.Equal(hexutil.EncodeUint64(nonce)))
	gomega.Expect(root.Block.Account.Code).To(gomega.Equal(hexutil.Encode(code)))
	gomega.Expect(root.Block.Account.Storage).To(
		gomega.Equal(fixture.value.Hex()),
		"GraphQL account storage",
	)
	for _, call := range []callResult{root.Block.Call, root.Pending.Call} {
		gomega.Expect(call.Data).To(gomega.Equal(hexutil.Encode(fixture.value[:])))
		gomega.Expect(call.Status).To(gomega.Equal("0x1"))
		gasUsed, err := hexutil.DecodeUint64(call.GasUsed)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(gasUsed).To(gomega.BeNumerically(">", 0))
	}
	gomega.Expect(root.Block.EstimateGas).To(gomega.Equal(blockEstimate.String()))
	gomega.Expect(root.Pending.EstimateGas).To(gomega.Equal(pendingEstimate.String()))
	gomega.Expect(root.GasPrice).NotTo(gomega.BeEmpty())
	gomega.Expect(root.PriorityFee).NotTo(gomega.BeEmpty())
	gomega.Expect(root.ChainID).To(gomega.Equal(hexutil.EncodeBig(suite.chainID)))
	gomega.Expect(root.Syncing).To(gomega.BeNil())
	if header.WithdrawalsHash == nil {
		gomega.Expect(root.Block.WithdrawalsRoot).To(gomega.BeNil())
		gomega.Expect(root.Block.Withdrawals).To(gomega.BeNil())
	} else {
		gomega.Expect(root.Block.WithdrawalsRoot).NotTo(gomega.BeNil())
		gomega.Expect(*root.Block.WithdrawalsRoot).To(gomega.Equal(header.WithdrawalsHash.Hex()))
		gomega.Expect(root.Block.Withdrawals).To(gomega.HaveLen(len(fixture.block.Withdrawals())))
		for index, want := range fixture.block.Withdrawals() {
			got := root.Block.Withdrawals[index]
			gomega.Expect(got.Index).To(gomega.Equal(hexutil.EncodeUint64(want.Index)))
			gomega.Expect(got.Validator).To(gomega.Equal(hexutil.EncodeUint64(want.Validator)))
			gomega.Expect(got.Address).To(gomega.Equal(want.Address.Hex()))
			gomega.Expect(got.Amount).To(gomega.Equal(hexutil.EncodeUint64(want.Amount)))
		}
	}

	var blockTransaction *graphQLTransaction
	for index := range root.Block.Transactions {
		if root.Block.Transactions[index].Hash == fixture.tx.Hash().Hex() {
			blockTransaction = &root.Block.Transactions[index]
			break
		}
	}
	gomega.Expect(blockTransaction).NotTo(gomega.BeNil())
	assertGraphQLTransaction(
		*blockTransaction,
		fixture.tx,
		fixture.receipt,
		fixture.block,
		suite.chainID,
	)
	assertGraphQLTransactionEnvelope(root.Transaction, fixture.tx, fixture.receipt)

	gomega.Expect(root.Block.Logs).To(gomega.HaveLen(1))
	assertGraphQLLog(root.Block.Logs[0], fixture.receipt.Logs[0])
	gomega.Expect(root.Logs).To(gomega.HaveLen(1))
	assertGraphQLLog(root.Logs[0], fixture.receipt.Logs[0])
}
