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
	"math/big"
	"strings"
	"testing"

	"github.com/theQRL/go-qrl/crypto"
)

const collidingErrorsJSON = `[
	{"type":"error","name":"E3","inputs":[]},
	{"type":"error","name":"E3","inputs":[{"name":"value","type":"uint512"}]},
	{"type":"error","name":"E30","inputs":[{"name":"tag","type":"bytes64"}]},
	{"type":"error","name":"E3","inputs":[{"name":"account","type":"address"}]},
	{"type":"error","name":"E3","inputs":[{"name":"value","type":"uint512"}]}
]`

func TestCustomErrorNameCollisionsPreserveEveryDefinition(t *testing.T) {
	t.Parallel()

	expected := map[string]struct {
		rawName string
		sig     string
	}{
		"E3":   {"E3", "E3()"},
		"E30":  {"E3", "E3(uint512)"},
		"E300": {"E30", "E30(bytes64)"},
		"E31":  {"E3", "E3(address)"},
		"E32":  {"E3", "E3(uint512)"},
	}
	for iteration := 0; iteration < 16; iteration++ {
		parsed, err := JSON(strings.NewReader(collidingErrorsJSON))
		if err != nil {
			t.Fatalf("iteration %d: failed to parse ABI: %v", iteration, err)
		}
		if len(parsed.Errors) != len(expected) {
			t.Fatalf("iteration %d: parsed %d errors, want %d", iteration, len(parsed.Errors), len(expected))
		}
		for key, want := range expected {
			got, ok := parsed.Errors[key]
			if !ok {
				t.Fatalf("iteration %d: normalized error key %q is missing", iteration, key)
			}
			if got.Name != want.rawName {
				t.Errorf("iteration %d, key %q: raw name = %q, want %q", iteration, key, got.Name, want.rawName)
			}
			if got.Sig != want.sig {
				t.Errorf("iteration %d, key %q: signature = %q, want %q", iteration, key, got.Sig, want.sig)
			}
			wantID := crypto.Keccak256Hash([]byte(want.sig))
			if got.ID != wantID {
				t.Errorf("iteration %d, key %q: ID = %x, want %x", iteration, key, got.ID, wantID)
			}
		}
	}
}

func TestCustomErrorNormalizedKeyPackAndUnpack(t *testing.T) {
	t.Parallel()

	parsed, err := JSON(strings.NewReader(collidingErrorsJSON))
	if err != nil {
		t.Fatal(err)
	}
	emptyError := parsed.Errors["E3"]
	emptyBody, err := emptyError.Inputs.Pack()
	if err != nil {
		t.Fatalf("failed to pack E3() through normalized key: %v", err)
	}
	if len(emptyBody) != 0 {
		t.Fatalf("E3() body is %d bytes, want zero", len(emptyBody))
	}
	emptyValues, err := parsed.Unpack("E3", emptyBody)
	if err != nil {
		t.Fatalf("failed to unpack E3() through normalized key: %v", err)
	}
	if len(emptyValues) != 0 {
		t.Fatalf("E3() unpacked to %d values, want zero", len(emptyValues))
	}

	maxUint512 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1))
	uintError := parsed.Errors["E30"]
	uintBody, err := uintError.Inputs.Pack(maxUint512)
	if err != nil {
		t.Fatalf("failed to pack %s through normalized key E30: %v", uintError.Sig, err)
	}
	if len(uintBody) != 64 {
		t.Fatalf("%s body is %d bytes, want one VM64 word", uintError.Sig, len(uintBody))
	}

	values, err := parsed.Unpack("E30", uintBody)
	if err != nil {
		t.Fatalf("Unpack by normalized key failed: %v", err)
	}
	if len(values) != 1 || values[0].(*big.Int).Cmp(maxUint512) != 0 {
		t.Fatalf("Unpack by normalized key = %#v, want %v", values, maxUint512)
	}

	mapped := make(map[string]any)
	if err := parsed.UnpackIntoMap(mapped, "E30", uintBody); err != nil {
		t.Fatalf("UnpackIntoMap by normalized key failed: %v", err)
	}
	if got, ok := mapped["value"].(*big.Int); !ok || got.Cmp(maxUint512) != 0 {
		t.Fatalf("UnpackIntoMap value = %#v, want %v", mapped["value"], maxUint512)
	}

	var decoded struct {
		Value *big.Int
	}
	if err := parsed.UnpackIntoInterface(&decoded, "E30", uintBody); err != nil {
		t.Fatalf("UnpackIntoInterface by normalized key failed: %v", err)
	}
	if decoded.Value == nil || decoded.Value.Cmp(maxUint512) != 0 {
		t.Fatalf("UnpackIntoInterface value = %v, want %v", decoded.Value, maxUint512)
	}

	fullError := append(append([]byte(nil), uintError.ID[:4]...), uintBody...)
	unpacked, err := uintError.Unpack(fullError)
	if err != nil {
		t.Fatalf("Error.Unpack failed for normalized-key value: %v", err)
	}
	errorValues, ok := unpacked.([]any)
	if !ok || len(errorValues) != 1 || errorValues[0].(*big.Int).Cmp(maxUint512) != 0 {
		t.Fatalf("Error.Unpack = %#v, want %v", unpacked, maxUint512)
	}

	// The repeated definition remains independently addressable even though it
	// has the same raw name, signature, and selector.
	repeated := parsed.Errors["E32"]
	repeatedBody, err := repeated.Inputs.Pack(maxUint512)
	if err != nil {
		t.Fatalf("failed to pack repeated definition: %v", err)
	}
	if string(repeatedBody) != string(uintBody) {
		t.Fatal("repeated custom-error definition produced a different encoding")
	}
}

