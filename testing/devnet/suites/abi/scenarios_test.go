// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.
//
// The go-qrl library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-qrl library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-qrl library. If not, see <http://www.gnu.org/licenses/>.

package abi

import (
	"context"
	_ "embed"
	"errors"
	"math/big"
	"reflect"
	"strings"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	qrl "github.com/theQRL/go-qrl"
	"github.com/theQRL/go-qrl/accounts/abi"
	"github.com/theQRL/go-qrl/accounts/abi/bind"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	qrlmath "github.com/theQRL/go-qrl/common/math"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/crypto"
	"github.com/theQRL/go-qrl/qrlclient"
	"github.com/theQRL/go-qrl/rpc"
	"github.com/theQRL/go-qrl/testing/devnet/internal/network"
)

// Regenerate the source-controlled Hyperion artifacts and generated binding.
// The compiler must be cyyber/hyperion@2b9a0f1d.
//
//go:generate sh -c "hypc --version 2>&1 | grep -Fq commit.2b9a0f1d || { echo 'hypc from cyyber/hyperion@2b9a0f1d is required; found:' >&2; hypc --version >&2; exit 1; }"
//go:generate hypc --abi --bin --no-cbor-metadata --overwrite -o testdata testdata/EventEmitter.hyp
//go:generate go -C ../../../.. run ./cmd/abigen --abi testing/devnet/suites/abi/testdata/EventEmitter.abi --bin testing/devnet/suites/abi/testdata/EventEmitter.bin --pkg abi --type EventEmitter --out testing/devnet/suites/abi/contract.go

//go:embed testdata/EventEmitter.abi
var eventEmitterABIJSON string

type liveSuite struct {
	client      *qrlclient.Client
	wsClient    *qrlclient.Client
	from        common.Address
	signer      bind.SignerFn
	contractABI abi.ABI
	inputs      scenarioInputs
}

func setupLiveSuite(ctx context.Context) *liveSuite {
	ginkgo.GinkgoHelper()

	environment, err := network.Inspect(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	client, err := qrlclient.DialContext(ctx, environment.RPCURL)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(client.Close)

	wsClient, err := qrlclient.DialContext(ctx, environment.WebSocketURL)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(wsClient.Close)

	wallet, err := network.UnsafeDevelopmentWallet()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	chainID, err := client.ChainID(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	transactor, err := bind.NewKeyedTransactorWithChainID(wallet, chainID)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	inputs := scenarioInputs{
		amount: new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 511), big.NewInt(0x1234)),
		delta:  new(big.Int).Add(new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 510)), big.NewInt(42)),
		note:   "VM string crosses the 64-byte ABI word boundary: 0123456789abcdef0123456789abcdef",
	}
	for index := range inputs.tag {
		inputs.tag[index] = byte(0x80 + index)
	}

	inputs.payload = make([]byte, 129)
	for index := range inputs.payload {
		inputs.payload[index] = byte((index*29 + 7) & 0xff)
	}

	parsed, err := abi.JSON(strings.NewReader(eventEmitterABIJSON))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return &liveSuite{
		client:      client,
		wsClient:    wsClient,
		from:        transactor.From,
		signer:      transactor.Signer,
		contractABI: parsed,
		inputs:      inputs,
	}
}

func (suite *liveSuite) deployEventEmitter(ctx context.Context) *liveFixture {
	ginkgo.GinkgoHelper()

	deploymentAuth := suite.transactOpts(ctx)
	initial := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 500), big.NewInt(1337))
	address, tx, binding, err := DeployEventEmitter(deploymentAuth, suite.client, initial)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	receipt := waitSuccessfulTransaction(ctx, suite.client, tx)
	gomega.Expect(receipt.ContractAddress).To(gomega.Equal(address))
	gomega.Expect(receipt.Logs).To(gomega.HaveLen(1))

	deployed := suite.contractABI.Events["Deployed"]
	log := receipt.Logs[0]
	gomega.Expect(log.Topics).To(gomega.Equal([]common.LogTopic{
		common.HashToLogTopic(deployed.ID),
	}))

	deployedValues, err := deployed.Inputs.Unpack(log.Data)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(deployedValues).To(gomega.HaveLen(1))

	deployedValue, ok := deployedValues[0].(*big.Int)
	gomega.Expect(ok).To(gomega.BeTrue(), "deployment event value has type %T", deployedValues[0])
	gomega.Expect(deployedValue.Cmp(initial)).To(gomega.Equal(0))

	return &liveFixture{
		liveSuite:       suite,
		deploymentBlock: receipt.BlockNumber,
		address:         address,
		contract: bind.NewBoundContract(
			address,
			suite.contractABI,
			suite.client,
			suite.client,
			suite.client,
		),
		binding: binding,
		initial: initial,
	}
}

type liveFixture struct {
	*liveSuite
	deploymentBlock *big.Int
	address         common.Address
	contract        *bind.BoundContract
	binding         *EventEmitter
	initial         *big.Int
}

type scenarioInputs struct {
	// Large uint512 value with upper-half bits set.
	amount *big.Int

	// Negative int512 value exercising signed 64-byte encoding.
	delta *big.Int

	// Fully populated bytes64 value.
	tag [64]byte

	// 129-byte dynamic value spanning three ABI data words.
	payload []byte

	// Dynamic string crossing the 64-byte ABI word boundary.
	note string
}

type eventExpectation struct {
	name        string
	log         types.Log
	data        []any
	exactTopics []common.LogTopic
	want        map[string]any
	filter      [][]any
	reject      [][]any
}

func (suite *liveSuite) transactOpts(ctx context.Context) *bind.TransactOpts {
	return &bind.TransactOpts{
		From:    suite.from,
		Signer:  suite.signer,
		Context: ctx,
	}
}

func waitTransaction(
	ctx context.Context,
	client *qrlclient.Client,
	tx *types.Transaction,
) *types.Receipt {
	ginkgo.GinkgoHelper()

	receipt, err := bind.WaitMined(ctx, client, tx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "wait for transaction %s", tx.Hash())
	gomega.Expect(receipt).NotTo(gomega.BeNil(), "transaction %s has no mined receipt", tx.Hash())
	gomega.Expect(receipt.BlockNumber).NotTo(gomega.BeNil(), "transaction %s has no block number", tx.Hash())
	return receipt
}

