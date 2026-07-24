// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.
//
// The go-qrl library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

//go:build local_testnet_runtime

// Command qrysm-deposit-runtime extracts the runtime bytecode from the exact
// Qrysm source revision used to build the local-testnet clients.
package main

import (
	"fmt"

	"github.com/theQRL/qrysm/contracts/deposit"
)

func main() {
	fmt.Println(deposit.DepositContractRuntimeCodeHex())
}
