// Copyright 2020 The go-ethereum Authors
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
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"

	fuzz "github.com/google/gofuzz"
)

// TestReplicate can be used to replicate crashers from the fuzzing tests.
// Just replace testString with the data in .quoted
func TestReplicate(t *testing.T) {
	t.Parallel()
	//t.Skip("Test only useful for reproducing issues")
	fuzzAbi([]byte("\x20\x20\x20\x20\x20\x20\x20\x20\x80\x00\x00\x00\x20\x20\x20\x20\x00"))
	//fuzzAbi([]byte("asdfasdfkadsf;lasdf;lasd;lfk"))
}

func TestCreateABIMultipleArguments(t *testing.T) {
	parsed, err := createABI("multi", nil, nil, []arg{
		{name: "count", typ: "uint512"},
		{name: "recipient", typ: "address"},
		{name: "note", typ: "string"},
	})
	if err != nil {
		t.Fatalf("create multi-argument ABI: %v", err)
	}
	method := parsed.Methods["multi"]
	if len(method.Inputs) != 3 || len(method.Outputs) != 3 {
		t.Fatalf("generated fuzz ABI inputs/outputs = %d/%d, want 3/3", len(method.Inputs), len(method.Outputs))
	}
	for i, want := range []string{"uint512", "address", "string"} {
		if method.Inputs[i].Type.String() != want || method.Outputs[i].Type.String() != want {
			t.Fatalf("generated fuzz ABI argument %d = %s/%s, want %s", i, method.Inputs[i].Type, method.Outputs[i].Type, want)
		}
	}
}

// FuzzABI is the main entrypoint for fuzzing
func FuzzABI(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("\x20\x20\x20\x20\x20\x20\x20\x20\x80\x00\x00\x00\x20\x20\x20\x20\x00"))
	f.Add(make([]byte, 64))
	f.Add(bytes.Repeat([]byte{0xff}, 128))
	f.Add(bytes.Repeat([]byte{0x00, 0x01, 0x3f, 0x40, 0x41, 0x7f, 0x80, 0xff}, 16))
	f.Fuzz(func(t *testing.T, data []byte) {
		fuzzAbi(data)
	})
}

// FuzzABIStructuredRecursive models Hyperion's recursive ABI-v2 fuzz input:
// bounded arrays and structs are generated as real ABI types, populated with
// matching Go values, round-tripped, and then subjected to hostile word-level
// mutations. The bounds keep each iteration useful for continuous fuzzing.
func FuzzABIStructuredRecursive(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7})
	f.Add(bytes.Repeat([]byte{0xff, 0x00, 0x40, 0x80}, 16))
	f.Fuzz(func(t *testing.T, input []byte) {
		input = input[:min(len(input), 4096)]
		fuzzer := fuzz.NewFromGoFuzz(input)
		argumentCount := 1 + upTo(fuzzer, 4)
		arguments := make(Arguments, argumentCount)
		values := make([]any, argumentCount)
		for i := range arguments {
			marshaling := generatedStructuredABIType(fuzzer, 0)
			typ, err := NewType(marshaling.Type, marshaling.InternalType, marshaling.Components)
			if err != nil {
				t.Fatalf("build generated recursive ABI type %+v: %v", marshaling, err)
			}
			arguments[i] = Argument{Name: fmt.Sprintf("arg%d", i), Type: typ}
			values[i] = generatedValue(fuzzer, typ)
		}

		packed, err := arguments.Pack(values...)
		if err != nil {
			t.Fatalf("pack generated recursive values: %v", err)
		}
		decoded, err := arguments.Unpack(packed)
		if err != nil {
			t.Fatalf("unpack generated recursive values: %v", err)
		}
		if !reflect.DeepEqual(decoded, values) {
			t.Fatalf("recursive ABI round trip differs\nhave: %#v\nwant: %#v", decoded, values)
		}
		repacked, err := arguments.Pack(decoded...)
		if err != nil {
			t.Fatalf("repack generated recursive values: %v", err)
		}
		if !bytes.Equal(repacked, packed) {
			t.Fatalf("recursive ABI repack is unstable\nhave: %x\nwant: %x", repacked, packed)
		}

		// Every mutation is allowed to decode or fail, but none may panic or
		// allocate from an attacker-controlled high word.
		mutations := [][]byte{
			append([]byte{}, packed[:upTo(fuzzer, len(packed)+1)]...),
			append(append([]byte{}, packed...), input...),
		}
		if len(packed) >= 64 {
			highWord := append([]byte{}, packed...)
			word := upTo(fuzzer, len(highWord)/64)
			highWord[word*64] |= 0x80
			mutations = append(mutations, highWord)

			dirtyLength := append([]byte{}, packed...)
			word = upTo(fuzzer, len(dirtyLength)/64)
			for i := 0; i < 63; i++ {
				dirtyLength[word*64+i] = 0xff
			}
			mutations = append(mutations, dirtyLength)
		}
		for _, mutated := range mutations {
			_, _ = arguments.Unpack(mutated)
		}
	})
}

