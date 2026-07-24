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
	"reflect"
	"strings"
	"testing"
)

func TestRecursiveCanonicalScalarPadding(t *testing.T) {
	tupleType, err := NewType("tuple[2][]", "struct Canonical.Record[2][]", []ArgumentMarshaling{
		{Name: "small", Type: "uint8"},
		{Name: "signed", Type: "int8"},
		{Name: "enabled", Type: "bool"},
		{Name: "tag", Type: "bytes3"},
	})
	if err != nil {
		t.Fatal(err)
	}
	value := reflect.MakeSlice(tupleType.GetType(), 1, 1)
	array := reflect.New(tupleType.GetType().Elem()).Elem()
	first := reflect.New(array.Type().Elem()).Elem()
	first.Field(0).SetUint(0x7a)
	first.Field(1).SetInt(-2)
	first.Field(2).SetBool(true)
	first.Field(3).Set(reflect.ValueOf([3]byte{0xaa, 0xbb, 0xcc}))
	second := reflect.New(array.Type().Elem()).Elem()
	second.Field(0).SetUint(0x21)
	second.Field(1).SetInt(-7)
	second.Field(2).SetBool(false)
	second.Field(3).Set(reflect.ValueOf([3]byte{0x11, 0x22, 0x33}))
	array.Index(0).Set(first)
	array.Index(1).Set(second)
	value.Index(0).Set(array)

	arguments := Arguments{{Name: "records", Type: tupleType}}
	encoded, err := arguments.Pack(value.Interface())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := arguments.Unpack(encoded); err != nil {
		t.Fatalf("canonical nested tuple-array-slice rejected: %v", err)
	}

	uint8Type, _ := NewType("uint8", "", nil)
	int8Type, _ := NewType("int8", "", nil)
	boolType, _ := NewType("bool", "", nil)
	bytes3Type, _ := NewType("bytes3", "", nil)
	tests := []struct {
		name    string
		typ     Type
		value   any
		corrupt func([]byte)
	}{
		{
			name:  "uint8 leading padding",
			typ:   uint8Type,
			value: uint8(0x7a),
			corrupt: func(word []byte) {
				word[0] = 1
			},
		},
		{
			name:  "int8 sign extension",
			typ:   int8Type,
			value: int8(-2),
			corrupt: func(word []byte) {
				word[0] = 0
			},
		},
		{
			name:  "bool leading padding",
			typ:   boolType,
			value: true,
			corrupt: func(word []byte) {
				word[0] = 1
			},
		},
		{
			name:  "bytes3 trailing padding",
			typ:   bytes3Type,
			value: [3]byte{0xaa, 0xbb, 0xcc},
			corrupt: func(word []byte) {
				word[len(word)-1] = 1
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			word, err := (Arguments{{Type: test.typ}}).Pack(test.value)
			if err != nil {
				t.Fatal(err)
			}
			malformed := append([]byte(nil), encoded...)
			index := alignedWordIndex(malformed, word)
			if index < 0 {
				t.Fatalf("canonical scalar word %x not found in nested payload", word)
			}
			test.corrupt(malformed[index : index+64])
			if _, err := arguments.Unpack(malformed); err == nil {
				t.Fatal("nested scalar with non-canonical padding was accepted")
			}
		})
	}
}

