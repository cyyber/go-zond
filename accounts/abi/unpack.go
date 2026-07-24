// Copyright 2017 The go-ethereum Authors
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
	"math"
	"math/big"
	"reflect"
	"strconv"

	"github.com/theQRL/go-qrl/common"
)

const maxZeroSizedArrayElements = 1 << 20

var (
	// MaxUint256 is the maximum value that can be represented by a uint256.
	MaxUint256 = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 256), common.Big1)
	// MaxInt256 is the maximum value that can be represented by a int256.
	MaxInt256 = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 255), common.Big1)
	// MaxUint512 is the maximum value that can be represented by a uint512.
	MaxUint512 = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 512), common.Big1)
)

// ReadInteger reads the integer based on its kind and returns the appropriate value.
func ReadInteger(typ Type, b []byte) (any, error) {
	if typ.T != IntTy && typ.T != UintTy {
		return nil, fmt.Errorf("abi: invalid type %v in call to ReadInteger", typ.T)
	}
	if !validIntegerSize(typ.Size) {
		return nil, fmt.Errorf("abi: invalid integer size %d", typ.Size)
	}
	if len(b) != 64 {
		return nil, fmt.Errorf("abi: invalid integer word length %d, want 64", len(b))
	}
	ret := new(big.Int).SetBytes(b)

	if typ.T == UintTy {
		u64, isu64 := ret.Uint64(), ret.IsUint64()
		switch typ.Size {
		case 8:
			if !isu64 || u64 > math.MaxUint8 {
				return nil, errBadUint8
			}
			return byte(u64), nil
		case 16:
			if !isu64 || u64 > math.MaxUint16 {
				return nil, errBadUint16
			}
			return uint16(u64), nil
		case 32:
			if !isu64 || u64 > math.MaxUint32 {
				return nil, errBadUint32
			}
			return uint32(u64), nil
		case 64:
			if !isu64 {
				return nil, errBadUint64
			}
			return u64, nil
		default:
			if !fitsUnsignedInteger(ret, typ.Size) {
				return nil, errBadUint(typ.Size)
			}
			return ret, nil
		}
	}

	// big.SetBytes can't tell if a number is negative or positive in itself.
	// Signed integers are sign-extended to fill the full 64-byte ABI slot,
	// so a value occupies the slot's MSB (bit 511) iff it is negative.
	if ret.Bit(511) == 1 {
		ret.Add(MaxUint512, new(big.Int).Neg(ret))
		ret.Add(ret, common.Big1)
		ret.Neg(ret)
	}
	i64, isi64 := ret.Int64(), ret.IsInt64()
	switch typ.Size {
	case 8:
		if !isi64 || i64 < math.MinInt8 || i64 > math.MaxInt8 {
			return nil, errBadInt8
		}
		return int8(i64), nil
	case 16:
		if !isi64 || i64 < math.MinInt16 || i64 > math.MaxInt16 {
			return nil, errBadInt16
		}
		return int16(i64), nil
	case 32:
		if !isi64 || i64 < math.MinInt32 || i64 > math.MaxInt32 {
			return nil, errBadInt32
		}
		return int32(i64), nil
	case 64:
		if !isi64 {
			return nil, errBadInt64
		}
		return i64, nil
	default:
		if !fitsSignedInteger(ret, typ.Size) {
			return nil, errBadInt(typ.Size)
		}
		return ret, nil
	}
}

// readBool reads a bool.
func readBool(word []byte) (bool, error) {
	for _, b := range word[:63] {
		if b != 0 {
			return false, errBadBool
		}
	}
	switch word[63] {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, errBadBool
	}
}

// A function type is simply the address with the function selection signature at the end.
//
// readFunctionType enforces that standard by always presenting it as an array of
// (AddressLength + 4) bytes (address + 4-byte selector).
func readFunctionType(t Type, word []byte) (funcTy [common.AddressLength + 4]byte, err error) {
	if t.T != FunctionTy {
		return [common.AddressLength + 4]byte{}, errors.New("abi: invalid type in call to make function type byte array")
	}
	if common.AddressLength+4 > len(word) {
		return [common.AddressLength + 4]byte{}, errors.New("abi: function type does not fit in a 64-byte ABI word with 64-byte addresses")
	}
	for _, b := range word[common.AddressLength+4:] {
		if b != 0 {
			err = fmt.Errorf("abi: got improperly encoded function type, got %v", word)
			return
		}
	}
	copy(funcTy[:], word[0:common.AddressLength+4])
	return
}