func generatedStructuredABIType(fuzzer *fuzz.Fuzzer, depth int) ArgumentMarshaling {
	leafTypes := []string{
		"bool", "address", "bytes", "string",
		"uint8", "uint512", "int8", "int512",
		"bytes1", "bytes32", "bytes64",
	}
	if depth >= 4 {
		return ArgumentMarshaling{Type: oneOf(fuzzer, leafTypes)}
	}
	switch upTo(fuzzer, 5) {
	case 0:
		return ArgumentMarshaling{Type: oneOf(fuzzer, leafTypes)}
	case 1, 2:
		element := generatedStructuredABIType(fuzzer, depth+1)
		suffixes := []string{"[]", "[0]", "[1]", "[2]"}
		suffix := oneOf(fuzzer, suffixes)
		element.Type += suffix
		if element.InternalType != "" {
			element.InternalType += suffix
		}
		return element
	default:
		count := 1 + upTo(fuzzer, 3)
		components := make([]ArgumentMarshaling, count)
		for i := range components {
			components[i] = generatedStructuredABIType(fuzzer, depth+1)
			components[i].Name = fmt.Sprintf("field%d", i)
		}
		suffixes := []string{"", "[]", "[0]", "[1]", "[2]"}
		suffix := oneOf(fuzzer, suffixes)
		return ArgumentMarshaling{
			Type:         "tuple" + suffix,
			InternalType: fmt.Sprintf("struct Fuzz.Record%d%s", depth, suffix),
			Components:   components,
		}
	}
}

var (
	names    = []string{"_name", "name", "NAME", "name_", "__", "_name_", "n"}
	stateMut = []string{"pure", "view", "payable"}
	pays     = []string{"true", "false"}
	vNames   = []string{"a", "b", "c", "d", "e", "f", "g"}
	varNames = append(vNames, names...)
	varTypes = []string{"bool", "address", "bytes", "string",
		"uint8", "int8", "uint16", "int16",
		"uint24", "int24", "uint32", "int32", "uint40", "int40", "uint48", "int48", "uint56", "int56",
		"uint64", "int64", "uint72", "int72", "uint80", "int80", "uint88", "int88", "uint96", "int96",
		"uint104", "int104", "uint112", "int112", "uint120", "int120", "uint128", "int128", "uint136", "int136",
		"uint144", "int144", "uint152", "int152", "uint160", "int160", "uint168", "int168", "uint176", "int176",
		"uint184", "int184", "uint192", "int192", "uint200", "int200", "uint208", "int208", "uint216", "int216",
		"uint224", "int224", "uint232", "int232", "uint240", "int240", "uint248", "int248", "uint256", "int256",
		"uint264", "int264", "uint272", "int272", "uint280", "int280", "uint288", "int288", "uint296", "int296",
		"uint304", "int304", "uint312", "int312", "uint320", "int320", "uint328", "int328", "uint336", "int336",
		"uint344", "int344", "uint352", "int352", "uint360", "int360", "uint368", "int368", "uint376", "int376",
		"uint384", "int384", "uint392", "int392", "uint400", "int400", "uint408", "int408", "uint416", "int416",
		"uint424", "int424", "uint432", "int432", "uint440", "int440", "uint448", "int448", "uint456", "int456",
		"uint464", "int464", "uint472", "int472", "uint480", "int480", "uint488", "int488", "uint496", "int496",
		"uint504", "int504", "uint512", "int512",
		"bytes1", "bytes2", "bytes3", "bytes4", "bytes5", "bytes6", "bytes7", "bytes8", "bytes9", "bytes10", "bytes11",
		"bytes12", "bytes13", "bytes14", "bytes15", "bytes16", "bytes17", "bytes18", "bytes19", "bytes20", "bytes21",
		"bytes22", "bytes23", "bytes24", "bytes25", "bytes26", "bytes27", "bytes28", "bytes29", "bytes30", "bytes31",
		"bytes32", "bytes33", "bytes34", "bytes35", "bytes36", "bytes37", "bytes38", "bytes39", "bytes40", "bytes41",
		"bytes42", "bytes43", "bytes44", "bytes45", "bytes46", "bytes47", "bytes48", "bytes49", "bytes50", "bytes51",
		"bytes52", "bytes53", "bytes54", "bytes55", "bytes56", "bytes57", "bytes58", "bytes59", "bytes60", "bytes61",
		"bytes62", "bytes63", "bytes64", "bytes"}
)

