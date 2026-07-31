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
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	qrlwallet "github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
	"github.com/theQRL/go-qrl/internal/qrlapi"
	"github.com/theQRL/go-qrl/params"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func (suite *liveSuite) assertTxPool(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()
	queuedWallet, err := qrlwallet.Generate(qrlwallet.ML_DSA_87)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	queuedAddress := common.Address(queuedWallet.GetAddress())

	funding, err := suite.signTransaction(
		ctx,
		&queuedAddress,
		big.NewInt(params.Quanta),
		nil,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(suite.submitAndWait(ctx, funding).Status).To(
		gomega.Equal(types.ReceiptStatusSuccessful),
	)

	queued, err := suite.signTransactionForWallet(
		ctx,
		queuedWallet,
		queuedAddress,
		1,
		&queuedAddress,
		new(big.Int),
		nil,
		nil,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(suite.client.SendTransaction(ctx, queued)).To(gomega.Succeed())

	var txpoolContent json.RawMessage
	gomega.Expect(raw.CallContext(ctx, &txpoolContent, "txpool_content")).To(
		gomega.Succeed(),
	)
	gomega.Expect(string(txpoolContent)).To(gomega.ContainSubstring(queued.Hash().Hex()))

	var accountContent json.RawMessage
	gomega.Expect(raw.CallContext(
		ctx,
		&accountContent,
		"txpool_contentFrom",
		queuedAddress,
	)).To(gomega.Succeed())
	gomega.Expect(string(accountContent)).To(gomega.ContainSubstring(queued.Hash().Hex()))

	var txpoolStatus map[string]hexutil.Uint
	gomega.Expect(raw.CallContext(ctx, &txpoolStatus, "txpool_status")).To(gomega.Succeed())
	gomega.Expect(uint64(txpoolStatus["queued"])).To(gomega.BeNumerically(">=", 1))

	var txpoolInspect json.RawMessage
	gomega.Expect(raw.CallContext(ctx, &txpoolInspect, "txpool_inspect")).To(gomega.Succeed())
	gomega.Expect(string(txpoolInspect)).To(gomega.ContainSubstring(queuedAddress.Hex()))

	var managedPending []json.RawMessage
	gomega.Expect(raw.CallContext(
		ctx,
		&managedPending,
		"qrl_pendingTransactions",
	)).To(gomega.Succeed())
	// The pool entries were signed by suite wallets, not the Clef-managed account.
	gomega.Expect(managedPending).To(gomega.BeEmpty())
	suite.assertManagedPendingTransaction(ctx)

	pending, err := suite.signTransactionForWallet(
		ctx,
		queuedWallet,
		queuedAddress,
		0,
		&queuedAddress,
		new(big.Int),
		nil,
		nil,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(suite.client.SendTransaction(ctx, pending)).To(gomega.Succeed())
	gomega.Expect(suite.waitReceipt(ctx, pending.Hash()).Status).To(
		gomega.Equal(types.ReceiptStatusSuccessful),
	)
	gomega.Expect(suite.waitReceipt(ctx, queued.Hash()).Status).To(
		gomega.Equal(types.ReceiptStatusSuccessful),
	)
}

func (suite *liveSuite) assertManagedPendingTransaction(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()
	var managedAccounts []common.Address
	gomega.Expect(raw.CallContext(ctx, &managedAccounts, "qrl_accounts")).To(gomega.Succeed())
	gomega.Expect(managedAccounts).To(gomega.HaveLen(1))
	managed := managedAccounts[0]

	blockNumber, err := suite.client.BlockNumber(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Eventually(func() uint64 {
		number, blockErr := suite.client.BlockNumber(ctx)
		gomega.Expect(blockErr).NotTo(gomega.HaveOccurred())
		return number
	}).WithContext(ctx).WithTimeout(2 * time.Minute).WithPolling(time.Second).Should(
		gomega.BeNumerically(">", blockNumber),
	)

	nonce, err := suite.client.PendingNonceAt(ctx, managed)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	feeCap, err := suite.client.SuggestGasPrice(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	tipCap, err := suite.client.SuggestGasTipCap(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	to := suite.from
	gas := hexutil.Uint64(params.TxGas)
	nonceValue := hexutil.Uint64(nonce)
	value := (*hexutil.Big)(big.NewInt(1))
	args := qrlapi.TransactionArgs{
		From:                 &managed,
		To:                   &to,
		Gas:                  &gas,
		MaxFeePerGas:         (*hexutil.Big)(feeCap),
		MaxPriorityFeePerGas: (*hexutil.Big)(tipCap),
		Value:                value,
		Nonce:                &nonceValue,
		ChainID:              (*hexutil.Big)(suite.chainID),
	}

	var hash common.Hash
	gomega.Expect(raw.CallContext(ctx, &hash, "qrl_sendTransaction", args)).To(gomega.Succeed())

	var pending []*qrlapi.RPCTransaction
	gomega.Expect(raw.CallContext(ctx, &pending, "qrl_pendingTransactions")).To(gomega.Succeed())
	var managedPending *qrlapi.RPCTransaction
	for _, transaction := range pending {
		if transaction.Hash == hash {
			managedPending = transaction
			break
		}
	}
	gomega.Expect(managedPending).NotTo(gomega.BeNil())
	gomega.Expect(managedPending.From).To(gomega.Equal(managed))
	gomega.Expect(managedPending.To).NotTo(gomega.BeNil())
	gomega.Expect(*managedPending.To).To(gomega.Equal(to))
	gomega.Expect(managedPending.Nonce).To(gomega.Equal(nonceValue))
	gomega.Expect(managedPending.Value).NotTo(gomega.BeNil())
	gomega.Expect(managedPending.Value.ToInt()).To(gomega.Equal(value.ToInt()))
	gomega.Expect(suite.waitReceipt(ctx, hash).Status).To(
		gomega.Equal(types.ReceiptStatusSuccessful),
	)
}
