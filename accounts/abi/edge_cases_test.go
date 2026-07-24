// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.
//
// The go-qrl library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package abi

import (
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/theQRL/go-qrl/common"
)

func mustNewEdgeType(t *testing.T, name string, components []ArgumentMarshaling) Type {
	t.Helper()
	typ, err := NewType(name, "", components)
	if err != nil {
		t.Fatalf("NewType(%q): %v", name, err)
	}
	return typ
}

func TestNewTypeStrictGrammarAndIntegerWidths(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"int8", "uint8", "int248", "uint248", "int504", "uint504", "int512", "uint512",
		"bytes64", "uint8[0]", "uint8[2][]", "string[2][3]",
	} {
		name := name
		t.Run("accept_"+name, func(t *testing.T) {
			t.Parallel()
			if _, err := NewType(name, "", nil); err != nil {
				t.Fatalf("canonical type rejected: %v", err)
			}
		})
	}

	for _, name := range []string{
		"int0", "uint0", "int513", "uint513",
		"int1", "uint1", "int7", "uint7", "int9", "uint9",
		"int17", "uint17", "int255", "uint255", "int511", "uint511",
		"int01", "uint08", "bytes01",
		"uint8garbage", "bytes32junk", "bool8", "address64", "string8", "tuple1", "function68",
		"uint8[00]", "uint8[01]", "uint8[0][01]",
		"uint8[2junk]", "uint8]2[", "uint8[2][x]", "bool8[]",
	} {
		name := name
		t.Run("reject_"+name, func(t *testing.T) {
			t.Parallel()
			if _, err := NewType(name, "", nil); err == nil {
				t.Fatalf("malformed or out-of-range type %q was accepted", name)
			}
		})
	}
}

func TestZeroSizedStaticValuesRoundTripAndAdjacency(t *testing.T) {
	t.Parallel()

	zeroArray := mustNewEdgeType(t, "uint16[0]", nil)
	nestedZeroArray := mustNewEdgeType(t, "uint16[0][2]", nil)
	emptyTuple := mustNewEdgeType(t, "tuple", nil)
	uint16Type := mustNewEdgeType(t, "uint16", nil)
	stringType := mustNewEdgeType(t, "string", nil)

	emptyTupleValue := reflect.New(emptyTuple.GetType()).Elem().Interface()
	arguments := Arguments{
		{Name: "zero", Type: zeroArray},
		{Name: "number", Type: uint16Type},
		{Name: "empty", Type: emptyTuple},
		{Name: "nested", Type: nestedZeroArray},
		{Name: "text", Type: stringType},
	}
	want := []any{[0]uint16{}, uint16(0x1234), emptyTupleValue, [2][0]uint16{}, "vm64"}
	encoded, err := arguments.Pack(want...)
	if err != nil {
		t.Fatalf("pack adjacent zero-sized values: %v", err)
	}
	got, err := arguments.Unpack(encoded)
	if err != nil {
		t.Fatalf("unpack adjacent zero-sized values: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip mismatch\ngot:  %#v\nwant: %#v", got, want)
	}

	emptyArguments := Arguments{
		{Name: "zero", Type: zeroArray},
		{Name: "empty", Type: emptyTuple},
		{Name: "nested", Type: nestedZeroArray},
	}
	emptyWant := []any{[0]uint16{}, emptyTupleValue, [2][0]uint16{}}
	emptyEncoding, err := emptyArguments.Pack(emptyWant...)
	if err != nil {
		t.Fatalf("pack zero-sized values: %v", err)
	}
	if len(emptyEncoding) != 0 {
		t.Fatalf("zero-sized encoding has %d bytes, want 0", len(emptyEncoding))
	}
	emptyGot, err := emptyArguments.Unpack(nil)
	if err != nil {
		t.Fatalf("unpack zero-sized values from empty encoding: %v", err)
	}
	if !reflect.DeepEqual(emptyGot, emptyWant) {
		t.Fatalf("empty round trip mismatch\ngot:  %#v\nwant: %#v", emptyGot, emptyWant)
	}
	mapped := make(map[string]any)
	if err := emptyArguments.UnpackIntoMap(mapped, nil); err != nil {
		t.Fatalf("unpack zero-sized values into map: %v", err)
	}
	for i, arg := range emptyArguments {
		if !reflect.DeepEqual(mapped[arg.Name], emptyWant[i]) {
			t.Fatalf("map[%q] = %#v, want %#v", arg.Name, mapped[arg.Name], emptyWant[i])
		}
	}
}

