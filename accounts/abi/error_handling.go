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
	"reflect"

	"github.com/theQRL/go-qrl/common"
)

var (
	errBadBool     = errors.New("abi: improperly encoded boolean value")
	errBadUint8    = errors.New("abi: improperly encoded uint8 value")
	errBadUint16   = errors.New("abi: improperly encoded uint16 value")
	errBadUint32   = errors.New("abi: improperly encoded uint32 value")
	errBadUint64   = errors.New("abi: improperly encoded uint64 value")
	errBadUint256  = errors.New("abi: improperly encoded uint256 value")
	errBadInt8     = errors.New("abi: improperly encoded int8 value")
	errBadInt16    = errors.New("abi: improperly encoded int16 value")
	errBadInt32    = errors.New("abi: improperly encoded int32 value")
	errBadInt64    = errors.New("abi: improperly encoded int64 value")
	errBadInt256   = errors.New("abi: improperly encoded int256 value")
	errInvalidSign = errors.New("abi: negatively-signed value cannot be packed into uint parameter")
)

func errBadUint(size int) error {
	switch size {
	case 8:
		return errBadUint8
	case 16:
		return errBadUint16
	case 32:
		return errBadUint32
	case 64:
		return errBadUint64
	case 256:
		return errBadUint256
	default:
		return fmt.Errorf("abi: improperly encoded uint%d value", size)
	}
}

func errBadInt(size int) error {
	switch size {
	case 8:
		return errBadInt8
	case 16:
		return errBadInt16
	case 32:
		return errBadInt32
	case 64:
		return errBadInt64
	case 256:
		return errBadInt256
	default:
		return fmt.Errorf("abi: improperly encoded int%d value", size)
	}
}

// formatSliceString formats the reflection kind with the given slice size
// and returns a formatted string representation.
func formatSliceString(kind reflect.Kind, sliceSize int) string {
	if sliceSize == -1 {
		return fmt.Sprintf("[]%v", kind)
	}
	return fmt.Sprintf("[%d]%v", sliceSize, kind)
}

// sliceTypeCheck checks that the given slice can by assigned to the reflection
// type in t.
func sliceTypeCheck(t Type, val reflect.Value) error {
	if !val.IsValid() {
		return typeErr(t.GetType(), "<nil>")
	}
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return typeErr(formatSliceString(t.GetType().Kind(), t.Size), val.Type())
	}

	if t.T == ArrayTy && val.Len() != t.Size {
		return typeErr(formatSliceString(t.Elem.GetType().Kind(), t.Size), formatSliceString(val.Type().Elem().Kind(), val.Len()))
	}

	if t.Elem.T == SliceTy || t.Elem.T == ArrayTy {
		if val.Len() > 0 {
			return sliceTypeCheck(*t.Elem, val.Index(0))
		}
	}

	if val.Type().Elem().Kind() != t.Elem.GetType().Kind() {
		return typeErr(formatSliceString(t.Elem.GetType().Kind(), t.Size), val.Type())
	}
	return nil
}

// typeCheck checks that the given reflection value can be assigned to the reflection
// type in t.
func typeCheck(t Type, value reflect.Value) error {
	if !value.IsValid() {
		return typeErr(t.GetType(), "<nil>")
	}
	if value.Kind() == reflect.Ptr && value.IsNil() {
		return typeErr(t.GetType(), "<nil>")
	}
	if t.T == SliceTy || t.T == ArrayTy {
		return sliceTypeCheck(t, value)
	}

	// Check base type validity. Element types will be checked later on.
	expected := t.GetType()
	if expected.Kind() != value.Kind() {
		return typeErr(expected.Kind(), value.Kind())
	}
	switch t.T {
	case IntTy, UintTy:
		// Integers wider than 64 bits use *big.Int. Merely checking the
		// pointer kind would let unrelated pointers reach a type assertion.
		if expected.Kind() == reflect.Ptr && value.Type() != expected {
			return typeErr(expected, value.Type())
		}
	case AddressTy:
		if value.Len() != common.AddressLength || value.Type().Elem().Kind() != reflect.Uint8 {
			return typeErr(expected, value.Type())
		}
	case FixedBytesTy, FunctionTy:
		if value.Len() != t.Size || value.Type().Elem().Kind() != reflect.Uint8 {
			return typeErr(expected, value.Type())
		}
	}
	if t.T == FixedBytesTy && t.Size != value.Len() {
		return typeErr(t.GetType(), value.Type())
	}
	return nil
}

// typeErr returns a formatted type casting error.
func typeErr(expected, got any) error {
	return fmt.Errorf("abi: cannot use %v as type %v as argument", got, expected)
}
