// Copyright 2020 The go-ethereum Authors
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

package t8ntool

import (
	"fmt"
	"math/big"

	"github.com/theQRL/go-zond/common"
	"github.com/theQRL/go-zond/common/math"
	"github.com/theQRL/go-zond/core"
	"github.com/theQRL/go-zond/core/rawdb"
	"github.com/theQRL/go-zond/core/state"
	"github.com/theQRL/go-zond/core/types"
	"github.com/theQRL/go-zond/core/vm"
	"github.com/theQRL/go-zond/crypto"
	"github.com/theQRL/go-zond/log"
	"github.com/theQRL/go-zond/params"
	"github.com/theQRL/go-zond/qrldb"
	"github.com/theQRL/go-zond/rlp"
	"github.com/theQRL/go-zond/trie"
	"golang.org/x/crypto/sha3"
)

type Prestate struct {
	Env stEnv             `json:"env"`
	Pre core.GenesisAlloc `json:"pre"`
}

// ExecutionResult contains the execution status after running a state test, any
// error that might have occurred and a dump of the final state if requested.
type ExecutionResult struct {
	StateRoot       common.Hash           `json:"stateRoot"`
	TxRoot          common.Hash           `json:"txRoot"`
	ReceiptRoot     common.Hash           `json:"receiptsRoot"`
	LogsHash        common.Hash           `json:"logsHash"`
	Bloom           types.Bloom           `json:"logsBloom"        gencodec:"required"`
	Receipts        types.Receipts        `json:"receipts"`
	Rejected        []*rejectedTx         `json:"rejected,omitempty"`
	GasUsed         math.HexOrDecimal64   `json:"gasUsed"`
	BaseFee         *math.HexOrDecimal256 `json:"currentBaseFee,omitempty"`
	WithdrawalsRoot *common.Hash          `json:"withdrawalsRoot,omitempty"`
}

//go:generate go run github.com/fjl/gencodec -type stEnv -field-override stEnvMarshaling -out gen_stenv.go
type stEnv struct {
	Coinbase        common.Address                      `json:"currentCoinbase"   gencodec:"required"`
	Random          *big.Int                            `json:"currentRandom"`
	ParentBaseFee   *big.Int                            `json:"parentBaseFee,omitempty"`
	ParentGasUsed   uint64                              `json:"parentGasUsed,omitempty"`
	ParentGasLimit  uint64                              `json:"parentGasLimit,omitempty"`
	GasLimit        uint64                              `json:"currentGasLimit"   gencodec:"required"`
	Number          uint64                              `json:"currentNumber"     gencodec:"required"`
	Timestamp       uint64                              `json:"currentTimestamp"  gencodec:"required"`
	ParentTimestamp uint64                              `json:"parentTimestamp,omitempty"`
	BlockHashes     map[math.HexOrDecimal64]common.Hash `json:"blockHashes,omitempty"`
	Withdrawals     []*types.Withdrawal                 `json:"withdrawals,omitempty"`
	BaseFee         *big.Int                            `json:"currentBaseFee,omitempty"`
}

type stEnvMarshaling struct {
	Coinbase        common.Address
	Random          *math.HexOrDecimal256
	ParentBaseFee   *math.HexOrDecimal256
	ParentGasUsed   math.HexOrDecimal64
	ParentGasLimit  math.HexOrDecimal64
	GasLimit        math.HexOrDecimal64
	Number          math.HexOrDecimal64
	Timestamp       math.HexOrDecimal64
	ParentTimestamp math.HexOrDecimal64
	BaseFee         *math.HexOrDecimal256
}

type rejectedTx struct {
	Index int    `json:"index"`
	Err   string `json:"error"`
}

