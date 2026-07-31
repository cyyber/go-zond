// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package vm

import (
	"context"
	"math/big"
	"time"

	"github.com/theQRL/go-qrl/accounts/abi"
	"github.com/theQRL/go-qrl/accounts/abi/bind"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	qrvm "github.com/theQRL/go-qrl/core/vm"
	"github.com/theQRL/go-qrl/internal/qrlapi"
	endtoendlive "github.com/theQRL/go-qrl/testing/endtoend/internal/live"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

const liveSpecTimeout = 5 * time.Minute

type liveSuite struct {
	session *endtoendlive.Session
	target  common.Address
}

var _ = ginkgo.Describe(
	"hand-written QRVM bytecode",
	ginkgo.Serial,
	ginkgo.Ordered,
	ginkgo.ContinueOnFailure,
	ginkgo.Label("e2e", "live", "vm", "mutates-chain"),
	func() {
		var suite *liveSuite

		ginkgo.BeforeAll(func(ctx ginkgo.SpecContext) {
			session, err := endtoendlive.Open(ctx, false)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			ginkgo.DeferCleanup(session.Close)
			suite = &liveSuite{
				session: session,
				target:  common.BytesToAddress([]byte{0xf0}),
			}
		})

		ginkgo.It("executes PUSH33 through PUSH64", func(ctx ginkgo.SpecContext) {
			for width := 33; width <= qrvm.WordBytes; width++ {
				code, want := pushCode(width)
				gomega.Expect(suite.callCode(ctx, code, nil)).To(gomega.Equal(want))
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("executes shifted DUP and SWAP ranges", func(ctx ginkgo.SpecContext) {
			for depth := 1; depth <= 16; depth++ {
				gomega.Expect(suite.callCode(ctx, dupCode(depth), nil)).To(
					gomega.Equal(common.LeftPadBytes([]byte{1}, qrvm.WordBytes)),
				)
				gomega.Expect(suite.callCode(ctx, swapCode(depth), nil)).To(
					gomega.Equal(common.LeftPadBytes([]byte{1}, qrvm.WordBytes)),
				)
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("loads and stores full 64-byte memory words", func(ctx ginkgo.SpecContext) {
			value := make([]byte, qrvm.WordBytes)
			for index := range value {
				value[index] = byte(index + 1)
			}
			gomega.Expect(suite.callCode(ctx, memoryCode(value), nil)).To(gomega.Equal(value))
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("executes CALL, STATICCALL, and DELEGATECALL", func(ctx ginkgo.SpecContext) {
			callee := common.BytesToAddress([]byte{0xf1})
			want := common.LeftPadBytes([]byte{0x2a}, qrvm.WordBytes)
			overrides := qrlapi.StateOverride{callee: codeOverride(returnWordCode([]byte{0x2a}))}
			for _, op := range []qrvm.OpCode{qrvm.CALL, qrvm.STATICCALL, qrvm.DELEGATECALL} {
				gomega.Expect(suite.callCode(ctx, callCode(op, callee), overrides)).To(gomega.Equal(want))
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("executes CREATE and CREATE2", func(ctx ginkgo.SpecContext) {
			for _, op := range []qrvm.OpCode{qrvm.CREATE, qrvm.CREATE2} {
				output := suite.callCode(ctx, createCode(op), nil)
				gomega.Expect(output).To(gomega.HaveLen(2 * qrvm.WordBytes))
				gomega.Expect(new(big.Int).SetBytes(output[:qrvm.WordBytes]).Uint64()).To(gomega.Equal(uint64(1)))
				gomega.Expect(common.BytesToAddress(output[qrvm.WordBytes:])).NotTo(gomega.Equal(common.Address{}))
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("mines a full-width log", func(ctx ginkgo.SpecContext) {
			data := make([]byte, qrvm.WordBytes)
			var topic common.LogTopic
			for index := range data {
				data[index] = byte(index + 1)
				topic[index] = byte(0xff - index)
			}

			auth, err := bind.NewKeyedTransactorWithChainID(suite.session.Wallet, suite.session.ChainID)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			auth.Context = ctx
			auth.NoSend = true
			_, tx, _, err := bind.DeployContract(
				auth,
				abi.ABI{},
				logInitCode(data, topic),
				suite.session.Client,
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(suite.session.Client.SendTransaction(ctx, tx)).To(gomega.Succeed())
			receipt, err := bind.WaitMined(ctx, suite.session.Client, tx)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(receipt.Status).To(gomega.Equal(types.ReceiptStatusSuccessful))
			gomega.Expect(receipt.Logs).To(gomega.HaveLen(1))
			gomega.Expect(receipt.Logs[0].Topics).To(gomega.Equal([]common.LogTopic{topic}))
			gomega.Expect(receipt.Logs[0].Data).To(gomega.Equal(data))
		}, ginkgo.SpecTimeout(liveSpecTimeout))
	},
)

func (suite *liveSuite) callCode(
	ctx context.Context,
	code []byte,
	extra qrlapi.StateOverride,
) []byte {
	ginkgo.GinkgoHelper()

	overrides := qrlapi.StateOverride{suite.target: codeOverride(code)}
	for address, account := range extra {
		overrides[address] = account
	}
	var output hexutil.Bytes
	err := suite.session.Client.Client().CallContext(
		ctx,
		&output,
		"qrl_call",
		map[string]any{
			"from": suite.session.Address,
			"to":   suite.target,
			"gas":  hexutil.Uint64(10_000_000),
		},
		"latest",
		overrides,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return output
}

func codeOverride(code []byte) qrlapi.OverrideAccount {
	encoded := hexutil.Bytes(code)
	return qrlapi.OverrideAccount{Code: &encoded}
}
