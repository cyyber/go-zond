// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.
//
// The go-qrl library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package abi

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/theQRL/go-qrl/common"
	qmath "github.com/theQRL/go-qrl/common/math"
)

func TestDeclaredIntegerBoundsEverySupportedWidth(t *testing.T) {
	t.Parallel()

	for width := 8; width <= 512; width += 8 {
		width := width
		t.Run(fmt.Sprintf("uint%d", width), func(t *testing.T) {
			testUnsignedIntegerWidth(t, width)
		})
		t.Run(fmt.Sprintf("int%d", width), func(t *testing.T) {
			testSignedIntegerWidth(t, width)
		})
	}
}

func TestDeclaredIntegerWidthsRejectNonByteAligned(t *testing.T) {
	t.Parallel()

	for width := 1; width <= 512; width++ {
		if width%8 == 0 {
			continue
		}
		for _, prefix := range []string{"uint", "int"} {
			name := fmt.Sprintf("%s%d", prefix, width)
			if _, err := NewType(name, "", nil); err == nil {
				t.Errorf("NewType(%q) accepted a non-byte-aligned width", name)
			}
		}
	}
}

func testUnsignedIntegerWidth(t *testing.T, width int) {
	t.Helper()
	typ := mustIntegerType(t, fmt.Sprintf("uint%d", width))
	args := Arguments{{Name: "value", Type: typ}}
	max := new(big.Int).Sub(new(big.Int).Lsh(common.Big1, uint(width)), common.Big1)

	for _, value := range []*big.Int{new(big.Int), new(big.Int).Set(common.Big1), max} {
		assertIntegerRoundTrip(t, args, typ, value)
	}

	overflow := new(big.Int).Add(new(big.Int).Set(max), common.Big1)
	if typ.GetType() == bigIntPointerType {
		input := new(big.Int).Set(overflow)
		if _, err := args.Pack(input); !abiErrorMatches(err, errBadUint(width)) {
			t.Fatalf("pack uint%d overflow error = %v, want %v", width, err, errBadUint(width))
		}
		if input.Cmp(overflow) != 0 {
			t.Fatalf("pack uint%d mutated input: got %s want %s", width, input, overflow)
		}
		if _, err := args.Pack(big.NewInt(-1)); !abiErrorMatches(err, errInvalidSign) {
			t.Fatalf("pack uint%d negative error = %v, want %v", width, err, errInvalidSign)
		}
	}
	if width < 512 {
		if _, err := args.Unpack(integerWord(overflow)); !abiErrorMatches(err, errBadUint(width)) {
			t.Fatalf("unpack uint%d overflow error = %v, want %v", width, err, errBadUint(width))
		}
	}
}

func testSignedIntegerWidth(t *testing.T, width int) {
	t.Helper()
	typ := mustIntegerType(t, fmt.Sprintf("int%d", width))
	args := Arguments{{Name: "value", Type: typ}}
	max := new(big.Int).Sub(new(big.Int).Lsh(common.Big1, uint(width-1)), common.Big1)
	min := new(big.Int).Neg(new(big.Int).Lsh(common.Big1, uint(width-1)))

	values := []*big.Int{min, big.NewInt(-1), new(big.Int)}
	if max.Sign() > 0 {
		values = append(values, new(big.Int).Set(common.Big1))
	}
	values = append(values, max)
	for _, value := range values {
		assertIntegerRoundTrip(t, args, typ, value)
	}

	if typ.GetType() == bigIntPointerType {
		overflow := new(big.Int).Add(new(big.Int).Set(max), common.Big1)
		underflow := new(big.Int).Sub(new(big.Int).Set(min), common.Big1)
		for _, value := range []*big.Int{overflow, underflow} {
			input := new(big.Int).Set(value)
			if _, err := args.Pack(input); !abiErrorMatches(err, errBadInt(width)) {
				t.Fatalf("pack int%d out-of-range %s error = %v, want %v", width, value, err, errBadInt(width))
			}
			if input.Cmp(value) != 0 {
				t.Fatalf("pack int%d mutated input: got %s want %s", width, input, value)
			}
		}
	}
	if width < 512 {
		for _, value := range []*big.Int{
			new(big.Int).Add(new(big.Int).Set(max), common.Big1),
			new(big.Int).Sub(new(big.Int).Set(min), common.Big1),
		} {
			if _, err := args.Unpack(integerWord(value)); !abiErrorMatches(err, errBadInt(width)) {
				t.Fatalf("unpack int%d out-of-range %s error = %v, want %v", width, value, err, errBadInt(width))
			}
		}
	}
}

var bigIntPointerType = reflect.TypeFor[*big.Int]()

func mustIntegerType(t *testing.T, name string) Type {
	t.Helper()
	typ, err := NewType(name, "", nil)
	if err != nil {
		t.Fatalf("NewType(%q): %v", name, err)
	}
	return typ
}

func assertIntegerRoundTrip(t *testing.T, args Arguments, typ Type, want *big.Int) {
	t.Helper()
	input := integerInput(typ, want)
	packed, err := args.PackValues([]any{input})
	if err != nil {
		t.Fatalf("pack %s value %s: %v", typ, want, err)
	}
	if len(packed) != 64 {
		t.Fatalf("pack %s value %s length = %d, want 64", typ, want, len(packed))
	}
	values, err := args.Unpack(packed)
	if err != nil {
		t.Fatalf("unpack %s value %s: %v", typ, want, err)
	}
	if len(values) != 1 {
		t.Fatalf("unpack %s value count = %d, want 1", typ, len(values))
	}
	if got := integerOutput(values[0]); got.Cmp(want) != 0 {
		t.Fatalf("round trip %s = %s, want %s", typ, got, want)
	}
}

func integerInput(typ Type, value *big.Int) any {
	if typ.T == UintTy {
		switch typ.Size {
		case 8:
			return uint8(value.Uint64())
		case 16:
			return uint16(value.Uint64())
		case 32:
			return uint32(value.Uint64())
		case 64:
			return value.Uint64()
		}
	} else {
		switch typ.Size {
		case 8:
			return int8(value.Int64())
		case 16:
			return int16(value.Int64())
		case 32:
			return int32(value.Int64())
		case 64:
			return value.Int64()
		}
	}
	return new(big.Int).Set(value)
}

func integerOutput(value any) *big.Int {
	switch value := value.(type) {
	case uint8:
		return new(big.Int).SetUint64(uint64(value))
	case uint16:
		return new(big.Int).SetUint64(uint64(value))
	case uint32:
		return new(big.Int).SetUint64(uint64(value))
	case uint64:
		return new(big.Int).SetUint64(value)
	case int8:
		return big.NewInt(int64(value))
	case int16:
		return big.NewInt(int64(value))
	case int32:
		return big.NewInt(int64(value))
	case int64:
		return big.NewInt(value)
	case *big.Int:
		return new(big.Int).Set(value)
	default:
		panic(fmt.Sprintf("unexpected integer output %T", value))
	}
}

func integerWord(value *big.Int) []byte {
	return qmath.U512Bytes(new(big.Int).Set(value))
}

func abiErrorMatches(got, want error) bool {
	return got != nil && (errors.Is(got, want) || got.Error() == want.Error())
}