// Apply applies a set of transactions to a pre-state
func (pre *Prestate) Apply(vmConfig vm.Config, chainConfig *params.ChainConfig,
	txs types.Transactions, miningReward int64,
	getTracerFn func(txIndex int, txHash common.Hash) (tracer vm.QRVMLogger, err error)) (*state.StateDB, *ExecutionResult, error) {
	// Capture errors for BLOCKHASH operation, if we haven't been supplied the
	// required blockhashes
	var hashError error
	getHash := func(num uint64) common.Hash {
		if pre.Env.BlockHashes == nil {
			hashError = fmt.Errorf("getHash(%d) invoked, no blockhashes provided", num)
			return common.Hash{}
		}
		h, ok := pre.Env.BlockHashes[math.HexOrDecimal64(num)]
		if !ok {
			hashError = fmt.Errorf("getHash(%d) invoked, blockhash for that block not provided", num)
		}
		return h
	}
	var (
		statedb     = MakePreState(rawdb.NewMemoryDatabase(), pre.Pre)
		signer      = types.MakeSigner(chainConfig)
		gaspool     = new(core.GasPool)
		blockHash   = common.Hash{0x13, 0x37}
		rejectedTxs []*rejectedTx
		includedTxs types.Transactions
		gasUsed     = uint64(0)
		receipts    = make(types.Receipts, 0)
		txIndex     = 0
	)
	gaspool.AddGas(pre.Env.GasLimit)
	vmContext := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Coinbase:    pre.Env.Coinbase,
		BlockNumber: new(big.Int).SetUint64(pre.Env.Number),
		Time:        pre.Env.Timestamp,
		GasLimit:    pre.Env.GasLimit,
		GetHash:     getHash,
	}
	// If currentBaseFee is defined, add it to the vmContext.
	if pre.Env.BaseFee != nil {
		vmContext.BaseFee = new(big.Int).Set(pre.Env.BaseFee)
	}
	// If random is defined, add it to the vmContext.
	if pre.Env.Random != nil {
		rnd := common.BigToHash(pre.Env.Random)
		vmContext.Random = &rnd
	}

	for i, tx := range txs {
		msg, err := core.TransactionToMessage(tx, signer, pre.Env.BaseFee)
		if err != nil {
			log.Warn("rejected tx", "index", i, "hash", tx.Hash(), "error", err)
			rejectedTxs = append(rejectedTxs, &rejectedTx{i, err.Error()})
			continue
		}
		tracer, err := getTracerFn(txIndex, tx.Hash())
		if err != nil {
			return nil, nil, err
		}
		vmConfig.Tracer = tracer
		statedb.SetTxContext(tx.Hash(), txIndex)

		var (
			txContext = core.NewQRVMTxContext(msg)
			snapshot  = statedb.Snapshot()
			prevGas   = gaspool.Gas()
		)
		qrvm := vm.NewQRVM(vmContext, txContext, statedb, chainConfig, vmConfig)

		// (ret []byte, usedGas uint64, failed bool, err error)
		msgResult, err := core.ApplyMessage(qrvm, msg, gaspool)
		if err != nil {
			statedb.RevertToSnapshot(snapshot)
			log.Info("rejected tx", "index", i, "hash", tx.Hash(), "from", msg.From, "error", err)
			rejectedTxs = append(rejectedTxs, &rejectedTx{i, err.Error()})
			gaspool.SetGas(prevGas)
			continue
		}
		includedTxs = append(includedTxs, tx)
		if hashError != nil {
			return nil, nil, NewError(ErrorMissingBlockhash, hashError)
		}
		gasUsed += msgResult.UsedGas

		// Receipt:
		{
			var root []byte
			statedb.Finalise(true)

			// Create a new receipt for the transaction, storing the intermediate root and
			// gas used by the tx.
			receipt := &types.Receipt{Type: tx.Type(), PostState: root, CumulativeGasUsed: gasUsed}
			if msgResult.Failed() {
				receipt.Status = types.ReceiptStatusFailed
			} else {
				receipt.Status = types.ReceiptStatusSuccessful
			}
			receipt.TxHash = tx.Hash()
			receipt.GasUsed = msgResult.UsedGas

			// If the transaction created a contract, store the creation address in the receipt.
			if msg.To == nil {
				receipt.ContractAddress = crypto.CreateAddress(qrvm.TxContext.Origin, tx.Nonce())
			}

			// Set the receipt logs and create the bloom filter.
			receipt.Logs = statedb.GetLogs(tx.Hash(), vmContext.BlockNumber.Uint64(), blockHash)
			receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
			// These three are non-consensus fields:
			//receipt.BlockHash
			//receipt.BlockNumber
			receipt.TransactionIndex = uint(txIndex)
			receipts = append(receipts, receipt)
		}

		txIndex++
	}
	statedb.IntermediateRoot(true)
	// Add mining reward? (-1 means rewards are disabled)
	if miningReward >= 0 {
		// Add mining reward. The mining reward may be `0`, which only makes a difference in the cases
		// where
		// - the coinbase self-destructed, or
		// - there are only 'bad' transactions, which aren't executed. In those cases,
		//   the coinbase gets no txfee, so isn't created, and thus needs to be touched
		var (
			blockReward = big.NewInt(miningReward)
			minerReward = new(big.Int).Set(blockReward)
		)
		statedb.AddBalance(pre.Env.Coinbase, minerReward)
	}
	// Apply withdrawals
	for _, w := range pre.Env.Withdrawals {
		// Amount is in gplanck, turn into planck
		amount := new(big.Int).Mul(new(big.Int).SetUint64(w.Amount), big.NewInt(params.GPlanck))
		statedb.AddBalance(w.Address, amount)
	}
	// Commit block
	root, err := statedb.Commit(vmContext.BlockNumber.Uint64(), true)
	if err != nil {
		return nil, nil, NewError(ErrorQRVM, fmt.Errorf("could not commit state: %v", err))
	}
	execRs := &ExecutionResult{
		StateRoot:   root,
		TxRoot:      types.DeriveSha(includedTxs, trie.NewStackTrie(nil)),
		ReceiptRoot: types.DeriveSha(receipts, trie.NewStackTrie(nil)),
		Bloom:       types.CreateBloom(receipts),
		LogsHash:    rlpHash(statedb.Logs()),
		Receipts:    receipts,
		Rejected:    rejectedTxs,
		GasUsed:     (math.HexOrDecimal64)(gasUsed),
		BaseFee:     (*math.HexOrDecimal256)(vmContext.BaseFee),
	}
	if pre.Env.Withdrawals != nil {
		h := types.DeriveSha(types.Withdrawals(pre.Env.Withdrawals), trie.NewStackTrie(nil))
		execRs.WithdrawalsRoot = &h
	}
	// Re-create statedb instance with new root upon the updated database
	// for accessing latest states.
	statedb, err = state.New(root, statedb.Database(), nil)
	if err != nil {
		return nil, nil, NewError(ErrorQRVM, fmt.Errorf("could not reopen state: %v", err))
	}
	return statedb, execRs, nil
}

func MakePreState(db qrldb.Database, accounts core.GenesisAlloc) *state.StateDB {
	sdb := state.NewDatabaseWithConfig(db, &trie.Config{Preimages: true})
	statedb, _ := state.New(types.EmptyRootHash, sdb, nil)
	for addr, a := range accounts {
		statedb.SetCode(addr, a.Code)
		statedb.SetNonce(addr, a.Nonce)
		statedb.SetBalance(addr, a.Balance)
		for k, v := range a.Storage {
			statedb.SetState(addr, k, v)
		}
	}
	// Commit and re-open to start with a clean state.
	root, _ := statedb.Commit(0, false)
	statedb, _ = state.New(root, sdb, nil)
	return statedb
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