func unpackPack(abi ABI, method string, input []byte) ([]any, bool) {
	if out, err := abi.Unpack(method, input); err == nil {
		_, err := abi.Pack(method, out...)
		if err != nil {
			// We have some false positives as we can unpack these type successfully, but not pack them
			if err.Error() == "abi: cannot use []uint8 as type [0]int8 as argument" ||
				err.Error() == "abi: cannot use uint8 as type int8 as argument" {
				return out, false
			}
			panic(err)
		}
		return out, true
	}
	return nil, false
}

func packUnpack(abi ABI, method string, input []any) bool {
	packed, err := abi.Pack(method, input...)
	if err != nil {
		return false
	}
	if len(packed) < 4 {
		panic(fmt.Sprintf("packed method call is missing its selector: %x", packed))
	}
	out, err := abi.Unpack(method, packed[4:])
	if err != nil {
		panic(err)
	}
	if len(input) != 0 || len(out) != 0 {
		if !reflect.DeepEqual(input, out) {
			panic(fmt.Sprintf("pack/unpack is not equal, \ninput : %#v\noutput: %#v", input, out))
		}
	}
	repacked, err := abi.Pack(method, out...)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(packed, repacked) {
		panic(fmt.Sprintf("pack/unpack/repack is not stable, \npacked  : %x\nrepacked: %x", packed, repacked))
	}
	return true
}

// generatedValue returns a small, valid Go value for typ. Keeping collections
// bounded ensures that every fuzz iteration reaches ABI packing instead of
// spending its budget allocating values.
func generatedValue(fuzzer *fuzz.Fuzzer, typ Type) any {
	switch typ.T {
	case IntTy, UintTy:
		goType := typ.GetType()
		if goType.Kind() != reflect.Ptr {
			value := reflect.New(goType).Elem()
			if typ.T == UintTy {
				var number uint64
				fuzzer.Fuzz(&number)
				value.SetUint(number)
			} else {
				var number int64
				fuzzer.Fuzz(&number)
				value.SetInt(number)
			}
			return value.Interface()
		}
		bits := typ.Size
		if typ.T == IntTy {
			bits--
		}
		limit := new(big.Int).Lsh(big.NewInt(1), uint(bits))
		raw := make([]byte, (bits+7)/8)
		for i := range raw {
			fuzzer.Fuzz(&raw[i])
		}
		number := new(big.Int).SetBytes(raw)
		number.Mod(number, limit)
		if typ.T == IntTy && number.Sign() != 0 && upTo(fuzzer, 2) == 1 {
			number.Neg(number)
		}
		return number
	case BoolTy:
		return upTo(fuzzer, 2) == 1
	case StringTy:
		value := make([]byte, upTo(fuzzer, 5))
		for i := range value {
			fuzzer.Fuzz(&value[i])
		}
		return string(value)
	case SliceTy:
		length := upTo(fuzzer, 4)
		value := reflect.MakeSlice(typ.GetType(), length, length)
		for i := 0; i < value.Len(); i++ {
			value.Index(i).Set(reflect.ValueOf(generatedValue(fuzzer, *typ.Elem)))
		}
		return value.Interface()
	case ArrayTy:
		value := reflect.New(typ.GetType()).Elem()
		for i := 0; i < value.Len(); i++ {
			value.Index(i).Set(reflect.ValueOf(generatedValue(fuzzer, *typ.Elem)))
		}
		return value.Interface()
	case TupleTy:
		value := reflect.New(typ.TupleType).Elem()
		for i, elem := range typ.TupleElems {
			value.Field(i).Set(reflect.ValueOf(generatedValue(fuzzer, *elem)))
		}
		return value.Interface()
	case AddressTy, FixedBytesTy, HashTy, FixedPointTy, FunctionTy:
		value := reflect.New(typ.GetType()).Elem()
		for i := 0; i < value.Len(); i++ {
			var b byte
			fuzzer.Fuzz(&b)
			value.Index(i).SetUint(uint64(b))
		}
		return value.Interface()
	case BytesTy:
		value := make([]byte, upTo(fuzzer, 5))
		for i := range value {
			fuzzer.Fuzz(&value[i])
		}
		return value
	default:
		panic(fmt.Sprintf("unsupported fuzz ABI type %v", typ.T))
	}
}

