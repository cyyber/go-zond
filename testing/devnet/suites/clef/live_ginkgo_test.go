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
	qrlbind "github.com/theQRL/go-qrl/accounts/abi/bind"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/qrlclient"
	"github.com/theQRL/go-qrl/testing/devnet/internal/build"
	"github.com/theQRL/go-qrl/testing/devnet/internal/network"
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
			session       *clefSession
			networkClient *qrlclient.Client
		)

		ginkgo.BeforeAll(func(ctx ginkgo.SpecContext) {
			live, err := network.Inspect(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			networkClient, err = qrlclient.DialContext(ctx, live.RPCURL)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			ginkgo.DeferCleanup(networkClient.Close)

			chainID, err := networkClient.ChainID(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			workDir := ginkgo.GinkgoT().TempDir()
			clefPath := filepath.Join(workDir, "clef")
			ginkgo.By("building the current Clef binary")
			gomega.Expect(build.Binary(ctx, "./cmd/clef", clefPath)).To(gomega.Succeed())

			developmentWallet, err := network.UnsafeDevelopmentWallet()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			session, err = newClefSession(
				ctx,
				context.WithoutCancel(ctx),
				clefPath,
				workDir,
				chainID,
				developmentWallet,
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

		ginkgo.It("signs and verifies plain-text data", func(ctx ginkgo.SpecContext) {
			gomega.Expect(
				verifyDataSigning(ctx, session.client, session.account, session.expectedWallet),
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

		ginkgo.It("signs, submits, and confirms a transaction", func(ctx ginkgo.SpecContext) {
			nonce, err := networkClient.PendingNonceAt(ctx, session.account)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			tip, err := networkClient.SuggestGasTipCap(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			feeCap, err := networkClient.SuggestGasPrice(ctx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			request := transactionArgs(session.account, session.chainID, nonce, tip, feeCap)
			signed, err := signTransaction(ctx, session.client, request)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(
				verifyTransaction(signed, request, session.account, session.expectedWallet),
			).To(gomega.Succeed())

			gomega.Expect(networkClient.SendTransaction(ctx, signed.Tx)).To(gomega.Succeed())
			receipt, err := qrlbind.WaitMined(ctx, networkClient, signed.Tx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(receipt.Status).To(gomega.Equal(types.ReceiptStatusSuccessful))
			gomega.Expect(receipt.TxHash).To(gomega.Equal(signed.Tx.Hash()))
		}, ginkgo.SpecTimeout(liveSpecTimeout))
	},
)
