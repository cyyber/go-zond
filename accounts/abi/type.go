// Copyright 2015 The go-ethereum Authors
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
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/theQRL/go-qrl/common"
)

// Type enumerator
const (
	IntTy byte = iota
	UintTy
	BoolTy
	StringTy
	SliceTy
	ArrayTy
	TupleTy
	AddressTy
	FixedBytesTy
	BytesTy
	HashTy
	FixedPointTy
	FunctionTy
)

// Type is the reflection of the supported argument type.
type Type struct {
	Elem *Type
	Size int
	T    byte // Our own type checking

	stringKind string // holds the unparsed string for deriving signatures

	// Tuple relative fields
	TupleRawName  string       // Raw struct name defined in source code, may be empty.
	TupleElems    []*Type      // Type information of all tuple fields
	TupleRawNames []string     // Raw field name of all tuple fields
	TupleType     reflect.Type // Underlying struct of the tuple
}

var (
	// typeRegex parses a complete ABI base type. Anchoring is important: a
	// prefix such as "uint8" must not make "uint8garbage" a valid type.
	typeRegex = regexp.MustCompile("^([a-zA-Z]+)(([0-9]+)(x([0-9]+))?)?$")

	// arraySuffixRegex matches one complete array/slice suffix.
	arraySuffixRegex = regexp.MustCompile(`^\[([0-9]*)\]$`)
)

const (
	maxABITypeNesting      = 64
	maxABIFixedArrayLength = 1 << 20
	maxABIStaticTypeSize   = 64 << 20
	maxABIReflectTypeSize  = 64 << 20
	maxABITupleFields      = 1 << 12
)

// NewType creates a new reflection type of abi type given in t.
func NewType(t string, internalType string, components []ArgumentMarshaling) (typ Type, err error) {
	return newType(t, internalType, components, 0)
}