func TestDynamicCollectionsOfZeroSizedValues(t *testing.T) {
	t.Parallel()

	zeroArraySlice := mustNewEdgeType(t, "uint16[0][]", nil)
	emptyTupleSlice := mustNewEdgeType(t, "tuple[]", nil)
	uint16Type := mustNewEdgeType(t, "uint16", nil)
	arguments := Arguments{
		{Name: "arrays", Type: zeroArraySlice},
		{Name: "tuples", Type: emptyTupleSlice},
		{Name: "tail", Type: uint16Type},
	}
	arrays := [][0]uint16{{}, {}, {}}
	tupleValue := reflect.New(emptyTupleSlice.Elem.GetType()).Elem().Interface()
	tuples := reflect.MakeSlice(emptyTupleSlice.GetType(), 2, 2)
	for i := 0; i < tuples.Len(); i++ {
		tuples.Index(i).Set(reflect.ValueOf(tupleValue))
	}
	want := []any{arrays, tuples.Interface(), uint16(0xbeef)}
	encoded, err := arguments.Pack(want...)
	if err != nil {
		t.Fatalf("pack dynamic zero-sized collections: %v", err)
	}
	decoded, err := arguments.Unpack(encoded)
	if err != nil {
		t.Fatalf("unpack dynamic zero-sized collections: %v", err)
	}
	if !reflect.DeepEqual(decoded, want) {
		t.Fatalf("dynamic zero-sized round trip mismatch\ngot:  %#v\nwant: %#v", decoded, want)
	}

	// A zero-sized element consumes no payload bytes, so constrain its declared
	// length independently to prevent a compact CPU-denial input.
	hostile := make([]byte, 2*64)
	hostile[63] = 64
	new(big.Int).SetUint64(maxZeroSizedArrayElements + 1).FillBytes(hostile[64:128])
	if _, err := (Arguments{{Type: zeroArraySlice}}).Unpack(hostile); err == nil {
		t.Fatal("unpack accepted an excessive zero-sized dynamic array length")
	}
}

func TestDynamicFixedArrayRejectsFullWidthHighOffset(t *testing.T) {
	t.Parallel()

	arrayType := mustNewEdgeType(t, "string[1]", nil)
	// The low 64 bits point to a valid array head at byte 64. The high byte
	// makes the actual 512-bit offset enormous. A truncated uint64 decoder
	// would incorrectly accept this as [1]string{""}.
	encoded := make([]byte, 3*64)
	encoded[0] = 1
	encoded[63] = 64
	encoded[127] = 64
	_, err := (Arguments{{Type: arrayType}}).Unpack(encoded)
	if err == nil || !strings.Contains(err.Error(), "offset") {
		t.Fatalf("high-bit array offset error = %v, want offset rejection", err)
	}
}