func TestNestedDynamicOffsetTablesLooseModeAndHostileLayouts(t *testing.T) {
	type aggregateCase struct {
		name      string
		arguments Arguments
		value     any
		table     func([]byte) int
		strings   func([]any) []string
	}

	stringSlice, err := NewType("string[]", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	stringTuple, err := NewType("tuple", "struct Offsets.Pair", []ArgumentMarshaling{
		{Name: "first", Type: "string"},
		{Name: "second", Type: "string"},
	})
	if err != nil {
		t.Fatal(err)
	}
	tupleValue := reflect.New(stringTuple.GetType()).Elem()
	tupleValue.Field(0).SetString("alpha")
	tupleValue.Field(1).SetString("bravo")
	stringArray, err := NewType("string[2]", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	string3D, err := NewType("string[][][]", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	cases := []aggregateCase{
		{
			name:      "slice",
			arguments: Arguments{{Type: stringSlice}},
			value:     []string{"alpha", "bravo"},
			table: func(data []byte) int {
				return wordInt(data, 0) + 64 // Skip the slice length.
			},
			strings: func(values []any) []string {
				return values[0].([]string)
			},
		},
		{
			name:      "tuple",
			arguments: Arguments{{Type: stringTuple}},
			value:     tupleValue.Interface(),
			table: func(data []byte) int {
				return wordInt(data, 0)
			},
			strings: func(values []any) []string {
				tuple := reflect.ValueOf(values[0])
				return []string{tuple.Field(0).String(), tuple.Field(1).String()}
			},
		},
		{
			name:      "fixed array of dynamic values",
			arguments: Arguments{{Type: stringArray}},
			value:     [2]string{"alpha", "bravo"},
			table: func(data []byte) int {
				return wordInt(data, 0)
			},
			strings: func(values []any) []string {
				array := values[0].([2]string)
				return array[:]
			},
		},
		{
			name:      "three dimensional dynamic array",
			arguments: Arguments{{Type: string3D}},
			value:     [][][]string{{{"alpha", "bravo"}}},
			table: func(data []byte) int {
				outerElements := wordInt(data, 0) + 64
				middle := outerElements + wordInt(data, outerElements)
				middleElements := middle + 64
				inner := middleElements + wordInt(data, middleElements)
				return inner + 64
			},
			strings: func(values []any) []string {
				return values[0].([][][]string)[0][0]
			},
		},
	}

	looseLayouts := []struct {
		name string
		edit func([]byte, int) []byte
		want []string
	}{
		{
			name: "offset into head",
			edit: func(data []byte, table int) []byte {
				putWordInt(data, table, 0)
				return data
			},
			want: []string{"", "bravo"},
		},
		{
			name: "nonmonotonic disjoint tails",
			edit: func(data []byte, table int) []byte {
				putWordInt(data, table, 256)
				putWordInt(data, table+64, 128)
				return data
			},
			want: []string{"bravo", "alpha"},
		},
		{
			name: "aliased tail",
			edit: func(data []byte, table int) []byte {
				putWordInt(data, table, 128)
				putWordInt(data, table+64, 128)
				return data
			},
			want: []string{"alpha", "alpha"},
		},
		{
			name: "gap between tails",
			edit: func(data []byte, table int) []byte {
				secondTail := table + 256
				withGap := make([]byte, 0, len(data)+64)
				withGap = append(withGap, data[:secondTail]...)
				withGap = append(withGap, make([]byte, 64)...)
				withGap = append(withGap, data[secondTail:]...)
				putWordInt(withGap, table+64, 320)
				return withGap
			},
			want: []string{"alpha", "bravo"},
		},
	}

	hostileLayouts := []struct {
		name string
		edit func([]byte, int)
		want string
	}{
		{
			name: "partially overlapping tail",
			edit: func(data []byte, table int) {
				putWordInt(data, table+64, 192)
			},
		},
		{
			name: "truncated tail",
			edit: func(data []byte, table int) {
				putWordInt(data, table+64, len(data)-table)
			},
		},
		{
			name: "high bit offset",
			edit: func(data []byte, table int) {
				clear(data[table : table+64])
				data[table] = 0x80
				data[table+63] = 0x80
			},
			want: "offset larger than int",
		},
	}

	for _, aggregate := range cases {
		t.Run(aggregate.name, func(t *testing.T) {
			encoded, err := aggregate.arguments.Pack(aggregate.value)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := aggregate.arguments.Unpack(encoded); err != nil {
				t.Fatalf("canonical payload rejected: %v", err)
			}
			table := aggregate.table(encoded)
			if table < 0 || table+128 > len(encoded) {
				t.Fatalf("offset table [%d:%d] outside payload length %d", table, table+128, len(encoded))
			}
			if got := wordInt(encoded, table); got != 128 {
				t.Fatalf("first canonical child offset = %d, want 128", got)
			}
			if got := wordInt(encoded, table+64); got != 256 {
				t.Fatalf("second canonical child offset = %d, want 256", got)
			}
			for _, layout := range looseLayouts {
				t.Run(layout.name, func(t *testing.T) {
					loose := layout.edit(append([]byte(nil), encoded...), table)
					values, err := aggregate.arguments.Unpack(loose)
					if err != nil {
						t.Fatalf("loose-mode payload rejected: %v", err)
					}
					if got := aggregate.strings(values); !reflect.DeepEqual(got, layout.want) {
						t.Fatalf("loose-mode decoded strings = %#v, want %#v", got, layout.want)
					}
				})
			}
			for _, layout := range hostileLayouts {
				t.Run(layout.name, func(t *testing.T) {
					malformed := append([]byte(nil), encoded...)
					layout.edit(malformed, table)
					_, err := aggregate.arguments.Unpack(malformed)
					if err == nil {
						t.Fatal("malformed payload was accepted")
					}
					if layout.want != "" && !strings.Contains(err.Error(), layout.want) {
						t.Fatalf("malformed payload error = %v, want error containing %q", err, layout.want)
					}
				})
			}
		})
	}
}

func alignedWordIndex(data, word []byte) int {
	for index := 0; index+64 <= len(data); index += 64 {
		if bytes.Equal(data[index:index+64], word) {
			return index
		}
	}
	return -1
}

func wordInt(data []byte, index int) int {
	value := 0
	for _, b := range data[index : index+64] {
		value = value<<8 | int(b)
	}
	return value
}

func putWordInt(data []byte, index, value int) {
	clear(data[index : index+64])
	for position := index + 63; value > 0; position-- {
		data[position] = byte(value)
		value >>= 8
	}
}
