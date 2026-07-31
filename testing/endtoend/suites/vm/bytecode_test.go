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

func callCode(op qrvm.OpCode, target common.Address) []byte {
	code := []byte{
		byte(qrvm.PUSH1), byte(3 * qrvm.WordBytes),
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
	}
	if op == qrvm.CALL {
		code = append(code, byte(qrvm.PUSH1), 0)
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

func createCode(op qrvm.OpCode) ([]byte, []byte) {
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
		code = append(code, byte(qrvm.PUSH1), 1)
	}
	code = append(code,
		byte(qrvm.PUSH1), byte(len(childInit)),
		byte(qrvm.PUSH1), byte(qrvm.WordBytes-len(childInit)),
		byte(qrvm.PUSH1), 0,
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
		byte(qrvm.PUSH2), 1, 0,
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
