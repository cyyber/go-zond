// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package vm

import (
	"github.com/theQRL/go-qrl/common"
	qrvm "github.com/theQRL/go-qrl/core/vm"
)

func push(data []byte) []byte {
	code := []byte{byte(qrvm.PUSH1) + byte(len(data)-1)}
	return append(code, data...)
}

func returnTop() []byte {
	return []byte{
		byte(qrvm.PUSH1), 0,
		byte(qrvm.MSTORE),
		byte(qrvm.PUSH1), byte(qrvm.WordBytes),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.RETURN),
	}
}

func pushCode(width int) ([]byte, []byte) {
	value := make([]byte, width)
	for index := range value {
		value[index] = byte(index + 1)
	}
	code := append(push(value), returnTop()...)
	return code, common.LeftPadBytes(value, qrvm.WordBytes)
}

func dupCode(depth int) []byte {
	code := make([]byte, 0, depth*2+8)
	for value := 1; value <= depth; value++ {
		code = append(code, byte(qrvm.PUSH1), byte(value))
	}
	code = append(code, byte(qrvm.DUP1)+byte(depth-1))
	return append(code, returnTop()...)
}

func swapCode(depth int) []byte {
	code := make([]byte, 0, (depth+1)*2+8)
	for value := 1; value <= depth+1; value++ {
		code = append(code, byte(qrvm.PUSH1), byte(value))
	}
	code = append(code, byte(qrvm.SWAP1)+byte(depth-1))
	return append(code, returnTop()...)
}

func memoryCode(value []byte) []byte {
	return memoryCodeAt(value, 0)
}

func memoryCodeAt(value []byte, offset byte) []byte {
	code := append(push(value),
		byte(qrvm.PUSH1), offset,
		byte(qrvm.MSTORE),
		byte(qrvm.PUSH1), offset,
		byte(qrvm.MLOAD),
	)
	return append(code, returnTop()...)
}

func returnWordCode(value []byte) []byte {
	return append(push(value), returnTop()...)
}

func echoCalldataCode() []byte {
	return []byte{
		byte(qrvm.CALLDATASIZE),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.CALLDATACOPY),
		byte(qrvm.CALLDATASIZE),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.RETURN),
	}
}

func calldataLoadCode(offset byte) []byte {
	return append([]byte{
		byte(qrvm.PUSH1), offset,
		byte(qrvm.CALLDATALOAD),
	}, returnTop()...)
}

func codeCopyCode(data []byte) []byte {
	code := []byte{
		byte(qrvm.PUSH1), byte(len(data)),
		byte(qrvm.PUSH2), 0, 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.CODECOPY),
		byte(qrvm.PUSH1), byte(len(data)),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.RETURN),
	}
	dataOffset := len(code)
	code[3] = byte(dataOffset >> 8)
	code[4] = byte(dataOffset)
	return append(code, data...)
}

func extCodeCopyCode(target common.Address, size byte) []byte {
	code := []byte{
		byte(qrvm.PUSH1), size,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH64),
	}
	code = append(code, target[:]...)
	return append(code,
		byte(qrvm.EXTCODECOPY),
		byte(qrvm.PUSH1), size,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.RETURN),
	)
}

func returnDataCopyCode(target common.Address) []byte {
	code := []byte{
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH64),
	}
	code = append(code, target[:]...)
	return append(code,
		byte(qrvm.GAS),
		byte(qrvm.CALL),
		byte(qrvm.POP),
		byte(qrvm.RETURNDATASIZE),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.RETURNDATACOPY),
		byte(qrvm.RETURNDATASIZE),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.RETURN),
	)
}

func keccakCalldataCode() []byte {
	return []byte{
		byte(qrvm.CALLDATASIZE),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.CALLDATACOPY),
		byte(qrvm.CALLDATASIZE),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.KECCAK256),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.MSTORE),
		byte(qrvm.PUSH1), byte(qrvm.WordBytes),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.RETURN),
	}
}

func staticCallPrecompileCode(address, gas byte) []byte {
	return []byte{
		byte(qrvm.CALLDATASIZE),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.CALLDATACOPY),
		byte(qrvm.PUSH1), 32,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.CALLDATASIZE),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), address,
		byte(qrvm.PUSH1), gas,
		byte(qrvm.STATICCALL),
		byte(qrvm.PUSH1), byte(qrvm.WordBytes),
		byte(qrvm.MSTORE),
		byte(qrvm.PUSH1), byte(2 * qrvm.WordBytes),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.RETURN),
	}
}

