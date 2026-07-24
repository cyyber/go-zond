// Copyright 2022 The go-ethereum Authors
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
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
)

type selectorTupleExpectation struct {
	suffix     string
	components []ArgumentMarshaling
}

func TestParseSelector(t *testing.T) {
	t.Parallel()
	mkType := func(types ...any) []ArgumentMarshaling {
		result := make([]ArgumentMarshaling, 0, len(types))
		for i, typeOrComponents := range types {
			name := fmt.Sprintf("name%d", i)
			if typeName, ok := typeOrComponents.(string); ok {
				result = append(result, ArgumentMarshaling{name, typeName, typeName, nil, false})
			} else if tuple, ok := typeOrComponents.(selectorTupleExpectation); ok {
				typeName := "tuple" + tuple.suffix
				result = append(result, ArgumentMarshaling{name, typeName, typeName, tuple.components, false})
			} else if components, ok := typeOrComponents.([]ArgumentMarshaling); ok {
				result = append(result, ArgumentMarshaling{name, "tuple", "tuple", components, false})
			} else if components, ok := typeOrComponents.([][]ArgumentMarshaling); ok {
				result = append(result, ArgumentMarshaling{name, "tuple[]", "tuple[]", components[0], false})
			} else {
				log.Fatalf("unexpected type %T", typeOrComponents)
			}
		}
		return result
	}
	tupleArray := func(suffix string, components []ArgumentMarshaling) selectorTupleExpectation {
		return selectorTupleExpectation{suffix: suffix, components: components}
	}
	tests := []struct {
		input string
		name  string
		args  []ArgumentMarshaling
	}{
		{"noargs()", "noargs", []ArgumentMarshaling{}},
		{"simple(uint256,uint256,uint256)", "simple", mkType("uint256", "uint256", "uint256")},
		{"other(uint256,address)", "other", mkType("uint256", "address")},
		{"withArray(uint256[],address[2],uint8[4][][5])", "withArray", mkType("uint256[]", "address[2]", "uint8[4][][5]")},
		{"singleNest(bytes32,uint8,(uint256,uint256),address)", "singleNest", mkType("bytes32", "uint8", mkType("uint256", "uint256"), "address")},
		{"multiNest(address,(uint256[],uint256),((address,bytes32),uint256))", "multiNest",
			mkType("address", mkType("uint256[]", "uint256"), mkType(mkType("address", "bytes32"), "uint256"))},
		{"arrayNest((uint256,uint256)[],bytes32)", "arrayNest", mkType([][]ArgumentMarshaling{mkType("uint256", "uint256")}, "bytes32")},
		{"multiArrayNest((uint256,uint256)[],(uint256,uint256)[])", "multiArrayNest",
			mkType([][]ArgumentMarshaling{mkType("uint256", "uint256")}, [][]ArgumentMarshaling{mkType("uint256", "uint256")})},
		{"singleArrayNestAndArray((uint256,uint256)[],bytes32[])", "singleArrayNestAndArray",
			mkType([][]ArgumentMarshaling{mkType("uint256", "uint256")}, "bytes32[]")},
		{"singleArrayNestWithArrayAndArray((uint256[],address[2],uint8[4][][5])[],bytes32[])", "singleArrayNestWithArrayAndArray",
			mkType([][]ArgumentMarshaling{mkType("uint256[]", "address[2]", "uint8[4][][5]")}, "bytes32[]")},
		{"emptyTuple(())", "emptyTuple", mkType(mkType())},
		{"emptyTupleArrays((),()[2][])", "emptyTupleArrays",
			mkType(mkType(), tupleArray("[2][]", mkType()))},
		{"fixedTupleArray((uint256,address)[2])", "fixedTupleArray",
			mkType(tupleArray("[2]", mkType("uint256", "address")))},
		{"tupleDimensions((uint256,address)[][3],(bytes32,bool)[2][],(uint8)[2][3])", "tupleDimensions",
			mkType(
				tupleArray("[][3]", mkType("uint256", "address")),
				tupleArray("[2][]", mkType("bytes32", "bool")),
				tupleArray("[2][3]", mkType("uint8")),
			)},
		{"nestedFixedTupleArray(((uint256,address)[2],bytes32)[3])", "nestedFixedTupleArray",
			mkType(tupleArray("[3]", mkType(
				tupleArray("[2]", mkType("uint256", "address")),
				"bytes32",
			)))},
		{"nestedMixedTupleArrays(((uint256,bool)[][2],(address,bytes64)[3][])[4][])", "nestedMixedTupleArrays",
			mkType(tupleArray("[4][]", mkType(
				tupleArray("[][2]", mkType("uint256", "bool")),
				tupleArray("[3][]", mkType("address", "bytes64")),
			)))},
	}
	for i, tt := range tests {
		selector, err := ParseSelector(tt.input)
		if err != nil {
			t.Errorf("test %d: failed to parse selector '%v': %v", i, tt.input, err)
		}
		if selector.Name != tt.name {
			t.Errorf("test %d: unexpected function name: '%s' != '%s'", i, selector.Name, tt.name)
		}

		if selector.Type != "function" {
			t.Errorf("test %d: unexpected type: '%s' != '%s'", i, selector.Type, "function")
		}
		if !reflect.DeepEqual(selector.Inputs, tt.args) {
			t.Errorf("test %d: unexpected args: '%v' != '%v'", i, selector.Inputs, tt.args)
		}
	}
}