func newType(t string, internalType string, components []ArgumentMarshaling, depth int) (typ Type, err error) {
	if depth > maxABITypeNesting {
		return Type{}, fmt.Errorf("abi: type nesting exceeds safety limit %d", maxABITypeNesting)
	}
	// check that array brackets are equal if they exist
	if strings.Count(t, "[") != strings.Count(t, "]") {
		return Type{}, errors.New("invalid arg type in abi")
	}
	typ.stringKind = t

	// if there are brackets, get ready to go into slice/array mode and
	// recursively create the type
	if strings.Count(t, "[") != 0 {
		// Note internalType can be empty here.
		subInternal := internalType
		if i := strings.LastIndex(internalType, "["); i != -1 {
			subInternal = subInternal[:i]
		}
		// recursively embed the type
		i := strings.LastIndex(t, "[")
		embeddedType, err := newType(t[:i], subInternal, components, depth+1)
		if err != nil {
			return Type{}, err
		}
		// grab the last cell and create a type from there
		sliced := t[i:]
		match := arraySuffixRegex.FindStringSubmatch(sliced)
		if match == nil {
			return Type{}, errors.New("invalid formatting of array type")
		}
		if len(match[1]) > 1 && match[1][0] == '0' {
			return Type{}, fmt.Errorf("invalid formatting of array type: non-canonical length %q", match[1])
		}

		if match[1] == "" {
			// is a slice
			typ.T = SliceTy
			typ.Elem = &embeddedType
			typ.stringKind = embeddedType.stringKind + sliced
		} else {
			// is an array
			typ.T = ArrayTy
			typ.Elem = &embeddedType
			typ.Size, err = strconv.Atoi(match[1])
			if err != nil {
				return Type{}, fmt.Errorf("abi: error parsing variable size: %v", err)
			}
			typ.stringKind = embeddedType.stringKind + sliced
		}
		if err := validateArrayType(typ, embeddedType); err != nil {
			return Type{}, err
		}
		return typ, err
	}
	// parse the type and size of the abi-type.
	parsedType := typeRegex.FindStringSubmatch(t)
	if parsedType == nil {
		return Type{}, fmt.Errorf("invalid type '%v'", t)
	}

	// varSize is the size of the variable
	var varSize int
	if len(parsedType[3]) > 0 {
		if len(parsedType[3]) > 1 && parsedType[3][0] == '0' {
			return Type{}, fmt.Errorf("unsupported arg type: %s (non-canonical numeric width)", t)
		}
		var err error
		varSize, err = strconv.Atoi(parsedType[2])
		if err != nil {
			return Type{}, fmt.Errorf("abi: error parsing variable size: %v", err)
		}
	} else {
		if parsedType[0] == "uint" || parsedType[0] == "int" {
			// this should fail because it means that there's something wrong with
			// the abi type (the compiler should always format it to the size...always)
			return Type{}, fmt.Errorf("unsupported arg type: %s", t)
		}
	}
	// varType is the parsed abi type
	switch varType := parsedType[1]; varType {
	case "int":
		if varSize < 8 || varSize > 512 || varSize%8 != 0 {
			return Type{}, fmt.Errorf("unsupported arg type: %s", t)
		}
		typ.Size = varSize
		typ.T = IntTy
	case "uint":
		if varSize < 8 || varSize > 512 || varSize%8 != 0 {
			return Type{}, fmt.Errorf("unsupported arg type: %s", t)
		}
		typ.Size = varSize
		typ.T = UintTy
	case "bool":
		if len(parsedType[3]) != 0 {
			return Type{}, fmt.Errorf("unsupported arg type: %s", t)
		}
		typ.T = BoolTy
	case "address":
		if len(parsedType[3]) != 0 {
			return Type{}, fmt.Errorf("unsupported arg type: %s", t)
		}
		typ.Size = common.AddressLength
		typ.T = AddressTy
	case "string":
		if len(parsedType[3]) != 0 {
			return Type{}, fmt.Errorf("unsupported arg type: %s", t)
		}
		typ.T = StringTy
	case "bytes":
		if len(parsedType[3]) == 0 {
			typ.T = BytesTy
		} else {
			// VM64 fixed bytes must fit in one 64-byte ABI word.
			if varSize == 0 || varSize > 64 {
				return Type{}, fmt.Errorf("unsupported arg type: %s", t)
			}
			typ.T = FixedBytesTy
			typ.Size = varSize
		}
	case "tuple":
		if len(parsedType[3]) != 0 {
			return Type{}, fmt.Errorf("unsupported arg type: %s", t)
		}
		if len(components) > maxABITupleFields {
			return Type{}, fmt.Errorf("abi: tuple field count %d exceeds safety limit %d", len(components), maxABITupleFields)
		}
		var (
			fields      []reflect.StructField
			elems       []*Type
			names       []string
			expression  strings.Builder // canonical parameter expression
			used        = make(map[string]bool)
			reflectSize uintptr
		)
		expression.WriteString("(")
		for idx, c := range components {
			cType, err := newType(c.Type, c.InternalType, c.Components, depth+1)
			if err != nil {
				return Type{}, err
			}
			name := ToCamelCase(c.Name)
			if name == "" {
				return Type{}, errors.New("abi: purely anonymous or underscored field is not supported")
			}
			fieldName := ResolveNameConflict(name, func(s string) bool { return used[s] })
			used[fieldName] = true
			if !isValidFieldName(fieldName) {
				return Type{}, fmt.Errorf("field %d has invalid name", idx)
			}
			fieldType := cType.GetType()
			fieldSize := fieldType.Size()
			if fieldSize > uintptr(maxABIReflectTypeSize) || reflectSize > uintptr(maxABIReflectTypeSize)-fieldSize {
				return Type{}, fmt.Errorf("abi: tuple reflected type exceeds safety limit %d", maxABIReflectTypeSize)
			}
			reflectSize += fieldSize
			fields = append(fields, reflect.StructField{
				Name: fieldName, // reflect.StructOf will panic for any exported field.
				Type: fieldType,
				Tag:  reflect.StructTag("json:\"" + c.Name + "\""),
			})
			elems = append(elems, &cType)
			names = append(names, c.Name)
			expression.WriteString(cType.stringKind)
			if idx != len(components)-1 {
				expression.WriteString(",")
			}
		}
		expression.WriteString(")")

		typ.TupleElems = elems
		typ.TupleRawNames = names
		typ.T = TupleTy
		typ.stringKind = expression.String()
		if size, err := getTypeSizeChecked(typ); err != nil {
			return Type{}, err
		} else if !isDynamicType(typ) && size > maxABIStaticTypeSize {
			return Type{}, fmt.Errorf("abi: tuple type exceeds static-size safety limit %d", maxABIStaticTypeSize)
		}
		typ.TupleType, err = makeTupleType(fields)
		if err != nil {
			return Type{}, err
		}
		if typ.TupleType.Size() > uintptr(maxABIReflectTypeSize) {
			return Type{}, fmt.Errorf("abi: tuple reflected type exceeds safety limit %d", maxABIReflectTypeSize)
		}

		const structPrefix = "struct "
		// We can obtain the struct name user defined in
		// the source code from "internalType"
		if internalType != "" && strings.HasPrefix(internalType, structPrefix) {
			// Foo.Bar type definition is not allowed in golang,
			// convert the format to FooBar
			typ.TupleRawName = strings.ReplaceAll(internalType[len(structPrefix):], ".", "")
		}

	case "function":
		if len(parsedType[3]) != 0 {
			return Type{}, fmt.Errorf("unsupported arg type: %s", t)
		}
		typ.T = FunctionTy
		typ.Size = common.AddressLength + 4
	default:
		if strings.HasPrefix(internalType, "contract ") {
			typ.Size = common.AddressLength
			typ.T = AddressTy
		} else {
			return Type{}, fmt.Errorf("unsupported arg type: %s", t)
		}
	}

	return
}