func patternedBytes(size int) []byte {
	data := make([]byte, size)
	for index := range data {
		data[index] = byte(index + 1)
	}
	return data
}

func patternedAddress(lastByte byte) common.Address {
	address := common.BytesToAddress(patternedBytes(common.AddressLength))
	address[common.AddressLength-1] = lastByte
	return address
}

func patternedCreate2Salt() [qrvm.WordBytes]byte {
	var salt [qrvm.WordBytes]byte
	copy(salt[:], patternedBytes(len(salt)))
	return salt
}

func operationCode(op qrvm.OpCode, operands ...[]byte) []byte {
	var code []byte
	for _, operand := range operands {
		code = append(code, push(operand)...)
	}
	code = append(code, byte(op))
	return append(code, returnTop()...)
}

func storageRoundTripCode(key, value []byte) []byte {
	code := append(push(value), push(key)...)
	code = append(code, byte(qrvm.SSTORE))
	code = append(code, push(key)...)
	code = append(code, byte(qrvm.SLOAD))
	return append(code, returnTop()...)
}

func addressOpcodeCode(op qrvm.OpCode, address common.Address) []byte {
	code := []byte{byte(qrvm.PUSH64)}
	code = append(code, address[:]...)
	code = append(code, byte(op))
	return append(code, returnTop()...)
}

func opcodeCode(op qrvm.OpCode) []byte {
	return append([]byte{byte(op)}, returnTop()...)
}

func jumpDestinationCode(width int, embedded bool) []byte {
	data := make([]byte, width)
	data[0] = byte(qrvm.JUMPDEST)
	target := 4 + 1 + width
	if embedded {
		target = 5
	}
	code := []byte{
		byte(qrvm.PUSH2), byte(target >> 8), byte(target),
		byte(qrvm.JUMP),
		byte(qrvm.PUSH1) + byte(width-1),
	}
	code = append(code, data...)
	if embedded {
		return append(code, byte(qrvm.STOP))
	}
	code = append(code, byte(qrvm.JUMPDEST), byte(qrvm.PUSH1), 1)
	return append(code, returnTop()...)
}

func revertingStorageCode() []byte {
	return []byte{
		byte(qrvm.PUSH1), 1,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.SSTORE),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.REVERT),
	}
}

func failingCallCode(op qrvm.OpCode, target common.Address) []byte {
	code := []byte{
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
	}
	if op == qrvm.CALL {
		code = append(code, byte(qrvm.PUSH1), 7)
	}
	code = append(code, byte(qrvm.PUSH64))
	code = append(code, target[:]...)
	code = append(code,
		byte(qrvm.GAS),
		byte(op),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.MSTORE),
	)
	switch op {
	case qrvm.CALL:
		code = append(code, byte(qrvm.PUSH64))
		code = append(code, target[:]...)
		code = append(code, byte(qrvm.BALANCE))
	case qrvm.DELEGATECALL:
		code = append(code, byte(qrvm.PUSH1), 0, byte(qrvm.SLOAD))
	default:
		code = append(code, byte(qrvm.RETURNDATASIZE))
	}
	return append(code,
		byte(qrvm.PUSH1), byte(qrvm.WordBytes),
		byte(qrvm.MSTORE),
		byte(qrvm.SELFBALANCE),
		byte(qrvm.PUSH1), byte(2*qrvm.WordBytes),
		byte(qrvm.MSTORE),
		byte(qrvm.PUSH1), byte(3*qrvm.WordBytes),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.RETURN),
	)
}

func failingCreateCode(op qrvm.OpCode, child common.Address, salt [qrvm.WordBytes]byte) ([]byte, []byte) {
	initCode := revertingStorageCode()
	code := append(push(initCode),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.MSTORE),
	)
	if op == qrvm.CREATE2 {
		code = append(code, push(salt[:])...)
	}
	code = append(code,
		byte(qrvm.PUSH1), byte(len(initCode)),
		byte(qrvm.PUSH1), byte(qrvm.WordBytes-len(initCode)),
		byte(qrvm.PUSH1), 7,
		byte(op),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.MSTORE),
		byte(qrvm.PUSH64),
	)
	code = append(code, child[:]...)
	code = append(code,
		byte(qrvm.EXTCODESIZE),
		byte(qrvm.PUSH1), byte(qrvm.WordBytes),
		byte(qrvm.MSTORE),
		byte(qrvm.PUSH64),
	)
	code = append(code, child[:]...)
	return append(code,
		byte(qrvm.BALANCE),
		byte(qrvm.PUSH1), byte(2*qrvm.WordBytes),
		byte(qrvm.MSTORE),
		byte(qrvm.SELFBALANCE),
		byte(qrvm.PUSH1), byte(3*qrvm.WordBytes),
		byte(qrvm.MSTORE),
		byte(qrvm.PUSH2), 1, 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.RETURN),
	), initCode
}

