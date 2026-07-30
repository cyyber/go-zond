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
	"math/big"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	qrl "github.com/theQRL/go-qrl"
	"github.com/theQRL/go-qrl/accounts/abi"
	"github.com/theQRL/go-qrl/accounts/abi/bind"
	"github.com/theQRL/go-qrl/common"
	qrlmath "github.com/theQRL/go-qrl/common/math"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/crypto"
)

type eventExpectation struct {
	name        string
	log         types.Log
	data        []any
	exactTopics []common.LogTopic
	want        map[string]any
	filter      [][]any
	reject      [][]any
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
	receipt := fixture.waitSuccessfulTransaction(ctx, storeTx)
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
	compositeReceipt := fixture.waitSuccessfulTransaction(ctx, compositeTx)
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
	indexedReceipt := fixture.waitSuccessfulTransaction(ctx, indexedTx)
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
	stringReceipt := fixture.waitSuccessfulTransaction(ctx, stringTx)
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
	integerReceipt := fixture.waitSuccessfulTransaction(ctx, integerTx)
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
