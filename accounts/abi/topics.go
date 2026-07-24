// Copyright 2018 The go-ethereum Authors
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
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/big"
	"reflect"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/math"
	"github.com/theQRL/go-qrl/crypto"
)

// MakeTopic constructs the topic for an indexed event value whose ABI type is
// known. Value types are encoded directly into one VM64 word. Strings, bytes,
// arrays and tuples are hashed using Hyperion's indexed-event encoding: array
// lengths and dynamic offsets are omitted, array and tuple members are
// concatenated in place, and string/bytes members inside an array or tuple are
// padded to a 64-byte boundary. The resulting 32-byte Keccak digest is
// left-aligned in the 64-byte log topic.
//
// MakeTopics remains the type-agnostic filter API. In particular, common.Hash
// and common.LogTopic passed to MakeTopics are still treated as precomputed
// topic rules. Use MakeTopic when the declared ABI type is available and a
// composite preimage must be encoded; the explicit Type also disambiguates
// signed and unsigned *big.Int values and [N]byte values that could mean either
// bytesN or an array of integers.
func MakeTopic(typ Type, value any) (common.LogTopic, error) {
	if err := validateIndexedTopicType(typ, 0); err != nil {
		return common.LogTopic{}, err
	}
	if typ.T == FunctionTy {
		return common.LogTopic{}, errors.New("abi: indexed function values do not fit in one VM64 topic")
	}

	val := indirect(reflect.ValueOf(value))
	if indexedTopicNeedsHash(typ) {
		hasher := crypto.NewKeccakState()
		if err := encodeIndexedTopicValue(hasher, typ, val, false); err != nil {
			return common.LogTopic{}, fmt.Errorf("abi: cannot encode indexed %s topic: %w", typ.String(), err)
		}
		var digest common.Hash
		if _, err := hasher.Read(digest[:]); err != nil {
			return common.LogTopic{}, fmt.Errorf("abi: cannot hash indexed %s topic: %w", typ.String(), err)
		}
		return common.HashToLogTopic(digest), nil
	}
	if err := typeCheck(typ, val); err != nil {
		return common.LogTopic{}, err
	}
	encoded, err := packElement(typ, val)
	if err != nil {
		return common.LogTopic{}, err
	}
	if len(encoded) != common.LogTopicLength {
		return common.LogTopic{}, fmt.Errorf("abi: indexed %s encodes to %d bytes, want one %d-byte topic", typ.String(), len(encoded), common.LogTopicLength)
	}
	var topic common.LogTopic
	copy(topic[:], encoded)
	return topic, nil
}

// MakeTopicHash returns the Keccak-256 digest used to filter an indexed string,
// bytes, array, slice, or tuple value.
//
// Hyperion stores this 32-byte digest in the high half of a 64-byte VM topic
// (digest || zero-padding). Do not recover it with common.BytesToHash(topic[:]):
// BytesToHash crops oversized input from the left and would retain the padded
// low half instead. This helper validates the VM64 padding before returning the
// high 32 bytes. Directly encoded scalar topics do not have a hash digest and
// are rejected.
func MakeTopicHash(typ Type, value any) (common.Hash, error) {
	topic, err := MakeTopic(typ, value)
	if err != nil {
		return common.Hash{}, err
	}
	if !indexedTopicNeedsHash(typ) {
		return common.Hash{}, fmt.Errorf("abi: indexed %s topic is encoded directly and has no hash digest", typ.String())
	}
	for _, b := range topic[common.HashLength:] {
		if b != 0 {
			return common.Hash{}, fmt.Errorf("abi: indexed %s hash has non-zero VM64 padding", typ.String())
		}
	}
	var digest common.Hash
	copy(digest[:], topic[:common.HashLength])
	return digest, nil
}

func indexedTopicNeedsHash(typ Type) bool {
	switch typ.T {
	case StringTy, BytesTy, SliceTy, ArrayTy, TupleTy:
		return true
	default:
		return false
	}
}

