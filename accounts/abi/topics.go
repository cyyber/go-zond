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
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/math"
	"github.com/theQRL/go-qrl/crypto"
)

// MakeTopics converts a filter query argument list into a filter topic set.
// Topics are 64-byte values; scalar arguments are right-aligned (big-endian).
func MakeTopics(query ...[]any) ([][]common.LogTopic, error) {
	topics := make([][]common.LogTopic, len(query))
	for i, filter := range query {
		for _, rule := range filter {
			var topic common.LogTopic

			// Try to generate the topic based on simple types
			switch rule := rule.(type) {
			case common.LogTopic:
				copy(topic[:], rule[:])
			case common.Hash:
				copy(topic[common.LogTopicLength-common.HashLength:], rule[:])
			case common.Address:
				copy(topic[common.LogTopicLength-common.AddressLength:], rule[:])
			case *big.Int:
				blob := math.U256Bytes(new(big.Int).Set(rule))
				copy(topic[common.LogTopicLength-len(blob):], blob)
			case bool:
				if rule {
					topic[common.LogTopicLength-1] = 1
				}
			case int8:
				copy(topic[:], genIntType(int64(rule), 1))
			case int16:
				copy(topic[:], genIntType(int64(rule), 2))
			case int32:
				copy(topic[:], genIntType(int64(rule), 4))
			case int64:
				copy(topic[:], genIntType(rule, 8))
			case uint8:
				blob := new(big.Int).SetUint64(uint64(rule)).Bytes()
				copy(topic[common.LogTopicLength-len(blob):], blob)
			case uint16:
				blob := new(big.Int).SetUint64(uint64(rule)).Bytes()
				copy(topic[common.LogTopicLength-len(blob):], blob)
			case uint32:
				blob := new(big.Int).SetUint64(uint64(rule)).Bytes()
				copy(topic[common.LogTopicLength-len(blob):], blob)
			case uint64:
				blob := new(big.Int).SetUint64(rule).Bytes()
				copy(topic[common.LogTopicLength-len(blob):], blob)
			case string:
				hash := crypto.Keccak256Hash([]byte(rule))
				copy(topic[common.LogTopicLength-common.HashLength:], hash[:])
			case []byte:
				hash := crypto.Keccak256Hash(rule)
				copy(topic[common.LogTopicLength-common.HashLength:], hash[:])

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
					reflect.Copy(reflect.ValueOf(topic[:val.Len()]), val)
				default:
					return nil, fmt.Errorf("unsupported indexed type: %T", rule)
				}
			}
			topics[i] = append(topics[i], topic)
		}
	}
	return topics, nil
}

func genIntType(rule int64, size uint) []byte {
	var topic [common.LogTopicLength]byte
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
	return topic[:]
}

// ParseTopics converts the indexed topic fields into actual log field values.
func ParseTopics(out any, fields Arguments, topics []common.LogTopic) error {
	return parseTopicWithSetter(fields, topics,
		func(arg Argument, reconstr any) {
			field := reflect.ValueOf(out).Elem().FieldByName(ToCamelCase(arg.Name))
			field.Set(reflect.ValueOf(reconstr))
		})
}

// ParseTopicsIntoMap converts the indexed topic field-value pairs into map key-value pairs.
func ParseTopicsIntoMap(out map[string]any, fields Arguments, topics []common.LogTopic) error {
	return parseTopicWithSetter(fields, topics,
		func(arg Argument, reconstr any) {
			out[arg.Name] = reconstr
		})
}

// parseTopicWithSetter converts the indexed topic field-value pairs and stores them using the
// provided set function.
//
// Note, dynamic types cannot be reconstructed since they get mapped to Keccak256
// hashes as the topic value!
func parseTopicWithSetter(fields Arguments, topics []common.LogTopic, setter func(Argument, any)) error {
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
		case TupleTy:
			return errors.New("tuple type in topic reconstruction")
		case StringTy, BytesTy, SliceTy, ArrayTy:
			// Array types (including strings and bytes) have their keccak256 hashes stored in the topic — returned verbatim.
			reconstr = topics[i]
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
		setter(arg, reconstr)
	}

	return nil
}