func TestPackAndCopyNilInputsReturnErrors(t *testing.T) {
	t.Parallel()

	uint256Type := mustNewEdgeType(t, "uint256", nil)
	uint16Type := mustNewEdgeType(t, "uint16", nil)
	uintArrayType := mustNewEdgeType(t, "uint256[1]", nil)
	tupleType := mustNewEdgeType(t, "tuple", []ArgumentMarshaling{{Name: "amount", Type: "uint256"}})

	var nilBig *big.Int
	var nilUint16 *uint16
	wrongPointer := "not an integer"
	tupleWithNil := reflect.New(tupleType.GetType()).Elem().Interface()
	packTests := []struct {
		name  string
		args  Arguments
		value any
	}{
		{name: "nil interface", args: Arguments{{Type: uint256Type}}, value: nil},
		{name: "nil big integer", args: Arguments{{Type: uint256Type}}, value: nilBig},
		{name: "pointer to nil big integer", args: Arguments{{Type: uint256Type}}, value: &nilBig},
		{name: "nil scalar pointer", args: Arguments{{Type: uint16Type}}, value: nilUint16},
		{name: "nil array element", args: Arguments{{Type: uintArrayType}}, value: [1]*big.Int{nil}},
		{name: "nil tuple field", args: Arguments{{Type: tupleType}}, value: tupleWithNil},
		{name: "unrelated pointer", args: Arguments{{Type: uint256Type}}, value: &wrongPointer},
	}
	addressType := mustNewEdgeType(t, "address", nil)
	if _, err := (Arguments{{Type: addressType}}).Pack([common.AddressLength]int{}); err == nil {
		t.Fatal("Pack accepted a non-byte address array")
	}
	bytes64Type := mustNewEdgeType(t, "bytes64", nil)
	if _, err := (Arguments{{Type: bytes64Type}}).Pack([64]int{}); err == nil {
		t.Fatal("Pack accepted a non-byte fixed array")
	}
	for _, test := range packTests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := test.args.Pack(test.value); err == nil {
				t.Fatal("Pack returned nil error for nil input")
			}
		})
	}

	arguments := Arguments{{Name: "value", Type: uint256Type}}
	if err := arguments.Copy(nil, []any{big.NewInt(1)}); err == nil {
		t.Fatal("Copy accepted nil destination")
	}
	if err := arguments.Copy(nilBig, []any{big.NewInt(1)}); err == nil {
		t.Fatal("Copy accepted typed-nil destination")
	}
	var destination *big.Int
	if err := arguments.Copy(&destination, []any{nil}); err == nil {
		t.Fatal("Copy accepted nil source value")
	}
	if err := arguments.Copy(&destination, []any{nilBig}); err == nil {
		t.Fatal("Copy accepted typed-nil source value")
	}
}

func TestPublicReadersRejectMalformedWords(t *testing.T) {
	t.Parallel()

	uint512Type := mustNewEdgeType(t, "uint512", nil)
	bytes3Type := mustNewEdgeType(t, "bytes3", nil)
	word := make([]byte, 64)
	word[63] = 1
	if value, err := ReadInteger(uint512Type, word); err != nil || value.(*big.Int).Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("valid ReadInteger = %v, %v", value, err)
	}
	for _, malformed := range [][]byte{nil, make([]byte, 63), make([]byte, 65)} {
		if _, err := ReadInteger(uint512Type, malformed); err == nil {
			t.Fatalf("ReadInteger accepted %d-byte word", len(malformed))
		}
	}
	if _, err := ReadInteger(Type{T: BoolTy}, word); err == nil {
		t.Fatal("ReadInteger accepted non-integer type")
	}

	fixedWord := make([]byte, 64)
	copy(fixedWord, []byte{1, 2, 3})
	if got, err := ReadFixedBytes(bytes3Type, fixedWord); err != nil || got != [3]byte{1, 2, 3} {
		t.Fatalf("valid ReadFixedBytes = %#v, %v", got, err)
	}
	if _, err := ReadFixedBytes(bytes3Type, fixedWord[:3]); err == nil {
		t.Fatal("ReadFixedBytes accepted short word")
	}
	nonCanonical := append([]byte(nil), fixedWord...)
	nonCanonical[3] = 1
	if _, err := ReadFixedBytes(bytes3Type, nonCanonical); err == nil {
		t.Fatal("ReadFixedBytes accepted non-zero right padding")
	}
	if _, err := ReadFixedBytes(uint512Type, word); err == nil {
		t.Fatal("ReadFixedBytes accepted non-fixed-bytes type")
	}

	var topic common.LogTopic
	copy(topic[:], nonCanonical)
	fields := Arguments{{Name: "value", Type: bytes3Type, Indexed: true}}
	if err := ParseTopics(new(struct{ Value [3]byte }), fields, []common.LogTopic{topic}); err == nil {
		t.Fatal("ParseTopics accepted non-canonical bytes3 topic padding")
	}
}