func callCode(op qrvm.OpCode, target common.Address) []byte {
	return callCodeWithValue(op, target, 0)
}

func callCodeWithValue(op qrvm.OpCode, target common.Address, value byte) []byte {
	code := []byte{
		byte(qrvm.PUSH1), byte(3 * qrvm.WordBytes),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
	}
	if op == qrvm.CALL {
		code = append(code, byte(qrvm.PUSH1), value)
	}
	code = append(code, byte(qrvm.PUSH64))
	code = append(code, target[:]...)
	code = append(code,
		byte(qrvm.GAS),
		byte(op),
		byte(qrvm.PUSH1), byte(3*qrvm.WordBytes),
		byte(qrvm.MSTORE),
		byte(qrvm.PUSH2), 1, 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.RETURN),
	)
	return code
}

func callContextCode() []byte {
	return []byte{
		byte(qrvm.ADDRESS),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.MSTORE),
		byte(qrvm.CALLER),
		byte(qrvm.PUSH1), byte(qrvm.WordBytes),
		byte(qrvm.MSTORE),
		byte(qrvm.CALLVALUE),
		byte(qrvm.PUSH1), byte(2 * qrvm.WordBytes),
		byte(qrvm.MSTORE),
		byte(qrvm.PUSH1), byte(3 * qrvm.WordBytes),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.RETURN),
	}
}

func callValueContextCode() []byte {
	return []byte{
		byte(qrvm.CALLVALUE),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.MSTORE),
		byte(qrvm.SELFBALANCE),
		byte(qrvm.PUSH1), byte(qrvm.WordBytes),
		byte(qrvm.MSTORE),
		byte(qrvm.PUSH1), byte(2 * qrvm.WordBytes),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.RETURN),
	}
}

func createCode(op qrvm.OpCode, value byte, salt [qrvm.WordBytes]byte) ([]byte, []byte) {
	childRuntime := returnWordCode([]byte{0x2a})
	childInit := append(push(childRuntime),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.MSTORE),
		byte(qrvm.PUSH1), byte(len(childRuntime)),
		byte(qrvm.PUSH1), byte(qrvm.WordBytes-len(childRuntime)),
		byte(qrvm.RETURN),
	)
	code := append(push(childInit),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.MSTORE),
	)
	if op == qrvm.CREATE2 {
		code = append(code, push(salt[:])...)
	}
	code = append(code,
		byte(qrvm.PUSH1), byte(len(childInit)),
		byte(qrvm.PUSH1), byte(qrvm.WordBytes-len(childInit)),
		byte(qrvm.PUSH1), value,
		byte(op),
		byte(qrvm.DUP1),
		byte(qrvm.PUSH1), byte(qrvm.WordBytes),
		byte(qrvm.MSTORE),
		byte(qrvm.DUP1),
		byte(qrvm.EXTCODESIZE),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.MSTORE),
		byte(qrvm.POP),
		byte(qrvm.PUSH1), byte(qrvm.WordBytes),
		byte(qrvm.MLOAD),
		byte(qrvm.BALANCE),
		byte(qrvm.PUSH2), 1, 0,
		byte(qrvm.MSTORE),
		byte(qrvm.PUSH1), byte(qrvm.WordBytes),
		byte(qrvm.PUSH1), byte(2*qrvm.WordBytes),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), byte(qrvm.WordBytes),
		byte(qrvm.MLOAD),
		byte(qrvm.GAS),
		byte(qrvm.CALL),
		byte(qrvm.PUSH1), byte(3*qrvm.WordBytes),
		byte(qrvm.MSTORE),
		byte(qrvm.PUSH2), 1, 64,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.RETURN),
	)
	return code, childInit
}

func logInitCode(data []byte, topics []common.LogTopic) []byte {
	code := append(push(data),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.MSTORE),
	)
	for index := len(topics) - 1; index >= 0; index-- {
		code = append(code, byte(qrvm.PUSH64))
		code = append(code, topics[index][:]...)
	}
	code = append(code,
		byte(qrvm.PUSH1), byte(qrvm.WordBytes),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.LOG0)+byte(len(topics)),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.MSTORE),
		byte(qrvm.PUSH1), 1,
		byte(qrvm.PUSH1), byte(qrvm.WordBytes-1),
		byte(qrvm.RETURN),
	)
	return code
}
