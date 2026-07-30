// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"fmt"
	"math/big"
	"time"

	qrl "github.com/theQRL/go-qrl"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/core/types"
	qrlwallet "github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
	"github.com/theQRL/go-qrl/qrlclient"
	"github.com/theQRL/go-qrl/testing/devnet"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

type liveSuite struct {
	graphQLURL string
	client     *qrlclient.Client
	wsClient   *qrlclient.Client
	wallet     qrlwallet.Wallet
	from       common.Address
	chainID    *big.Int
	fixture    *liveFixture
}

type liveFixture struct {
	address common.Address
	tx      *types.Transaction
	receipt *types.Receipt
	block   *types.Block
	value   common.StorageValue64
	topic   common.LogTopic
}

func setupLiveSuite(ctx context.Context) *liveSuite {
	ginkgo.GinkgoHelper()

	environment, err := devnet.Inspect(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	client, err := qrlclient.DialContext(ctx, environment.RPCURL)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(client.Close)

	wsClient, err := qrlclient.DialContext(ctx, environment.WebSocketURL)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(wsClient.Close)

	wallet, err := devnet.UnsafeDevelopmentWallet()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	chainID, err := client.ChainID(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	suite := &liveSuite{
		graphQLURL: environment.GraphQLURL,
		client:     client,
		wsClient:   wsClient,
		wallet:     wallet,
		from:       common.Address(wallet.GetAddress()),
		chainID:    chainID,
	}
	suite.fixture = suite.deployFixture(ctx)
	return suite
}

func (suite *liveSuite) deployFixture(ctx context.Context) *liveFixture {
	ginkgo.GinkgoHelper()

	var value common.StorageValue64
	for index := range value {
		value[index] = byte(index + 1)
	}
	var topic common.LogTopic
	for index := range topic {
		topic[index] = byte(0xff - index)
	}

	tx, err := suite.signTransaction(ctx, nil, new(big.Int), apiContractCode(value, topic))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	receipt := suite.submitAndWait(ctx, tx)
	gomega.Expect(receipt.Status).To(gomega.Equal(types.ReceiptStatusSuccessful))
	gomega.Expect(receipt.ContractAddress).NotTo(gomega.Equal(common.Address{}))
	gomega.Expect(receipt.Logs).To(gomega.HaveLen(1))
	gomega.Expect(receipt.Logs[0].Address).To(gomega.Equal(receipt.ContractAddress))
	gomega.Expect(receipt.Logs[0].Topics).To(gomega.Equal([]common.LogTopic{topic}))
	gomega.Expect(receipt.Logs[0].Data).To(gomega.Equal(value[:]))

	block, err := suite.client.BlockByNumber(ctx, receipt.BlockNumber)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(block).NotTo(gomega.BeNil())

	return &liveFixture{
		address: receipt.ContractAddress,
		tx:      tx,
		receipt: receipt,
		block:   block,
		value:   value,
		topic:   topic,
	}
}

func (suite *liveSuite) signTransaction(
	ctx context.Context,
	to *common.Address,
	value *big.Int,
	data []byte,
) (*types.Transaction, error) {
	if value == nil {
		value = new(big.Int)
	}
	nonce, err := suite.client.PendingNonceAt(ctx, suite.from)
	if err != nil {
		return nil, fmt.Errorf("read pending nonce: %w", err)
	}
	feeCap, err := suite.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("suggest gas price: %w", err)
	}
	tipCap, err := suite.client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, fmt.Errorf("suggest gas tip: %w", err)
	}
	feeCap = new(big.Int).Mul(feeCap, big.NewInt(4))
	if feeCap.Cmp(tipCap) < 0 {
		feeCap = new(big.Int).Set(tipCap)
	}
	gas, err := suite.client.EstimateGas(ctx, qrl.CallMsg{
		From:  suite.from,
		To:    to,
		Value: value,
		Data:  data,
	})
	if err != nil {
		return nil, fmt.Errorf("estimate gas: %w", err)
	}
	gas += gas / 5

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   suite.chainID,
		Nonce:     nonce,
		GasTipCap: tipCap,
		GasFeeCap: feeCap,
		Gas:       gas,
		To:        to,
		Value:     value,
		Data:      data,
	})
	signed, err := types.SignTx(tx, types.LatestSignerForChainID(suite.chainID), suite.wallet)
	if err != nil {
		return nil, fmt.Errorf("sign transaction: %w", err)
	}
	return signed, nil
}

func (suite *liveSuite) submitAndWait(ctx context.Context, tx *types.Transaction) *types.Receipt {
	ginkgo.GinkgoHelper()

	gomega.Expect(suite.client.SendTransaction(ctx, tx)).To(gomega.Succeed())
	return suite.waitReceipt(ctx, tx.Hash())
}

func (suite *liveSuite) waitReceipt(ctx context.Context, hash common.Hash) *types.Receipt {
	ginkgo.GinkgoHelper()

	var receipt *types.Receipt
	gomega.Eventually(func() error {
		var err error
		receipt, err = suite.client.TransactionReceipt(ctx, hash)
		return err
	}).WithContext(ctx).WithTimeout(2 * time.Minute).WithPolling(time.Second).Should(
		gomega.Succeed(),
	)
	gomega.Expect(receipt).NotTo(gomega.BeNil())
	gomega.Expect(receipt.BlockNumber).NotTo(gomega.BeNil())
	return receipt
}
