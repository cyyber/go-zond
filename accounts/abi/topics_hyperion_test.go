// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package abi

import (
	"encoding/hex"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/crypto"
)

func mustIndexedTopicType(t *testing.T, name string, components []ArgumentMarshaling) Type {
	t.Helper()
	typ, err := NewType(name, "", components)
	if err != nil {
		t.Fatalf("NewType(%q): %v", name, err)
	}
	return typ
}

func goldenIndexedTopic(t *testing.T, digest string) common.LogTopic {
	t.Helper()
	decoded, err := hex.DecodeString(digest)
	if err != nil {
		t.Fatalf("invalid golden digest %q: %v", digest, err)
	}
	if len(decoded) != common.HashLength {
		t.Fatalf("golden digest has %d bytes, want %d", len(decoded), common.HashLength)
	}
	var topic common.LogTopic
	copy(topic[:common.HashLength], decoded)
	return topic
}

func TestMakeTopicValueTypesVM64(t *testing.T) {
	t.Parallel()

	t.Run("uint512 resolves big.Int signedness from the ABI type", func(t *testing.T) {
		t.Parallel()
		typ := mustIndexedTopicType(t, "uint512", nil)
		value := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1))
		got, err := MakeTopic(typ, value)
		if err != nil {
			t.Fatalf("MakeTopic: %v", err)
		}
		var want common.LogTopic
		for i := range want {
			want[i] = 0xff
		}
		if got != want {
			t.Fatalf("uint512 topic = %x, want %x", got, want)
		}
	})

	t.Run("int512 minimum is sign extended to one VM64 word", func(t *testing.T) {
		t.Parallel()
		typ := mustIndexedTopicType(t, "int512", nil)
		value := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 511))
		got, err := MakeTopic(typ, value)
		if err != nil {
			t.Fatalf("MakeTopic: %v", err)
		}
		var want common.LogTopic
		want[0] = 0x80
		if got != want {
			t.Fatalf("int512 topic = %x, want %x", got, want)
		}
	})

	t.Run("narrow signed integer is sign extended to VM64", func(t *testing.T) {
		t.Parallel()
		typ := mustIndexedTopicType(t, "int16", nil)
		got, err := MakeTopic(typ, int16(-2))
		if err != nil {
			t.Fatalf("MakeTopic: %v", err)
		}
		var want common.LogTopic
		for i := range want {
			want[i] = 0xff
		}
		want[len(want)-1] = 0xfe
		if got != want {
			t.Fatalf("int16 topic = %x, want %x", got, want)
		}
	})

	t.Run("bool and full-width address use direct ABI words", func(t *testing.T) {
		t.Parallel()
		boolType := mustIndexedTopicType(t, "bool", nil)
		gotBool, err := MakeTopic(boolType, true)
		if err != nil {
			t.Fatalf("MakeTopic(bool): %v", err)
		}
		var wantBool common.LogTopic
		wantBool[common.LogTopicLength-1] = 1
		if gotBool != wantBool {
			t.Fatalf("bool topic = %x, want %x", gotBool, wantBool)
		}

		addressType := mustIndexedTopicType(t, "address", nil)
		var address common.Address
		for i := range address {
			address[i] = byte(0x80 + i)
		}
		gotAddress, err := MakeTopic(addressType, address)
		if err != nil {
			t.Fatalf("MakeTopic(address): %v", err)
		}
		if gotAddress != common.LogTopic(address) {
			t.Fatalf("address topic = %x, want %x", gotAddress, address)
		}
	})

	t.Run("bytes64 is direct and not confused with uint8 array", func(t *testing.T) {
		t.Parallel()
		var value [64]byte
		for i := range value {
			value[i] = byte(i)
		}
		bytesType := mustIndexedTopicType(t, "bytes64", nil)
		gotBytes, err := MakeTopic(bytesType, value)
		if err != nil {
			t.Fatalf("MakeTopic(bytes64): %v", err)
		}
		if gotBytes != common.LogTopic(value) {
			t.Fatalf("bytes64 topic = %x, want direct bytes %x", gotBytes, value)
		}

		arrayType := mustIndexedTopicType(t, "uint8[64]", nil)
		gotArray, err := MakeTopic(arrayType, value)
		if err != nil {
			t.Fatalf("MakeTopic(uint8[64]): %v", err)
		}
		wantArray := goldenIndexedTopic(t, "a1e2ed8e31ddabce15defc4d5cb5f1f5ce37d0740ac8404a931483f98f66891d")
		if gotArray != wantArray {
			t.Fatalf("uint8[64] topic = %x, want golden %x", gotArray, wantArray)
		}
	})

	t.Run("top-level string and bytes hash raw contents", func(t *testing.T) {
		t.Parallel()
		want := goldenIndexedTopic(t, "4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45")
		for _, test := range []struct {
			name  string
			typ   Type
			value any
		}{
			{"string", mustIndexedTopicType(t, "string", nil), "abc"},
			{"bytes", mustIndexedTopicType(t, "bytes", nil), []byte("abc")},
		} {
			test := test
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()
				got, err := MakeTopic(test.typ, test.value)
				if err != nil {
					t.Fatalf("MakeTopic: %v", err)
				}
				if got != want {
					t.Fatalf("topic = %x, want raw-content golden %x", got, want)
				}
			})
		}
	})
}