func TestCustomErrorNormalizedKeysRemainDiscoverableByID(t *testing.T) {
	t.Parallel()

	parsed, err := JSON(strings.NewReader(collidingErrorsJSON))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"E3", "E30", "E300", "E31", "E32"} {
		want := parsed.Errors[key]
		var selector [4]byte
		copy(selector[:], want.ID[:4])
		got, err := parsed.ErrorByID(selector)
		if err != nil {
			t.Fatalf("ErrorByID(%s) failed: %v", key, err)
		}
		// E30 and E32 intentionally have the same signature and selector, so
		// either normalized entry is an equivalent successful lookup.
		if got.Name != want.Name || got.Sig != want.Sig || got.ID != want.ID {
			t.Fatalf("ErrorByID(%s) = {%q, %q, %x}, want {%q, %q, %x}",
				key, got.Name, got.Sig, got.ID, want.Name, want.Sig, want.ID)
		}
	}
}

func TestUnpackRejectsCrossCategoryNameCollisions(t *testing.T) {
	t.Parallel()

	const (
		method  = `{"type":"function","name":"Collision","inputs":[],"outputs":[{"name":"value","type":"uint512"}]}`
		event   = `{"type":"event","name":"Collision","inputs":[{"name":"value","type":"uint512","indexed":false}]}`
		failure = `{"type":"error","name":"Collision","inputs":[{"name":"tag","type":"bytes64"}]}`
	)
	tests := []struct {
		name       string
		definition string
		want       string
	}{
		{
			name:       "method and error",
			definition: `[` + method + `,` + failure + `]`,
			want:       `abi: name "Collision" is ambiguous between method and error`,
		},
		{
			name:       "event and error",
			definition: `[` + event + `,` + failure + `]`,
			want:       `abi: name "Collision" is ambiguous between event and error`,
		},
		{
			name:       "method and event",
			definition: `[` + method + `,` + event + `]`,
			want:       `abi: name "Collision" is ambiguous between method and event`,
		},
		{
			name:       "method, event, and error",
			definition: `[` + method + `,` + event + `,` + failure + `]`,
			want:       `abi: name "Collision" is ambiguous between method and event and error`,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := JSON(strings.NewReader(test.definition))
			if err != nil {
				t.Fatal(err)
			}
			payload := make([]byte, 64)
			payload[len(payload)-1] = 1

			unpackCalls := []struct {
				name string
				call func() error
			}{
				{
					name: "Unpack",
					call: func() error {
						_, err := parsed.Unpack("Collision", payload)
						return err
					},
				},
				{
					name: "UnpackIntoInterface",
					call: func() error {
						var decoded struct {
							Value *big.Int
							Tag   [64]byte
						}
						return parsed.UnpackIntoInterface(&decoded, "Collision", payload)
					},
				},
				{
					name: "UnpackIntoMap",
					call: func() error {
						return parsed.UnpackIntoMap(make(map[string]any), "Collision", payload)
					},
				},
			}
			for _, unpackCall := range unpackCalls {
				unpackCall := unpackCall
				t.Run(unpackCall.name, func(t *testing.T) {
					t.Parallel()

					if err := unpackCall.call(); err == nil {
						t.Fatalf("ambiguous name was accepted")
					} else if err.Error() != test.want {
						t.Fatalf("error = %q, want %q", err, test.want)
					}
				})
			}
		})
	}
}

func TestABIJSONRejectsUnsupportedFixedPointTypes(t *testing.T) {
	t.Parallel()

	for _, typeName := range []string{"fixed", "fixed128x18", "ufixed", "ufixed128x18"} {
		typeName := typeName
		t.Run(typeName, func(t *testing.T) {
			t.Parallel()
			definition := fmt.Sprintf(
				`[{"type":"error","name":"FixedPoint","inputs":[{"name":"value","type":%q}]}]`,
				typeName,
			)
			if _, err := JSON(strings.NewReader(definition)); err == nil {
				t.Fatalf("ABI JSON accepted unsupported fixed-point type %q", typeName)
			}
		})
	}
}
