// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"encoding/json"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func (suite *liveSuite) assertGraphQLMutation(ctx context.Context) {
	ginkgo.GinkgoHelper()

	storageKey := common.HexToHash("0x01")
	accessList := types.AccessList{{
		Address:     suite.from,
		StorageKeys: []common.Hash{storageKey},
	}}
	tx, err := suite.signTransactionWithAccessList(
		ctx,
		&suite.from,
		nil,
		nil,
		accessList,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	encoded, err := tx.MarshalBinary()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	mutationData := suite.queryGraphQL(ctx, `
			mutation Send($raw: Bytes!) {
				sendRawTransaction(data: $raw)
			}
		`, map[string]any{"raw": hexutil.Encode(encoded)})
	var mutation struct {
		Hash string `json:"sendRawTransaction"`
	}
	gomega.Expect(json.Unmarshal(mutationData, &mutation)).To(gomega.Succeed())
	gomega.Expect(mutation.Hash).To(gomega.Equal(tx.Hash().Hex()))
	receipt := suite.waitReceipt(ctx, tx.Hash())
	gomega.Expect(receipt.Status).To(
		gomega.Equal(types.ReceiptStatusSuccessful),
	)

	transactionData := suite.queryGraphQL(ctx, `
		query Transaction($hash: Bytes32!) {
			transaction(hash: $hash) {
				hash
				to { address }
				accessList { address storageKeys }
			}
		}
	`, map[string]any{"hash": tx.Hash().Hex()})
	var transactionRoot struct {
		Transaction graphQLTransaction `json:"transaction"`
	}
	gomega.Expect(json.Unmarshal(transactionData, &transactionRoot)).To(gomega.Succeed())
	gomega.Expect(transactionRoot.Transaction.Hash).To(gomega.Equal(tx.Hash().Hex()))
	gomega.Expect(transactionRoot.Transaction.To).NotTo(gomega.BeNil())
	gomega.Expect(transactionRoot.Transaction.To.Address).To(gomega.Equal(suite.from.Hex()))
	gomega.Expect(transactionRoot.Transaction.AccessList).To(gomega.HaveLen(1))
	gomega.Expect(transactionRoot.Transaction.AccessList[0].Address).To(
		gomega.Equal(suite.from.Hex()),
	)
	gomega.Expect(transactionRoot.Transaction.AccessList[0].StorageKeys).To(
		gomega.Equal([]string{storageKey.Hex()}),
	)
}