// validateArrayType rejects array shapes which cannot be represented by Go's
// reflection package or whose ABI head table would overflow an int. Without
// this validation, reflect.ArrayOf and later size arithmetic can panic for
// hostile, but syntactically valid, ABI JSON.
func validateArrayType(array, elem Type) error {
	if array.T != ArrayTy {
		return nil
	}
	if array.Size > maxABIFixedArrayLength {
		return fmt.Errorf("abi: fixed array length %d exceeds safety limit %d", array.Size, maxABIFixedArrayLength)
	}
	if elemSize, err := getTypeSizeChecked(elem); err != nil {
		return err
	} else if headSize, ok := checkedTypeSizeMul(array.Size, elemSize); !ok {
		return fmt.Errorf("abi: array type %s is too large", array.stringKind)
	} else if headSize > maxABIStaticTypeSize {
		return fmt.Errorf("abi: array type %s exceeds static-size safety limit %d", array.stringKind, maxABIStaticTypeSize)
	} else if elemSize == 0 && array.Size > maxZeroSizedArrayElements {
		return fmt.Errorf("abi: zero-sized array length %d exceeds safety limit %d", array.Size, maxZeroSizedArrayElements)
	}
	if _, err := makeArrayType(array.Size, elem.GetType()); err != nil {
		return fmt.Errorf("abi: invalid array type %s: %w", array.stringKind, err)
	}
	return nil
}

func makeArrayType(length int, elem reflect.Type) (typ reflect.Type, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%v", recovered)
		}
	}()
	return reflect.ArrayOf(length, elem), nil
}

func makeTupleType(fields []reflect.StructField) (typ reflect.Type, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("abi: invalid tuple type: %v", recovered)
		}
	}()
	return reflect.StructOf(fields), nil
}

