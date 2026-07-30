// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package abi

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/theQRL/go-qrl/common"
)

// TestPack tests the general pack/unpack tests in packing_test.go
func TestPack(t *testing.T) {
	t.Parallel()
	for i, test := range packUnpackTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			encb, err := hex.DecodeString(test.packed)
			if err != nil {
				t.Fatalf("invalid hex %s: %v", test.packed, err)
			}
			inDef := fmt.Sprintf(`[{ "name" : "method", "type": "function", "inputs": %s}]`, test.def)
			inAbi, err := JSON(strings.NewReader(inDef))
			if err != nil {
				t.Fatalf("invalid ABI definition %s, %v", inDef, err)
			}
			var packed []byte
			packed, err = inAbi.Pack("method", test.unpacked)

			if err != nil {
				t.Fatalf("test %d (%v) failed: %v", i, test.def, err)
			}
			if !reflect.DeepEqual(packed[4:], encb) {
				t.Errorf("test %d (%v) failed: expected %v, got %v", i, test.def, encb, packed[4:])
			}
		})
	}
}

func TestPackFunctionHeadOffsets(t *testing.T) {
	t.Parallel()

	function := functionValue(1, [4]byte{0xde, 0xad, 0xbe, 0xef})
	tests := []struct {
		name       string
		inputs     string
		values     []any
		offsetWord int
		wantOffset int
	}{
		{
			name:       "function",
			inputs:     `[{"type":"function"},{"type":"string"}]`,
			values:     []any{function, "hyperion"},
			offsetWord: 2,
			wantOffset: 3 * 64,
		},
		{
			name:   "fixed array",
			inputs: `[{"type":"function[2]"},{"type":"string"}]`,
			values: []any{
				[2][common.AddressLength + 4]byte{function, function},
				"hyperion",
			},
			offsetWord: 4,
			wantOffset: 5 * 64,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := fmt.Sprintf(
				`[{"name":"method","type":"function","inputs":%s}]`,
				test.inputs,
			)
			parsed, err := JSON(strings.NewReader(definition))
			if err != nil {
				t.Fatalf("invalid ABI definition: %v", err)
			}
			packed, err := parsed.Pack("method", test.values...)
			if err != nil {
				t.Fatalf("pack: %v", err)
			}

			start := 4 + test.offsetWord*64
			want := common.LeftPadBytes(big.NewInt(int64(test.wantOffset)).Bytes(), 64)
			if got := packed[start : start+64]; !bytes.Equal(got, want) {
				t.Fatalf("dynamic offset = %x, want %x", got, want)
			}
		})
	}
}

func TestPackFunctionRejectsWrongLength(t *testing.T) {
	t.Parallel()

	definition := `[{"name":"method","type":"function","inputs":[{"type":"function"}]}]`
	parsed, err := JSON(strings.NewReader(definition))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parsed.Pack("method", [common.AddressLength]byte{}); err == nil {
		t.Fatal("packed an address without a function selector")
	}
}

