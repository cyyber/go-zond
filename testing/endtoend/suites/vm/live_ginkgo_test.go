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
	"github.com/theQRL/go-qrl/crypto"
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

		ginkgo.It("executes 512-bit arithmetic, signed, shift, byte, and sign-extension boundaries", func(ctx ginkgo.SpecContext) {
			modulus := new(big.Int).Lsh(big.NewInt(1), qrvm.WordBits)
			max := new(big.Int).Sub(new(big.Int).Set(modulus), big.NewInt(1))
			minSigned := new(big.Int).Lsh(big.NewInt(1), qrvm.WordBits-1)
			zero := []byte{0}
			one := []byte{1}
			shift511 := new(big.Int).SetUint64(qrvm.WordBits - 1).Bytes()

			cases := []struct {
				name     string
				op       qrvm.OpCode
				operands [][]byte
				want     *big.Int
			}{
				{"ADD wraps at 512 bits", qrvm.ADD, [][]byte{vmWord(max), one}, new(big.Int)},
				{"SUB wraps at 512 bits", qrvm.SUB, [][]byte{one, zero}, max},
				{"SLT compares the signed minimum", qrvm.SLT, [][]byte{zero, vmWord(minSigned)}, big.NewInt(1)},
				{"SGT compares minus one", qrvm.SGT, [][]byte{vmWord(max), zero}, big.NewInt(1)},
				{"SHL reaches bit 511", qrvm.SHL, [][]byte{one, shift511}, minSigned},
				{"SHR crosses the upper 256 bits", qrvm.SHR, [][]byte{vmWord(minSigned), shift511}, big.NewInt(1)},
				{"SAR preserves the sign", qrvm.SAR, [][]byte{vmWord(minSigned), shift511}, max},
			}
			for _, test := range cases {
				ginkgo.By(test.name)
				gomega.Expect(suite.callCode(ctx, operationCode(test.op, test.operands...), nil)).To(
					gomega.Equal(vmWord(test.want)),
				)
			}

			pattern := patternedBytes(qrvm.WordBytes)
			for _, test := range []struct {
				index byte
				want  byte
			}{{0, pattern[0]}, {qrvm.WordBytes - 1, pattern[qrvm.WordBytes-1]}, {qrvm.WordBytes, 0}} {
				gomega.Expect(suite.callCode(ctx, operationCode(
					qrvm.BYTE,
					pattern,
					[]byte{test.index},
				), nil)).To(gomega.Equal(vmWord(new(big.Int).SetUint64(uint64(test.want)))))
			}

			for _, index := range []uint{31, 32} {
				signBit := index*8 + 7
				value := new(big.Int).Lsh(big.NewInt(1), signBit)
				upper := new(big.Int).Sub(
					new(big.Int).Set(modulus),
					new(big.Int).Lsh(big.NewInt(1), signBit+1),
				)
				want := new(big.Int).Or(value, upper)
				gomega.Expect(suite.callCode(ctx, operationCode(
					qrvm.SIGNEXTEND,
					vmWord(value),
					[]byte{byte(index)},
				), nil)).To(gomega.Equal(vmWord(want)))
			}
			gomega.Expect(suite.callCode(ctx, operationCode(
				qrvm.SIGNEXTEND,
				vmWord(minSigned),
				[]byte{qrvm.WordBytes - 1},
			), nil)).To(gomega.Equal(vmWord(minSigned)))
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("loads and stores full 64-byte memory words", func(ctx ginkgo.SpecContext) {
			value := make([]byte, qrvm.WordBytes)
			for index := range value {
				value[index] = byte(index + 1)
			}
			gomega.Expect(suite.callCode(ctx, memoryCode(value), nil)).To(gomega.Equal(value))
			for _, offset := range []byte{1, qrvm.WordBytes - 1} {
				gomega.Expect(suite.callCode(ctx, memoryCodeAt(value, offset), nil)).To(gomega.Equal(value))
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("stores and loads a full-width storage value", func(ctx ginkgo.SpecContext) {
			key := patternedBytes(common.HashLength)
			value := patternedBytes(qrvm.WordBytes)
			gomega.Expect(suite.callCode(ctx, storageRoundTripCode(key, value), nil)).To(
				gomega.Equal(value),
			)
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("does not treat JUMPDEST bytes inside PUSH33 through PUSH64 as code", func(ctx ginkgo.SpecContext) {
			for width := 33; width <= qrvm.WordBytes; width++ {
				_, err := suite.callCodeResult(ctx, jumpDestinationCode(width, true), nil, nil)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(suite.callCode(ctx, jumpDestinationCode(width, false), nil)).To(
					gomega.Equal(vmWord(big.NewInt(1))),
				)
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("copies calldata across 64-byte word boundaries", func(ctx ginkgo.SpecContext) {
			for _, size := range []int{63, 64, 65} {
				input := patternedBytes(size)
				gomega.Expect(suite.callCodeWithInput(ctx, echoCalldataCode(), input, nil)).To(
					gomega.Equal(input),
				)
				for _, offset := range []int{0, 1} {
					want := make([]byte, qrvm.WordBytes)
					if offset < len(input) {
						copy(want, input[offset:])
					}
					gomega.Expect(suite.callCodeWithInput(
						ctx,
						calldataLoadCode(byte(offset)),
						input,
						nil,
					)).To(gomega.Equal(want))
				}
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("copies code and return data across 64-byte word boundaries", func(ctx ginkgo.SpecContext) {
			callee := common.BytesToAddress([]byte{0xf2})
			for _, size := range []int{63, 64, 65} {
				input := patternedBytes(size)
				gomega.Expect(suite.callCode(ctx, codeCopyCode(input), nil)).To(gomega.Equal(input))

				overrides := qrlapi.StateOverride{callee: codeOverride(input)}
				gomega.Expect(suite.callCode(
					ctx,
					extCodeCopyCode(callee, byte(size)),
					overrides,
				)).To(gomega.Equal(input))

				overrides[callee] = codeOverride(codeCopyCode(input))
				gomega.Expect(suite.callCode(ctx, returnDataCopyCode(callee), overrides)).To(
					gomega.Equal(input),
				)
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("hashes memory across 64-byte word boundaries", func(ctx ginkgo.SpecContext) {
			for _, size := range []int{63, 64, 65} {
				input := patternedBytes(size)
				want := common.LeftPadBytes(crypto.Keccak256(input), qrvm.WordBytes)
				gomega.Expect(suite.callCodeWithInput(ctx, keccakCalldataCode(), input, nil)).To(
					gomega.Equal(want),
				)
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("executes CALL, STATICCALL, and DELEGATECALL", func(ctx ginkgo.SpecContext) {
			callee := common.BytesToAddress([]byte{0xf1})
			overrides := qrlapi.StateOverride{callee: codeOverride(callContextCode())}
			for _, op := range []qrvm.OpCode{qrvm.CALL, qrvm.STATICCALL, qrvm.DELEGATECALL} {
				output := suite.callCode(ctx, callCode(op, callee), overrides)
				gomega.Expect(output).To(gomega.HaveLen(4 * qrvm.WordBytes))

				wantAddress, wantCaller := callee, suite.target
				if op == qrvm.DELEGATECALL {
					wantAddress = suite.target
					wantCaller = suite.session.Address
				}
				gomega.Expect(common.BytesToAddress(output[:qrvm.WordBytes])).To(gomega.Equal(wantAddress))
				gomega.Expect(common.BytesToAddress(output[qrvm.WordBytes : 2*qrvm.WordBytes])).To(gomega.Equal(wantCaller))
				gomega.Expect(new(big.Int).SetBytes(output[2*qrvm.WordBytes : 3*qrvm.WordBytes]).Sign()).To(gomega.BeZero())
				gomega.Expect(new(big.Int).SetBytes(output[3*qrvm.WordBytes:]).Uint64()).To(gomega.Equal(uint64(1)))
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("reverts failed CALL, STATICCALL, and DELEGATECALL effects", func(ctx ginkgo.SpecContext) {
			callee := common.BytesToAddress([]byte{0xf4})
			callerBalance := big.NewInt(100)
			calleeBalance := big.NewInt(5)
			for _, op := range []qrvm.OpCode{qrvm.CALL, qrvm.STATICCALL, qrvm.DELEGATECALL} {
				code := failingCallCode(op, callee)
				overrides := qrlapi.StateOverride{
					suite.target: codeAndBalanceOverride(code, callerBalance),
					callee:       codeAndBalanceOverride(revertingStorageCode(), calleeBalance),
				}
				output := suite.callCode(ctx, code, overrides)
				gomega.Expect(output).To(gomega.HaveLen(3 * qrvm.WordBytes))
				gomega.Expect(new(big.Int).SetBytes(output[:qrvm.WordBytes]).Sign()).To(gomega.BeZero())
				if op == qrvm.CALL {
					gomega.Expect(new(big.Int).SetBytes(output[qrvm.WordBytes : 2*qrvm.WordBytes])).To(
						gomega.Equal(calleeBalance),
					)
				} else {
					gomega.Expect(new(big.Int).SetBytes(output[qrvm.WordBytes : 2*qrvm.WordBytes]).Sign()).To(gomega.BeZero())
				}
				gomega.Expect(new(big.Int).SetBytes(output[2*qrvm.WordBytes:])).To(
					gomega.Equal(callerBalance),
				)
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("executes CREATE and CREATE2", func(ctx ginkgo.SpecContext) {
			for _, op := range []qrvm.OpCode{qrvm.CREATE, qrvm.CREATE2} {
				code, childInit := createCode(op)
				output := suite.callCode(ctx, code, nil)
				gomega.Expect(output).To(gomega.HaveLen(4 * qrvm.WordBytes))
				gomega.Expect(new(big.Int).SetBytes(output[:qrvm.WordBytes]).Uint64()).To(gomega.Equal(uint64(len(returnWordCode([]byte{0x2a})))))

				var wantAddress common.Address
				if op == qrvm.CREATE {
					wantAddress = crypto.CreateAddress(suite.target, 0)
				} else {
					var salt [qrvm.WordBytes]byte
					salt[len(salt)-1] = 1
					initHash := crypto.Keccak256Hash(childInit)
					wantAddress = crypto.CreateAddress2(suite.target, salt, initHash[:])
				}
				gomega.Expect(common.BytesToAddress(output[qrvm.WordBytes : 2*qrvm.WordBytes])).To(gomega.Equal(wantAddress))
				gomega.Expect(new(big.Int).SetBytes(output[2*qrvm.WordBytes : 3*qrvm.WordBytes]).Uint64()).To(gomega.Equal(uint64(0x2a)))
				gomega.Expect(new(big.Int).SetBytes(output[3*qrvm.WordBytes:]).Uint64()).To(gomega.Equal(uint64(1)))
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("reverts failed CREATE and CREATE2 state and value", func(ctx ginkgo.SpecContext) {
			callerBalance := big.NewInt(100)
			initCode := revertingStorageCode()
			for _, op := range []qrvm.OpCode{qrvm.CREATE, qrvm.CREATE2} {
				var child common.Address
				if op == qrvm.CREATE {
					child = crypto.CreateAddress(suite.target, 0)
				} else {
					var salt [qrvm.WordBytes]byte
					salt[len(salt)-1] = 1
					hash := crypto.Keccak256Hash(initCode)
					child = crypto.CreateAddress2(suite.target, salt, hash[:])
				}
				code, encodedInit := failingCreateCode(op, child)
				gomega.Expect(encodedInit).To(gomega.Equal(initCode))
				output := suite.callCode(ctx, code, qrlapi.StateOverride{
					suite.target: codeAndBalanceOverride(code, callerBalance),
				})
				gomega.Expect(output).To(gomega.HaveLen(4 * qrvm.WordBytes))
				for index := 0; index < 3; index++ {
					gomega.Expect(new(big.Int).SetBytes(output[index*qrvm.WordBytes : (index+1)*qrvm.WordBytes]).Sign()).To(gomega.BeZero())
				}
				gomega.Expect(new(big.Int).SetBytes(output[3*qrvm.WordBytes:])).To(
					gomega.Equal(callerBalance),
				)
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("returns full-width address and block context values", func(ctx ginkgo.SpecContext) {
			header, err := suite.session.Client.HeaderByNumber(ctx, nil)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			balance, err := suite.session.Client.BalanceAt(ctx, suite.session.Address, nil)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(suite.callCode(ctx, opcodeCode(qrvm.ORIGIN), nil)).To(
				gomega.Equal(suite.session.Address.Bytes()),
			)
			gomega.Expect(suite.callCode(ctx, opcodeCode(qrvm.COINBASE), nil)).To(
				gomega.Equal(header.Coinbase.Bytes()),
			)
			gomega.Expect(suite.callCode(ctx, addressOpcodeCode(qrvm.BALANCE, suite.session.Address), nil)).To(
				gomega.Equal(vmWord(balance)),
			)

			callee := common.BytesToAddress([]byte{0xf5})
			code := returnWordCode([]byte{0x2a})
			gomega.Expect(suite.callCode(ctx, addressOpcodeCode(qrvm.EXTCODEHASH, callee), qrlapi.StateOverride{
				callee: codeOverride(code),
			})).To(gomega.Equal(vmWord(new(big.Int).SetBytes(crypto.Keccak256(code)))))
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("mines LOG0 through LOG4 with full-width values", func(ctx ginkgo.SpecContext) {
			data := make([]byte, qrvm.WordBytes)
			for index := range data {
				data[index] = byte(index + 1)
			}

			for count := 0; count <= 4; count++ {
				topics := make([]common.LogTopic, count)
				for topicIndex := range topics {
					for byteIndex := range topics[topicIndex] {
						topics[topicIndex][byteIndex] = byte((topicIndex+1)*17 + byteIndex)
					}
				}
				receipt := suite.mineLog(ctx, data, topics)
				gomega.Expect(receipt.Logs).To(gomega.HaveLen(1))
				gomega.Expect(receipt.Logs[0].Topics).To(gomega.Equal(topics))
				gomega.Expect(receipt.Logs[0].Data).To(gomega.Equal(data))
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))
	},
)

func (suite *liveSuite) callCode(
	ctx context.Context,
	code []byte,
	extra qrlapi.StateOverride,
) []byte {
	return suite.callCodeWithInput(ctx, code, nil, extra)
}

func (suite *liveSuite) callCodeWithInput(
	ctx context.Context,
	code []byte,
	input []byte,
	extra qrlapi.StateOverride,
) []byte {
	ginkgo.GinkgoHelper()
	output, err := suite.callCodeResult(ctx, code, input, extra)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return output
}

func (suite *liveSuite) callCodeResult(
	ctx context.Context,
	code []byte,
	input []byte,
	extra qrlapi.StateOverride,
) ([]byte, error) {
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
			"from":  suite.session.Address,
			"to":    suite.target,
			"gas":   hexutil.Uint64(10_000_000),
			"input": hexutil.Bytes(input),
		},
		"latest",
		overrides,
	)
	return output, err
}

func codeOverride(code []byte) qrlapi.OverrideAccount {
	encoded := hexutil.Bytes(code)
	nonce := hexutil.Uint64(0)
	return qrlapi.OverrideAccount{Nonce: &nonce, Code: &encoded}
}

func codeAndBalanceOverride(code []byte, balance *big.Int) qrlapi.OverrideAccount {
	override := codeOverride(code)
	rpcBalance := (*hexutil.Big)(balance)
	override.Balance = &rpcBalance
	return override
}

func vmWord(value *big.Int) []byte {
	return value.FillBytes(make([]byte, qrvm.WordBytes))
}

func (suite *liveSuite) mineLog(ctx context.Context, data []byte, topics []common.LogTopic) *types.Receipt {
	ginkgo.GinkgoHelper()

	auth, err := bind.NewKeyedTransactorWithChainID(suite.session.Wallet, suite.session.ChainID)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	auth.Context = ctx
	auth.NoSend = true
	_, tx, _, err := bind.DeployContract(
		auth,
		abi.ABI{},
		logInitCode(data, topics),
		suite.session.Client,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(suite.session.Client.SendTransaction(ctx, tx)).To(gomega.Succeed())
	receipt, err := bind.WaitMined(ctx, suite.session.Client, tx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(receipt.Status).To(gomega.Equal(types.ReceiptStatusSuccessful))
	return receipt
}
