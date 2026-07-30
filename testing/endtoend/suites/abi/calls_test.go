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
	"github.com/theQRL/go-qrl/accounts/abi/bind"
)

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
	nestedValues := []any{nested, records, cube}
	fixture.assertCall(ctx, "echoNested", nestedValues, nestedValues)

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
	boundaryValues := []any{
		smallUnsigned,
		smallSigned,
		wideUnsigned,
		wideSigned,
		shortBytes,
		fixedNumbers,
		fixedStrings,
		mixed,
	}
	fixture.assertCall(ctx, "echoBoundaries", boundaryValues, boundaryValues)

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

	fixture.assertBoundaryRoundTrips(ctx)
}
