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

import "errors"

// generateBlocks is a stub. Implement genesis + block RLP/JSON fixtures so
// core/types, core/rawdb, cmd/qrvm t8ntool testdata can be regenerated
// instead of parked behind t.Skip.
//
// See hack/gen/README.md for the generator contract.
func generateBlocks(outDir string) error {
	return errors.New("blocks fixture generator is not implemented")
}