type arg struct {
	name string
	typ  string
}

func createABI(name string, stateMutability, payable *string, inputs []arg) (ABI, error) {
	var sig strings.Builder
	sig.WriteString(fmt.Sprintf(`[{ "type" : "function", "name" : "%v" `, name))
	if stateMutability != nil {
		sig.WriteString(fmt.Sprintf(`, "stateMutability": "%v" `, *stateMutability))
	}
	if payable != nil {
		sig.WriteString(fmt.Sprintf(`, "payable": %v `, *payable))
	}
	for _, field := range []string{"inputs", "outputs"} {
		fmt.Fprintf(&sig, `, %q: [`, field)
		for i, inp := range inputs {
			if i > 0 {
				sig.WriteByte(',')
			}
			fmt.Fprintf(&sig, `{"name":%q,"type":%q}`, inp.name, inp.typ)
		}
		sig.WriteByte(']')
	}
	sig.WriteString(`}]`)
	//fmt.Printf("sig: %s\n", sig)
	return JSON(strings.NewReader(sig.String()))
}

func fuzzAbi(input []byte) {
	var (
		fuzzer    = fuzz.NewFromGoFuzz(input)
		name      = oneOf(fuzzer, names)
		stateM    = oneOfOrNil(fuzzer, stateMut)
		payable   = oneOfOrNil(fuzzer, pays)
		arguments []arg
	)
	for i := 0; i < upTo(fuzzer, 10); i++ {
		argName := oneOf(fuzzer, varNames)
		argTyp := oneOf(fuzzer, varTypes)
		switch upTo(fuzzer, 10) {
		case 0: // 10% chance to make it a slice
			argTyp += "[]"
		case 1: // 10% chance to make it an array
			argTyp += fmt.Sprintf("[%d]", 1+upTo(fuzzer, 30))
		case 2: // Exercise zero-sized static values.
			argTyp += "[0]"
		case 3: // An outer array containing dynamic slices.
			argTyp += "[][2]"
		case 4: // A dynamic slice containing static arrays.
			argTyp += "[2][]"
		default:
		}
		arguments = append(arguments, arg{name: argName, typ: argTyp})
	}
	abi, err := createABI(name, stateM, payable, arguments)
	if err != nil {
		//fmt.Printf("err: %v\n", err)
		panic(err)
	}
	if decoded, ok := unpackPack(abi, name, input); ok {
		if !packUnpack(abi, name, decoded) {
			panic("values accepted by ABI unpacking could not be packed")
		}
	}
	method := abi.Methods[name]
	values := make([]any, len(method.Inputs))
	for i, argument := range method.Inputs {
		values[i] = generatedValue(fuzzer, argument.Type)
	}
	if !packUnpack(abi, name, values) {
		panic("generated valid ABI values could not be packed")
	}
}

func upTo(fuzzer *fuzz.Fuzzer, max int) int {
	var i int
	fuzzer.Fuzz(&i)
	if i < 0 {
		return (-1 - i) % max
	}
	return i % max
}

func oneOf(fuzzer *fuzz.Fuzzer, options []string) string {
	return options[upTo(fuzzer, len(options))]
}

func oneOfOrNil(fuzzer *fuzz.Fuzzer, options []string) *string {
	if i := upTo(fuzzer, len(options)+1); i < len(options) {
		return &options[i]
	}
	return nil
}
