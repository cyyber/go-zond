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
	"fmt"
	"strings"
	"testing"

	"github.com/theQRL/go-qrl/crypto"
)

func pinnedSelector(signature string) [4]byte {
	var selector [4]byte
	copy(selector[:], crypto.Keccak256([]byte(signature))[:4])
	return selector
}

func TestABIJSONRejectsPinnedHyperionSelectorCollisions(t *testing.T) {
	t.Parallel()

	const selectorHex = "67e43e43"
	if got := fmt.Sprintf("%x", pinnedSelector("gsf()")); got != selectorHex {
		t.Fatalf("gsf() selector = %s, want %s", got, selectorHex)
	}
	if got := fmt.Sprintf("%x", pinnedSelector("tgeo()")); got != selectorHex {
		t.Fatalf("tgeo() selector = %s, want %s", got, selectorHex)
	}

	tests := []struct {
		name   string
		fields [2]string
	}{
		{
			name: "functions",
			fields: [2]string{
				`{"type":"function","name":"gsf","inputs":[],"outputs":[]}`,
				`{"type":"function","name":"tgeo","inputs":[],"outputs":[]}`,
			},
		},
		{
			name: "custom errors",
			fields: [2]string{
				`{"type":"error","name":"gsf","inputs":[]}`,
				`{"type":"error","name":"tgeo","inputs":[]}`,
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			for _, fields := range [][2]string{test.fields, {test.fields[1], test.fields[0]}} {
				definition := "[" + fields[0] + "," + fields[1] + "]"
				_, err := JSON(strings.NewReader(definition))
				if err == nil {
					t.Fatalf("ABI JSON accepted colliding definitions: %s", definition)
				}
				for _, want := range []string{"selector collision", "gsf()", "tgeo()", "0x" + selectorHex} {
					if !strings.Contains(err.Error(), want) {
						t.Fatalf("collision error %q does not contain %q", err, want)
					}
				}
			}
		})
	}
}

func TestABIJSONCustomErrorSelectorRestrictions(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name      string
		signature string
		inputs    string
		selector  [4]byte
	}{
		{
			name:      "zero",
			signature: "buyAndFree22457070633(uint256)",
			inputs:    `[{"name":"value","type":"uint256"}]`,
			selector:  [4]byte{},
		},
		{
			name:      "all ones",
			signature: "test266151307()",
			inputs:    `[]`,
			selector:  [4]byte{0xff, 0xff, 0xff, 0xff},
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := pinnedSelector(test.signature); got != test.selector {
				t.Fatalf("%s selector = %#x, want %#x", test.signature, got, test.selector)
			}
			errorName := test.signature[:strings.IndexByte(test.signature, '(')]
			definition := fmt.Sprintf(
				`[{"type":"error","name":%q,"inputs":%s}]`,
				errorName,
				test.inputs,
			)
			_, err := JSON(strings.NewReader(definition))
			if err == nil {
				t.Fatalf("ABI JSON accepted reserved custom-error selector for %s", test.signature)
			}
			for _, want := range []string{"reserved selector", test.signature, fmt.Sprintf("%#x", test.selector)} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("reserved-selector error %q does not contain %q", err, want)
				}
			}
		})
	}
}

func TestABIJSONRejectsReservedCustomErrorNames(t *testing.T) {
	t.Parallel()

	for _, definition := range []string{
		`[{"type":"error","name":"Error","inputs":[{"name":"reason","type":"string"}]}]`,
		`[{"type":"error","name":"Error","inputs":[{"name":"tag","type":"bytes64"}]}]`,
		`[{"type":"error","name":"Panic","inputs":[{"name":"code","type":"uint256"}]}]`,
		`[{"type":"error","name":"Panic","inputs":[]}]`,
	} {
		if _, err := JSON(strings.NewReader(definition)); err == nil || !strings.Contains(err.Error(), "is reserved") {
			t.Fatalf("reserved custom-error name result = %v, want rejection for %s", err, definition)
		}
	}
}

