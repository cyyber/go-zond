// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"encoding/json"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"

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
	var root struct {
		Block struct {
			Hash         string               `json:"hash"`
			Transactions []graphQLTransaction `json:"transactions"`
			Logs         []graphQLLog         `json:"logs"`
			Withdrawals  []json.RawMessage    `json:"withdrawals"`
			Account      struct {
				Address string `json:"address"`
				Storage string `json:"storage"`
			} `json:"account"`
		} `json:"block"`
		BlockByHash struct {
			Hash string `json:"hash"`
		} `json:"blockByHash"`
		Blocks      []json.RawMessage  `json:"blocks"`
		Pending     json.RawMessage    `json:"pending"`
		Transaction graphQLTransaction `json:"transaction"`
		Logs        []graphQLLog       `json:"logs"`
		GasPrice    string             `json:"gasPrice"`
		PriorityFee string             `json:"maxPriorityFeePerGas"`
		Syncing     *json.RawMessage   `json:"syncing"`
		ChainID     string             `json:"chainID"`
	}
	gomega.Expect(json.Unmarshal(data, &root)).To(gomega.Succeed())
	gomega.Expect(root.Block.Hash).To(gomega.Equal(fixture.block.Hash().Hex()))
	gomega.Expect(root.BlockByHash.Hash).To(gomega.Equal(fixture.block.Hash().Hex()))
	gomega.Expect(root.Blocks).To(gomega.HaveLen(1))
	gomega.Expect(root.Pending).NotTo(gomega.BeEmpty())
	gomega.Expect(root.Block.Account.Address).To(gomega.Equal(fixture.address.Hex()))
	gomega.Expect(root.Block.Account.Storage).To(
		gomega.Equal(fixture.value.Hex()),
		"GraphQL account storage",
	)
	gomega.Expect(root.GasPrice).NotTo(gomega.BeEmpty())
	gomega.Expect(root.PriorityFee).NotTo(gomega.BeEmpty())
	gomega.Expect(root.ChainID).To(gomega.Equal(hexutil.EncodeBig(suite.chainID)))
	gomega.Expect(root.Syncing).To(gomega.BeNil())
	gomega.Expect(root.Block.Withdrawals).NotTo(gomega.BeNil())

	var blockTransaction *graphQLTransaction
	for index := range root.Block.Transactions {
		if root.Block.Transactions[index].Hash == fixture.tx.Hash().Hex() {
			blockTransaction = &root.Block.Transactions[index]
			break
		}
	}
	gomega.Expect(blockTransaction).NotTo(gomega.BeNil())
	assertGraphQLTransaction(*blockTransaction, fixture.tx, fixture.receipt)
	assertGraphQLTransaction(root.Transaction, fixture.tx, fixture.receipt)

	gomega.Expect(root.Block.Logs).To(gomega.HaveLen(1))
	assertGraphQLLog(root.Block.Logs[0], fixture)
	gomega.Expect(root.Logs).To(gomega.HaveLen(1))
	assertGraphQLLog(root.Logs[0], fixture)
}