func TestMakeTopicHyperionCompositeGoldens(t *testing.T) {
	t.Parallel()

	t.Run("fixed array omits length and offsets", func(t *testing.T) {
		t.Parallel()
		typ := mustIndexedTopicType(t, "uint16[2]", nil)
		got, err := MakeTopic(typ, [2]uint16{1, 0x1234})
		if err != nil {
			t.Fatalf("MakeTopic: %v", err)
		}
		// Oracle preimage:
		//   VM64Word(1) || VM64Word(0x1234)
		// It contains neither an array length nor dynamic offsets.
		want := goldenIndexedTopic(t, "a43a66ce96d0c82ddae11ba9f8a7ad285dfca977fd1bd80032149e989dc23eb6")
		if got != want {
			t.Fatalf("fixed-array topic = %x, want golden %x", got, want)
		}
	})

	t.Run("dynamic string array pads members but omits length", func(t *testing.T) {
		t.Parallel()
		typ := mustIndexedTopicType(t, "string[]", nil)
		got, err := MakeTopic(typ, []string{"a", "bc"})
		if err != nil {
			t.Fatalf("MakeTopic: %v", err)
		}
		// Oracle preimage:
		//   rightPad64("a") || rightPad64("bc")
		// The strings have no individual length words or offsets.
		want := goldenIndexedTopic(t, "46b31a5d947cf77c53ac05ab04db93a60631ae4ddc1f2698ee43fca44dc176ea")
		if got != want {
			t.Fatalf("string-array topic = %x, want golden %x", got, want)
		}
	})

	type dynamicBoundaries struct {
		Text string
		Raw  []byte
	}
	t.Run("nested dynamic values pad at the VM64 boundary", func(t *testing.T) {
		t.Parallel()
		typ := mustIndexedTopicType(t, "tuple", []ArgumentMarshaling{
			{Name: "text", Type: "string"},
			{Name: "raw", Type: "bytes"},
		})
		raw := make([]byte, 65)
		for i := range raw {
			raw[i] = byte(i)
		}
		got, err := MakeTopic(typ, dynamicBoundaries{
			Text: strings.Repeat("s", common.LogTopicLength),
			Raw:  raw,
		})
		if err != nil {
			t.Fatalf("MakeTopic: %v", err)
		}
		// Oracle preimage is 192 bytes: the exact 64-byte string followed by
		// the 65-byte byte slice and 63 zero pad bytes.
		want := goldenIndexedTopic(t, "114dec3c551bc9142da99f5b021ebb7b592fc130a5382fc0b8b728b0b120b7e9")
		if got != want {
			t.Fatalf("dynamic-boundary topic = %x, want golden %x", got, want)
		}
	})

	type hyperionChild struct {
		Text string
		Raw  []byte
	}
	type hyperionRecord struct {
		Negative *big.Int
		Positive *big.Int
		Fixed    [64]byte
		Matrix   [2][]uint16
		Child    hyperionChild
		Empty    [0]uint8
	}
	t.Run("tuple concatenates nested dynamic members in place", func(t *testing.T) {
		t.Parallel()
		childComponents := []ArgumentMarshaling{
			{Name: "text", Type: "string"},
			{Name: "raw", Type: "bytes"},
		}
		components := []ArgumentMarshaling{
			{Name: "negative", Type: "int512"},
			{Name: "positive", Type: "uint512"},
			{Name: "fixed", Type: "bytes64"},
			{Name: "matrix", Type: "uint16[][2]"},
			{Name: "child", Type: "tuple", Components: childComponents},
			{Name: "empty", Type: "uint8[0]"},
		}
		typ := mustIndexedTopicType(t, "tuple", components)
		positive := new(big.Int).Lsh(big.NewInt(1), 511)
		positive.Add(positive, big.NewInt(5))
		var fixed [64]byte
		for i := range fixed {
			fixed[i] = byte(i)
		}
		value := hyperionRecord{
			Negative: big.NewInt(-2),
			Positive: positive,
			Fixed:    fixed,
			Matrix:   [2][]uint16{{1, 0x1234}, {0xffff}},
			Child:    hyperionChild{Text: "hyperion", Raw: []byte{0, 1, 2, 3, 4}},
			Empty:    [0]uint8{},
		}
		got, err := MakeTopic(typ, value)
		if err != nil {
			t.Fatalf("MakeTopic: %v", err)
		}
		// The independently generated 512-byte oracle is:
		//   word(-2) || word(2^511+5) || bytes64(0..63)
		//   || word(1) || word(0x1234) || word(0xffff)
		//   || rightPad64("hyperion") || rightPad64(0x0001020304)
		// Empty uint8[0] contributes no bytes.
		want := goldenIndexedTopic(t, "1b5938d7eb8e2f74f36120a4bd7d40df752ff21983a1c5c999d9da60afbeaf5a")
		if got != want {
			t.Fatalf("tuple topic = %x, want golden %x", got, want)
		}
	})

	type hyperionTupleElement struct {
		Number  int16
		Payload []byte
	}
	t.Run("fixed and dynamic arrays of tuples recurse without boundaries", func(t *testing.T) {
		t.Parallel()
		components := []ArgumentMarshaling{
			{Name: "number", Type: "int16"},
			{Name: "payload", Type: "bytes"},
		}
		typ := mustIndexedTopicType(t, "tuple[][2]", components)
		value := [2][]hyperionTupleElement{
			{{Number: -1, Payload: []byte{0xaa}}},
			{
				{Number: 2, Payload: nil},
				{Number: -3, Payload: []byte{0xbb, 0xcc}},
			},
		}
		got, err := MakeTopic(typ, value)
		if err != nil {
			t.Fatalf("MakeTopic: %v", err)
		}
		// Oracle preimage:
		//   word(-1) || rightPad64(0xaa)
		//   || word(2) || "" || word(-3) || rightPad64(0xbbcc)
		// The empty nested bytes value contributes zero bytes.
		want := goldenIndexedTopic(t, "67555d053d2bc715fed9ac579abdec9c35f9d9120807e4fc31cd1873b40da454")
		if got != want {
			t.Fatalf("tuple-array topic = %x, want golden %x", got, want)
		}
	})
}