// encodeIndexedTopicValue writes Hyperion's special in-place indexed-event
// encoding. At the top level, string and bytes are hashed without padding.
// Within arrays and tuples they are padded to the VM64 word width. Every other
// elementary value already occupies exactly one word.
func encodeIndexedTopicValue(dst io.Writer, typ Type, value reflect.Value, nested bool) error {
	value = indirect(value)
	if err := typeCheck(typ, value); err != nil {
		return err
	}
	switch typ.T {
	case StringTy:
		return writeIndexedTopicBytes(dst, []byte(value.String()), nested)
	case BytesTy:
		if value.Type() != reflect.TypeFor[[]byte]() {
			return errors.New("bytes type is neither slice nor array")
		}
		return writeIndexedTopicBytes(dst, value.Bytes(), nested)
	case SliceTy, ArrayTy:
		for i := 0; i < value.Len(); i++ {
			if err := encodeIndexedTopicValue(dst, *typ.Elem, value.Index(i), true); err != nil {
				return fmt.Errorf("element %d: %w", i, err)
			}
		}
		return nil
	case TupleTy:
		fieldMap, err := mapTupleRawNamesToStructFields(typ.TupleRawNames, value)
		if err != nil {
			return err
		}
		for i, elem := range typ.TupleElems {
			field := value.FieldByIndex(fieldMap[i])
			if !field.IsValid() {
				return fmt.Errorf("field %s for tuple not found in the given struct", typ.TupleRawNames[i])
			}
			if err := encodeIndexedTopicValue(dst, *elem, field, true); err != nil {
				return fmt.Errorf("tuple field %q: %w", typ.TupleRawNames[i], err)
			}
		}
		return nil
	default:
		encoded, err := packElement(typ, value)
		if err != nil {
			return err
		}
		if len(encoded) != common.LogTopicLength {
			return fmt.Errorf("indexed %s member encodes to %d bytes, want %d", typ.String(), len(encoded), common.LogTopicLength)
		}
		return writeFull(dst, encoded)
	}
}

func writeIndexedTopicBytes(dst io.Writer, value []byte, nested bool) error {
	if err := writeFull(dst, value); err != nil {
		return err
	}
	if !nested {
		return nil
	}
	padding := (common.LogTopicLength - len(value)%common.LogTopicLength) % common.LogTopicLength
	if padding == 0 {
		return nil
	}
	var zeroWord [common.LogTopicLength]byte
	return writeFull(dst, zeroWord[:padding])
}

func writeFull(dst io.Writer, value []byte) error {
	for len(value) > 0 {
		written, err := dst.Write(value)
		if err != nil {
			return err
		}
		if written <= 0 || written > len(value) {
			return io.ErrShortWrite
		}
		value = value[written:]
	}
	return nil
}

func validateIndexedTopicType(typ Type, depth int) error {
	if depth > maxABITypeNesting {
		return fmt.Errorf("abi: indexed topic type nesting exceeds safety limit %d", maxABITypeNesting)
	}
	switch typ.T {
	case IntTy, UintTy:
		if typ.Size < 1 || typ.Size > 512 {
			return fmt.Errorf("abi: invalid indexed integer width %d", typ.Size)
		}
	case BoolTy, StringTy, BytesTy:
	case AddressTy:
		if typ.Size != common.AddressLength {
			return fmt.Errorf("abi: invalid indexed address width %d", typ.Size)
		}
	case FixedBytesTy:
		if typ.Size < 1 || typ.Size > common.LogTopicLength {
			return fmt.Errorf("abi: invalid indexed fixed-bytes width %d", typ.Size)
		}
	case SliceTy, ArrayTy:
		if typ.Elem == nil {
			return errors.New("abi: indexed array type has no element type")
		}
		if typ.T == ArrayTy && (typ.Size < 0 || typ.Size > maxABIFixedArrayLength) {
			return fmt.Errorf("abi: invalid indexed array length %d", typ.Size)
		}
		return validateIndexedTopicType(*typ.Elem, depth+1)
	case TupleTy:
		if len(typ.TupleElems) != len(typ.TupleRawNames) {
			return errors.New("abi: indexed tuple type has inconsistent component metadata")
		}
		if len(typ.TupleElems) > maxABITupleFields {
			return fmt.Errorf("abi: indexed tuple field count %d exceeds safety limit %d", len(typ.TupleElems), maxABITupleFields)
		}
		if typ.TupleType == nil || typ.TupleType.Kind() != reflect.Struct {
			return errors.New("abi: indexed tuple type has no struct representation")
		}
		for _, elem := range typ.TupleElems {
			if elem == nil {
				return errors.New("abi: indexed tuple type has a nil component type")
			}
			if err := validateIndexedTopicType(*elem, depth+1); err != nil {
				return err
			}
		}
	case FunctionTy:
		if typ.Size != common.AddressLength+4 {
			return fmt.Errorf("abi: invalid indexed function width %d", typ.Size)
		}
	case HashTy, FixedPointTy:
		return fmt.Errorf("abi: unsupported indexed ABI type %d", typ.T)
	default:
		return fmt.Errorf("abi: unknown indexed ABI type %d", typ.T)
	}
	return nil
}

