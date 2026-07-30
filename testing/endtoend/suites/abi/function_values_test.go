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
	"reflect"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/theQRL/go-qrl/accounts/abi/bind"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/core/types"
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