func waitSuccessfulTransaction(
	ctx context.Context,
	client *qrlclient.Client,
	tx *types.Transaction,
) *types.Receipt {
	ginkgo.GinkgoHelper()

	receipt := waitTransaction(ctx, client, tx)
	gomega.Expect(receipt.Status).To(
		gomega.Equal(types.ReceiptStatusSuccessful),
		"transaction %s status",
		tx.Hash(),
	)
	return receipt
}

func (fixture *liveFixture) callRevertData(
	ctx context.Context,
	method string,
	args ...any,
) []byte {
	ginkgo.GinkgoHelper()

	var output []any
	callErr := fixture.contract.Call(
		&bind.CallOpts{Context: ctx, BlockNumber: fixture.deploymentBlock},
		&output,
		method,
		args...,
	)
	gomega.Expect(callErr).To(gomega.HaveOccurred(), "%s unexpectedly succeeded", method)
	var dataError rpc.DataError
	gomega.Expect(errors.As(callErr, &dataError)).To(
		gomega.BeTrue(),
		"%s returned %T, want rpc.DataError",
		method,
		callErr,
	)
	encoded, ok := dataError.ErrorData().(string)
	gomega.Expect(ok).To(
		gomega.BeTrue(),
		"%s revert data has type %T, want hex string",
		method,
		dataError.ErrorData(),
	)
	revertData, err := hexutil.Decode(encoded)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "decode %s revert data %q", method, encoded)
	return revertData
}

func makeFunctionValue(address common.Address, selector []byte) [common.AddressLength + 4]byte {
	var value [common.AddressLength + 4]byte
	copy(value[:common.AddressLength], address[:])
	copy(value[common.AddressLength:], selector)
	return value
}

// assertCall checks BoundContract decoding and independently proves that
// the compiler returned the canonical ABI bytes.
func (fixture *liveFixture) assertCall(
	ctx context.Context,
	method string,
	args, want []any,
) {
	ginkgo.GinkgoHelper()

	var decoded []any
	err := fixture.contract.Call(
		&bind.CallOpts{Context: ctx, BlockNumber: fixture.deploymentBlock},
		&decoded,
		method,
		args...,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "%s through BoundContract", method)
	wantOutput, err := fixture.contractABI.Methods[method].Outputs.Pack(want...)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "pack canonical %s output", method)
	repacked, err := fixture.contractABI.Methods[method].Outputs.Pack(decoded...)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "repack BoundContract %s output", method)
	gomega.Expect(repacked).To(gomega.Equal(wantOutput), "BoundContract %s output", method)
	input, err := fixture.contractABI.Pack(method, args...)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "pack %s input", method)
	raw, err := fixture.client.CallContract(
		ctx,
		qrl.CallMsg{From: fixture.from, To: &fixture.address, Data: input},
		fixture.deploymentBlock,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "raw %s call", method)
	gomega.Expect(raw).To(gomega.Equal(wantOutput), "compiler %s output", method)
}

