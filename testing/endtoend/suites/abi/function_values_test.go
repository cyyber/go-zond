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
	"github.com/theQRL/go-qrl/accounts/abi/bind"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/crypto"
)

func makeFunctionValue(address common.Address, selector []byte) [common.AddressLength + 4]byte {
	var value [common.AddressLength + 4]byte
	copy(value[:common.AddressLength], address[:])
	copy(value[common.AddressLength:], selector)
	return value
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
	functionRecord := EventEmitterFunctionRecord{
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
	callOpts := fixture.callOpts(ctx)
	gotCallback, gotNote, gotFixedCallbacks, gotCallbacks, gotRecord, err :=
		fixture.binding.EchoFunctions(
			callOpts,
			callback,
			fixture.inputs.note,
			fixedCallbacks,
			callbacks,
			functionRecord,
		)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(gotCallback).To(gomega.Equal(callback))
	gomega.Expect(gotNote).To(gomega.Equal(fixture.inputs.note))
	gomega.Expect(gotFixedCallbacks).To(gomega.Equal(fixedCallbacks))
	gomega.Expect(gotCallbacks).To(gomega.Equal(callbacks))
	gomega.Expect(gotRecord).To(gomega.Equal(functionRecord))

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
	functionTx, err := fixture.binding.ExerciseFunction(auth, callback, functionInput)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "generated ExerciseFunction transaction")
	receipt := fixture.waitSuccessfulTransaction(ctx, functionTx)
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

	parsedEvent, err := fixture.binding.ParseFunctionObserved(*receipt.Logs[0])
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(parsedEvent.IndexedCallback).To(gomega.Equal(callbackHash))
	gomega.Expect(parsedEvent.Callback).To(gomega.Equal(callback))
	gomega.Expect(parsedEvent.Result).To(gomega.Equal(functionResult))

	block := receipt.BlockNumber.Uint64()
	iterator, err := fixture.binding.FilterFunctionObserved(
		&bind.FilterOpts{Start: block, End: &block, Context: ctx},
		[][common.AddressLength + 4]byte{callback},
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer iterator.Close()
	gomega.Expect(iterator.Next()).To(gomega.BeTrue())
	gomega.Expect(iterator.Event.Raw.TxHash).To(gomega.Equal(receipt.TxHash))
	gomega.Expect(iterator.Next()).To(gomega.BeFalse())
	gomega.Expect(iterator.Error()).NotTo(gomega.HaveOccurred())
}