func TestMakeTopicHyperionCompositeBoundaryAmbiguity(t *testing.T) {
	t.Parallel()

	type ambiguousArrays struct {
		First  []uint16
		Second []uint16
	}
	typ := mustIndexedTopicType(t, "tuple", []ArgumentMarshaling{
		{Name: "first", Type: "uint16[]"},
		{Name: "second", Type: "uint16[]"},
	})
	left := ambiguousArrays{First: []uint16{1}, Second: []uint16{2, 3}}
	right := ambiguousArrays{First: []uint16{1, 2}, Second: []uint16{3}}

	leftTopic, err := MakeTopic(typ, left)
	if err != nil {
		t.Fatalf("MakeTopic(left): %v", err)
	}
	rightTopic, err := MakeTopic(typ, right)
	if err != nil {
		t.Fatalf("MakeTopic(right): %v", err)
	}
	if leftTopic != rightTopic {
		t.Fatalf("indexed composite boundary ambiguity was lost: left=%x right=%x", leftTopic, rightTopic)
	}

	// Hyperion's indexed-event encoding deliberately omits dynamic-array
	// lengths and member boundaries. Both values therefore hash the independent
	// preimage VM64Word(1) || VM64Word(2) || VM64Word(3).
	preimage := make([]byte, 3*common.LogTopicLength)
	preimage[common.LogTopicLength-1] = 1
	preimage[2*common.LogTopicLength-1] = 2
	preimage[3*common.LogTopicLength-1] = 3
	digest := crypto.Keccak256Hash(preimage)
	var want common.LogTopic
	copy(want[:common.HashLength], digest[:])
	if leftTopic != want {
		t.Fatalf("ambiguous indexed topic = %x, want independent preimage digest %x", leftTopic, want)
	}
}