// MakeTopics converts a filter query argument list into a filter topic set.
// Topics are 64-byte ABI slots: numeric values are right-aligned, while
// fixed bytes and Keccak hash topics are left-aligned like ABI bytes32 values.
func MakeTopics(query ...[]any) ([][]common.LogTopic, error) {
	topics := make([][]common.LogTopic, len(query))
	for i, filter := range query {
		for _, rule := range filter {
			var topic common.LogTopic

			// Try to generate the topic based on simple types
			switch rule := rule.(type) {
			case common.LogTopic:
				topic = rule
			case common.Hash:
				topic = common.HashToLogTopic(rule)
			case common.Address:
				topic = common.AddressToLogTopic(rule)
			case *big.Int:
				if rule == nil {
					return nil, errors.New("unsupported indexed type: nil *big.Int")
				}
				if (rule.Sign() < 0 && !fitsSignedInteger(rule, 512)) || (rule.Sign() >= 0 && !fitsUnsignedInteger(rule, 512)) {
					return nil, fmt.Errorf("indexed integer does not fit in a 512-bit topic: %v", rule)
				}
				topic = common.BytesToRightAlignedLogTopic(math.U512Bytes(new(big.Int).Set(rule)))
			case bool:
				if rule {
					topic[common.LogTopicLength-1] = 1
				}
			case int8:
				topic = genIntType(int64(rule), 1)
			case int16:
				topic = genIntType(int64(rule), 2)
			case int32:
				topic = genIntType(int64(rule), 4)
			case int64:
				topic = genIntType(rule, 8)
			case uint8:
				topic = genUintType(uint64(rule))
			case uint16:
				topic = genUintType(uint64(rule))
			case uint32:
				topic = genUintType(uint64(rule))
			case uint64:
				topic = genUintType(rule)
			case string:
				topic = common.HashToLogTopic(crypto.Keccak256Hash([]byte(rule)))
			case []byte:
				topic = common.HashToLogTopic(crypto.Keccak256Hash(rule))

			default:
				// todo(rjl493456442) according to hyperion documentation, indexed event
				// parameters that are not value types i.e. arrays and structs are not
				// stored directly but instead a keccak256-hash of an encoding is stored.
				//
				// We only convert stringS and bytes to hash, still need to deal with
				// array(both fixed-size and dynamic-size) and struct.

				// Attempt to generate the topic from funky types
				val := reflect.ValueOf(rule)
				switch {
				// static byte array
				case val.Kind() == reflect.Array && reflect.TypeOf(rule).Elem().Kind() == reflect.Uint8:
					if val.Len() > common.LogTopicLength {
						return nil, fmt.Errorf("unsupported indexed type: %T exceeds the %d-byte topic width", rule, common.LogTopicLength)
					}
					b := make([]byte, val.Len())
					reflect.Copy(reflect.ValueOf(b), val)
					topic = common.BytesToLeftAlignedLogTopic(b)
				default:
					return nil, fmt.Errorf("unsupported indexed type: %T", rule)
				}
			}
			topics[i] = append(topics[i], topic)
		}
	}
	return topics, nil
}

func genUintType(rule uint64) common.LogTopic {
	var topic common.LogTopic
	binary.BigEndian.PutUint64(topic[common.LogTopicLength-8:], rule)
	return topic
}

func genIntType(rule int64, size uint) common.LogTopic {
	var topic common.LogTopic
	if rule < 0 {
		// if a rule is negative, we need to put it into two's complement,
		// extended to common.LogTopicLength bytes.
		for i := range topic {
			topic[i] = 0xff
		}
	}
	for i := range size {
		topic[common.LogTopicLength-i-1] = byte(rule >> (i * 8))
	}
	return topic
}

// ParseTopics converts the indexed topic fields into actual log field values.
func ParseTopics(out any, fields Arguments, topics []common.LogTopic) error {
	value := reflect.ValueOf(out)
	if value.Kind() != reflect.Ptr || value.IsNil() || value.Elem().Kind() != reflect.Struct {
		return errors.New("abi: ParseTopics output must be a non-nil pointer to a struct")
	}
	value = value.Elem()
	fieldNames, err := topicStructFieldNames(fields, value)
	if err != nil {
		return err
	}
	fieldIndex := 0
	return parseTopicWithSetter(fields, topics,
		func(arg Argument, reconstr any) error {
			fieldName := fieldNames[fieldIndex]
			fieldIndex++
			field := value.FieldByName(fieldName)
			if !field.IsValid() {
				return fmt.Errorf("abi: struct field %s not found", fieldName)
			}
			if !field.CanSet() {
				return fmt.Errorf("abi: struct field %s cannot be set", fieldName)
			}
			reconstructed := reflect.ValueOf(reconstr)
			if !reconstructed.Type().AssignableTo(field.Type()) {
				return fmt.Errorf("abi: cannot unmarshal indexed %s into %s (topic decodes to %s)", arg.Type.String(), field.Type(), reconstructed.Type())
			}
			field.Set(reconstructed)
			return nil
		})
}

