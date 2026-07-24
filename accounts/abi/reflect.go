// Copyright 2016 The go-ethereum Authors
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
	"math/big"
	"reflect"
	"strings"
)

// ConvertType converts an interface of a runtime type into an interface of the
// given type, e.g. turn this code:
//
//	var fields []reflect.StructField
//
//	fields = append(fields, reflect.StructField{
//			Name: "X",
//			Type: reflect.TypeOf(new(big.Int)),
//			Tag:  reflect.StructTag("json:\"" + "x" + "\""),
//	})
//
// into:
//
//	type TupleT struct { X *big.Int }
func ConvertType(in any, proto any) any {
	protoType := reflect.TypeOf(proto)
	if reflect.TypeOf(in).ConvertibleTo(protoType) {
		return reflect.ValueOf(in).Convert(protoType).Interface()
	}
	// Use set as a last ditch effort
	if err := set(reflect.ValueOf(proto), reflect.ValueOf(in)); err != nil {
		panic(err)
	}
	return proto
}

// indirect recursively dereferences the value until it either gets the value
// or finds a big.Int
func indirect(v reflect.Value) reflect.Value {
	if !v.IsValid() || (v.Kind() == reflect.Ptr && v.IsNil()) {
		return v
	}
	if v.Kind() == reflect.Ptr && v.Elem().Type() != reflect.TypeFor[big.Int]() {
		return indirect(v.Elem())
	}
	return v
}

// reflectIntType returns the reflect using the given size and
// unsignedness.
func reflectIntType(unsigned bool, size int) reflect.Type {
	if unsigned {
		switch size {
		case 8:
			return reflect.TypeFor[uint8]()
		case 16:
			return reflect.TypeFor[uint16]()
		case 32:
			return reflect.TypeFor[uint32]()
		case 64:
			return reflect.TypeFor[uint64]()
		}
	}
	switch size {
	case 8:
		return reflect.TypeFor[int8]()
	case 16:
		return reflect.TypeFor[int16]()
	case 32:
		return reflect.TypeFor[int32]()
	case 64:
		return reflect.TypeFor[int64]()
	}
	return reflect.TypeFor[*big.Int]()
}

// mustArrayToByteSlice creates a new byte slice with the exact same size as value
// and copies the bytes in value to the new slice.
func mustArrayToByteSlice(value reflect.Value) reflect.Value {
	slice := reflect.ValueOf(make([]byte, value.Len()))
	reflect.Copy(slice, value)
	return slice
}

// set attempts to assign src to dst by either setting, copying or otherwise.
//
// set is a bit more lenient when it comes to assignment and doesn't force an as
// strict ruleset as bare `reflect` does.
func set(dst, src reflect.Value) error {
	if !dst.IsValid() {
		return errors.New("abi: cannot unmarshal into a nil destination")
	}
	if !src.IsValid() || (src.Kind() == reflect.Ptr && src.IsNil()) {
		return errors.New("abi: cannot unmarshal a nil value")
	}
	if dst.Kind() == reflect.Ptr && dst.IsNil() && dst.Type().Elem() != reflect.TypeFor[big.Int]() {
		return fmt.Errorf("abi: cannot unmarshal into nil pointer %v", dst.Type())
	}
	dstType, srcType := dst.Type(), src.Type()
	switch {
	case dstType.Kind() == reflect.Interface && dst.Elem().IsValid() && (dst.Elem().Type().Kind() == reflect.Ptr || dst.Elem().CanSet()):
		return set(dst.Elem(), src)
	case dstType.Kind() == reflect.Ptr && dstType.Elem() != reflect.TypeFor[big.Int]():
		return set(dst.Elem(), src)
	case srcType.AssignableTo(dstType) && dst.CanSet():
		dst.Set(src)
	case dstType.Kind() == reflect.Slice && srcType.Kind() == reflect.Slice && dst.CanSet():
		return setSlice(dst, src)
	case dstType.Kind() == reflect.Array:
		return setArray(dst, src)
	case dstType.Kind() == reflect.Struct:
		return setStruct(dst, src)
	default:
		return fmt.Errorf("abi: cannot unmarshal %v in to %v", src.Type(), dst.Type())
	}
	return nil
}

// setSlice attempts to assign src to dst when slices are not assignable by default
// e.g. src: [][]byte -> dst: [][15]byte
// setSlice ignores if we cannot copy all of src' elements.
func setSlice(dst, src reflect.Value) error {
	slice := reflect.MakeSlice(dst.Type(), src.Len(), src.Len())
	for i := 0; i < src.Len(); i++ {
		if err := set(slice.Index(i), src.Index(i)); err != nil {
			return err
		}
	}
	if dst.CanSet() {
		dst.Set(slice)
		return nil
	}
	return errors.New("cannot set slice, destination not settable")
}

func setArray(dst, src reflect.Value) error {
	if src.Kind() == reflect.Ptr {
		return set(dst, indirect(src))
	}
	array := reflect.New(dst.Type()).Elem()
	min := src.Len()
	if src.Len() > dst.Len() {
		min = dst.Len()
	}
	for i := range min {
		if err := set(array.Index(i), src.Index(i)); err != nil {
			return err
		}
	}
	if dst.CanSet() {
		dst.Set(array)
		return nil
	}
	return errors.New("cannot set array, destination not settable")
}

func setStruct(dst, src reflect.Value) error {
	if dst.NumField() < src.NumField() {
		return fmt.Errorf("abi: destination struct has %d fields, need %d", dst.NumField(), src.NumField())
	}
	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		dstField := dst.Field(i)
		if !dstField.IsValid() || !srcField.IsValid() {
			return fmt.Errorf("could not find src field: %v value: %v in destination", srcField.Type().Name(), srcField)
		}
		if err := set(dstField, srcField); err != nil {
			return err
		}
	}
	return nil
}

