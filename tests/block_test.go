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

package tests

import (
	"math/rand"
	"testing"

	"github.com/theQRL/go-zond/common"
	"github.com/theQRL/go-zond/core/rawdb"
)

func TestBlockchain(t *testing.T) {
	bt := new(testMatcher)

	// Slow tests
	bt.slow(`.*bcExploitTest/DelegateCallSpam.json`)
	bt.slow(`.*bcExploitTest/ShanghaiLove.json`)
	bt.slow(`.*/bcWalletTest/`)

	bt.walk(t, blockTestDir, func(t *testing.T, name string, test *BlockTest) {
		execBlockTest(t, bt, test)
	})
}

// TestExecutionSpecBlocktests runs the test fixtures from execution-spec-tests.
func TestExecutionSpecBlocktests(t *testing.T) {
	if !common.FileExist(executionSpecBlockchainTestDir) {
		t.Skipf("directory %s does not exist", executionSpecBlockchainTestDir)
	}
	bt := new(testMatcher)

	bt.walk(t, executionSpecBlockchainTestDir, func(t *testing.T, name string, test *BlockTest) {
		execBlockTest(t, bt, test)
	})
}

func execBlockTest(t *testing.T, bt *testMatcher, test *BlockTest) {
	// Define all the different flag combinations we should run the tests with,
	// picking only one for short tests.
	//
	// Note, witness building and self-testing is always enabled as it's a very
	// good test to ensure that we don't break it.
	var (
		snapshotConf = []bool{false, true}
		dbschemeConf = []string{rawdb.HashScheme, rawdb.PathScheme}
	)
	if testing.Short() {
		snapshotConf = []bool{snapshotConf[rand.Int()%2]}
		dbschemeConf = []string{dbschemeConf[rand.Int()%2]}
	}
	for _, snapshot := range snapshotConf {
		for _, dbscheme := range dbschemeConf {
			if err := bt.checkFailure(t, test.Run(snapshot, dbscheme, nil)); err != nil {
				t.Errorf("test with config {snapshotter:%v, scheme:%v} failed: %v", snapshot, dbscheme, err)
				return
			}
		}
	}
}