func TestMethodPack(t *testing.T) {
	t.Parallel()
	abi, err := JSON(strings.NewReader(jsondata))
	if err != nil {
		t.Fatal(err)
	}

	sig := abi.Methods["slice"].ID
	sig = append(sig, common.LeftPadBytes([]byte{1}, 64)...)
	sig = append(sig, common.LeftPadBytes([]byte{2}, 64)...)

	packed, err := abi.Pack("slice", []uint32{1, 2})
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(packed, sig) {
		t.Errorf("expected %x got %x", sig, packed)
	}

	var addrA, addrB = common.Address{1}, common.Address{2}
	sig = abi.Methods["sliceAddress"].ID
	sig = append(sig, common.LeftPadBytes([]byte{64}, 64)...) // offset = 1 slot * 64 B
	sig = append(sig, common.LeftPadBytes([]byte{2}, 64)...)
	sig = append(sig, common.LeftPadBytes(addrA[:], 64)...)
	sig = append(sig, common.LeftPadBytes(addrB[:], 64)...)

	packed, err = abi.Pack("sliceAddress", []common.Address{addrA, addrB})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(packed, sig) {
		t.Errorf("expected %x got %x", sig, packed)
	}

	var addrC, addrD = common.Address{3}, common.Address{4}
	sig = abi.Methods["sliceMultiAddress"].ID
	sig = append(sig, common.LeftPadBytes([]byte{0x80}, 64)...)       // 2 head slots * 64 = 128
	sig = append(sig, common.LeftPadBytes([]byte{0x01, 0x40}, 64)...) // 128 + count + 2 elems = 320
	sig = append(sig, common.LeftPadBytes([]byte{2}, 64)...)
	sig = append(sig, common.LeftPadBytes(addrA[:], 64)...)
	sig = append(sig, common.LeftPadBytes(addrB[:], 64)...)
	sig = append(sig, common.LeftPadBytes([]byte{2}, 64)...)
	sig = append(sig, common.LeftPadBytes(addrC[:], 64)...)
	sig = append(sig, common.LeftPadBytes(addrD[:], 64)...)

	packed, err = abi.Pack("sliceMultiAddress", []common.Address{addrA, addrB}, []common.Address{addrC, addrD})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(packed, sig) {
		t.Errorf("expected %x got %x", sig, packed)
	}

	sig = abi.Methods["slice256"].ID
	sig = append(sig, common.LeftPadBytes([]byte{1}, 64)...)
	sig = append(sig, common.LeftPadBytes([]byte{2}, 64)...)

	packed, err = abi.Pack("slice256", []*big.Int{big.NewInt(1), big.NewInt(2)})
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(packed, sig) {
		t.Errorf("expected %x got %x", sig, packed)
	}

	a := [2][2]*big.Int{{big.NewInt(1), big.NewInt(1)}, {big.NewInt(2), big.NewInt(0)}}
	sig = abi.Methods["nestedArray"].ID
	sig = append(sig, common.LeftPadBytes([]byte{1}, 64)...)
	sig = append(sig, common.LeftPadBytes([]byte{1}, 64)...)
	sig = append(sig, common.LeftPadBytes([]byte{2}, 64)...)
	sig = append(sig, common.LeftPadBytes([]byte{0}, 64)...)
	sig = append(sig, common.LeftPadBytes([]byte{0x01, 0x40}, 64)...) // offset 0xa0 doubled → 0x140
	sig = append(sig, common.LeftPadBytes([]byte{2}, 64)...)
	sig = append(sig, common.LeftPadBytes(addrC[:], 64)...)
	sig = append(sig, common.LeftPadBytes(addrD[:], 64)...)
	packed, err = abi.Pack("nestedArray", a, []common.Address{addrC, addrD})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(packed, sig) {
		t.Errorf("expected %x got %x", sig, packed)
	}

	sig = abi.Methods["nestedArray2"].ID
	sig = append(sig, common.LeftPadBytes([]byte{0x40}, 64)...)       // 0x20 → 0x40
	sig = append(sig, common.LeftPadBytes([]byte{0x80}, 64)...)       // 0x40 → 0x80
	sig = append(sig, common.LeftPadBytes([]byte{0x01, 0x00}, 64)...) // 0x80 → 0x100
	sig = append(sig, common.LeftPadBytes([]byte{1}, 64)...)
	sig = append(sig, common.LeftPadBytes([]byte{1}, 64)...)
	sig = append(sig, common.LeftPadBytes([]byte{1}, 64)...)
	sig = append(sig, common.LeftPadBytes([]byte{1}, 64)...)
	packed, err = abi.Pack("nestedArray2", [2][]uint8{{1}, {1}})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(packed, sig) {
		t.Errorf("expected %x got %x", sig, packed)
	}

	sig = abi.Methods["nestedSlice"].ID
	sig = append(sig, common.LeftPadBytes([]byte{0x40}, 64)...) // 0x20 → 0x40
	sig = append(sig, common.LeftPadBytes([]byte{0x02}, 64)...)
	sig = append(sig, common.LeftPadBytes([]byte{0x80}, 64)...)       // 0x40 → 0x80
	sig = append(sig, common.LeftPadBytes([]byte{0x01, 0x40}, 64)...) // 0xa0 → 0x140
	sig = append(sig, common.LeftPadBytes([]byte{2}, 64)...)
	sig = append(sig, common.LeftPadBytes([]byte{1}, 64)...)
	sig = append(sig, common.LeftPadBytes([]byte{2}, 64)...)
	sig = append(sig, common.LeftPadBytes([]byte{2}, 64)...)
	sig = append(sig, common.LeftPadBytes([]byte{1}, 64)...)
	sig = append(sig, common.LeftPadBytes([]byte{2}, 64)...)
	packed, err = abi.Pack("nestedSlice", [][]uint8{{1, 2}, {1, 2}})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(packed, sig) {
		t.Errorf("expected %x got %x", sig, packed)
	}

	// test that we can't pack a negative value for a parameter that is specified as a uint
	if _, err := abi.Pack("send", big.NewInt(-1)); err == nil {
		t.Fatal("expected error when trying to pack negative big.Int into uint256 value")
	}
}

func TestPackNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value  reflect.Value
		packed []byte
	}{
		// Protocol limits
		{reflect.ValueOf(0), common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")},
		{reflect.ValueOf(1), common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001")},
		{reflect.ValueOf(-1), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")},

		// Type corner cases
		{reflect.ValueOf(uint8(math.MaxUint8)), common.Hex2Bytes("000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ff")},
		{reflect.ValueOf(uint16(math.MaxUint16)), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ffff")},
		{reflect.ValueOf(uint32(math.MaxUint32)), common.Hex2Bytes("000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ffffffff")},
		{reflect.ValueOf(uint64(math.MaxUint64)), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000ffffffffffffffff")},

		{reflect.ValueOf(int8(math.MaxInt8)), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007f")},
		{reflect.ValueOf(int16(math.MaxInt16)), common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007fff")},
		{reflect.ValueOf(int32(math.MaxInt32)), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007fffffff")},
		{reflect.ValueOf(int64(math.MaxInt64)), common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007fffffffffffffff")},

		{reflect.ValueOf(int8(math.MinInt8)), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80")},
		{reflect.ValueOf(int16(math.MinInt16)), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8000")},
		{reflect.ValueOf(int32(math.MinInt32)), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80000000")},
		{reflect.ValueOf(int64(math.MinInt64)), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8000000000000000")},
	}
	for i, tt := range tests {
		packed := packNum(tt.value)
		if !bytes.Equal(packed, tt.packed) {
			t.Errorf("test %d: pack mismatch: have %x, want %x", i, packed, tt.packed)
		}
	}
}
