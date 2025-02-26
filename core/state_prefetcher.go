// Copyright 2019 The go-ethereum Authors
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
	"sync/atomic"

	"github.com/theQRL/go-zond/consensus"
	"github.com/theQRL/go-zond/core/state"
	"github.com/theQRL/go-zond/core/types"
	"github.com/theQRL/go-zond/core/vm"
	"github.com/theQRL/go-zond/params"
)

// statePrefetcher is a basic Prefetcher, which blindly executes a block on top
// of an arbitrary state with the goal of prefetching potentially useful state
// data from disk before the main block processor start executing.
type statePrefetcher struct {
	config *params.ChainConfig // Chain configuration options
	bc     *BlockChain         // Canonical block chain
	engine consensus.Engine    // Consensus engine used for block rewards
}

// newStatePrefetcher initialises a new statePrefetcher.
func newStatePrefetcher(config *params.ChainConfig, bc *BlockChain, engine consensus.Engine) *statePrefetcher {
	return &statePrefetcher{
		config: config,
		bc:     bc,
		engine: engine,
	}
}

// Prefetch processes the state changes according to the Ethereum rules by running
// the transaction messages using the statedb, but any changes are discarded. The
// only goal is to pre-cache transaction signatures and state trie nodes.
func (p *statePrefetcher) Prefetch(block *types.Block, statedb *state.StateDB, cfg vm.Config, interrupt *atomic.Bool) {
	var (
		header       = block.Header()
		gaspool      = new(GasPool).AddGas(block.GasLimit())
		blockContext = NewZVMBlockContext(header, p.bc, nil)
		zvm          = vm.NewZVM(blockContext, vm.TxContext{}, statedb, p.config, cfg)
		signer       = types.MakeSigner(p.config)
	)
	// Iterate over and process the individual transactions
	for i, tx := range block.Transactions() {
		// If block precaching was interrupted, abort
		if interrupt != nil && interrupt.Load() {
			return
		}
		// Convert the transaction into an executable message and pre-cache its sender
		msg, err := TransactionToMessage(tx, signer, header.BaseFee)
		if err != nil {
			return // Also invalid block, bail out
		}
		statedb.SetTxContext(tx.Hash(), i)
		if err := precacheTransaction(msg, gaspool, statedb, zvm); err != nil {
			return // Ugh, something went horribly wrong, bail out
		}
	}
	// pre-load trie nodes for the final root hash
	statedb.IntermediateRoot(true)
}

// precacheTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. The goal is not to execute
// the transaction successfully, rather to warm up touched data slots.
func precacheTransaction(msg *Message, gaspool *GasPool, statedb *state.StateDB, zvm *vm.ZVM) error {
	// Update the zvm with the new transaction context.
	zvm.Reset(NewZVMTxContext(msg), statedb)
	// Add addresses to access list if applicable
	_, err := ApplyMessage(zvm, msg, gaspool)
	return err
}