func (fixture *liveFixture) assertEvent(
	ctx context.Context,
	expectation eventExpectation,
) {
	ginkgo.GinkgoHelper()

	definition, ok := fixture.contractABI.Events[expectation.name]
	gomega.Expect(ok).To(gomega.BeTrue(), "ABI has no event %q", expectation.name)
	wantData, err := definition.Inputs.NonIndexed().Pack(expectation.data...)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "pack canonical %s data", expectation.name)
	gomega.Expect(expectation.log.Data).To(gomega.Equal(wantData), "%s data", expectation.name)
	gomega.Expect(expectation.log.Topics).To(gomega.Equal(expectation.exactTopics), "%s topics", expectation.name)

	assertDecoded := func(log types.Log) {
		ginkgo.GinkgoHelper()
		have := make(map[string]any)
		err := fixture.contract.UnpackLogIntoMap(have, expectation.name, log)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(have).To(gomega.Equal(expectation.want), "%s decode", expectation.name)
	}
	assertDecoded(expectation.log)

	filter := func(rules [][]any) []types.Log {
		ginkgo.GinkgoHelper()
		topics, err := abi.MakeTopics(append([][]any{{definition.ID}}, rules...)...)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		block := new(big.Int).SetUint64(expectation.log.BlockNumber)
		logs, err := fixture.client.FilterLogs(ctx, qrl.FilterQuery{
			FromBlock: block,
			ToBlock:   block,
			Addresses: []common.Address{expectation.log.Address},
			Topics:    topics,
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return logs
	}
	filtered := filter(expectation.filter)
	gomega.Expect(filtered).To(gomega.HaveLen(1))
	gomega.Expect(filtered[0].TxHash).To(gomega.Equal(expectation.log.TxHash))
	assertDecoded(filtered[0])
	if expectation.reject != nil {
		gomega.Expect(filter(expectation.reject)).To(gomega.BeEmpty())
	}
}

func (fixture *liveFixture) assertCallRoundTrips(ctx context.Context) {
	ginkgo.GinkgoHelper()

	inputs := fixture.inputs

	// Hyperion:
	// function echo(uint512, int512, bytes64, address, bytes, string, bool)
	//     external pure returns (uint512, int512, bytes64, address, bytes, string, bool);
	// Goal: generic ABI decoding and raw RPC return exactly the values sent.
	ginkgo.By("round-tripping scalar and dynamic values through generic ABI and raw RPC")
	echoValues := []any{
		inputs.amount,
		inputs.delta,
		inputs.tag,
		fixture.from,
		inputs.payload,
		inputs.note,
		true,
	}
	fixture.assertCall(ctx, "echo", echoValues, echoValues)

	callOpts := &bind.CallOpts{
		Context:     ctx,
		From:        fixture.from,
		BlockNumber: fixture.deploymentBlock,
	}

	// Hyperion:
	// function echoNested(DynamicRecord, DynamicRecord[], uint16[][][])
	//     external pure returns (DynamicRecord, DynamicRecord[], uint16[][][]);
	// Goal: generated bindings preserve nested tuples, arrays, and empty values.
	ginkgo.By("round-tripping nested tuples and arrays through generated bindings")
	nested := EventEmitterDynamicRecord{
		Amount:  inputs.amount,
		Note:    inputs.note,
		Payload: inputs.payload,
		Values:  [][]uint16{{1, 2}, {}, {3}},
	}
	records := []EventEmitterDynamicRecord{
		nested,
		{Amount: new(big.Int), Note: "", Payload: []byte{}, Values: [][]uint16{}},
	}
	cube := [][][]uint16{{{1}, {2, 3}}, {}, {{4}}}
	gotNested, gotRecords, gotCube, err := fixture.binding.EchoNested(
		callOpts,
		nested,
		records,
		cube,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	outputs := fixture.contractABI.Methods["echoNested"].Outputs
	want, err := outputs.Pack(nested, records, cube)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "pack canonical echoNested output")
	got, err := outputs.Pack(gotNested, gotRecords, gotCube)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "repack generated echoNested output")
	gomega.Expect(got).To(gomega.Equal(want), "generated echoNested output")

	// Hyperion:
	// function echoBoundaries(uint8, int8, uint256, int256, bytes5, uint16[3], string[2], uint16[][2])
	//     external pure returns (uint8, int8, uint256, int256, bytes5, uint16[3], string[2], uint16[][2]);
	// Goal: generated bindings preserve fixed/dynamic values plus uint256 zero-extension
	// and int256 sign-extension in 64-byte ABI words.
	ginkgo.By("round-tripping integer widths and fixed/dynamic boundary values through generated bindings")
	wideUnsigned := new(big.Int).Add(
		new(big.Int).Lsh(big.NewInt(1), 255),
		big.NewInt(0x1234),
	)
	wideSigned := new(big.Int).Add(
		new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 254)),
		big.NewInt(42),
	)
	shortBytes := [5]byte{0x00, 0x7f, 0x80, 0xfe, 0xff}
	fixedNumbers := [3]uint16{0, 0xffff, 0x1234}
	fixedStrings := [2]string{"", inputs.note}
	mixed := [2][]uint16{{}, {1, 0xffff, 0x1234}}
	smallUnsigned, smallSigned := uint8(0xff), int8(-128)
	gotUnsigned, gotSigned, gotWideUnsigned, gotWideSigned,
		gotShortBytes, gotFixedNumbers, gotFixedStrings, gotMixed, err :=
		fixture.binding.EchoBoundaries(
			callOpts,
			smallUnsigned,
			smallSigned,
			wideUnsigned,
			wideSigned,
			shortBytes,
			fixedNumbers,
			fixedStrings,
			mixed,
		)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect([]any{
		gotUnsigned,
		gotSigned,
		gotWideUnsigned,
		gotWideSigned,
		gotShortBytes,
		gotFixedNumbers,
		gotFixedStrings,
		gotMixed,
	}).To(gomega.Equal([]any{
		smallUnsigned,
		smallSigned,
		wideUnsigned,
		wideSigned,
		shortBytes,
		fixedNumbers,
		fixedStrings,
		mixed,
	}))

	// Hyperion: function observe() external view returns (uint512 value, address caller);
	// Goal: the generated view returns the constructor state and original caller.
	ginkgo.By("reading constructor state through a generated view")
	observed, err := fixture.binding.Observe(callOpts)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(observed.Value).To(gomega.Equal(fixture.initial))
	gomega.Expect(observed.Caller).To(gomega.Equal(fixture.from))

	// Hyperion:
	// function transform(string) external pure returns (string);
	// function transform(uint16) external pure returns (uint16);
	// Goal: ABI lookup and generated wrappers select the correct overload.
	ginkgo.By("resolving overloaded methods through their generated wrappers")
	stringMethod := fixture.contractABI.Methods["transform"]
	integerMethod := fixture.contractABI.Methods["transform0"]
	gomega.Expect(stringMethod.Sig).To(gomega.Equal("transform(string)"))
	gomega.Expect(integerMethod.Sig).To(gomega.Equal("transform(uint16)"))
	transformedString, err := fixture.binding.Transform(callOpts, inputs.note)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(transformedString).To(gomega.Equal(inputs.note))
	transformedInteger, err := fixture.binding.Transform0(callOpts, 0xfffe)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(transformedInteger).To(gomega.Equal(uint16(0xffff)))
}

func (fixture *liveFixture) assertErrors(ctx context.Context) {
	ginkgo.GinkgoHelper()

	inputs := fixture.inputs
	record := EventEmitterRecord{
		Amount:    inputs.amount,
		Recipient: fixture.from,
		Tag:       inputs.tag,
	}
	complexArguments := []any{
		inputs.amount,
		"unique custom error across a VM word: " + inputs.note,
		inputs.payload,
		record,
		[][]uint16{{}, {1, 0xffff}, {0x1234}},
	}

	// Hyperion:
	// error ComplexFailure(uint512, string, bytes, Record, uint16[][]);
	// function failComplex(...) external pure {
	//     revert ComplexFailure(code, reason, payload, record, nested);
	// }
	// Goal: RPC revert data equals the four-byte selector followed by the
	// canonical VM encoding of every custom-error argument; selector lookup,
	// decoding, and re-encoding then reproduce the same payload.
	ginkgo.By("round-tripping a complex custom error through raw revert data and selector-based ABI decoding")
	definition, ok := fixture.contractABI.Errors["ComplexFailure"]
	gomega.Expect(ok).To(gomega.BeTrue(), "ABI has no ComplexFailure error")
	signature := definition.Sig
	revertData := fixture.callRevertData(ctx, "failComplex", complexArguments...)
	encodedArguments, err := definition.Inputs.Pack(complexArguments...)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "pack %s", signature)
	wantRevertData := append([]byte{}, definition.ID[:4]...)
	wantRevertData = append(wantRevertData, encodedArguments...)
	gomega.Expect(revertData).To(gomega.Equal(wantRevertData), "%s compiler revert", signature)

	var errorSelector [4]byte
	copy(errorSelector[:], revertData)
	resolvedError, err := fixture.contractABI.ErrorByID(errorSelector)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "ErrorByID(%s)", signature)
	gomega.Expect(resolvedError.Sig).To(gomega.Equal(signature))
	decoded, err := resolvedError.Unpack(revertData)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "decode %s", signature)
	decodedArguments, ok := decoded.([]any)
	gomega.Expect(ok).To(gomega.BeTrue(), "decoded %s has type %T", signature, decoded)
	repackedArguments, err := resolvedError.Inputs.Pack(decodedArguments...)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "repack %s", signature)
	gomega.Expect(repackedArguments).To(
		gomega.Equal(encodedArguments),
		"decoded %s did not round-trip",
		signature,
	)

	// Hyperion:
	// function failReason() external pure { revert("VM standard revert reason"); }
	// function failPanic() external pure { assert(false); }
	// Goal: standard Error(string) and Panic(uint256) payloads decode to their
	// human-readable reason.
	ginkgo.By("decoding standard Error(string) and Panic(uint256) revert payloads")
	for _, standardError := range []struct {
		method string
		want   string
	}{
		{method: "failReason", want: "VM standard revert reason"},
		{method: "failPanic", want: "assert(false)"},
	} {
		reason, err := abi.UnpackRevert(fixture.callRevertData(ctx, standardError.method))
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "decode %s", standardError.method)
		gomega.Expect(reason).To(gomega.Equal(standardError.want), standardError.method)
	}

	// Hyperion:
	// function failReason() external pure { revert("VM standard revert reason"); }
	// Goal: submitting the reverting call as a transaction produces a mined
	// receipt whose status is explicitly failed.
	ginkgo.By("requiring a failed receipt for a reverting transaction")
	auth := fixture.transactOpts(ctx)
	auth.GasLimit = 1_000_000
	failedTx, err := fixture.contract.Transact(auth, "failReason")
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "submit reverting transaction")
	failedReceipt := waitTransaction(ctx, fixture.client, failedTx)
	gomega.Expect(failedReceipt.Status).To(
		gomega.Equal(types.ReceiptStatusFailed),
		"reverting transaction %s status",
		failedTx.Hash(),
	)

}