// ReadFixedBytes uses reflection to create a fixed array to be read from.
func ReadFixedBytes(t Type, word []byte) (any, error) {
	if t.T != FixedBytesTy {
		return nil, errors.New("abi: invalid type in call to make fixed byte array")
	}
	if t.Size < 1 || t.Size > 64 {
		return nil, fmt.Errorf("abi: invalid fixed byte size %d", t.Size)
	}
	if len(word) != 64 {
		return nil, fmt.Errorf("abi: invalid fixed byte word length %d, want 64", len(word))
	}
	for _, b := range word[t.Size:] {
		if b != 0 {
			return nil, fmt.Errorf("abi: improperly encoded bytes%d value", t.Size)
		}
	}
	// convert
	array := reflect.New(t.GetType()).Elem()

	reflect.Copy(array, reflect.ValueOf(word[0:t.Size]))
	return array.Interface(), nil
}

// forEachUnpack iteratively unpack elements.
func forEachUnpack(t Type, output []byte, start, size int) (any, error) {
	if size < 0 {
		return nil, fmt.Errorf("cannot marshal input to array, size is negative (%d)", size)
	}
	// Arrays can contain multi-word or zero-sized static elements, while slices
	// contain one head word per element. getTypeSize captures all three cases.
	elemSize := getTypeSize(*t.Elem)
	if elemSize == 0 && size > maxZeroSizedArrayElements {
		return nil, fmt.Errorf("abi: zero-sized array length %d exceeds safety limit %d", size, maxZeroSizedArrayElements)
	}
	if start < 0 || start > len(output) || (elemSize > 0 && size > (len(output)-start)/elemSize) {
		return nil, fmt.Errorf("abi: cannot marshal into go array: offset %d would go over slice boundary (len=%d)", start+elemSize*size, len(output))
	}

	// this value will become our slice or our array, depending on the type
	var refSlice reflect.Value

	switch t.T {
	case SliceTy:
		// declare our slice
		refSlice = reflect.MakeSlice(t.GetType(), size, size)
	case ArrayTy:
		// declare our array
		refSlice = reflect.New(t.GetType()).Elem()
	default:
		return nil, errors.New("abi: invalid type in array/slice unpacking stage")
	}

	for i, j := start, 0; j < size; i, j = i+elemSize, j+1 {
		inter, err := toGoType(i, *t.Elem, output)
		if err != nil {
			return nil, err
		}

		// append the item to our reflect slice
		refSlice.Index(j).Set(reflect.ValueOf(inter))
	}

	// return the interface
	return refSlice.Interface(), nil
}

func forTupleUnpack(t Type, output []byte) (any, error) {
	retval := reflect.New(t.GetType()).Elem()
	virtualArgs := 0
	for index, elem := range t.TupleElems {
		marshalledValue, err := toGoType((index+virtualArgs)*64, *elem, output)
		if err != nil {
			return nil, err
		}
		if elem.T == ArrayTy && !isDynamicType(*elem) {
			// If we have a static array, like [3]uint256, these are coded as
			// just like uint256,uint256,uint256.
			// This means that we need to add two 'virtual' arguments when
			// we count the index from now on.
			//
			// Array values nested multiple levels deep are also encoded inline:
			// [2][3]uint256: uint256,uint256,uint256,uint256,uint256,uint256
			//
			// Calculate the full array size to get the correct offset for the next argument.
			// Decrement it by 1, as the normal index increment is still applied.
			virtualArgs += getTypeSize(*elem)/64 - 1
		} else if elem.T == TupleTy && !isDynamicType(*elem) {
			// If we have a static tuple, like (uint256, bool, uint256), these are
			// coded as just like uint256,bool,uint256
			virtualArgs += getTypeSize(*elem)/64 - 1
		}
		retval.Field(index).Set(reflect.ValueOf(marshalledValue))
	}
	return retval.Interface(), nil
}

