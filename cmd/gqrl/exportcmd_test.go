// Copyright 2022 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/theQRL/go-qrl/common"
)

func initGqrl(t *testing.T) string {
	args := []string{"--networkid=42", "init", "./testdata/genesis.json"}
	t.Logf("Initializing gqrl: %v ", args)
	g := runGqrl(t, args...)
	datadir := g.Datadir
	g.WaitExit()
	return datadir
}

// TestExport does a basic test of "gqrl export", exporting the test-genesis.
func TestExport(t *testing.T) {
	outfile := fmt.Sprintf("%v/testExport.out", t.TempDir())
	defer os.Remove(outfile)
	gqrl := runGqrl(t, "--datadir", initGqrl(t), "export", outfile)
	gqrl.WaitExit()
	if have, want := gqrl.ExitStatus(), 0; have != want {
		t.Errorf("exit error, have %d want %d", have, want)
	}
	have, err := os.ReadFile(outfile)
	if err != nil {
		t.Fatal(err)
	}
	// Exported RLP regenerated for 64-byte addresses: the coinbase field
	// is now a 64-byte RLP string (b840 prefix instead of b0), and the
	// state root shifts accordingly because the alloc map key widens.
	want := common.FromHex("0xf90294f9028fa00000000000000000000000000000000000000000000000000000000000000000b84000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a00e6fde8528e5386707f12b77e28ab213ad2f30fa35266cba24c60ed111574838a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421b901000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000080837a12008080b875000000000000000000000000000000000000000000000000000000000000000002f0d131f1f97aef08aec6e3291b957d9efe71050000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000000000000000000000000000000000000000000085174876e800a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421c0c0")
	if !bytes.Equal(have, want) {
		t.Fatalf("wrong content exported\nhave: %x\nwant: %x", have, want)
	}
}