// GetType returns the reflection type of the ABI type.
func (t Type) GetType() reflect.Type {
	switch t.T {
	case IntTy:
		return reflectIntType(false, t.Size)
	case UintTy:
		return reflectIntType(true, t.Size)
	case BoolTy:
		return reflect.TypeFor[bool]()
	case StringTy:
		return reflect.TypeFor[string]()
	case SliceTy:
		return reflect.SliceOf(t.Elem.GetType())
	case ArrayTy:
		return reflect.ArrayOf(t.Size, t.Elem.GetType())
	case TupleTy:
		return t.TupleType
	case AddressTy:
		return reflect.TypeFor[common.Address]()
	case FixedBytesTy:
		return reflect.ArrayOf(t.Size, reflect.TypeFor[byte]())
	case BytesTy:
		return reflect.TypeFor[[]byte]()
	case HashTy, FixedPointTy: // currently not used
		return reflect.TypeFor[[64]byte]()
	case FunctionTy:
		return reflect.TypeFor[[common.AddressLength + 4]byte]()
	default:
		panic("Invalid type")
	}
}

// String implements Stringer.
func (t Type) String() (out string) {
	return t.stringKind
}

func (t Type) pack(v reflect.Value) ([]byte, error) {
	// dereference pointer first if it's a pointer
	v = indirect(v)
	if err := typeCheck(t, v); err != nil {
		return nil, err
	}

	switch t.T {
	case SliceTy, ArrayTy:
		var ret []byte

		if t.requiresLengthPrefix() {
			// append length
			ret = append(ret, packNum(reflect.ValueOf(v.Len()))...)
		}

		// calculate offset if any
		offset := 0
		offsetReq := isDynamicType(*t.Elem)
		if offsetReq {
			var ok bool
			offset, ok = checkedTypeSizeMul(getTypeSize(*t.Elem), v.Len())
			if !ok {
				return nil, fmt.Errorf("abi: array head size overflows int")
			}
		}
		var tail []byte
		for i := 0; i < v.Len(); i++ {
			val, err := t.Elem.pack(v.Index(i))
			if err != nil {
				return nil, err
			}
			if !offsetReq {
				ret = append(ret, val...)
				continue
			}
			ret = append(ret, packNum(reflect.ValueOf(offset))...)
			var ok bool
			offset, ok = checkedTypeSizeAdd(offset, len(val))
			if !ok {
				return nil, fmt.Errorf("abi: array encoding size overflows int")
			}
			tail = append(tail, val...)
		}
		return append(ret, tail...), nil
	case TupleTy:
		// (T1,...,Tk) for k >= 0 and any types T1, …, Tk
		// enc(X) = head(X(1)) ... head(X(k)) tail(X(1)) ... tail(X(k))
		// where X = (X(1), ..., X(k)) and head and tail are defined for Ti being a static
		// type as
		//     head(X(i)) = enc(X(i)) and tail(X(i)) = "" (the empty string)
		// and as
		//     head(X(i)) = enc(len(head(X(1)) ... head(X(k)) tail(X(1)) ... tail(X(i-1))))
		//     tail(X(i)) = enc(X(i))
		// otherwise, i.e. if Ti is a dynamic type.
		fieldmap, err := mapTupleRawNamesToStructFields(t.TupleRawNames, v)
		if err != nil {
			return nil, err
		}
		// Calculate prefix occupied size.
		offset := 0
		for _, elem := range t.TupleElems {
			var ok bool
			offset, ok = checkedTypeSizeAdd(offset, getTypeSize(*elem))
			if !ok {
				return nil, fmt.Errorf("abi: tuple head size overflows int")
			}
		}
		var ret, tail []byte
		for i, elem := range t.TupleElems {
			field := v.FieldByIndex(fieldmap[i])
			if !field.IsValid() {
				return nil, fmt.Errorf("field %s for tuple not found in the given struct", t.TupleRawNames[i])
			}
			val, err := elem.pack(field)
			if err != nil {
				return nil, err
			}
			if isDynamicType(*elem) {
				ret = append(ret, packNum(reflect.ValueOf(offset))...)
				tail = append(tail, val...)
				var ok bool
				offset, ok = checkedTypeSizeAdd(offset, len(val))
				if !ok {
					return nil, fmt.Errorf("abi: tuple encoding size overflows int")
				}
			} else {
				ret = append(ret, val...)
			}
		}
		return append(ret, tail...), nil

	default:
		return packElement(t, v)
	}
}