func TestMakeTopicEmptyCompositeGoldens(t *testing.T) {
	t.Parallel()
	want := goldenIndexedTopic(t, "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")

	emptyTupleType := mustIndexedTopicType(t, "tuple", nil)
	emptyBytesTupleType := mustIndexedTopicType(t, "tuple", []ArgumentMarshaling{{Name: "value", Type: "bytes"}})
	type emptyBytesTuple struct {
		Value []byte
	}
	emptyTupleArrayType := mustIndexedTopicType(t, "tuple[2]", nil)
	tests := []struct {
		name  string
		typ   Type
		value any
	}{
		{"fixed zero array", mustIndexedTopicType(t, "uint8[0]", nil), [0]uint8{}},
		{"nested fixed zero arrays", mustIndexedTopicType(t, "uint8[0][3]", nil), [3][0]uint8{}},
		{"nil dynamic array", mustIndexedTopicType(t, "uint8[]", nil), []uint8(nil)},
		{"empty dynamic array", mustIndexedTopicType(t, "uint8[]", nil), []uint8{}},
		{"empty tuple", emptyTupleType, struct{}{}},
		{"tuple containing empty bytes", emptyBytesTupleType, emptyBytesTuple{}},
		{"array of empty tuples", emptyTupleArrayType, [2]struct{}{}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := MakeTopic(test.typ, test.value)
			if err != nil {
				t.Fatalf("MakeTopic: %v", err)
			}
			if got != want {
				t.Fatalf("empty topic = %x, want Keccak(empty) %x", got, want)
			}
		})
	}
}

func TestMakeTopicHashBridgesHashedValuesToFilterRules(t *testing.T) {
	t.Parallel()

	type hashRecord struct {
		Count uint16
	}
	tests := []struct {
		name  string
		typ   Type
		value any
	}{
		{"string", mustIndexedTopicType(t, "string", nil), "hyperion"},
		{"bytes", mustIndexedTopicType(t, "bytes", nil), []byte{0, 1, 2}},
		{"slice", mustIndexedTopicType(t, "uint16[]", nil), []uint16{1, 0x1234}},
		{"array", mustIndexedTopicType(t, "uint16[2]", nil), [2]uint16{1, 0x1234}},
		{
			"tuple",
			mustIndexedTopicType(t, "tuple", []ArgumentMarshaling{{Name: "count", Type: "uint16"}}),
			hashRecord{Count: 0x1234},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			topic, err := MakeTopic(test.typ, test.value)
			if err != nil {
				t.Fatalf("MakeTopic: %v", err)
			}
			digest, err := MakeTopicHash(test.typ, test.value)
			if err != nil {
				t.Fatalf("MakeTopicHash: %v", err)
			}
			if got := common.HashToLogTopic(digest); got != topic {
				t.Fatalf("HashToLogTopic(MakeTopicHash) = %x, want MakeTopic %x", got, topic)
			}
			var want common.Hash
			copy(want[:], topic[:common.HashLength])
			if digest != want {
				t.Fatalf("MakeTopicHash = %x, want high-half digest %x", digest, want)
			}
			if wrong := common.BytesToHash(topic[:]); wrong == digest {
				t.Fatalf("BytesToHash unexpectedly recovered left-aligned digest %x", digest)
			}
		})
	}
}

func TestMakeTopicHashRejectsDirectAndInvalidTopics(t *testing.T) {
	t.Parallel()

	var fixed [64]byte
	var address common.Address
	for _, test := range []struct {
		name  string
		typ   Type
		value any
	}{
		{"uint512", mustIndexedTopicType(t, "uint512", nil), big.NewInt(1)},
		{"bool", mustIndexedTopicType(t, "bool", nil), true},
		{"address", mustIndexedTopicType(t, "address", nil), address},
		{"bytes64", mustIndexedTopicType(t, "bytes64", nil), fixed},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := MakeTopicHash(test.typ, test.value); err == nil || !strings.Contains(err.Error(), "encoded directly") {
				t.Fatalf("MakeTopicHash error = %v, want direct-topic rejection", err)
			}
		})
	}

	stringType := mustIndexedTopicType(t, "string", nil)
	if _, err := MakeTopicHash(stringType, []byte("wrong type")); err == nil || !strings.Contains(err.Error(), "cannot use") {
		t.Fatalf("MakeTopicHash invalid value error = %v", err)
	}
}