func TestMakeTopicsRejectsInvalidBigIntegers(t *testing.T) {
	t.Parallel()

	var nilBig *big.Int
	tooLarge := new(big.Int).Lsh(big.NewInt(1), 512)
	tooSmall := new(big.Int).Sub(new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 511)), big.NewInt(1))
	for _, value := range []*big.Int{nilBig, tooLarge, tooSmall} {
		if _, err := MakeTopics([]any{value}); err == nil {
			t.Fatalf("MakeTopics accepted invalid integer %v", value)
		}
	}
	maxUint := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1))
	minInt := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 511))
	if _, err := MakeTopics([]any{maxUint}, []any{minInt}); err != nil {
		t.Fatalf("MakeTopics rejected 512-bit boundaries: %v", err)
	}
}

func TestParseTopicsUsesOriginalNamesAndTags(t *testing.T) {
	t.Parallel()

	uint16Type := mustNewEdgeType(t, "uint16", nil)
	fields := Arguments{
		{Name: "range", Type: uint16Type, Indexed: true},
		{Name: "_msg", Type: uint16Type, Indexed: true},
		{Name: "msg", Type: uint16Type, Indexed: true},
		{Name: "", Type: uint16Type, Indexed: true},
		{Name: "", Type: uint16Type, Indexed: true},
	}
	topics, err := MakeTopics(
		[]any{uint16(1)}, []any{uint16(2)}, []any{uint16(3)}, []any{uint16(4)}, []any{uint16(5)},
	)
	if err != nil {
		t.Fatal(err)
	}
	type result struct {
		Keyword    uint16 `abi:"range"`
		First      uint16 `abi:"_msg"`
		Second     uint16 `abi:"msg"`
		Anonymous0 uint16 `abi:""`
		Anonymous1 uint16 `abi:""`
		Other      bool   `abi:"notIndexed"`
	}
	var out result
	flattened := make([]common.LogTopic, len(topics))
	for i := range topics {
		flattened[i] = topics[i][0]
	}
	if err := ParseTopics(&out, fields, flattened); err != nil {
		t.Fatalf("ParseTopics with original-name tags: %v", err)
	}
	want := result{Keyword: 1, First: 2, Second: 3, Anonymous0: 4, Anonymous1: 5}
	if out != want {
		t.Fatalf("parsed tagged fields = %#v, want %#v", out, want)
	}
}

func TestParseTopicsRejectsConflictingFallbackTag(t *testing.T) {
	t.Parallel()

	uint16Type := mustNewEdgeType(t, "uint16", nil)
	fields := Arguments{{Name: "value", Type: uint16Type, Indexed: true}}
	topics, err := MakeTopics([]any{uint16(7)})
	if err != nil {
		t.Fatal(err)
	}
	out := struct {
		Value uint16 `abi:"other"`
	}{}
	if err := ParseTopics(&out, fields, []common.LogTopic{topics[0][0]}); err == nil || !strings.Contains(err.Error(), "conflicting abi tag") {
		t.Fatalf("ParseTopics conflicting-tag error = %v", err)
	}
	if out.Value != 0 {
		t.Fatalf("ParseTopics mutated destination on conflicting tag: %d", out.Value)
	}
}

func TestFunctionTypeCannotFitVM64Word(t *testing.T) {
	t.Parallel()

	functionType := mustNewEdgeType(t, "function", nil)
	value := [common.AddressLength + 4]byte{}
	if _, err := (Arguments{{Type: functionType}}).Pack(value); err == nil || !strings.Contains(err.Error(), "does not fit") {
		t.Fatalf("function Pack error = %v, want VM64 width error", err)
	}
	if _, err := (Arguments{{Type: functionType}}).Unpack(make([]byte, 64)); err == nil || !strings.Contains(err.Error(), "does not fit") {
		t.Fatalf("function Unpack error = %v, want VM64 width error", err)
	}
}

