// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"encoding/json"
	"math/big"
	"time"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/core/types"
	qrlwallet "github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
	"github.com/theQRL/go-qrl/params"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func (suite *liveSuite) assertGraphQLPending(ctx context.Context) {
	ginkgo.GinkgoHelper()

	pendingWallet, err := qrlwallet.Generate(qrlwallet.ML_DSA_87)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	pendingAddress := common.Address(pendingWallet.GetAddress())

	funding, err := suite.signTransaction(
		ctx,
		&pendingAddress,
		big.NewInt(params.Quanta),
		nil,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(suite.submitAndWait(ctx, funding).Status).To(
		gomega.Equal(types.ReceiptStatusSuccessful),
	)

	headers := make(chan *types.Header, 1)
	subscription, err := suite.wsClient.SubscribeNewHead(ctx, headers)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer subscription.Unsubscribe()

	select {
	case <-headers:
	case err := <-subscription.Err():
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	case <-time.After(30 * time.Second):
		ginkgo.Fail("timed out waiting for a fresh block before pending query")
	case <-ctx.Done():
		ginkgo.Fail("pending GraphQL context ended: " + ctx.Err().Error())
	}

	tx, err := suite.signTransactionForWallet(
		ctx,
		pendingWallet,
		pendingAddress,
		0,
		&pendingAddress,
		new(big.Int),
		nil,
		nil,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(suite.client.SendTransaction(ctx, tx)).To(gomega.Succeed())

	data := suite.queryGraphQL(ctx, `{
		pending {
			transactions {
				hash
				to { address }
				accessList { address storageKeys }
			}
		}
	}`, nil)
	var root struct {
		Pending struct {
			Transactions []graphQLTransaction `json:"transactions"`
		} `json:"pending"`
	}
	gomega.Expect(json.Unmarshal(data, &root)).To(gomega.Succeed())

	var found *graphQLTransaction
	for index := range root.Pending.Transactions {
		if root.Pending.Transactions[index].Hash == tx.Hash().Hex() {
			found = &root.Pending.Transactions[index]
			break
		}
	}
	gomega.Expect(found).NotTo(gomega.BeNil())
	gomega.Expect(found.To).NotTo(gomega.BeNil())
	gomega.Expect(found.To.Address).To(gomega.Equal(pendingAddress.Hex()))
	gomega.Expect(found.AccessList).NotTo(gomega.BeNil())

	gomega.Expect(suite.waitReceipt(ctx, tx.Hash()).Status).To(
		gomega.Equal(types.ReceiptStatusSuccessful),
	)
}
