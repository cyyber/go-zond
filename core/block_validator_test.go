// Copyright 2015 The go-ethereum Authors
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

package core

import (
	"testing"
	"time"

	"github.com/theQRL/go-zond/consensus/beacon"
	"github.com/theQRL/go-zond/core/rawdb"
	"github.com/theQRL/go-zond/core/types"
	"github.com/theQRL/go-zond/core/vm"
	"github.com/theQRL/go-zond/params"
)

// Tests that simple header verification works, for both good and bad blocks.
func TestHeaderVerification(t *testing.T) {
	testHeaderVerification(t, rawdb.HashScheme)
	testHeaderVerification(t, rawdb.PathScheme)
}

func testHeaderVerification(t *testing.T, scheme string) {
	// Create a simple chain to verify
	var (
		gspec        = &Genesis{Config: params.TestChainConfig}
		_, blocks, _ = GenerateChainWithGenesis(gspec, beacon.NewFaker(), 8, nil)
	)
	headers := make([]*types.Header, len(blocks))
	for i, block := range blocks {
		headers[i] = block.Header()
	}
	// Run the header checker for blocks one-by-one, checking for both valid and invalid nonces
	chain, _ := NewBlockChain(rawdb.NewMemoryDatabase(), DefaultCacheConfigWithScheme(scheme), gspec, beacon.NewFaker(), vm.Config{}, nil)
	defer chain.Stop()

	for i := 0; i < len(blocks); i++ {
		for j, valid := range []bool{true, false} {
			var results <-chan error

			if valid {
				engine := beacon.NewFaker()
				_, results = engine.VerifyHeaders(chain, []*types.Header{headers[i]})
			} else {
				engine := beacon.NewFakeFailer(headers[i].Number.Uint64())
				_, results = engine.VerifyHeaders(chain, []*types.Header{headers[i]})
			}
			// Wait for the verification result
			select {
			case result := <-results:
				if (result == nil) != valid {
					t.Errorf("test %d.%d: validity mismatch: have %v, want %v", i, j, result, valid)
				}
			case <-time.After(time.Second):
				t.Fatalf("test %d.%d: verification timeout", i, j)
			}
			// Make sure no more data is returned
			select {
			case result := <-results:
				t.Fatalf("test %d.%d: unexpected result returned: %v", i, j, result)
			case <-time.After(25 * time.Millisecond):
			}
		}
		chain.InsertChain(blocks[i : i+1])
	}
}

func TestCalcGasLimit(t *testing.T) {
	for i, tc := range []struct {
		pGasLimit uint64
		max       uint64
		min       uint64
	}{
		{20000000, 20019530, 19980470},
		{40000000, 40039061, 39960939},
	} {
		// Increase
		if have, want := CalcGasLimit(tc.pGasLimit, 2*tc.pGasLimit), tc.max; have != want {
			t.Errorf("test %d: have %d want <%d", i, have, want)
		}
		// Decrease
		if have, want := CalcGasLimit(tc.pGasLimit, 0), tc.min; have != want {
			t.Errorf("test %d: have %d want >%d", i, have, want)
		}
		// Small decrease
		if have, want := CalcGasLimit(tc.pGasLimit, tc.pGasLimit-1), tc.pGasLimit-1; have != want {
			t.Errorf("test %d: have %d want %d", i, have, want)
		}
		// Small increase
		if have, want := CalcGasLimit(tc.pGasLimit, tc.pGasLimit+1), tc.pGasLimit+1; have != want {
			t.Errorf("test %d: have %d want %d", i, have, want)
		}
		// No change
		if have, want := CalcGasLimit(tc.pGasLimit, tc.pGasLimit), tc.pGasLimit; have != want {
			t.Errorf("test %d: have %d want %d", i, have, want)
		}
	}
}
