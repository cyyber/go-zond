// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package api

import (
	"bytes"
	"testing"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/core/vm"
	"github.com/theQRL/go-qrl/core/vm/runtime"
)

func TestAPIContractCode(t *testing.T) {
	var value common.StorageValue64
	for index := range value {
		value[index] = byte(index + 1)
	}
	var topic common.LogTopic
	for index := range topic {
		topic[index] = byte(0xff - index)
	}

	runtimeCode, state, err := runtime.Execute(apiContractCode(value, topic), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(runtimeCode) == 0 {
		t.Fatal("constructor returned empty runtime code")
	}

	contract := common.BytesToAddress([]byte("contract"))
	if got := state.GetState(contract, common.Hash{}); got != value {
		t.Fatalf("stored value mismatch: got %x, want %x", got, value)
	}
	logs := state.Logs()
	if len(logs) != 1 {
		t.Fatalf("log count mismatch: got %d, want 1", len(logs))
	}
	if logs[0].Address != contract {
		t.Fatalf("log address mismatch: got %s, want %s", logs[0].Address, contract)
	}
	if len(logs[0].Topics) != 1 || logs[0].Topics[0] != topic {
		t.Fatalf("log topic mismatch: got %x, want %x", logs[0].Topics, topic)
	}
	if !bytes.Equal(logs[0].Data, value[:]) {
		t.Fatalf("log data mismatch: got %x, want %x", logs[0].Data, value)
	}

	state.SetCode(contract, runtimeCode)
	output, _, err := runtime.Call(contract, nil, &runtime.Config{State: state})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output, value[:]) {
		t.Fatalf("runtime output mismatch: got %x, want %x", output, value)
	}
}

func apiContractCode(value common.StorageValue64, topic common.LogTopic) []byte {
	runtimeCode := []byte{
		byte(vm.PUSH1), 0,
		byte(vm.SLOAD),
		byte(vm.PUSH1), 0,
		byte(vm.MSTORE),
		byte(vm.PUSH1), byte(vm.WordBytes),
		byte(vm.PUSH1), 0,
		byte(vm.RETURN),
	}

	code := []byte{byte(vm.PUSH64)}
	code = append(code, value[:]...)
	code = append(code,
		byte(vm.PUSH1), 0,
		byte(vm.SSTORE),
		byte(vm.PUSH64),
	)
	code = append(code, value[:]...)
	code = append(code,
		byte(vm.PUSH1), 0,
		byte(vm.MSTORE),
		byte(vm.PUSH64),
	)
	code = append(code, topic[:]...)
	code = append(code,
		byte(vm.PUSH1), byte(vm.WordBytes),
		byte(vm.PUSH1), 0,
		byte(vm.LOG1),
		byte(vm.PUSH1)+byte(len(runtimeCode))-1,
	)
	code = append(code, runtimeCode...)
	code = append(code,
		byte(vm.PUSH1), 0,
		byte(vm.MSTORE),
		byte(vm.PUSH1), byte(len(runtimeCode)),
		byte(vm.PUSH1), byte(vm.WordBytes-len(runtimeCode)),
		byte(vm.RETURN),
	)
	return code
}