func (fixture *liveFixture) assertEventsAndFilters(ctx context.Context) {
	ginkgo.GinkgoHelper()

	// Hyperion:
	// event Stored(
	//     address indexed recipient,
	//     uint512 indexed amount,
	//     int512 indexed delta,
	//     bytes64 tag,
	//     bytes payload,
	//     string note,
	//     bool enabled
	// );
	// function store(...) external {
	//     emit Stored(recipient, amount, delta, tag, payload, note, enabled);
	// }
	// Goal: the generated transaction emits the exact VM topics and data, and
	// generated plus raw filters recover the same event while honoring OR,
	// wildcard, and rejection rules.
	ginkgo.By("round-tripping a Stored event through generated transactions, decoding, topics, and filters")
	auth := fixture.transactOpts(ctx)
	inputs := fixture.inputs
	storeTx, err := fixture.binding.Store(
		auth,
		inputs.amount,
		inputs.delta,
		inputs.tag,
		auth.From,
		inputs.payload,
		inputs.note,
		true,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	receipt := waitSuccessfulTransaction(ctx, fixture.client, storeTx)
	gomega.Expect(receipt.Logs).To(gomega.HaveLen(2))

	end := receipt.BlockNumber.Uint64()
	filterOpts := &bind.FilterOpts{Start: end, End: &end, Context: ctx}
	wrongRecipient := auth.From
	wrongRecipient[0] ^= 0xff
	iterator, err := fixture.binding.FilterStored(
		filterOpts,
		[]common.Address{wrongRecipient, auth.From},
		nil,
		[]*big.Int{inputs.delta},
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer iterator.Close()
	gomega.Expect(iterator.Next()).To(gomega.BeTrue(), "generated Stored OR/wildcard filter missed the transaction")
	stored := iterator.Event
	gomega.Expect(stored.Recipient).To(gomega.Equal(auth.From))
	gomega.Expect(stored.Amount.Cmp(inputs.amount)).To(gomega.Equal(0))
	gomega.Expect(stored.Delta.Cmp(inputs.delta)).To(gomega.Equal(0))
	gomega.Expect(stored.Tag).To(gomega.Equal(inputs.tag))
	gomega.Expect(stored.Payload).To(gomega.Equal(inputs.payload))
	gomega.Expect(stored.Note).To(gomega.Equal(inputs.note))
	gomega.Expect(stored.Enabled).To(gomega.BeTrue())
	gomega.Expect(stored.Raw.TxHash).To(gomega.Equal(receipt.TxHash))
	gomega.Expect(iterator.Next()).To(gomega.BeFalse())
	gomega.Expect(iterator.Error()).NotTo(gomega.HaveOccurred())

	fixture.assertEvent(ctx, eventExpectation{
		name: "Stored",
		log:  *receipt.Logs[0],
		data: []any{inputs.tag, inputs.payload, inputs.note, true},
		exactTopics: []common.LogTopic{
			common.HashToLogTopic(fixture.contractABI.Events["Stored"].ID),
			common.BytesToLeftAlignedLogTopic(auth.From[:]),
			common.BytesToRightAlignedLogTopic(qrlmath.U512Bytes(new(big.Int).Set(inputs.amount))),
			common.BytesToRightAlignedLogTopic(qrlmath.U512Bytes(new(big.Int).Set(inputs.delta))),
		},
		want: map[string]any{
			"recipient": auth.From,
			"amount":    inputs.amount,
			"delta":     inputs.delta,
			"tag":       inputs.tag,
			"payload":   inputs.payload,
			"note":      inputs.note,
			"enabled":   true,
		},
		filter: [][]any{{auth.From}, {inputs.amount}, {inputs.delta}},
		reject: [][]any{nil, nil, {big.NewInt(0)}},
	})

	// Hyperion:
	// event Dynamic(bytes indexed payload, string indexed note, uint512 amount);
	// function store(...) external { emit Dynamic(payload, note, amount); }
	// Goal: dynamic indexed values use their Keccak-256 hashes as topics, and
	// generated parsing plus filtering reproduce those hashes and event data.
	ginkgo.By("hashing and filtering indexed dynamic event values")
	payloadHash := crypto.Keccak256Hash(inputs.payload)
	noteHash := crypto.Keccak256Hash([]byte(inputs.note))
	fixture.assertEvent(ctx, eventExpectation{
		name: "Dynamic",
		log:  *receipt.Logs[1],
		data: []any{inputs.amount},
		exactTopics: []common.LogTopic{
			common.HashToLogTopic(fixture.contractABI.Events["Dynamic"].ID),
			common.HashToLogTopic(payloadHash),
			common.HashToLogTopic(noteHash),
		},
		want: map[string]any{
			"payload": payloadHash,
			"note":    noteHash,
			"amount":  inputs.amount,
		},
		filter: [][]any{{inputs.payload}, {inputs.note}},
	})
	dynamic, err := fixture.binding.ParseDynamic(*receipt.Logs[1])
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(dynamic.Payload).To(gomega.Equal(payloadHash))
	gomega.Expect(dynamic.Note).To(gomega.Equal(noteHash))
	gomega.Expect(dynamic.Amount).To(gomega.Equal(inputs.amount))
	dynamicIterator, err := fixture.binding.FilterDynamic(
		filterOpts,
		[][]byte{[]byte("not the payload"), inputs.payload},
		[]string{inputs.note},
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer dynamicIterator.Close()
	gomega.Expect(dynamicIterator.Next()).To(gomega.BeTrue(), "generated Dynamic OR filter missed the transaction")
	gomega.Expect(dynamicIterator.Event.Raw.TxHash).To(gomega.Equal(receipt.TxHash))
	gomega.Expect(dynamicIterator.Next()).To(gomega.BeFalse())
	gomega.Expect(dynamicIterator.Error()).NotTo(gomega.HaveOccurred())

	// Hyperion:
	// event Composite(
	//     DynamicRecord record,
	//     uint16[3] fixedNumbers,
	//     string[2] fixedStrings,
	//     uint16[][2] mixed
	// );
	// function emitComposite(...) external {
	//     emit Composite(record, fixedNumbers, fixedStrings, mixed);
	// }
	// Goal: tuples and nested fixed/dynamic arrays survive event encoding,
	// generic decoding, and generated parsing without changing their shape.
	ginkgo.By("round-tripping composite event data through generic and generated decoders")
	record := EventEmitterDynamicRecord{
		Amount:  inputs.amount,
		Note:    inputs.note,
		Payload: inputs.payload,
		Values:  [][]uint16{{}, {1, 0xffff}, {0x1234}},
	}
	fixedNumbers := [3]uint16{0, 0xffff, 0x1234}
	fixedStrings := [2]string{"", inputs.note}
	mixed := [2][]uint16{{}, {1, 0xffff}}
	compositeTx, err := fixture.binding.EmitComposite(
		fixture.transactOpts(ctx),
		record,
		fixedNumbers,
		fixedStrings,
		mixed,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	compositeReceipt := waitSuccessfulTransaction(ctx, fixture.client, compositeTx)
	gomega.Expect(compositeReceipt.Logs).To(gomega.HaveLen(1))
	fixture.assertEvent(ctx, eventExpectation{
		name: "Composite",
		log:  *compositeReceipt.Logs[0],
		data: []any{record, fixedNumbers, fixedStrings, mixed},
		exactTopics: []common.LogTopic{
			common.HashToLogTopic(fixture.contractABI.Events["Composite"].ID),
		},
		want: map[string]any{
			"record": struct {
				Amount  *big.Int   `json:"amount"`
				Note    string     `json:"note"`
				Payload []byte     `json:"payload"`
				Values  [][]uint16 `json:"values"`
			}{
				Amount: record.Amount, Note: record.Note,
				Payload: record.Payload, Values: record.Values,
			},
			"fixedNumbers": fixedNumbers,
			"fixedStrings": fixedStrings,
			"mixed":        mixed,
		},
	})
	composite, err := fixture.binding.ParseComposite(*compositeReceipt.Logs[0])
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(composite.Record).To(gomega.Equal(record))
	gomega.Expect(composite.FixedNumbers).To(gomega.Equal(fixedNumbers))
	gomega.Expect(composite.FixedStrings).To(gomega.Equal(fixedStrings))
	gomega.Expect(composite.Mixed).To(gomega.Equal(mixed))

	// Hyperion:
	// event IndexedScalars(bool indexed flag, bytes5 indexed code, int16 indexed delta);
	// function emitIndexedScalars(bool flag, bytes5 code, int16 delta) external {
	//     emit IndexedScalars(flag, code, delta);
	// }
	// Goal: indexed scalar topics use the correct VM padding and sign
	// extension, and generated parsing plus filters recover their values.
	ginkgo.By("encoding and filtering indexed scalar event values")
	code, delta := [5]byte{0x00, 0x7f, 0x80, 0xfe, 0xff}, int16(-321)
	indexedTx, err := fixture.binding.EmitIndexedScalars(
		fixture.transactOpts(ctx),
		false,
		code,
		delta,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	indexedReceipt := waitSuccessfulTransaction(ctx, fixture.client, indexedTx)
	gomega.Expect(indexedReceipt.Logs).To(gomega.HaveLen(1))
	fixture.assertEvent(ctx, eventExpectation{
		name: "IndexedScalars",
		log:  *indexedReceipt.Logs[0],
		exactTopics: []common.LogTopic{
			common.HashToLogTopic(fixture.contractABI.Events["IndexedScalars"].ID),
			{},
			common.BytesToLeftAlignedLogTopic(code[:]),
			common.BytesToRightAlignedLogTopic(qrlmath.U512Bytes(big.NewInt(int64(delta)))),
		},
		want: map[string]any{
			"flag":  false,
			"code":  code,
			"delta": delta,
		},
		filter: [][]any{{true, false}, {code}, {delta}},
		reject: [][]any{{false}, {code}, {int16(321)}},
	})
	indexed, err := fixture.binding.ParseIndexedScalars(*indexedReceipt.Logs[0])
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(indexed.Flag).To(gomega.BeFalse())
	gomega.Expect(indexed.Code).To(gomega.Equal(code))
	gomega.Expect(indexed.Delta).To(gomega.Equal(delta))

	// Hyperion:
	// event Transformed(uint16 value);
	// event Transformed(string value);
	// function emitTransformed(uint16 value) external { emit Transformed(value); }
	// function emitTransformed(string calldata value) external { emit Transformed(value); }
	// Goal: overloaded event lookup and generated parsers retain the correct
	// canonical signature and decode each overload independently.
	ginkgo.By("resolving and decoding overloaded events")
	gomega.Expect(fixture.contractABI.Events["Transformed"].Sig).To(
		gomega.Equal("Transformed(uint16)"),
	)
	gomega.Expect(fixture.contractABI.Events["Transformed0"].Sig).To(
		gomega.Equal("Transformed(string)"),
	)
	stringTx, err := fixture.binding.EmitTransformed(fixture.transactOpts(ctx), inputs.note)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	stringReceipt := waitSuccessfulTransaction(ctx, fixture.client, stringTx)
	gomega.Expect(stringReceipt.Logs).To(gomega.HaveLen(1))
	fixture.assertEvent(ctx, eventExpectation{
		name: "Transformed0",
		log:  *stringReceipt.Logs[0],
		data: []any{inputs.note},
		exactTopics: []common.LogTopic{
			common.HashToLogTopic(fixture.contractABI.Events["Transformed0"].ID),
		},
		want: map[string]any{"value": inputs.note},
	})
	stringEvent, err := fixture.binding.ParseTransformed0(*stringReceipt.Logs[0])
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(stringEvent.Value).To(gomega.Equal(inputs.note))

	const transformedInteger = uint16(0x1234)
	integerTx, err := fixture.binding.EmitTransformed0(
		fixture.transactOpts(ctx),
		transformedInteger,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	integerReceipt := waitSuccessfulTransaction(ctx, fixture.client, integerTx)
	gomega.Expect(integerReceipt.Logs).To(gomega.HaveLen(1))
	fixture.assertEvent(ctx, eventExpectation{
		name: "Transformed",
		log:  *integerReceipt.Logs[0],
		data: []any{transformedInteger},
		exactTopics: []common.LogTopic{
			common.HashToLogTopic(fixture.contractABI.Events["Transformed"].ID),
		},
		want: map[string]any{"value": transformedInteger},
	})
	integerEvent, err := fixture.binding.ParseTransformed(*integerReceipt.Logs[0])
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(integerEvent.Value).To(gomega.Equal(transformedInteger))
}

func (fixture *liveFixture) assertFunctionValues(ctx context.Context) {
	ginkgo.GinkgoHelper()

	// Hyperion:
	// function echoFunctions(
	//     function(uint512) external pure returns (uint512) callback,
	//     string note,
	//     function(uint512) external pure returns (uint512)[2] fixedCallbacks,
	//     function(uint512) external pure returns (uint512)[] callbacks,
	//     FunctionRecord record
	// ) external pure returns (...);
	// Goal: standalone 68-byte function values and function values nested in a
	// fixed array, dynamic array, and tuple return exactly the values sent.
	ginkgo.By("round-tripping function values and their containers through generic ABI and raw RPC")
	callback := makeFunctionValue(
		fixture.address,
		fixture.contractABI.Methods["plusOne"].ID,
	)
	secondCallback := callback
	secondCallback[len(secondCallback)-1] ^= 0xff
	fixedCallbacks := [2][common.AddressLength + 4]byte{callback, secondCallback}
	callbacks := [][common.AddressLength + 4]byte{secondCallback, callback}
	functionRecord := struct {
		Callback [common.AddressLength + 4]byte `json:"callback"`
		Note     string                         `json:"note"`
	}{
		Callback: callback,
		Note:     fixture.inputs.note,
	}
	functionValues := []any{
		callback,
		fixture.inputs.note,
		fixedCallbacks,
		callbacks,
		functionRecord,
	}
	fixture.assertCall(ctx, "echoFunctions", functionValues, functionValues)

	// Hyperion:
	// function echoFunctions(...) external pure returns (...);
	// Goal: abigen represents every function value as [68]byte, including
	// values inside fixed arrays, dynamic arrays, and generated tuple types.
	ginkgo.By("round-tripping function values through generated binding types")
	generatedMethod := reflect.ValueOf(fixture.binding).MethodByName("EchoFunctions")
	gomega.Expect(generatedMethod.IsValid()).To(gomega.BeTrue())
	callbackType := reflect.TypeOf(callback)
	generatedType := generatedMethod.Type()
	gomega.Expect(generatedType.In(1)).To(gomega.Equal(callbackType), "generated callback input")
	gomega.Expect(generatedType.In(3)).To(gomega.Equal(reflect.TypeOf(fixedCallbacks)), "generated fixed function array")
	gomega.Expect(generatedType.In(4)).To(gomega.Equal(reflect.TypeOf(callbacks)), "generated function slice")
	recordType := generatedType.In(5)
	recordCallback, ok := recordType.FieldByName("Callback")
	gomega.Expect(ok).To(gomega.BeTrue(), "generated function record has no Callback field")
	gomega.Expect(recordCallback.Type).To(gomega.Equal(callbackType), "generated tuple function field")
	gomega.Expect(generatedType.Out(0)).To(gomega.Equal(callbackType), "generated callback output")
	gomega.Expect(generatedType.Out(2)).To(gomega.Equal(reflect.TypeOf(fixedCallbacks)), "generated fixed function output")
	gomega.Expect(generatedType.Out(3)).To(gomega.Equal(reflect.TypeOf(callbacks)), "generated function-slice output")

	generatedRecord := reflect.New(recordType).Elem()
	generatedRecord.FieldByName("Callback").Set(reflect.ValueOf(callback))
	generatedRecord.FieldByName("Note").SetString(fixture.inputs.note)
	callOpts := &bind.CallOpts{
		Context:     ctx,
		From:        fixture.from,
		BlockNumber: fixture.deploymentBlock,
	}
	generatedResults := generatedMethod.Call([]reflect.Value{
		reflect.ValueOf(callOpts),
		reflect.ValueOf(callback),
		reflect.ValueOf(fixture.inputs.note),
		reflect.ValueOf(fixedCallbacks),
		reflect.ValueOf(callbacks),
		generatedRecord,
	})
	gomega.Expect(generatedResults).To(gomega.HaveLen(6))
	gomega.Expect(generatedResults[5].IsNil()).To(gomega.BeTrue(), "generated EchoFunctions returned an error")
	gomega.Expect(generatedResults[0].Interface()).To(gomega.Equal(callback))
	gomega.Expect(generatedResults[1].String()).To(gomega.Equal(fixture.inputs.note))
	gomega.Expect(generatedResults[2].Interface()).To(gomega.Equal(fixedCallbacks))
	gomega.Expect(generatedResults[3].Interface()).To(gomega.Equal(callbacks))
	gomega.Expect(generatedResults[4].FieldByName("Callback").Interface()).To(gomega.Equal(callback))
	gomega.Expect(generatedResults[4].FieldByName("Note").String()).To(gomega.Equal(fixture.inputs.note))

	// Hyperion:
	// function exerciseFunction(
	//     function(uint512) external pure returns (uint512) callback,
	//     uint512 value
	// ) external returns (function(uint512) external pure returns (uint512), uint512);
	// Goal: a decoded 68-byte function value calls the encoded contract and
	// selector, then returns the same callback and the callback result.
	ginkgo.By("executing a function value through generic ABI and raw RPC")
	functionInput := new(big.Int).Add(
		new(big.Int).Lsh(big.NewInt(1), 500),
		big.NewInt(42),
	)
	functionResult := new(big.Int).Add(functionInput, big.NewInt(1))
	fixture.assertCall(
		ctx,
		"exerciseFunction",
		[]any{callback, functionInput},
		[]any{callback, functionResult},
	)

	// Hyperion:
	// event FunctionObserved(
	//     function(uint512) external pure returns (uint512) indexed indexedCallback,
	//     function(uint512) external pure returns (uint512) callback,
	//     uint512 result
	// );
	// function exerciseFunction(...) external {
	//     uint512 result = callback(value);
	//     emit FunctionObserved(callback, callback, result);
	// }
	// Goal: generated transactions execute the callback, indexed function
	// values hash to the expected topic, and generated parsing plus filtering
	// recover the callback and result.
	ginkgo.By("executing and filtering a function value through generated bindings")
	auth := fixture.transactOpts(ctx)
	exerciseMethod := reflect.ValueOf(fixture.binding).MethodByName("ExerciseFunction")
	gomega.Expect(exerciseMethod.IsValid()).To(gomega.BeTrue())
	gomega.Expect(exerciseMethod.Type().In(1)).To(
		gomega.Equal(callbackType),
		"generated ExerciseFunction callback",
	)
	exerciseResults := exerciseMethod.Call([]reflect.Value{
		reflect.ValueOf(auth),
		reflect.ValueOf(callback),
		reflect.ValueOf(functionInput),
	})
	gomega.Expect(exerciseResults[1].IsNil()).To(gomega.BeTrue(), "generated ExerciseFunction returned an error")
	functionTx, ok := exerciseResults[0].Interface().(*types.Transaction)
	gomega.Expect(ok).To(gomega.BeTrue(), "generated ExerciseFunction returned %T", exerciseResults[0].Interface())
	receipt := waitSuccessfulTransaction(ctx, fixture.client, functionTx)
	gomega.Expect(receipt.Logs).To(gomega.HaveLen(1))
	callbackHash := crypto.Keccak256Hash(callback[:])
	fixture.assertEvent(
		ctx,
		eventExpectation{
			name: "FunctionObserved",
			log:  *receipt.Logs[0],
			data: []any{callback, functionResult},
			exactTopics: []common.LogTopic{
				common.HashToLogTopic(fixture.contractABI.Events["FunctionObserved"].ID),
				common.HashToLogTopic(callbackHash),
			},
			want: map[string]any{
				"indexedCallback": callbackHash,
				"callback":        callback,
				"result":          functionResult,
			},
			filter: [][]any{{callback}},
		},
	)

	parseMethod := reflect.ValueOf(fixture.binding).MethodByName("ParseFunctionObserved")
	gomega.Expect(parseMethod.IsValid()).To(gomega.BeTrue())
	eventType := parseMethod.Type().Out(0).Elem()
	indexedField, ok := eventType.FieldByName("IndexedCallback")
	gomega.Expect(ok).To(gomega.BeTrue())
	gomega.Expect(indexedField.Type).To(
		gomega.Equal(reflect.TypeOf(common.Hash{})),
		"generated indexed function representation",
	)
	callbackField, ok := eventType.FieldByName("Callback")
	gomega.Expect(ok).To(gomega.BeTrue())
	gomega.Expect(callbackField.Type).To(gomega.Equal(callbackType), "generated event function representation")
	parsedResults := parseMethod.Call([]reflect.Value{reflect.ValueOf(*receipt.Logs[0])})
	gomega.Expect(parsedResults[1].IsNil()).To(gomega.BeTrue(), "generated FunctionObserved parser returned an error")
	parsedEvent := parsedResults[0].Elem()
	gomega.Expect(parsedEvent.FieldByName("IndexedCallback").Interface()).To(gomega.Equal(callbackHash))
	gomega.Expect(parsedEvent.FieldByName("Callback").Interface()).To(gomega.Equal(callback))
	gomega.Expect(parsedEvent.FieldByName("Result").Interface()).To(gomega.Equal(functionResult))

	filterMethod := reflect.ValueOf(fixture.binding).MethodByName("FilterFunctionObserved")
	gomega.Expect(filterMethod.IsValid()).To(gomega.BeTrue())
	ruleType := reflect.SliceOf(callbackType)
	gomega.Expect(filterMethod.Type().In(1)).To(gomega.Equal(ruleType), "generated function filter rule")
	rules := reflect.MakeSlice(ruleType, 1, 1)
	rules.Index(0).Set(reflect.ValueOf(callback))
	block := receipt.BlockNumber.Uint64()
	filterResults := filterMethod.Call([]reflect.Value{
		reflect.ValueOf(&bind.FilterOpts{Start: block, End: &block, Context: ctx}),
		rules,
	})
	gomega.Expect(filterResults[1].IsNil()).To(gomega.BeTrue(), "generated FunctionObserved filter returned an error")
	iterator := filterResults[0]
	gomega.Expect(iterator.MethodByName("Next").Call(nil)[0].Bool()).To(gomega.BeTrue())
	filteredEvent := iterator.Elem().FieldByName("Event").Elem()
	gomega.Expect(filteredEvent.FieldByName("Raw").Interface().(types.Log).TxHash).To(gomega.Equal(receipt.TxHash))
	gomega.Expect(iterator.MethodByName("Next").Call(nil)[0].Bool()).To(gomega.BeFalse())
	gomega.Expect(iterator.MethodByName("Error").Call(nil)[0].IsNil()).To(gomega.BeTrue())
	gomega.Expect(iterator.MethodByName("Close").Call(nil)[0].IsNil()).To(gomega.BeTrue())
}

func (fixture *liveFixture) assertPayableEntrypoints(ctx context.Context) {
	ginkgo.GinkgoHelper()

	// Hyperion:
	// event Received(uint256 amount);
	// receive() external payable { emit Received(msg.value); }
	// Goal: the generated receive entrypoint sends the requested value with
	// empty calldata, and its receipt plus generated parser reproduce that value.
	ginkgo.By("sending value through the generated receive entrypoint")
	amount := big.NewInt(11)
	auth := fixture.transactOpts(ctx)
	auth.Value = amount
	auth.GasLimit = 1_000_000
	tx, err := fixture.binding.Receive(auth)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "generated receive transaction")
	gomega.Expect(tx.To()).NotTo(gomega.BeNil())
	gomega.Expect(*tx.To()).To(gomega.Equal(fixture.address))
	gomega.Expect(tx.Data()).To(gomega.BeEmpty())
	gomega.Expect(tx.Value()).To(gomega.Equal(amount))
	receipt := waitSuccessfulTransaction(ctx, fixture.client, tx)
	gomega.Expect(receipt.Logs).To(gomega.HaveLen(1))
	fixture.assertEvent(ctx, eventExpectation{
		name: "Received",
		log:  *receipt.Logs[0],
		data: []any{amount},
		exactTopics: []common.LogTopic{
			common.HashToLogTopic(fixture.contractABI.Events["Received"].ID),
		},
		want: map[string]any{"amount": amount},
	})
	received, err := fixture.binding.ParseReceived(*receipt.Logs[0])
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(received.Amount).To(gomega.Equal(amount))

	// Hyperion:
	// event FallbackCalled(bytes payload, uint256 amount);
	// fallback() external payable { emit FallbackCalled(msg.data, msg.value); }
	// Goal: the generated fallback entrypoint preserves calldata larger than
	// one VM word and the transferred value in both the transaction and log.
	ginkgo.By("sending calldata and value through the generated fallback entrypoint")
	payload := []byte(strings.Repeat("\x5a", 65))
	amount = big.NewInt(13)
	auth = fixture.transactOpts(ctx)
	auth.Value = amount
	auth.GasLimit = 1_000_000
	tx, err = fixture.binding.Fallback(auth, payload)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "generated fallback transaction")
	gomega.Expect(tx.To()).NotTo(gomega.BeNil())
	gomega.Expect(*tx.To()).To(gomega.Equal(fixture.address))
	gomega.Expect(tx.Data()).To(gomega.Equal(payload))
	gomega.Expect(tx.Value()).To(gomega.Equal(amount))
	receipt = waitSuccessfulTransaction(ctx, fixture.client, tx)
	gomega.Expect(receipt.Logs).To(gomega.HaveLen(1))
	fixture.assertEvent(ctx, eventExpectation{
		name: "FallbackCalled",
		log:  *receipt.Logs[0],
		data: []any{payload, amount},
		exactTopics: []common.LogTopic{
			common.HashToLogTopic(fixture.contractABI.Events["FallbackCalled"].ID),
		},
		want: map[string]any{
			"payload": payload,
			"amount":  amount,
		},
	})
	fallback, err := fixture.binding.ParseFallbackCalled(*receipt.Logs[0])
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(fallback.Payload).To(gomega.Equal(payload))
	gomega.Expect(fallback.Amount).To(gomega.Equal(amount))

	// Hyperion:
	// event Paid(address indexed sender, uint16 indexed marker, uint256 amount);
	// function pay(uint16 marker) external payable {
	//     emit Paid(msg.sender, marker, msg.value);
	// }
	// Goal: a named generated payable method preserves its argument and value,
	// and its indexed event can be decoded and positively or negatively filtered.
	ginkgo.By("sending value through a named generated payable method")
	const marker = uint16(0xbabe)
	amount = big.NewInt(17)
	auth = fixture.transactOpts(ctx)
	auth.Value = amount
	payTx, err := fixture.binding.Pay(auth, marker)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "generated named payable transaction")
	payReceipt := waitSuccessfulTransaction(ctx, fixture.client, payTx)
	gomega.Expect(payReceipt.Logs).To(gomega.HaveLen(1))
	fixture.assertEvent(ctx, eventExpectation{
		name: "Paid",
		log:  *payReceipt.Logs[0],
		data: []any{amount},
		exactTopics: []common.LogTopic{
			common.HashToLogTopic(fixture.contractABI.Events["Paid"].ID),
			common.BytesToLeftAlignedLogTopic(fixture.from[:]),
			common.BytesToRightAlignedLogTopic(qrlmath.U512Bytes(new(big.Int).SetUint64(uint64(marker)))),
		},
		want: map[string]any{
			"sender": fixture.from,
			"marker": marker,
			"amount": amount,
		},
		filter: [][]any{{fixture.from}, {marker}},
		reject: [][]any{{fixture.from}, {uint16(marker + 1)}},
	})
	paid, err := fixture.binding.ParsePaid(*payReceipt.Logs[0])
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(paid.Sender).To(gomega.Equal(fixture.from))
	gomega.Expect(paid.Marker).To(gomega.Equal(marker))
	gomega.Expect(paid.Amount).To(gomega.Equal(amount))

}

func (fixture *liveFixture) assertWebSocketWatcher(ctx context.Context) {
	ginkgo.GinkgoHelper()

	// Hyperion:
	// event IndexedScalars(bool indexed flag, bytes5 indexed code, int16 indexed delta);
	// function emitIndexedScalars(bool flag, bytes5 code, int16 delta) external {
	//     emit IndexedScalars(flag, code, delta);
	// }
	// Goal: the generated WebSocket watcher ignores an event that fails one
	// indexed-topic rule, then delivers and decodes the event matching all rules.
	ginkgo.By("watching a filtered event through the generated WebSocket binding")
	auth := fixture.transactOpts(ctx)
	watched, err := NewEventEmitter(fixture.address, fixture.wsClient)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	events := make(chan *EventEmitterIndexedScalars, 1)
	code, delta := [5]byte{1, 2, 3, 4, 5}, int16(-777)
	subscription, err := watched.WatchIndexedScalars(
		&bind.WatchOpts{Context: ctx},
		events,
		[]bool{false},
		[][5]byte{code},
		[]int16{delta},
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(subscription.Unsubscribe)

	nonMatchingTx, err := fixture.binding.EmitIndexedScalars(
		auth,
		true,
		code,
		delta,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	waitSuccessfulTransaction(ctx, fixture.client, nonMatchingTx)
	matchingTx, err := fixture.binding.EmitIndexedScalars(auth, false, code, delta)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	receipt := waitSuccessfulTransaction(ctx, fixture.client, matchingTx)

	select {
	case event, open := <-events:
		gomega.Expect(open).To(gomega.BeTrue(), "generated IndexedScalars event channel closed")
		gomega.Expect(event).NotTo(gomega.BeNil())
		gomega.Expect(event.Raw.TxHash).To(gomega.Equal(receipt.TxHash))
		gomega.Expect(event.Raw.Address).To(gomega.Equal(fixture.address))
		gomega.Expect(event.Flag).To(gomega.BeFalse())
		gomega.Expect(event.Code).To(gomega.Equal(code))
		gomega.Expect(event.Delta).To(gomega.Equal(delta))
	case err, open := <-subscription.Err():
		gomega.Expect(open).To(gomega.BeTrue(), "generated IndexedScalars subscription closed")
		gomega.Expect(err).NotTo(gomega.BeNil(), "generated IndexedScalars subscription closed without an error")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	case <-time.After(90 * time.Second):
		ginkgo.Fail("timed out waiting for generated filtered IndexedScalars event")
	case <-ctx.Done():
		gomega.Expect(ctx.Err()).NotTo(gomega.HaveOccurred())
	}
}
