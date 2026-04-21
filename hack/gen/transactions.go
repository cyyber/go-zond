// Copyright 2026 The go-qrl Authors
// This file is part of go-qrl.
//
// go-qrl is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-qrl is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

package main

import "fmt"

// generateTransactions is a stub. Implement signed tx fixtures (legacy,
// DynamicFee, access-list) using the wallets from addresses.json so
// internal/qrlapi, core/types, and signer tests can drop their t.Skip
// markers.
func generateTransactions(outDir string) error {
	fmt.Println("  transactions TODO: implement signed transaction fixtures")
	return nil
}