func TestABIJSONAllowsCanonicalDuplicateCustomErrors(t *testing.T) {
	t.Parallel()

	const canonicalDuplicates = `[
		{"type":"error","name":"Repeated","inputs":[
			{"name":"payload","type":"tuple","internalType":"struct First.Payload","components":[
				{"name":"amount","type":"uint512"},
				{"name":"note","type":"string"}
			]}
		]},
		{"type":"error","name":"Repeated","inputs":[
			{"name":"other","type":"tuple","internalType":"struct First.Payload","components":[
				{"name":"amount","type":"uint512"},
				{"name":"note","type":"string"}
			]}
		]},
		{"type":"error","name":"Repeated","inputs":[
			{"name":"payload","type":"tuple","internalType":"struct First.Payload","components":[
				{"name":"value","type":"uint512"},
				{"name":"memo","type":"string"}
			]}
		]},
		{"type":"error","name":"Repeated","inputs":[
			{"name":"payload","type":"tuple","internalType":"struct Second.Record","components":[
				{"name":"amount","type":"uint512"},
				{"name":"note","type":"string"}
			]}
		]}
	]`
	parsed, err := JSON(strings.NewReader(canonicalDuplicates))
	if err != nil {
		t.Fatalf("canonical inherited custom-error duplicates were rejected: %v", err)
	}
	if len(parsed.Errors) != 4 {
		t.Fatalf("parsed %d canonical duplicate errors, want 4", len(parsed.Errors))
	}
	for _, key := range []string{"Repeated", "Repeated0", "Repeated1", "Repeated2"} {
		if got := parsed.Errors[key].Sig; got != "Repeated((uint512,string))" {
			t.Fatalf("%s signature = %q, want canonical tuple signature", key, got)
		}
	}
	if got := parsed.Errors["Repeated0"].Inputs[0].Name; got != "other" {
		t.Fatalf("duplicate argument name = %q, want preserved metadata", got)
	}
	if got := parsed.Errors["Repeated1"].Inputs[0].Type.TupleRawNames; fmt.Sprint(got) != "[value memo]" {
		t.Fatalf("duplicate tuple component names = %v, want preserved metadata", got)
	}
	if got := parsed.Errors["Repeated2"].Inputs[0].Type.TupleRawName; got != "SecondRecord" {
		t.Fatalf("duplicate internal type = %q, want preserved metadata", got)
	}
}

func TestABISelectorLookupsAreStableForAcceptedDuplicates(t *testing.T) {
	t.Parallel()

	const definition = `[
		{"type":"function","name":"repeat","inputs":[{"name":"value","type":"uint512"}],"outputs":[]},
		{"type":"function","name":"repeat","inputs":[{"name":"value","type":"uint512"}],"outputs":[]},
		{"type":"error","name":"Repeated","inputs":[{"name":"firstDeclaration","type":"uint512"}]},
		{"type":"error","name":"Repeated","inputs":[{"name":"laterDeclaration","type":"uint512"}]}
	]`
	for parseIteration := 0; parseIteration < 32; parseIteration++ {
		parsed, err := JSON(strings.NewReader(definition))
		if err != nil {
			t.Fatalf("parse iteration %d: %v", parseIteration, err)
		}
		if got := parsed.Methods["repeat0"].Sig; got != "repeat(uint512)" {
			t.Fatalf("parse iteration %d: duplicate method signature = %q", parseIteration, got)
		}
		if got := parsed.Errors["Repeated0"].Sig; got != "Repeated(uint512)" {
			t.Fatalf("parse iteration %d: duplicate error signature = %q", parseIteration, got)
		}

		methodSelector := parsed.Methods["repeat"].ID
		var errorSelector [4]byte
		repeatedError := parsed.Errors["Repeated"]
		copy(errorSelector[:], repeatedError.ID[:4])
		for lookupIteration := 0; lookupIteration < 64; lookupIteration++ {
			method, err := parsed.MethodById(methodSelector)
			if err != nil {
				t.Fatalf("method lookup iteration %d: %v", lookupIteration, err)
			}
			if method.Name != "repeat" || method.Sig != "repeat(uint512)" {
				t.Fatalf("method lookup iteration %d = {%q, %q}, want first normalized definition",
					lookupIteration, method.Name, method.Sig)
			}

			abiError, err := parsed.ErrorByID(errorSelector)
			if err != nil {
				t.Fatalf("error lookup iteration %d: %v", lookupIteration, err)
			}
			if abiError.Name != "Repeated" || abiError.Sig != "Repeated(uint512)" ||
				abiError.Inputs[0].Name != "firstDeclaration" {
				t.Fatalf("error lookup iteration %d = {%q, %q, %q}, want first declaration",
					lookupIteration, abiError.Name, abiError.Sig, abiError.Inputs[0].Name)
			}
		}
	}
}