// topicStructFieldNames maps indexed ABI arguments to destination fields. Event
// structs can contain both indexed and non-indexed fields, so tags belonging to
// arguments outside this ParseTopics call are ignored. Matching tags by
// occurrence also supports multiple unnamed indexed arguments.
func topicStructFieldNames(fields Arguments, value reflect.Value) ([]string, error) {
	fieldNames := make([]string, len(fields))
	usedArgs := make([]bool, len(fields))
	usedStruct := make(map[string]bool)
	typ := value.Type()

	// Prefer explicit original ABI names, including an empty name.
	for i := range typ.NumField() {
		structField := typ.Field(i)
		if structField.PkgPath != "" { // unexported
			continue
		}
		tagName, ok := structField.Tag.Lookup("abi")
		if !ok {
			continue
		}
		for argIndex, arg := range fields {
			if !usedArgs[argIndex] && arg.Name == tagName {
				fieldNames[argIndex] = structField.Name
				usedArgs[argIndex] = true
				usedStruct[structField.Name] = true
				break
			}
		}
	}

	// Preserve the historical camel-case mapping for fields without tags.
	for i, arg := range fields {
		if usedArgs[i] {
			continue
		}
		name := ToCamelCase(arg.Name)
		if name == "" {
			return nil, errors.New("abi: purely underscored topic cannot unpack to struct")
		}
		if usedStruct[name] {
			return nil, fmt.Errorf("abi: multiple indexed arguments mapping to the same struct field %s", name)
		}
		field, ok := typ.FieldByName(name)
		if !ok {
			return nil, fmt.Errorf("abi: struct field %s not found", name)
		}
		if tag, tagged := field.Tag.Lookup("abi"); tagged && tag != arg.Name {
			return nil, fmt.Errorf("abi: struct field %s has conflicting abi tag %q for indexed argument %q", name, tag, arg.Name)
		}
		fieldNames[i] = name
		usedStruct[name] = true
	}
	return fieldNames, nil
}

// ParseTopicsIntoMap converts the indexed topic field-value pairs into map key-value pairs.
func ParseTopicsIntoMap(out map[string]any, fields Arguments, topics []common.LogTopic) error {
	if out == nil {
		return errors.New("abi: ParseTopicsIntoMap output map is nil")
	}
	return parseTopicWithSetter(fields, topics,
		func(arg Argument, reconstr any) error {
			out[arg.Name] = reconstr
			return nil
		})
}

// parseTopicWithSetter converts the indexed topic field-value pairs and stores them using the
// provided set function.
//
// Note, dynamic types cannot be reconstructed since they get mapped to Keccak256
// hashes as the topic value!
func parseTopicWithSetter(fields Arguments, topics []common.LogTopic, setter func(Argument, any) error) error {
	// Sanity check that the fields and topics match up
	if len(fields) != len(topics) {
		return errors.New("topic/field count mismatch")
	}
	// Iterate over all the fields and reconstruct them from topics
	for i, arg := range fields {
		if !arg.Indexed {
			return errors.New("non-indexed field in topic reconstruction")
		}
		var reconstr any
		switch arg.Type.T {
		case StringTy, BytesTy, SliceTy, ArrayTy, TupleTy:
			// Indexed composite values cannot be reconstructed from a log. Their
			// Keccak-256 digest occupies the left-aligned bytes32 portion of the
			// 64-byte VM64 topic, which is the common.Hash type emitted by abigen.
			for _, b := range topics[i][common.HashLength:] {
				if b != 0 {
					return fmt.Errorf("abi: improperly encoded indexed %s hash, got %x", arg.Type.String(), topics[i])
				}
			}
			var hash common.Hash
			copy(hash[:], topics[i][:common.HashLength])
			reconstr = hash
		case FunctionTy:
			// Functions are AddressLength+4 bytes and fit right-aligned in the
			// 64-byte topic. Reject topics with non-zero bytes in the leading
			// padding — matches the go-ethereum invariant adapted to QRL
			// addresses.
			fnLen := common.AddressLength + 4
			if fnLen > common.LogTopicLength {
				return errors.New("abi: function type does not fit in a 64-byte topic with 64-byte addresses")
			}
			prefix := topics[i][:common.LogTopicLength-fnLen]
			for _, b := range prefix {
				if b != 0 {
					return fmt.Errorf("abi: improperly encoded function type, got %x", topics[i])
				}
			}
			var tmp [common.AddressLength + 4]byte
			copy(tmp[:], topics[i][common.LogTopicLength-fnLen:])
			reconstr = tmp
		default:
			// Topic is already the width of an ABI slot (64 bytes); decode directly.
			var err error
			reconstr, err = toGoType(0, arg.Type, topics[i][:])
			if err != nil {
				return err
			}
		}
		// Use the setter function to store the value
		if err := setter(arg, reconstr); err != nil {
			return err
		}
	}

	return nil
}
