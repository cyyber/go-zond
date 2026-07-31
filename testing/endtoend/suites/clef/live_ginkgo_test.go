// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package clef

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	qrl "github.com/theQRL/go-qrl"
	qrlbind "github.com/theQRL/go-qrl/accounts/abi/bind"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/core/types"
	qrvm "github.com/theQRL/go-qrl/core/vm"
	"github.com/theQRL/go-qrl/testing/endtoend/internal/build"
	endtoendlive "github.com/theQRL/go-qrl/testing/endtoend/internal/live"
	qrllibwallet "github.com/theQRL/go-qrllib/wallet/common"
)

const liveSpecTimeout = 10 * time.Minute

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Clef live E2E suite")
}

var _ = ginkgo.Describe(
	"standalone Clef against a live qrl-package network",
	ginkgo.Serial,
	ginkgo.Ordered,
	ginkgo.ContinueOnFailure,
	ginkgo.Label("e2e", "live", "clef", "mutates-chain"),
	func() {
		var (
			session *clefSession
			network *endtoendlive.Session
		)

		ginkgo.BeforeAll(func(ctx ginkgo.SpecContext) {
			var err error
			network, err = endtoendlive.Open(ctx, false)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			ginkgo.DeferCleanup(network.Close)

			workDir := ginkgo.GinkgoT().TempDir()
			clefPath := filepath.Join(workDir, "clef")
			ginkgo.By("building the current Clef binary")
			gomega.Expect(build.Binary(ctx, "./cmd/clef", clefPath)).To(gomega.Succeed())

			session, err = newClefSession(
				ctx,
				context.WithoutCancel(ctx),
				clefPath,
				workDir,
				network.ChainID,
				network.Wallet,
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			ginkgo.DeferCleanup(func() {
				gomega.Expect(session.close()).To(gomega.Succeed())
			})
		})

		ginkgo.It("lists the imported QRL account", func(ctx ginkgo.SpecContext) {
			gomega.Expect(
				verifyAccountListing(ctx, session.client, session.account),
			).To(gomega.Succeed())
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("reports the external API version", func(ctx ginkgo.SpecContext) {
			gomega.Expect(verifyVersion(ctx, session.client)).To(gomega.Succeed())
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("signs and verifies plain-text data", func(ctx ginkgo.SpecContext) {
			gomega.Expect(
				verifyDataSigning(ctx, session.client, session.account, session.expectedWallet),
			).To(gomega.Succeed())
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("rejects data denied by the ruleset", func(ctx ginkgo.SpecContext) {
			gomega.Expect(
				verifyDataRejection(ctx, session.client, session.account),
			).To(gomega.Succeed())
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("signs and verifies validator-bound data", func(ctx ginkgo.SpecContext) {
			gomega.Expect(
				verifyValidatorDataSigning(
					ctx,
					session.client,
					session.account,
					session.expectedWallet,
				),
			).To(gomega.Succeed())
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("signs and verifies QRL typed data", func(ctx ginkgo.SpecContext) {
			gomega.Expect(
				verifyTypedDataSigning(
					ctx,
					session.client,
					session.account,
					session.chainID,
					session.expectedWallet,
				),
			).To(gomega.Succeed())
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("rejects typed data for a different chain", func(ctx ginkgo.SpecContext) {
			gomega.Expect(
				verifyTypedDataChainIDRejection(
					ctx,
					session.client,
					session.account,
					session.chainID,
				),
			).To(gomega.Succeed())
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("verifies a Clef typed-data signature through the precompile", func(ctx ginkgo.SpecContext) {
			signature, digest, err := signTypedData(
				ctx,
				session.client,
				session.account,
				session.chainID,
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			context := qrllibwallet.SigningContext(session.expectedWallet.GetDescriptor())
			input := make([]byte, 0, len(digest)+len(session.expectedWallet.GetPK())+len(signature)+1+len(context))
			input = append(input, digest...)
			input = append(input, session.expectedWallet.GetPK()...)
			input = append(input, signature...)
			input = append(input, byte(len(context)))
			input = append(input, context...)

			address := common.BytesToAddress([]byte{3})
			output, err := network.Client.CallContract(ctx, qrl.CallMsg{
				From: session.account,
				To:   &address,
				Gas:  500_000,
				Data: input,
			}, nil)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(output).To(gomega.Equal(common.LeftPadBytes([]byte{1}, qrvm.WordBytes)))
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("signs, submits, and confirms a transaction", func(ctx ginkgo.SpecContext) {
			nonce, err := network.Client.PendingNonceAt(ctx, session.account)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			tip, err := network.Client.SuggestGasTipCap(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			feeCap, err := network.Client.SuggestGasPrice(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			request := transactionArgs(session.account, session.chainID, nonce, tip, feeCap)
			signed, err := signTransaction(ctx, session.client, request)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(
				verifyTransaction(signed, request, session.account, session.expectedWallet),
			).To(gomega.Succeed())

			gomega.Expect(network.Client.SendTransaction(ctx, signed.Tx)).To(gomega.Succeed())
			receipt, err := qrlbind.WaitMined(ctx, network.Client, signed.Tx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(receipt.Status).To(gomega.Equal(types.ReceiptStatusSuccessful))
			gomega.Expect(receipt.TxHash).To(gomega.Equal(signed.Tx.Hash()))
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("rejects a transaction denied by the ruleset", func(ctx ginkgo.SpecContext) {
			nonce, err := network.Client.PendingNonceAt(ctx, session.account)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			tip, err := network.Client.SuggestGasTipCap(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			feeCap, err := network.Client.SuggestGasPrice(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(verifyTransactionRejection(
				ctx,
				session.client,
				transactionArgs(session.account, session.chainID, nonce, tip, feeCap),
			)).To(gomega.Succeed())
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("persists a new password-protected account across restart", func(ctx ginkgo.SpecContext) {
			account, err := verifyNewAccount(ctx, session)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(session.restart(ctx, context.WithoutCancel(ctx))).To(gomega.Succeed())
			gomega.Expect(verifyAccountPresent(ctx, session.client, account)).To(gomega.Succeed())

			nonce, err := network.Client.PendingNonceAt(ctx, account)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			tip, err := network.Client.SuggestGasTipCap(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			feeCap, err := network.Client.SuggestGasPrice(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			signed, err := signTransaction(
				ctx,
				session.client,
				transactionArgs(account, session.chainID, nonce, tip, feeCap),
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(verifyTransactionSender(signed, account)).To(gomega.Succeed())
		}, ginkgo.SpecTimeout(liveSpecTimeout))
	},
)