func TestParseSelectorRejectsLiteralTupleKeyword(t *testing.T) {
	t.Parallel()

	for _, selector := range []string{"f(tuple)", "f(tuple[])", "f(tuple[2])"} {
		selector := selector
		t.Run(selector, func(t *testing.T) {
			t.Parallel()
			if _, err := ParseSelector(selector); err == nil {
				t.Fatalf("ParseSelector(%q) unexpectedly accepted ABI-JSON tuple keyword", selector)
			}
		})
	}
	if _, err := ParseSelector("f(())"); err != nil {
		t.Fatalf("ParseSelector parenthesized empty tuple: %v", err)
	}
}

func TestParseSelectorNestingLimit(t *testing.T) {
	t.Parallel()

	nested := func(levels int, close bool) string {
		selector := "deep(" + strings.Repeat("(", levels) + "uint512"
		if close {
			selector += strings.Repeat(")", levels) + ")"
		}
		return selector
	}

	t.Run("boundary", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("ParseSelector(valid nesting boundary) panicked: %v", recovered)
			}
		}()
		if _, err := ParseSelector(nested(maxABITypeNesting, true)); err != nil {
			t.Fatalf("ParseSelector(valid nesting boundary): %v", err)
		}
	})

	for name, selector := range map[string]string{
		"over-limit": nested(maxABITypeNesting+1, true),
		"malformed":  nested(maxABITypeNesting+1, false),
	} {
		name, selector := name, selector
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("ParseSelector(%s) panicked: %v", name, recovered)
				}
			}()
			if _, err := ParseSelector(selector); err == nil || !strings.Contains(err.Error(), "nesting exceeds safety limit") {
				t.Fatalf("ParseSelector(%s) error = %v, want nesting-limit error", name, err)
			}
		})
	}
}

func FuzzParseSelectorNeverPanics(f *testing.F) {
	for _, selector := range []string{
		"noargs()",
		"nested(((uint512,address)[2],bytes64)[][3])",
		"deep(" + strings.Repeat("(", maxABITypeNesting+1) + "uint512" + strings.Repeat(")", maxABITypeNesting+1) + ")",
		"broken(" + strings.Repeat("(", maxABITypeNesting+1),
	} {
		f.Add(selector)
	}
	f.Fuzz(func(t *testing.T, selector string) {
		selector = selector[:min(len(selector), 4096)]
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("ParseSelector(%q) panicked: %v", selector, recovered)
			}
		}()
		_, _ = ParseSelector(selector)
	})
}