func TestMakeTopicRejectsMalformedAndWrongValues(t *testing.T) {
	t.Parallel()

	uint512Type := mustIndexedTopicType(t, "uint512", nil)
	int504Type := mustIndexedTopicType(t, "int504", nil)
	bytes64Type := mustIndexedTopicType(t, "bytes64", nil)
	arrayType := mustIndexedTopicType(t, "uint16[2]", nil)
	tupleType := mustIndexedTopicType(t, "tuple", []ArgumentMarshaling{{Name: "count", Type: "uint16"}})
	functionType := mustIndexedTopicType(t, "function", nil)

	tooLargeUint := new(big.Int).Lsh(big.NewInt(1), 512)
	tooLargeInt := new(big.Int).Lsh(big.NewInt(1), 503)
	var nilBigInt *big.Int
	var nilArray *[2]uint16
	var nilTuple *struct{ Count uint16 }
	tests := []struct {
		name  string
		typ   Type
		value any
		want  string
	}{
		{"nil interface", uint512Type, nil, "cannot use <nil>"},
		{"nil big integer", uint512Type, nilBigInt, "cannot use <nil>"},
		{"negative unsigned integer", uint512Type, big.NewInt(-1), "negatively-signed"},
		{"unsigned overflow", uint512Type, tooLargeUint, "uint512"},
		{"signed overflow", int504Type, tooLargeInt, "int504"},
		{"wrong fixed bytes width", bytes64Type, [63]byte{}, "cannot use"},
		{"wrong fixed array value", arrayType, []uint16{1}, "cannot use"},
		{"nil fixed array pointer", arrayType, nilArray, "cannot use <nil>"},
		{"tuple missing field", tupleType, struct{ Other uint16 }{}, "field Count"},
		{"nil tuple pointer", tupleType, nilTuple, "cannot use <nil>"},
		{"function wider than topic", functionType, [common.AddressLength + 4]byte{}, "do not fit"},
		{"array type without element", Type{T: SliceTy}, []uint8{}, "no element type"},
		{"tuple metadata mismatch", Type{T: TupleTy, TupleType: reflect.TypeFor[struct{}](), TupleElems: []*Type{{T: BoolTy}}}, struct{}{}, "inconsistent"},
		{"unknown type", Type{T: 0xff}, uint8(0), "unknown"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("MakeTopic panicked: %v", recovered)
				}
			}()
			if _, err := MakeTopic(test.typ, test.value); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("MakeTopic error = %v, want substring %q", err, test.want)
			}
		})
	}

	t.Run("cyclic malformed type is rejected without recursion overflow", func(t *testing.T) {
		t.Parallel()
		cyclic := Type{T: SliceTy}
		cyclic.Elem = &cyclic
		if _, err := MakeTopic(cyclic, nil); err == nil || !strings.Contains(err.Error(), "nesting exceeds") {
			t.Fatalf("MakeTopic cyclic type error = %v", err)
		}
	})

	t.Run("non-nil pointers are accepted", func(t *testing.T) {
		t.Parallel()
		value := [2]uint16{1, 0x1234}
		got, err := MakeTopic(arrayType, &value)
		if err != nil {
			t.Fatalf("MakeTopic(pointer): %v", err)
		}
		want := goldenIndexedTopic(t, "a43a66ce96d0c82ddae11ba9f8a7ad285dfca977fd1bd80032149e989dc23eb6")
		if got != want {
			t.Fatalf("pointer topic = %x, want %x", got, want)
		}
	})
}

func TestMakeTopicPreservesMakeTopicsRawRules(t *testing.T) {
	t.Parallel()
	hash := common.Hash{0x11, 0x22, 0x33}
	raw := common.LogTopic{0xaa, 0xbb, 0xcc}
	var fixed [64]byte
	fixed[0], fixed[63] = 0x44, 0x55

	got, err := MakeTopics([]any{hash}, []any{raw}, []any{fixed})
	if err != nil {
		t.Fatalf("MakeTopics: %v", err)
	}
	want := [][]common.LogTopic{
		{common.HashToLogTopic(hash)},
		{raw},
		{common.LogTopic(fixed)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MakeTopics raw rules = %x, want %x", got, want)
	}
}