func TestParseSelectorMalformedInputsReturnErrors(t *testing.T) {
	t.Parallel()

	for _, selector := range []string{
		"", "missing", "broken(", "trailing()x", "bad(unknown)",
		"bad(uint512,,address)", "bad(uint8garbage)", "bad(uint8[2junk])",
		"bad(uint08)", "bad(bytes01)", "bad(uint8[01])",
		"bad(uint512 address)",
		"bad((uint256)[)", "bad((uint256)[1)", "bad((uint256)[x])",
		"bad((uint256)[-1])", "bad((uint256)[01])", "bad((uint256)[][2)",
		"bad((uint256)[]])", "bad((uint256)[2][][3]garbage)",
		"bad((uint256,,address)[2])", "bad((uint256,address)[2",
		"bad((uint256,address)[1048577])", "bad((uint256)[18446744073709551616])",
		"bad((),)", "bad((,))", "bad(((uint256)[2])", "bad((uint256)[2][[]])",
		"bad()[]", "bad()[2]", "bad((uint256))[]",
	} {
		selector := selector
		t.Run(selector, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("ParseSelector(%q) panicked: %v", selector, recovered)
				}
			}()
			if _, err := ParseSelector(selector); err == nil {
				t.Fatalf("ParseSelector(%q) unexpectedly succeeded", selector)
			}
		})
	}
}

func TestUnpackRejectsUnalignedBodies(t *testing.T) {
	t.Parallel()

	bytesType := mustNewEdgeType(t, "bytes", nil)
	arguments := Arguments{{Name: "payload", Type: bytesType}}
	encoded, err := arguments.Pack([]byte{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := arguments.Unpack(encoded[:len(encoded)-1]); err == nil || !strings.Contains(err.Error(), "multiple of 64") {
		t.Fatalf("Arguments.Unpack truncated error = %v, want alignment error", err)
	}
	if _, err := arguments.UnpackValues(encoded[:len(encoded)-1]); err == nil || !strings.Contains(err.Error(), "multiple of 64") {
		t.Fatalf("Arguments.UnpackValues truncated error = %v, want alignment error", err)
	}

	customError := NewError("PayloadError", arguments)
	errorData := append(append([]byte(nil), customError.ID[:4]...), encoded...)
	if _, err := customError.Unpack(errorData); err != nil {
		t.Fatalf("valid custom error unpack: %v", err)
	}
	if _, err := customError.Unpack(errorData[:len(errorData)-1]); err == nil || !strings.Contains(err.Error(), "multiple of 64") {
		t.Fatalf("Error.Unpack truncated error = %v, want alignment error", err)
	}

	emptyError := NewError("EmptyError", nil)
	emptyData := append(append([]byte(nil), emptyError.ID[:4]...), 1)
	if _, err := emptyError.Unpack(emptyData); err == nil || !strings.Contains(err.Error(), "multiple of 64") {
		t.Fatalf("empty Error.Unpack trailing-byte error = %v, want alignment error", err)
	}
}

func TestLegacyMethodMetadataCompatibility(t *testing.T) {
	t.Parallel()

	definition := `[
		{"type":"constructor","constant":true,"payable":true},
		{"type":"function","name":"legacyRead","constant":true,"payable":false},
		{"type":"function","name":"legacyPay","constant":false,"payable":true,"stateMutability":"nonpayable"},
		{"type":"function","name":"modernRead","constant":false,"payable":false,"stateMutability":"view"},
		{"type":"function","name":"modernPay","constant":false,"payable":false,"stateMutability":"payable"},
		{"type":"fallback","constant":false,"payable":true}
	]`
	parsed, err := JSON(strings.NewReader(definition))
	if err != nil {
		t.Fatal(err)
	}
	legacyRead := parsed.Methods["legacyRead"]
	if !legacyRead.Constant || legacyRead.Payable || !legacyRead.IsConstant() {
		t.Fatalf("legacy read metadata lost: %#v", legacyRead)
	}
	legacyPay := parsed.Methods["legacyPay"]
	if legacyPay.Constant || !legacyPay.Payable || !legacyPay.IsPayable() {
		t.Fatalf("legacy payable metadata lost: %#v", legacyPay)
	}
	if !parsed.Methods["modernRead"].IsConstant() || !parsed.Methods["modernPay"].IsPayable() {
		t.Fatal("modern stateMutability semantics were not preserved")
	}
	if !parsed.Constructor.Constant || !parsed.Constructor.Payable || !parsed.Constructor.IsConstant() || !parsed.Constructor.IsPayable() {
		t.Fatalf("constructor legacy metadata lost: %#v", parsed.Constructor)
	}
	if !parsed.Fallback.Payable || !parsed.Fallback.IsPayable() {
		t.Fatalf("fallback legacy metadata lost: %#v", parsed.Fallback)
	}
}
