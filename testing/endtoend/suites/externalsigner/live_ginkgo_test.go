// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package externalsigner

import (
	"context"
	"math/big"
	"time"

	qrl "github.com/theQRL/go-qrl"
	"github.com/theQRL/go-qrl/accounts"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/crypto/pqcrypto"
	qrlwallet "github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
	"github.com/theQRL/go-qrl/internal/qrlapi"
	"github.com/theQRL/go-qrl/testing/endtoend/internal/fixture"
	endtoendlive "github.com/theQRL/go-qrl/testing/endtoend/internal/live"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

const (
	liveSpecTimeout = 5 * time.Minute
)

type liveSuite struct {
	session *endtoendlive.Session
	wallet  qrlwallet.Wallet
	account common.Address
}

var _ = ginkgo.Describe(
	"go-qrl configured with Clef",
	ginkgo.Serial,
	ginkgo.Ordered,
	ginkgo.ContinueOnFailure,
	ginkgo.Label("e2e", "live", "external-signer", "mutates-chain"),
	func() {
		var suite *liveSuite

		ginkgo.BeforeAll(func(ctx ginkgo.SpecContext) {
			suite = newLiveSuite(ctx)
			gomega.Expect(suite).NotTo(gomega.BeNil())
			ginkgo.DeferCleanup(suite.session.Close)

			var accounts []common.Address
			err := suite.session.Client.Client().CallContext(ctx, &accounts, "qrl_accounts")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(accounts).To(gomega.ContainElement(suite.account))
		})

		ginkgo.It("discovers the node-managed Clef account", func(ctx ginkgo.SpecContext) {
			var managed []common.Address
			gomega.Expect(suite.session.Client.Client().CallContext(
				ctx,
				&managed,
				"qrl_accounts",
			)).To(gomega.Succeed())
			gomega.Expect(managed).To(gomega.Equal([]common.Address{suite.account}))
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("signs text through the node", func(ctx ginkgo.SpecContext) {
			message := []byte("go-qrl external signer E2E")
			var signature hexutil.Bytes
			gomega.Expect(suite.session.Client.Client().CallContext(
				ctx,
				&signature,
				"qrl_sign",
				suite.account,
				hexutil.Bytes(message),
			)).To(gomega.Succeed())

			valid, err := pqcrypto.MLDSA87VerifySignature(
				signature,
				accounts.TextHash(message),
				suite.wallet.GetPK(),
				suite.wallet.GetDescriptor(),
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(valid).To(gomega.BeTrue())
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("signs a transaction through the node", func(ctx ginkgo.SpecContext) {
			args := suite.transactionArgs(ctx)
			var signed qrlapi.SignTransactionResult
			gomega.Expect(suite.session.Client.Client().CallContext(
				ctx,
				&signed,
				"qrl_signTransaction",
				args,
			)).To(gomega.Succeed())
			gomega.Expect(signed.Tx).NotTo(gomega.BeNil())
			gomega.Expect(signed.Raw).NotTo(gomega.BeEmpty())
			gomega.Expect(transactionSender(signed.Tx, suite.session.ChainID)).To(gomega.Equal(suite.account))

			var decoded types.Transaction
			gomega.Expect(decoded.UnmarshalBinary(signed.Raw)).To(gomega.Succeed())
			gomega.Expect(decoded.Hash()).To(gomega.Equal(signed.Tx.Hash()))
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("signs, submits, and mines a transaction through the node", func(ctx ginkgo.SpecContext) {
			args := suite.transactionArgs(ctx)
			var hash common.Hash
			gomega.Expect(suite.session.Client.Client().CallContext(
				ctx,
				&hash,
				"qrl_sendTransaction",
				args,
			)).To(gomega.Succeed())
			gomega.Expect(hash).NotTo(gomega.Equal(common.Hash{}))

			var receipt *types.Receipt
			gomega.Eventually(func() error {
				var err error
				receipt, err = suite.session.Client.TransactionReceipt(ctx, hash)
				return err
			}).WithContext(ctx).WithTimeout(2 * time.Minute).WithPolling(time.Second).Should(
				gomega.Succeed(),
			)
			gomega.Expect(receipt.Status).To(gomega.Equal(types.ReceiptStatusSuccessful))
			gomega.Expect(receipt.TxHash).To(gomega.Equal(hash))

			tx, pending, err := suite.session.Client.TransactionByHash(ctx, hash)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(pending).To(gomega.BeFalse())
			gomega.Expect(transactionSender(tx, suite.session.ChainID)).To(gomega.Equal(suite.account))
		}, ginkgo.SpecTimeout(liveSpecTimeout))
	},
)

func newLiveSuite(ctx context.Context) *liveSuite {
	ginkgo.GinkgoHelper()

	session, err := endtoendlive.Open(ctx, false)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	wallet, err := qrlwallet.RestoreFromSeedHex(fixture.RemoteSignerSeed)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	account := common.Address(wallet.GetAddress())

	return &liveSuite{session: session, wallet: wallet, account: account}
}

func (suite *liveSuite) transactionArgs(ctx context.Context) map[string]any {
	ginkgo.GinkgoHelper()

	nonce, err := suite.session.Client.PendingNonceAt(ctx, suite.account)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	feeCap, err := suite.session.Client.SuggestGasPrice(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	tipCap, err := suite.session.Client.SuggestGasTipCap(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gas, err := suite.session.Client.EstimateGas(ctx, qrl.CallMsg{
		From: suite.account,
		To:   &suite.session.Address,
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return map[string]any{
		"from":                 suite.account,
		"to":                   suite.session.Address,
		"gas":                  hexutil.Uint64(gas),
		"maxFeePerGas":         (*hexutil.Big)(feeCap),
		"maxPriorityFeePerGas": (*hexutil.Big)(tipCap),
		"value":                (*hexutil.Big)(big.NewInt(1)),
		"nonce":                hexutil.Uint64(nonce),
		"chainId":              (*hexutil.Big)(suite.session.ChainID),
	}
}

func transactionSender(tx *types.Transaction, chainID *big.Int) common.Address {
	ginkgo.GinkgoHelper()

	from, err := types.Sender(types.LatestSignerForChainID(chainID), tx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return from
}