// mapArgNamesToStructFields maps a slice of argument names to struct fields.
//
// first round: for each Exportable field that contains a `abi:""` tag and this field name
// exists in the given argument name list, pair them together.
//
// second round: for each argument name that has not been already linked, find what
// variable is expected to be mapped into, if it exists and has not been used, pair them.
//
// Note this function assumes the given value is a struct value.
func mapArgNamesToStructFields(argNames []string, value reflect.Value) (map[string]string, error) {
	typ := value.Type()

	abi2struct := make(map[string]string)
	struct2abi := make(map[string]string)

	// first round ~~~
	for i := range typ.NumField() {
		structFieldName := typ.Field(i).Name

		// skip private struct fields.
		if structFieldName[:1] != strings.ToUpper(structFieldName[:1]) {
			continue
		}
		// skip fields that have no abi:"" tag.
		tagName, ok := typ.Field(i).Tag.Lookup("abi")
		if !ok {
			continue
		}
		// check if tag is empty.
		if tagName == "" {
			return nil, fmt.Errorf("struct: abi tag in '%s' is empty", structFieldName)
		}
		// check which argument field matches with the abi tag.
		found := false
		for _, arg := range argNames {
			if arg == tagName {
				if abi2struct[arg] != "" {
					return nil, fmt.Errorf("struct: abi tag in '%s' already mapped", structFieldName)
				}
				// pair them
				abi2struct[arg] = structFieldName
				struct2abi[structFieldName] = arg
				found = true
			}
		}
		// check if this tag has been mapped.
		if !found {
			return nil, fmt.Errorf("struct: abi tag '%s' defined but not found in abi", tagName)
		}
	}

	// second round ~~~
	for _, argName := range argNames {
		structFieldName := ToCamelCase(argName)

		if structFieldName == "" {
			return nil, errors.New("abi: purely underscored output cannot unpack to struct")
		}

		// this abi has already been paired, skip it... unless there exists another, yet unassigned
		// struct field with the same field name. If so, raise an error:
		//    abi: [ { "name": "value" } ]
		//    struct { Value  *big.Int , Value1 *big.Int `abi:"value"`}
		if abi2struct[argName] != "" {
			if abi2struct[argName] != structFieldName &&
				struct2abi[structFieldName] == "" &&
				value.FieldByName(structFieldName).IsValid() {
				return nil, fmt.Errorf("abi: multiple variables maps to the same abi field '%s'", argName)
			}
			continue
		}

		// return an error if this struct field has already been paired.
		if struct2abi[structFieldName] != "" {
			return nil, fmt.Errorf("abi: multiple outputs mapping to the same struct field '%s'", structFieldName)
		}

		if value.FieldByName(structFieldName).IsValid() {
			// pair them
			abi2struct[argName] = structFieldName
			struct2abi[structFieldName] = argName
		} else {
			// not paired, but annotate as used, to detect cases like
			//   abi : [ { "name": "value" }, { "name": "_value" } ]
			//   struct { Value *big.Int }
			struct2abi[structFieldName] = argName
		}
	}
	return abi2struct, nil
}

// mapTupleRawNamesToStructFields maps every tuple component occurrence to one
// concrete struct field. A map keyed by the ABI name cannot represent tuples
// such as (data,_data,data), because all three names normalize to Data and ABI
// component names are not required to be unique. Generated structs resolve
// these collisions as Data, Data0, Data1; explicit abi tags are matched to
// duplicate names in stable declaration order.
func mapTupleRawNamesToStructFields(argNames []string, value reflect.Value) ([][]int, error) {
	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("abi: cannot map tuple fields onto %v", value.Type())
	}
	typ := value.Type()
	mapped := make([][]int, len(argNames))
	usedFields := make(map[string]bool)

	// Explicit tags take precedence. When a tuple contains duplicate raw names,
	// successive tagged fields map to successive unmatched occurrences.
	for fieldIndex := range typ.NumField() {
		field := typ.Field(fieldIndex)
		if field.PkgPath != "" { // unexported
			continue
		}
		tagName, ok := field.Tag.Lookup("abi")
		if !ok {
			continue
		}
		if tagName == "" {
			return nil, fmt.Errorf("struct: abi tag in '%s' is empty", field.Name)
		}
		match := -1
		for argIndex, argName := range argNames {
			if mapped[argIndex] == nil && argName == tagName {
				match = argIndex
				break
			}
		}
		if match == -1 {
			return nil, fmt.Errorf("struct: abi tag '%s' defined but not found in abi", tagName)
		}
		mapped[match] = []int{fieldIndex}
		usedFields[fmt.Sprint(mapped[match])] = true
	}

	// Reproduce the collision resolution used by NewType and abigen. Resolve
	// names for every occurrence, including those already mapped by tags, so
	// suffix numbering remains stable.
	usedNames := make(map[string]bool)
	for argIndex, argName := range argNames {
		name := ToCamelCase(argName)
		if name == "" {
			return nil, errors.New("abi: purely underscored tuple field is not supported")
		}
		name = ResolveNameConflict(name, func(candidate string) bool { return usedNames[candidate] })
		usedNames[name] = true
		if mapped[argIndex] != nil {
			continue
		}
		field, ok := typ.FieldByName(name)
		if !ok || field.PkgPath != "" || usedFields[fmt.Sprint(field.Index)] {
			return nil, fmt.Errorf("abi: field %s for tuple component %d not found in %v", name, argIndex, typ)
		}
		mapped[argIndex] = field.Index
		usedFields[fmt.Sprint(field.Index)] = true
	}
	return mapped, nil
}