// requiresLengthPrefix returns whether the type requires any sort of length
// prefixing.
func (t Type) requiresLengthPrefix() bool {
	return t.T == StringTy || t.T == BytesTy || t.T == SliceTy
}

// isDynamicType returns true if the type is dynamic.
// The following types are called “dynamic”:
// * bytes
// * string
// * T[] for any T
// * T[k] for any dynamic T and any k >= 0
// * (T1,...,Tk) if Ti is dynamic for some 1 <= i <= k
func isDynamicType(t Type) bool {
	if t.T == TupleTy {
		for _, elem := range t.TupleElems {
			if isDynamicType(*elem) {
				return true
			}
		}
		return false
	}
	return t.T == StringTy || t.T == BytesTy || t.T == SliceTy || (t.T == ArrayTy && isDynamicType(*t.Elem))
}

// getTypeSize returns the size that this type needs to occupy.
// We distinguish static and dynamic types. Static types are encoded in-place
// and dynamic types are encoded at a separately allocated location after the
// current block.
// So for a static variable, the size returned represents the size that the
// variable actually occupies.
// For a dynamic variable, the returned size is fixed 64 bytes, which is used
// to store the location reference for actual value storage.
func getTypeSize(t Type) int {
	size, err := getTypeSizeChecked(t)
	if err != nil {
		// Types returned by NewType have already passed this validation. Keep
		// this helper total for manually assembled Type values so arithmetic at
		// call sites cannot wrap into a negative size.
		return int(^uint(0) >> 1)
	}
	return size
}

func getTypeSizeChecked(t Type) (int, error) {
	if t.T == ArrayTy && !isDynamicType(*t.Elem) {
		// Recursively calculate type size if it is a nested array
		if t.Elem.T == ArrayTy || t.Elem.T == TupleTy {
			elemSize, err := getTypeSizeChecked(*t.Elem)
			if err != nil {
				return 0, err
			}
			if size, ok := checkedTypeSizeMul(t.Size, elemSize); ok {
				return size, nil
			}
			return 0, fmt.Errorf("abi: static type %s is too large", t.stringKind)
		}
		if size, ok := checkedTypeSizeMul(t.Size, 64); ok {
			return size, nil
		}
		return 0, fmt.Errorf("abi: static type %s is too large", t.stringKind)
	} else if t.T == TupleTy && !isDynamicType(t) {
		total := 0
		for _, elem := range t.TupleElems {
			elemSize, err := getTypeSizeChecked(*elem)
			if err != nil {
				return 0, err
			}
			if elemSize > int(^uint(0)>>1)-total {
				return 0, fmt.Errorf("abi: static type %s is too large", t.stringKind)
			}
			total += elemSize
		}
		return total, nil
	}
	return 64, nil
}

func checkedTypeSizeMul(left, right int) (int, bool) {
	if left < 0 || right < 0 {
		return 0, false
	}
	if left != 0 && right > int(^uint(0)>>1)/left {
		return 0, false
	}
	return left * right, true
}

func checkedTypeSizeAdd(left, right int) (int, bool) {
	if left < 0 || right < 0 || right > int(^uint(0)>>1)-left {
		return 0, false
	}
	return left + right, true
}

// isLetter reports whether a given 'rune' is classified as a Letter.
// This method is copied from reflect/type.go
func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

// isValidFieldName checks if a string is a valid (struct) field name or not.
//
// According to the language spec, a field name should be an identifier.
//
// identifier = letter { letter | unicode_digit } .
// letter = unicode_letter | "_" .
// This method is copied from reflect/type.go
func isValidFieldName(fieldName string) bool {
	for i, c := range fieldName {
		if i == 0 && !isLetter(c) {
			return false
		}

		if !(isLetter(c) || unicode.IsDigit(c)) {
			return false
		}
	}

	return len(fieldName) > 0
}
