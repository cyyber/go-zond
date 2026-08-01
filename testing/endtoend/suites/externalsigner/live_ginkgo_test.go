// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package externalsigner

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"time"

	qrl "github.com/theQRL/go-qrl"
	"github.com/theQRL/go-qrl/accounts"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/crypto/pqcrypto"
	qrlwallet "github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
	"github.com/theQRL/go-qrl/internal/qrlapi"
	"github.com/theQRL/go-qrl/testing/devnet"
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

		ginkgo.It("propagates Clef signing rejection through the node", func(ctx ginkgo.SpecContext) {
			var signature hexutil.Bytes
			err := suite.session.Client.Client().CallContext(
				ctx,
				&signature,
				"qrl_sign",
				suite.account,
				hexutil.Bytes(fixture.RemoteSignerRejectedText),
			)
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("request denied")))
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
			expectTransactionMatchesArgs(signed.Tx, args)

			var decoded types.Transaction
			gomega.Expect(decoded.UnmarshalBinary(signed.Raw)).To(gomega.Succeed())
			gomega.Expect(decoded.Hash()).To(gomega.Equal(signed.Tx.Hash()))
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("propagates Clef transaction rejection through the node", func(ctx ginkgo.SpecContext) {
			before, err := suite.session.Client.PendingNonceAt(ctx, suite.account)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			args := suite.transactionArgs(ctx)
			args.Value = (*hexutil.Big)(big.NewInt(fixture.RemoteSignerRejectedTransaction))

			var signed qrlapi.SignTransactionResult
			err = suite.session.Client.Client().CallContext(
				ctx,
				&signed,
				"qrl_signTransaction",
				args,
			)
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("request denied")))

			var hash common.Hash
			err = suite.session.Client.Client().CallContext(
				ctx,
				&hash,
				"qrl_sendTransaction",
				args,
			)
			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("request denied")))

			expectNoTransactionAtNonce(ctx, suite, before)
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
			expectTransactionMatchesArgs(tx, args)
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("does not submit a transaction after signing is canceled", func(ctx ginkgo.SpecContext) {
			before, err := suite.session.Client.PendingNonceAt(ctx, suite.account)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			args := suite.transactionArgs(ctx)
			args.Value = (*hexutil.Big)(big.NewInt(fixture.RemoteSignerDelayedTransaction))
			requestCtx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()

			var hash common.Hash
			err = suite.session.Client.Client().CallContext(
				requestCtx,
				&hash,
				"qrl_sendTransaction",
				args,
			)
			gomega.Expect(err).To(gomega.MatchError(context.DeadlineExceeded))

			expectNoTransactionAtNonce(ctx, suite, before)
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("fails while Clef is unavailable and recovers after restart", func(ctx ginkgo.SpecContext) {
			gomega.Expect(clefService(ctx, "stop")).To(gomega.Succeed())
			stopped := true
			defer func() {
				if stopped {
					_ = clefService(context.Background(), "start")
				}
			}()

			requestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			var signature hexutil.Bytes
			err := suite.session.Client.Client().CallContext(
				requestCtx,
				&signature,
				"qrl_sign",
				suite.account,
				hexutil.Bytes("Clef unavailable"),
			)
			gomega.Expect(err).To(gomega.HaveOccurred())

			gomega.Expect(clefService(ctx, "start")).To(gomega.Succeed())
			stopped = false
			gomega.Eventually(func() error {
				var managed []common.Address
				if err := suite.session.Client.Client().CallContext(ctx, &managed, "qrl_accounts"); err != nil {
					return err
				}
				if len(managed) != 1 || managed[0] != suite.account {
					return fmt.Errorf("unexpected managed accounts after restart: %v", managed)
				}
				return nil
			}).WithContext(ctx).WithTimeout(time.Minute).WithPolling(time.Second).Should(gomega.Succeed())
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("reconnects to Clef after the signer restarts", func(ctx ginkgo.SpecContext) {
			gomega.Expect(restartClef(ctx)).To(gomega.Succeed())

			message := []byte("go-qrl external signer restart E2E")
			gomega.Eventually(func() error {
				var managed []common.Address
				if err := suite.session.Client.Client().CallContext(ctx, &managed, "qrl_accounts"); err != nil {
					return err
				}
				if len(managed) != 1 || managed[0] != suite.account {
					return fmt.Errorf("unexpected managed accounts after restart: %v", managed)
				}

				var signature hexutil.Bytes
				if err := suite.session.Client.Client().CallContext(
					ctx,
					&signature,
					"qrl_sign",
					suite.account,
					hexutil.Bytes(message),
				); err != nil {
					return err
				}
				valid, err := pqcrypto.MLDSA87VerifySignature(
					signature,
					accounts.TextHash(message),
					suite.wallet.GetPK(),
					suite.wallet.GetDescriptor(),
				)
				if err != nil {
					return err
				}
				if !valid {
					return fmt.Errorf("invalid signature after Clef restart")
				}
				return nil
			}).WithContext(ctx).WithTimeout(time.Minute).WithPolling(time.Second).Should(gomega.Succeed())

			args := suite.transactionArgs(ctx)
			var signed qrlapi.SignTransactionResult
			gomega.Eventually(func() error {
				return suite.session.Client.Client().CallContext(
					ctx,
					&signed,
					"qrl_signTransaction",
					args,
				)
			}).WithContext(ctx).WithTimeout(time.Minute).WithPolling(time.Second).Should(gomega.Succeed())
			gomega.Expect(signed.Tx).NotTo(gomega.BeNil())
			gomega.Expect(transactionSender(signed.Tx, suite.session.ChainID)).To(
				gomega.Equal(suite.account),
			)
			expectTransactionMatchesArgs(signed.Tx, args)
		}, ginkgo.SpecTimeout(liveSpecTimeout))
	},
)

func restartClef(ctx context.Context) error {
	for _, action := range []string{"stop", "start"} {
		if err := clefService(ctx, action); err != nil {
			return err
		}
	}
	return nil
}

func clefService(ctx context.Context, action string) error {
	enclave := os.Getenv("DEVNET_ENCLAVE_NAME")
	if enclave == "" {
		enclave = devnet.DefaultEnclaveName
	}
	output, err := exec.CommandContext(
		ctx,
		"kurtosis",
		"service",
		action,
		enclave,
		"signer-clef",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s Clef service: %w: %s", action, err, output)
	}
	return nil
}

func newLiveSuite(ctx context.Context) *liveSuite {
	ginkgo.GinkgoHelper()

	session, err := endtoendlive.Open(ctx, false)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	wallet, err := qrlwallet.RestoreFromSeedHex(fixture.RemoteSignerSeed)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	account := common.Address(wallet.GetAddress())

	return &liveSuite{session: session, wallet: wallet, account: account}
}

func (suite *liveSuite) transactionArgs(ctx context.Context) qrlapi.TransactionArgs {
	ginkgo.GinkgoHelper()

	nonce, err := suite.session.Client.PendingNonceAt(ctx, suite.account)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	feeCap, err := suite.session.Client.SuggestGasPrice(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	tipCap, err := suite.session.Client.SuggestGasTipCap(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	input := make(hexutil.Bytes, 65)
	for index := range input {
		input[index] = byte(index + 1)
	}
	accessList := types.AccessList{{
		Address:     suite.session.Address,
		StorageKeys: []common.Hash{{0x01}},
	}}
	gas, err := suite.session.Client.EstimateGas(ctx, qrl.CallMsg{
		From:       suite.account,
		To:         &suite.session.Address,
		Data:       input,
		AccessList: accessList,
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	from := suite.account
	to := suite.session.Address
	gasLimit := hexutil.Uint64(gas)
	nonceValue := hexutil.Uint64(nonce)
	return qrlapi.TransactionArgs{
		From:                 &from,
		To:                   &to,
		Gas:                  &gasLimit,
		MaxFeePerGas:         (*hexutil.Big)(feeCap),
		MaxPriorityFeePerGas: (*hexutil.Big)(tipCap),
		Value:                (*hexutil.Big)(big.NewInt(1)),
		Nonce:                &nonceValue,
		Input:                &input,
		AccessList:           &accessList,
		ChainID:              (*hexutil.Big)(suite.session.ChainID),
	}
}

func expectTransactionMatchesArgs(tx *types.Transaction, args qrlapi.TransactionArgs) {
	ginkgo.GinkgoHelper()

	gomega.Expect(tx.To()).NotTo(gomega.BeNil())
	gomega.Expect(*tx.To()).To(gomega.Equal(*args.To))
	gomega.Expect(tx.Nonce()).To(gomega.Equal(uint64(*args.Nonce)))
	gomega.Expect(tx.Gas()).To(gomega.Equal(uint64(*args.Gas)))
	gomega.Expect(tx.GasFeeCap()).To(gomega.Equal(args.MaxFeePerGas.ToInt()))
	gomega.Expect(tx.GasTipCap()).To(gomega.Equal(args.MaxPriorityFeePerGas.ToInt()))
	gomega.Expect(tx.Value()).To(gomega.Equal(args.Value.ToInt()))
	gomega.Expect(tx.ChainId()).To(gomega.Equal(args.ChainID.ToInt()))
	gomega.Expect(args.Input).NotTo(gomega.BeNil())
	gomega.Expect(tx.Data()).To(gomega.Equal([]byte(*args.Input)))
	gomega.Expect(args.AccessList).NotTo(gomega.BeNil())
	gomega.Expect(tx.AccessList()).To(gomega.Equal(*args.AccessList))
}

func expectNoTransactionAtNonce(ctx context.Context, suite *liveSuite, nonce uint64) {
	ginkgo.GinkgoHelper()

	gomega.Consistently(func() uint64 {
		pending, err := suite.session.Client.PendingNonceAt(ctx, suite.account)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return pending
	}).WithContext(ctx).WithTimeout(5 * time.Second).WithPolling(time.Second).Should(
		gomega.Equal(nonce),
	)

	var content map[string]map[string]*qrlapi.RPCTransaction
	gomega.Expect(suite.session.Client.Client().CallContext(
		ctx,
		&content,
		"txpool_contentFrom",
		suite.account,
	)).To(gomega.Succeed())
	gomega.Expect(content["pending"]).NotTo(gomega.HaveKey(fmt.Sprint(nonce)))
	gomega.Expect(content["queued"]).NotTo(gomega.HaveKey(fmt.Sprint(nonce)))
}

func transactionSender(tx *types.Transaction, chainID *big.Int) common.Address {
	ginkgo.GinkgoHelper()

	from, err := types.Sender(types.LatestSignerForChainID(chainID), tx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return from
}
