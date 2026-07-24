// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.
//
// The go-qrl library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package abi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type collidingTupleFields struct {
	Data  []byte
	Data0 []byte
	Data1 []byte
}

func TestTupleCollidingNamesRoundTripNestedArrays(t *testing.T) {
	contract, err := JSON(strings.NewReader(`[
		{"inputs":[{"components":[
			{"name":"data","type":"bytes"},
			{"name":"_data","type":"bytes"},
			{"name":"data","type":"bytes"}
		],"internalType":"struct Collision.Record[][2]","name":"records","type":"tuple[][2]"}],
		"name":"roundTrip","outputs":[{"components":[
			{"name":"data","type":"bytes"},
			{"name":"_data","type":"bytes"},
			{"name":"data","type":"bytes"}
		],"internalType":"struct Collision.Record[][2]","name":"records","type":"tuple[][2]"}],
		"stateMutability":"pure","type":"function"}
	]`))
	if err != nil {
		t.Fatalf("JSON() failed: %v", err)
	}
	want := [2][]collidingTupleFields{
		{
			{Data: []byte{0x01}, Data0: []byte{0x02, 0x03}, Data1: []byte{0x04}},
			{Data: []byte{}, Data0: []byte{}, Data1: []byte{0x05, 0x06, 0x07}},
		},
		{
			{Data: bytes.Repeat([]byte{0xaa}, 65), Data0: []byte{0xbb}, Data1: bytes.Repeat([]byte{0xcc}, 64)},
		},
	}
	method := contract.Methods["roundTrip"]
	encoded, err := method.Inputs.Pack(want)
	if err != nil {
		t.Fatalf("Pack() failed for colliding tuple names: %v", err)
	}
	values, err := method.Outputs.Unpack(encoded)
	if err != nil {
		t.Fatalf("Unpack() failed for colliding tuple names: %v", err)
	}
	var got [2][]collidingTupleFields
	if err := method.Outputs.Copy(&got, values); err != nil {
		t.Fatalf("Copy() failed for colliding tuple names: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip mismatch\n got: %#v\nwant: %#v", got, want)
	}
	encodedAgain, err := method.Inputs.Pack(got)
	if err != nil {
		t.Fatalf("second Pack() failed: %v", err)
	}
	if !bytes.Equal(encodedAgain, encoded) {
		t.Fatalf("repacked bytes differ\n got: %x\nwant: %x", encodedAgain, encoded)
	}
}

func TestTupleDuplicateTagsMapByOccurrence(t *testing.T) {
	tuple, err := NewType("tuple", "struct Tagged.Record", []ArgumentMarshaling{
		{Name: "data", Type: "bytes"},
		{Name: "_data", Type: "bytes"},
		{Name: "data", Type: "bytes"},
	})
	if err != nil {
		t.Fatalf("NewType() failed: %v", err)
	}
	tagged := struct {
		First  []byte `abi:"data"`
		Second []byte `abi:"_data"`
		Third  []byte `abi:"data"`
	}{[]byte{1}, []byte{2}, []byte{3}}
	encoded, err := (Arguments{{Name: "record", Type: tuple}}).Pack(tagged)
	if err != nil {
		t.Fatalf("Pack() failed for repeated abi tags: %v", err)
	}
	values, err := (Arguments{{Name: "record", Type: tuple}}).Unpack(encoded)
	if err != nil {
		t.Fatalf("Unpack() failed: %v", err)
	}
	decoded := reflect.ValueOf(values[0])
	for i, want := range [][]byte{{1}, {2}, {3}} {
		if got := decoded.Field(i).Bytes(); !bytes.Equal(got, want) {
			t.Fatalf("component %d = %x, want %x", i, got, want)
		}
	}
}

func TestNewTypeRejectsHostileArrayShapes(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	cases := []string{
		"uint512[" + strconv.Itoa(maxInt) + "]",
		"bytes[" + strconv.Itoa(maxInt/64+1) + "]",
		"uint512[" + strconv.Itoa(maxABIFixedArrayLength+1) + "]",
		"uint512[1073741824][1073741824]",
		"tuple[" + strconv.Itoa(maxZeroSizedArrayElements+1) + "]",
		"uint512" + strings.Repeat("[1]", maxABITypeNesting+1),
		"uint512[999999999999999999999999999999999999]",
	}
	for _, typeName := range cases {
		t.Run(typeName, func(t *testing.T) {
			_, err := NewType(typeName, "", nil)
			if err == nil {
				t.Fatalf("NewType(%q) succeeded, want a bounded error", typeName)
			}
		})
	}
}