// toGoType parses the output bytes and recursively assigns the value of these bytes
// into a go type with accordance with the ABI spec.
func toGoType(index int, t Type, output []byte) (any, error) {
	if index < 0 || index > len(output) {
		return nil, fmt.Errorf("abi: cannot marshal in to go type: offset %d would go over slice boundary (len=%d)", index, len(output))
	}
	// Zero-length static arrays and empty tuples have an empty encoding. Decode
	// them before requiring a full word so that they also work next to ordinary
	// arguments (where they consume no head slot).
	if !isDynamicType(t) && getTypeSize(t) == 0 {
		switch t.T {
		case ArrayTy:
			return forEachUnpack(t, output[index:], 0, t.Size)
		case TupleTy:
			return forTupleUnpack(t, output[index:])
		}
	}
	if len(output)-index < 64 {
		return nil, fmt.Errorf("abi: cannot marshal in to go type: length insufficient %d require %d", len(output), index+64)
	}

	var (
		returnOutput  []byte
		begin, length int
		err           error
	)

	// if we require a length prefix, find the beginning word and size returned.
	if t.requiresLengthPrefix() {
		begin, length, err = lengthPrefixPointsTo(index, output)
		if err != nil {
			return nil, err
		}
	} else {
		returnOutput = output[index : index+64]
	}

	switch t.T {
	case TupleTy:
		if isDynamicType(t) {
			begin, err := tuplePointsTo(index, output)
			if err != nil {
				return nil, err
			}
			return forTupleUnpack(t, output[begin:])
		}
		return forTupleUnpack(t, output[index:])
	case SliceTy:
		return forEachUnpack(t, output[begin:], 0, length)
	case ArrayTy:
		if isDynamicType(*t.Elem) {
			offset, err := tuplePointsTo(index, output)
			if err != nil {
				return nil, err
			}
			return forEachUnpack(t, output[offset:], 0, t.Size)
		}
		return forEachUnpack(t, output[index:], 0, t.Size)
	case StringTy: // variable arrays are written at the end of the return bytes
		if length > len(output)-begin {
			return nil, fmt.Errorf("abi: cannot marshal in to go string: length %d would go over slice boundary (len=%d)", length, len(output)-begin)
		}
		return string(output[begin : begin+length]), nil
	case IntTy, UintTy:
		return ReadInteger(t, returnOutput)
	case BoolTy:
		return readBool(returnOutput)
	case AddressTy:
		return common.BytesToAddress(returnOutput), nil
	case HashTy:
		return common.BytesToHash(returnOutput), nil
	case BytesTy:
		if length > len(output)-begin {
			return nil, fmt.Errorf("abi: cannot marshal in to go bytes: length %d would go over slice boundary (len=%d)", length, len(output)-begin)
		}
		return output[begin : begin+length], nil
	case FixedBytesTy:
		return ReadFixedBytes(t, returnOutput)
	case FunctionTy:
		return readFunctionType(t, returnOutput)
	default:
		return nil, fmt.Errorf("abi: unknown type %v", t.T)
	}
}

// lengthPrefixPointsTo interprets a 64 byte slice as an offset and then determines which indices to look to decode the type.
func lengthPrefixPointsTo(index int, output []byte) (start int, length int, err error) {
	bigOffsetEnd := new(big.Int).SetBytes(output[index : index+64])
	bigOffsetEnd.Add(bigOffsetEnd, common.Big64)
	outputLength := big.NewInt(int64(len(output)))

	if bigOffsetEnd.BitLen() > strconv.IntSize-1 {
		return 0, 0, fmt.Errorf("abi offset larger than int%d: %v", strconv.IntSize, bigOffsetEnd)
	}
	if bigOffsetEnd.Cmp(outputLength) > 0 {
		return 0, 0, fmt.Errorf("abi: cannot marshal in to go slice: offset %v would go over slice boundary (len=%v)", bigOffsetEnd, outputLength)
	}

	offsetEnd := int(bigOffsetEnd.Uint64())
	lengthBig := new(big.Int).SetBytes(output[offsetEnd-64 : offsetEnd])

	if lengthBig.BitLen() > strconv.IntSize-1 {
		return 0, 0, fmt.Errorf("abi: length larger than int%d: %v", strconv.IntSize, lengthBig)
	}
	start = int(bigOffsetEnd.Uint64())
	length = int(lengthBig.Uint64())
	return
}

// tuplePointsTo resolves the location reference for dynamic tuple.
func tuplePointsTo(index int, output []byte) (start int, err error) {
	offset := new(big.Int).SetBytes(output[index : index+64])
	outputLen := big.NewInt(int64(len(output)))

	if offset.Cmp(outputLen) > 0 {
		return 0, fmt.Errorf("abi: cannot marshal in to go slice: offset %v would go over slice boundary (len=%v)", offset, outputLen)
	}
	if offset.BitLen() > strconv.IntSize-1 {
		return 0, fmt.Errorf("abi offset larger than int%d: %v", strconv.IntSize, offset)
	}
	return int(offset.Uint64()), nil
}
