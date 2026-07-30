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
	"strings"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/theQRL/go-qrl/common"
)

func (fixture *liveFixture) assertBoundaryRoundTrips(ctx context.Context) {
	ginkgo.GinkgoHelper()

	callOpts := fixture.callOpts(ctx)
	inputs := fixture.inputs

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

	// Hyperion:
	// function echoBoundaryEdges(BoundaryEdges edges)
	//     external pure returns (BoundaryEdges);
	// Goal: widths immediately below and above 256 bits and immediately below
	// and at 512 bits preserve zero, signed, and unsigned boundary values.
	ginkgo.By("round-tripping integer transition widths and extrema")
	for _, test := range []struct {
		name  string
		edges EventEmitterBoundaryEdges
	}{
		{
			name:  "zero",
			edges: zeroBoundaryEdges(),
		},
		{
			name: "unsigned maxima and signed minima",
			edges: EventEmitterBoundaryEdges{
				Unsigned248: unsignedMaximum(248),
				Signed248:   signedMinimum(248),
				Unsigned256: unsignedMaximum(256),
				Signed256:   signedMinimum(256),
				Unsigned264: unsignedMaximum(264),
				Signed264:   signedMinimum(264),
				Unsigned504: unsignedMaximum(504),
				Signed504:   signedMinimum(504),
				Unsigned512: unsignedMaximum(512),
				Signed512:   signedMinimum(512),
			},
		},
		{
			name: "signed maxima",
			edges: EventEmitterBoundaryEdges{
				Unsigned248: big.NewInt(1),
				Signed248:   signedMaximum(248),
				Unsigned256: big.NewInt(1),
				Signed256:   signedMaximum(256),
				Unsigned264: big.NewInt(1),
				Signed264:   signedMaximum(264),
				Unsigned504: big.NewInt(1),
				Signed504:   signedMaximum(504),
				Unsigned512: big.NewInt(1),
				Signed512:   signedMaximum(512),
			},
		},
		{
			name: "negative one",
			edges: EventEmitterBoundaryEdges{
				Unsigned248: big.NewInt(1),
				Signed248:   big.NewInt(-1),
				Unsigned256: big.NewInt(1),
				Signed256:   big.NewInt(-1),
				Unsigned264: big.NewInt(1),
				Signed264:   big.NewInt(-1),
				Unsigned504: big.NewInt(1),
				Signed504:   big.NewInt(-1),
				Unsigned512: big.NewInt(1),
				Signed512:   big.NewInt(-1),
			},
		},
	} {
		ginkgo.By("checking integer edges: " + test.name)
		fixture.assertCall(
			ctx,
			"echoBoundaryEdges",
			[]any{test.edges},
			[]any{test.edges},
		)
		got, err := fixture.binding.EchoBoundaryEdges(callOpts, test.edges)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), test.name)
		gomega.Expect(got).To(gomega.Equal(test.edges), test.name)
	}

	// Hyperion:
	// function echoBoundaryEdges(BoundaryEdges edges)
	//     external pure returns (BoundaryEdges);
	// Goal: fixed bytes remain left-aligned across the legacy 32-byte boundary
	// and at the 64-byte VM word boundary.
	ginkgo.By("round-tripping fixed bytes across ABI word boundaries")
	fixedBytes := zeroBoundaryEdges()
	fillPattern(fixedBytes.Bytes31Value[:], 0x11)
	fillPattern(fixedBytes.Bytes32Value[:], 0x22)
	fillPattern(fixedBytes.Bytes33Value[:], 0x33)
	fillPattern(fixedBytes.Bytes63Value[:], 0x44)
	fillPattern(fixedBytes.Bytes64Value[:], 0x55)
	fixture.assertCall(
		ctx,
		"echoBoundaryEdges",
		[]any{fixedBytes},
		[]any{fixedBytes},
	)
	gotFixedBytes, err := fixture.binding.EchoBoundaryEdges(callOpts, fixedBytes)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(gotFixedBytes).To(gomega.Equal(fixedBytes))

	// Hyperion:
	// function echo(
	//     uint512 amount, int512 delta, bytes64 tag, address recipient,
	//     bytes payload, string note, bool enabled
	// ) external pure returns (...);
	// Goal: dynamic lengths immediately around each 64-byte boundary preserve
	// their offsets, length words, payload bytes, and padding.
	ginkgo.By("round-tripping dynamic bytes and strings around VM word boundaries")
	for _, length := range []int{0, 1, 63, 64, 65, 127, 128, 129} {
		payload := make([]byte, length)
		fillPattern(payload, byte(length))
		note := strings.Repeat(string(rune('a'+length%26)), length)
		values := []any{
			fixture.inputs.amount,
			fixture.inputs.delta,
			fixture.inputs.tag,
			fixture.from,
			payload,
			note,
			true,
		}
		fixture.assertCall(ctx, "echo", values, values)
		gotAmount, gotDelta, gotTag, gotRecipient, gotPayload, gotNote, gotEnabled, err :=
			fixture.binding.Echo(
				callOpts,
				fixture.inputs.amount,
				fixture.inputs.delta,
				fixture.inputs.tag,
				fixture.from,
				payload,
				note,
				true,
			)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "length %d", length)
		gomega.Expect([]any{
			gotAmount,
			gotDelta,
			gotTag,
			gotRecipient,
			gotPayload,
			gotNote,
			gotEnabled,
		}).To(gomega.Equal(values), "length %d", length)
	}

	// Hyperion:
	// function echoLeafContainers(address[2], address[], bytes64[2], bytes64[])
	//     external pure returns (...);
	// Goal: fixed and dynamic containers of full-word leaves preserve zero and
	// distinct nonzero 64-byte addresses and bytes64 values.
	ginkgo.By("round-tripping 64-byte address and fixed-bytes containers")
	secondAddress := fixture.from
	secondAddress[0] ^= 0x80
	secondAddress[len(secondAddress)-1] ^= 0x01
	secondTag := fixture.inputs.tag
	secondTag[0] ^= 0xff
	secondTag[len(secondTag)-1] ^= 0xff
	fixedAddresses := [2]common.Address{{}, fixture.from}
	addresses := []common.Address{secondAddress, {}, fixture.from}
	fixedTags := [2][64]byte{{}, fixture.inputs.tag}
	tags := [][64]byte{secondTag, {}, fixture.inputs.tag}
	leafValues := []any{fixedAddresses, addresses, fixedTags, tags}
	fixture.assertCall(ctx, "echoLeafContainers", leafValues, leafValues)
	gotFixedAddresses, gotAddresses, gotFixedTags, gotTags, err :=
		fixture.binding.EchoLeafContainers(
			callOpts,
			fixedAddresses,
			addresses,
			fixedTags,
			tags,
		)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(gotFixedAddresses).To(gomega.Equal(fixedAddresses))
	gomega.Expect(gotAddresses).To(gomega.Equal(addresses))
	gomega.Expect(gotFixedTags).To(gomega.Equal(fixedTags))
	gomega.Expect(gotTags).To(gomega.Equal(tags))

	// Hyperion:
	// function echoDynamicContainers(bytes[2], bytes[], string[])
	//     external pure returns (...);
	// Goal: fixed and dynamic containers whose elements have dynamic tails
	// preserve empty and multi-word members.
	ginkgo.By("round-tripping fixed and dynamic containers with dynamic elements")
	fixedDynamicBytes := [2][]byte{
		{},
		patternedBytes(65, 0x61),
	}
	byteSlices := [][]byte{
		{},
		patternedBytes(1, 0x71),
		patternedBytes(65, 0x72),
		patternedBytes(129, 0x73),
	}
	stringValues := []string{
		"",
		"x",
		strings.Repeat("s", 65),
		strings.Repeat("t", 129),
	}
	dynamicContainerValues := []any{fixedDynamicBytes, byteSlices, stringValues}
	fixture.assertCall(
		ctx,
		"echoDynamicContainers",
		dynamicContainerValues,
		dynamicContainerValues,
	)
	gotFixedDynamicBytes, gotByteSlices, gotStrings, err :=
		fixture.binding.EchoDynamicContainers(
			callOpts,
			fixedDynamicBytes,
			byteSlices,
			stringValues,
		)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(gotFixedDynamicBytes).To(gomega.Equal(fixedDynamicBytes))
	gomega.Expect(gotByteSlices).To(gomega.Equal(byteSlices))
	gomega.Expect(gotStrings).To(gomega.Equal(stringValues))

	// Hyperion:
	// function echoCompositeContainers(
	//     uint16[2][2], uint16[2][], DynamicRecord[2], NestedRecord
	// ) external pure returns (...);
	// Goal: fixed-of-fixed arrays, dynamic arrays of fixed arrays, fixed arrays
	// of dynamic tuples, and tuples nested inside tuples retain their shape.
	ginkgo.By("round-tripping offset-heavy arrays and nested tuples")
	fixedMatrix := [2][2]uint16{{0, 0xffff}, {1, 0x1234}}
	rows := [][2]uint16{{}, {1, 0xffff}, {0x1234, 0x4321}}
	records := [2]EventEmitterDynamicRecord{
		{
			Amount:  fixture.inputs.amount,
			Note:    fixture.inputs.note,
			Payload: fixture.inputs.payload,
			Values:  [][]uint16{{1, 2}, {}, {3}},
		},
		{
			Amount:  new(big.Int),
			Note:    "",
			Payload: []byte{},
			Values:  [][]uint16{},
		},
	}
	nested := EventEmitterNestedRecord{
		FixedRecord: EventEmitterRecord{
			Amount:    fixture.inputs.amount,
			Recipient: secondAddress,
			Tag:       secondTag,
		},
		DynamicRecord: records[0],
		Extra:         patternedBytes(65, 0x91),
	}
	compositeValues := []any{fixedMatrix, rows, records, nested}
	fixture.assertCall(
		ctx,
		"echoCompositeContainers",
		compositeValues,
		compositeValues,
	)
	gotFixedMatrix, gotRows, gotRecords, gotNested, err :=
		fixture.binding.EchoCompositeContainers(
			callOpts,
			fixedMatrix,
			rows,
			records,
			nested,
		)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(gotFixedMatrix).To(gomega.Equal(fixedMatrix))
	gomega.Expect(gotRows).To(gomega.Equal(rows))
	gomega.Expect(gotRecords).To(gomega.Equal(records))
	gomega.Expect(gotNested).To(gomega.Equal(nested))
}

func unsignedMaximum(bits uint) *big.Int {
	return new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), bits), big.NewInt(1))
}

func zeroBoundaryEdges() EventEmitterBoundaryEdges {
	return EventEmitterBoundaryEdges{
		Unsigned248: new(big.Int),
		Signed248:   new(big.Int),
		Unsigned256: new(big.Int),
		Signed256:   new(big.Int),
		Unsigned264: new(big.Int),
		Signed264:   new(big.Int),
		Unsigned504: new(big.Int),
		Signed504:   new(big.Int),
		Unsigned512: new(big.Int),
		Signed512:   new(big.Int),
	}
}

func signedMinimum(bits uint) *big.Int {
	return new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), bits-1))
}

func signedMaximum(bits uint) *big.Int {
	return new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), bits-1), big.NewInt(1))
}

func fillPattern(destination []byte, seed byte) {
	for index := range destination {
		destination[index] = seed + byte(index*29)
	}
}

func patternedBytes(length int, seed byte) []byte {
	value := make([]byte, length)
	fillPattern(value, seed)
	return value
}