func TestNewTypeRejectsDeepTupleJSON(t *testing.T) {
	component := ArgumentMarshaling{Name: "value", Type: "uint512"}
	for i := 0; i <= maxABITypeNesting; i++ {
		component = ArgumentMarshaling{
			Name:       fmt.Sprintf("level%d", i),
			Type:       "tuple",
			Components: []ArgumentMarshaling{component},
		}
	}
	raw, err := json.Marshal([]ArgumentMarshaling{component})
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}
	var arguments Arguments
	if err := json.Unmarshal(raw, &arguments); err == nil {
		t.Fatal("deep tuple JSON unexpectedly parsed")
	} else if !strings.Contains(err.Error(), "nesting exceeds safety limit") {
		t.Fatalf("unexpected deep tuple error: %v", err)
	}
}

func TestNewTypeRejectsOversizedTupleShape(t *testing.T) {
	components := make([]ArgumentMarshaling, maxABITupleFields+1)
	for i := range components {
		components[i] = ArgumentMarshaling{Name: fmt.Sprintf("field%d", i), Type: "uint512"}
	}
	if _, err := NewType("tuple", "struct Hostile.Fields", components); err == nil {
		t.Fatal("tuple with excessive field count unexpectedly parsed")
	}

	largeArray := "uint512[" + strconv.Itoa(maxABIFixedArrayLength) + "]"
	if _, err := NewType("tuple", "struct Hostile.Static", []ArgumentMarshaling{
		{Name: "first", Type: largeArray},
		{Name: "second", Type: largeArray},
	}); err == nil {
		t.Fatal("tuple exceeding the static-size limit unexpectedly parsed")
	}

	// A tuple containing any dynamic member has a small ABI head, but its Go
	// reflected struct can still be enormous. Nine maximum-length string arrays
	// exceed the reflected-size cap on both 32-bit and 64-bit platforms.
	dynamicComponents := make([]ArgumentMarshaling, 9)
	for i := range dynamicComponents {
		dynamicComponents[i] = ArgumentMarshaling{
			Name: fmt.Sprintf("strings%d", i),
			Type: "string[" + strconv.Itoa(maxABIFixedArrayLength) + "]",
		}
	}
	if _, err := NewType("tuple", "struct Hostile.Dynamic", dynamicComponents); err == nil || !strings.Contains(err.Error(), "reflected type exceeds safety limit") {
		t.Fatalf("dynamic tuple with oversized reflected type error = %v", err)
	}
}

func TestCheckedTypeSizeRejectsOverflow(t *testing.T) {
	elem := Type{T: UintTy, Size: 512, stringKind: "uint512"}
	hostile := Type{
		T:          ArrayTy,
		Size:       int(^uint(0)>>1)/64 + 1,
		Elem:       &elem,
		stringKind: "uint512[huge]",
	}
	if _, err := getTypeSizeChecked(hostile); err == nil {
		t.Fatal("getTypeSizeChecked() accepted an overflowing static array")
	}
}

func FuzzABITypeConstructionNeverPanics(f *testing.F) {
	for _, typeName := range []string{
		"uint512[2][3]",
		"bytes[][0][1]",
		"tuple[1048577]",
		"uint512" + strings.Repeat("[1]", maxABITypeNesting+1),
		"uint512[999999999999999999999999999999999999]",
		"[[]]",
	} {
		f.Add(typeName)
	}
	f.Fuzz(func(t *testing.T, typeName string) {
		// Bound work without a Skip-only coverage edge. Native fuzzing otherwise
		// spends most of a short run minimizing the first input just above the
		// threshold instead of exercising the parser.
		typeName = typeName[:min(len(typeName), 4096)]
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("type construction panicked for %q: %v", typeName, recovered)
			}
		}()
		if typ, err := NewType(typeName, "", nil); err == nil {
			_ = typ.GetType()
			if _, err := getTypeSizeChecked(typ); err != nil {
				t.Fatalf("successfully constructed type %q has invalid size: %v", typeName, err)
			}
		}
		raw, err := json.Marshal([]ArgumentMarshaling{{Name: "value", Type: typeName}})
		if err != nil {
			t.Fatalf("json.Marshal() failed: %v", err)
		}
		var arguments Arguments
		if err := json.Unmarshal(raw, &arguments); err == nil {
			for _, argument := range arguments {
				_ = argument.Type.GetType()
			}
		}
	})
}
