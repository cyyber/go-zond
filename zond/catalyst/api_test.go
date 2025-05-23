// Copyright 2021 The go-ethereum Authors
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

package catalyst

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/theQRL/go-zond/beacon/engine"
	"github.com/theQRL/go-zond/common"
	"github.com/theQRL/go-zond/common/hexutil"
	beaconConsensus "github.com/theQRL/go-zond/consensus/beacon"
	"github.com/theQRL/go-zond/core"
	"github.com/theQRL/go-zond/core/types"
	"github.com/theQRL/go-zond/crypto"
	"github.com/theQRL/go-zond/crypto/pqcrypto"
	"github.com/theQRL/go-zond/miner"
	"github.com/theQRL/go-zond/node"
	"github.com/theQRL/go-zond/p2p"
	"github.com/theQRL/go-zond/params"
	"github.com/theQRL/go-zond/rpc"
	"github.com/theQRL/go-zond/trie"
	"github.com/theQRL/go-zond/zond"
	"github.com/theQRL/go-zond/zond/downloader"
	"github.com/theQRL/go-zond/zond/zondconfig"
)

var (
	// testKey is a private key to use for funding a tester account.
	testKey, _ = pqcrypto.HexToDilithium("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")

	// testAddr is the Zond address of the tester account.
	testAddr = common.Address(testKey.GetAddress())

	testBalance = big.NewInt(2e18)
)

func generateChain(n int) (*core.Genesis, []*types.Block) {
	config := *params.AllBeaconProtocolChanges
	engine := beaconConsensus.NewFaker()
	genesis := &core.Genesis{
		Config: &config,
		Alloc: core.GenesisAlloc{
			testAddr: {Balance: testBalance},
		},
		ExtraData: []byte("test genesis"),
		Timestamp: 9000,
		BaseFee:   big.NewInt(params.InitialBaseFee),
	}
	testNonce := uint64(0)
	generate := func(i int, g *core.BlockGen) {
		g.OffsetTime(5)
		g.SetExtra([]byte("test"))
		to, _ := common.NewAddressFromString("Z9a9070028361F7AAbeB3f2F2Dc07F82C4a98A02a")
		tx, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
			Nonce:     testNonce,
			To:        &to,
			Value:     big.NewInt(1),
			Gas:       params.TxGas,
			GasFeeCap: big.NewInt(8750000000),
			GasTipCap: big.NewInt(params.GPlanck),
			Data:      nil}), types.LatestSigner(&config), testKey)
		g.AddTx(tx)
		testNonce++
	}
	_, blocks, _ := core.GenerateChainWithGenesis(genesis, engine, n, generate)

	return genesis, blocks
}

func TestEth2AssembleBlock(t *testing.T) {
	genesis, blocks := generateChain(10)
	n, zondservice := startZondService(t, genesis, blocks)
	defer n.Close()

	api := NewConsensusAPI(zondservice)
	signer := types.NewShanghaiSigner(zondservice.BlockChain().Config().ChainID)
	to := blocks[9].Coinbase()
	tx, err := types.SignTx(types.NewTx(&types.DynamicFeeTx{Nonce: uint64(10), To: &to, Value: big.NewInt(1000), Gas: params.TxGas, GasFeeCap: big.NewInt(875000000), Data: nil}), signer, testKey)
	if err != nil {
		t.Fatalf("error signing transaction, err=%v", err)
	}
	zondservice.TxPool().Add([]*types.Transaction{tx}, true, false)
	blockParams := engine.PayloadAttributes{
		Timestamp: blocks[9].Time() + 5,
	}
	// The miner needs to pick up on the txs in the pool, so a few retries might be
	// needed.
	if _, testErr := assembleWithTransactions(api, blocks[9].Hash(), &blockParams, 1); testErr != nil {
		t.Fatal(testErr)
	}
}

// assembleWithTransactions tries to assemble a block, retrying until it has 'want',
// number of transactions in it, or it has retried three times.
func assembleWithTransactions(api *ConsensusAPI, parentHash common.Hash, params *engine.PayloadAttributes, want int) (execData *engine.ExecutableData, err error) {
	for retries := 3; retries > 0; retries-- {
		execData, err = assembleBlock(api, parentHash, params)
		if err != nil {
			return nil, err
		}
		if have, want := len(execData.Transactions), want; have != want {
			err = fmt.Errorf("invalid number of transactions, have %d want %d", have, want)
			continue
		}
		return execData, nil
	}
	return nil, err
}

func TestEth2AssembleBlockWithAnotherBlocksTxs(t *testing.T) {
	genesis, blocks := generateChain(10)
	n, zondservice := startZondService(t, genesis, blocks[:9])
	defer n.Close()

	api := NewConsensusAPI(zondservice)

	// Put the 10th block's tx in the pool and produce a new block
	txs := blocks[9].Transactions()
	api.zond.TxPool().Add(txs, false, true)
	blockParams := engine.PayloadAttributes{
		Timestamp: blocks[8].Time() + 5,
	}
	// The miner needs to pick up on the txs in the pool, so a few retries might be
	// needed.
	if _, err := assembleWithTransactions(api, blocks[8].Hash(), &blockParams, blocks[9].Transactions().Len()); err != nil {
		t.Fatal(err)
	}
}

func TestEth2PrepareAndGetPayload(t *testing.T) {
	genesis, blocks := generateChain(10)
	n, zondservice := startZondService(t, genesis, blocks[:9])
	defer n.Close()

	api := NewConsensusAPI(zondservice)

	// Put the 10th block's tx in the pool and produce a new block
	txs := blocks[9].Transactions()
	zondservice.TxPool().Add(txs, true, false)
	blockParams := engine.PayloadAttributes{
		Timestamp:   blocks[8].Time() + 5,
		Withdrawals: []*types.Withdrawal{},
	}
	fcState := engine.ForkchoiceStateV1{
		HeadBlockHash:      blocks[8].Hash(),
		SafeBlockHash:      common.Hash{},
		FinalizedBlockHash: common.Hash{},
	}
	_, err := api.ForkchoiceUpdatedV2(fcState, &blockParams)
	if err != nil {
		t.Fatalf("error preparing payload, err=%v", err)
	}
	// give the payload some time to be built
	time.Sleep(100 * time.Millisecond)
	payloadID := (&miner.BuildPayloadArgs{
		Parent:       fcState.HeadBlockHash,
		Timestamp:    blockParams.Timestamp,
		FeeRecipient: blockParams.SuggestedFeeRecipient,
		Random:       blockParams.Random,
	}).Id()
	execData, err := api.GetPayloadV2(payloadID)
	if err != nil {
		t.Fatalf("error getting payload, err=%v", err)
	}
	if len(execData.ExecutionPayload.Transactions) != blocks[9].Transactions().Len() {
		t.Fatalf("invalid number of transactions %d != 1", len(execData.ExecutionPayload.Transactions))
	}
	// Test invalid payloadID
	var invPayload engine.PayloadID
	copy(invPayload[:], payloadID[:])
	invPayload[0] = ^invPayload[0]
	_, err = api.GetPayloadV2(invPayload)
	if err == nil {
		t.Fatal("expected error retrieving invalid payload")
	}
}

func checkLogEvents(t *testing.T, logsCh <-chan []*types.Log, rmLogsCh <-chan core.RemovedLogsEvent, wantNew, wantRemoved int) {
	t.Helper()

	if len(logsCh) != wantNew {
		t.Fatalf("wrong number of log events: got %d, want %d", len(logsCh), wantNew)
	}
	if len(rmLogsCh) != wantRemoved {
		t.Fatalf("wrong number of removed log events: got %d, want %d", len(rmLogsCh), wantRemoved)
	}
	// Drain events.
	for i := 0; i < len(logsCh); i++ {
		<-logsCh
	}
	for i := 0; i < len(rmLogsCh); i++ {
		<-rmLogsCh
	}
}

func TestInvalidPayloadTimestamp(t *testing.T) {
	genesis, blocks := generateChain(10)
	n, zondservice := startZondService(t, genesis, blocks)
	defer n.Close()
	var (
		api    = NewConsensusAPI(zondservice)
		parent = zondservice.BlockChain().CurrentBlock()
	)
	tests := []struct {
		time      uint64
		shouldErr bool
	}{
		{0, true},
		{parent.Time, true},
		{parent.Time - 1, true},
		{parent.Time + 1, false},
		{uint64(time.Now().Unix()) + uint64(time.Minute), false},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Timestamp test: %v", i), func(t *testing.T) {
			params := engine.PayloadAttributes{
				Timestamp:             test.time,
				Random:                crypto.Keccak256Hash([]byte{byte(123)}),
				SuggestedFeeRecipient: parent.Coinbase,
				Withdrawals:           []*types.Withdrawal{},
			}
			fcState := engine.ForkchoiceStateV1{
				HeadBlockHash:      parent.Hash(),
				SafeBlockHash:      common.Hash{},
				FinalizedBlockHash: common.Hash{},
			}
			_, err := api.ForkchoiceUpdatedV2(fcState, &params)
			if test.shouldErr && err == nil {
				t.Fatalf("expected error preparing payload with invalid timestamp, err=%v", err)
			} else if !test.shouldErr && err != nil {
				t.Fatalf("error preparing payload with valid timestamp, err=%v", err)
			}
		})
	}
}

func TestEth2NewBlock(t *testing.T) {
	genesis, blocks := generateChain(10)
	n, zondservice := startZondService(t, genesis, blocks)
	defer n.Close()

	var (
		api    = NewConsensusAPI(zondservice)
		parent = blocks[len(blocks)-1]

		// This ZVM code generates a log when the contract is created.
		logCode = common.Hex2Bytes("60606040525b7f24ec1d3ff24c2f6ff210738839dbc339cd45a5294d85c79361016243157aae7b60405180905060405180910390a15b600a8060416000396000f360606040526008565b00")
	)
	// The event channels.
	newLogCh := make(chan []*types.Log, 10)
	rmLogsCh := make(chan core.RemovedLogsEvent, 10)
	zondservice.BlockChain().SubscribeLogsEvent(newLogCh)
	zondservice.BlockChain().SubscribeRemovedLogsEvent(rmLogsCh)

	for i := 0; i < 10; i++ {
		statedb, _ := zondservice.BlockChain().StateAt(parent.Root())
		nonce := statedb.GetNonce(testAddr)
		signer := types.LatestSigner(zondservice.BlockChain().Config())
		tx := types.NewTx(&types.DynamicFeeTx{
			Nonce:     nonce,
			Value:     new(big.Int),
			Gas:       1000000,
			GasFeeCap: big.NewInt(2 * params.InitialBaseFee),
			Data:      logCode,
		})
		signedTx, _ := types.SignTx(tx, signer, testKey)
		zondservice.TxPool().Add([]*types.Transaction{signedTx}, true, false)

		execData, err := assembleWithTransactions(api, parent.Hash(), &engine.PayloadAttributes{
			Timestamp: parent.Time() + 5,
		}, 1)
		if err != nil {
			t.Fatalf("Failed to create the executable data %v", err)
		}
		block, err := engine.ExecutableDataToBlock(*execData)
		if err != nil {
			t.Fatalf("Failed to convert executable data to block %v", err)
		}
		newResp, err := api.NewPayloadV2(*execData)
		switch {
		case err != nil:
			t.Fatalf("Failed to insert block: %v", err)
		case newResp.Status != "VALID":
			t.Fatalf("Failed to insert block: %v", newResp.Status)
		case zondservice.BlockChain().CurrentBlock().Number.Uint64() != block.NumberU64()-1:
			t.Fatalf("Chain head shouldn't be updated")
		}
		checkLogEvents(t, newLogCh, rmLogsCh, 0, 0)
		fcState := engine.ForkchoiceStateV1{
			HeadBlockHash:      block.Hash(),
			SafeBlockHash:      block.Hash(),
			FinalizedBlockHash: block.Hash(),
		}
		if _, err := api.ForkchoiceUpdatedV2(fcState, nil); err != nil {
			t.Fatalf("Failed to insert block: %v", err)
		}
		if have, want := zondservice.BlockChain().CurrentBlock().Number.Uint64(), block.NumberU64(); have != want {
			t.Fatalf("Chain head should be updated, have %d want %d", have, want)
		}
		checkLogEvents(t, newLogCh, rmLogsCh, 1, 0)

		parent = block
	}

	// Introduce fork chain
	var (
		head = zondservice.BlockChain().CurrentBlock().Number.Uint64()
	)
	parent = blocks[len(blocks)-1]
	for i := 0; i < 10; i++ {
		execData, err := assembleBlock(api, parent.Hash(), &engine.PayloadAttributes{
			Timestamp: parent.Time() + 6,
		})
		if err != nil {
			t.Fatalf("Failed to create the executable data %v", err)
		}
		block, err := engine.ExecutableDataToBlock(*execData)
		if err != nil {
			t.Fatalf("Failed to convert executable data to block %v", err)
		}
		newResp, err := api.NewPayloadV2(*execData)
		if err != nil || newResp.Status != "VALID" {
			t.Fatalf("Failed to insert block: %v", err)
		}
		if zondservice.BlockChain().CurrentBlock().Number.Uint64() != head {
			t.Fatalf("Chain head shouldn't be updated")
		}

		fcState := engine.ForkchoiceStateV1{
			HeadBlockHash:      block.Hash(),
			SafeBlockHash:      block.Hash(),
			FinalizedBlockHash: block.Hash(),
		}
		if _, err := api.ForkchoiceUpdatedV2(fcState, nil); err != nil {
			t.Fatalf("Failed to insert block: %v", err)
		}
		if zondservice.BlockChain().CurrentBlock().Number.Uint64() != block.NumberU64() {
			t.Fatalf("Chain head should be updated")
		}
		parent, head = block, block.NumberU64()
	}
}

func TestEth2DeepReorg(t *testing.T) {
	// TODO (MariusVanDerWijden) TestEth2DeepReorg is currently broken, because it tries to reorg
	// before the totalTerminalDifficulty threshold
	/*
		genesis, blocks := generateChain(core.TriesInMemory * 2)
		n, zondservice := startZondService(t, genesis, blocks)
		defer n.Close()

		var (
			api    = NewConsensusAPI(zondservice)
			parent = blocks[len(blocks)-core.TriesInMemory-1]
			head   = zondservice.BlockChain().CurrentBlock().Number.Uint64()
		)
		if zondservice.BlockChain().HasBlockAndState(parent.Hash(), parent.NumberU64()) {
			t.Errorf("Block %d not pruned", parent.NumberU64())
		}
		for i := 0; i < 10; i++ {
			execData, err := api.assembleBlock(AssembleBlockParams{
				ParentHash: parent.Hash(),
				Timestamp:  parent.Time() + 5,
			})
			if err != nil {
				t.Fatalf("Failed to create the executable data %v", err)
			}
			block, err := ExecutableDataToBlock(ethservice.BlockChain().Config(), parent.Header(), *execData)
			if err != nil {
				t.Fatalf("Failed to convert executable data to block %v", err)
			}
			newResp, err := api.ExecutePayload(*execData)
			if err != nil || newResp.Status != "VALID" {
				t.Fatalf("Failed to insert block: %v", err)
			}
			if zondservice.BlockChain().CurrentBlock().Number.Uint64() != head {
				t.Fatalf("Chain head shouldn't be updated")
			}
			if err := api.setHead(block.Hash()); err != nil {
				t.Fatalf("Failed to set head: %v", err)
			}
			if zondservice.BlockChain().CurrentBlock().Number.Uint64() != block.NumberU64() {
				t.Fatalf("Chain head should be updated")
			}
			parent, head = block, block.NumberU64()
		}
	*/
}

// startZondService creates a full node instance for testing.
func startZondService(t *testing.T, genesis *core.Genesis, blocks []*types.Block) (*node.Node, *zond.Zond) {
	t.Helper()

	n, err := node.New(&node.Config{
		P2P: p2p.Config{
			ListenAddr:  "0.0.0.0:0",
			NoDiscovery: true,
			MaxPeers:    25,
		}})
	if err != nil {
		t.Fatal("can't create node:", err)
	}

	mcfg := miner.DefaultConfig
	mcfg.PendingFeeRecipient = testAddr
	zondcfg := &zondconfig.Config{Genesis: genesis, SyncMode: downloader.FullSync, TrieTimeout: time.Minute, TrieDirtyCache: 256, TrieCleanCache: 256, Miner: mcfg}
	zondservice, err := zond.New(n, zondcfg)
	if err != nil {
		t.Fatal("can't create zond service:", err)
	}
	if err := n.Start(); err != nil {
		t.Fatal("can't start node:", err)
	}
	if _, err := zondservice.BlockChain().InsertChain(blocks); err != nil {
		n.Close()
		t.Fatal("can't import test blocks:", err)
	}

	zondservice.SetSynced()
	return n, zondservice
}

func TestFullAPI(t *testing.T) {
	genesis, blocks := generateChain(10)
	n, zondservice := startZondService(t, genesis, blocks)
	defer n.Close()
	var (
		parent = zondservice.BlockChain().CurrentBlock()
		// This ZVM code generates a log when the contract is created.
		logCode = common.Hex2Bytes("60606040525b7f24ec1d3ff24c2f6ff210738839dbc339cd45a5294d85c79361016243157aae7b60405180905060405180910390a15b600a8060416000396000f360606040526008565b00")
	)

	callback := func(parent *types.Header) {
		statedb, _ := zondservice.BlockChain().StateAt(parent.Root)
		nonce := statedb.GetNonce(testAddr)
		signer := types.LatestSigner(zondservice.BlockChain().Config())
		tx := types.NewTx(&types.DynamicFeeTx{
			Nonce: nonce,
			Value: new(big.Int),
			Gas:   1000000,
			Data:  logCode,
		})
		signedTx, _ := types.SignTx(tx, signer, testKey)
		zondservice.TxPool().Add([]*types.Transaction{signedTx}, true, false)
	}

	setupBlocks(t, zondservice, 10, parent, callback, nil)
}

func setupBlocks(t *testing.T, zondservice *zond.Zond, n int, parent *types.Header, callback func(parent *types.Header), withdrawals [][]*types.Withdrawal) []*types.Header {
	api := NewConsensusAPI(zondservice)
	var blocks []*types.Header
	for i := 0; i < n; i++ {
		callback(parent)
		var w []*types.Withdrawal
		if withdrawals != nil {
			w = withdrawals[i]
		}

		payload := getNewPayload(t, api, parent, w)
		execResp, err := api.NewPayloadV2(*payload)
		if err != nil {
			t.Fatalf("can't execute payload: %v", err)
		}
		if execResp.Status != engine.VALID {
			t.Fatalf("invalid status: %v", execResp.Status)
		}
		fcState := engine.ForkchoiceStateV1{
			HeadBlockHash:      payload.BlockHash,
			SafeBlockHash:      payload.ParentHash,
			FinalizedBlockHash: payload.ParentHash,
		}
		if _, err := api.ForkchoiceUpdatedV2(fcState, nil); err != nil {
			t.Fatalf("Failed to insert block: %v", err)
		}
		if zondservice.BlockChain().CurrentBlock().Number.Uint64() != payload.Number {
			t.Fatal("Chain head should be updated")
		}
		if zondservice.BlockChain().CurrentFinalBlock().Number.Uint64() != payload.Number-1 {
			t.Fatal("Finalized block should be updated")
		}
		parent = zondservice.BlockChain().CurrentBlock()
		blocks = append(blocks, parent)
	}
	return blocks
}

/*
TestNewPayloadOnInvalidChain sets up a valid chain and tries to feed blocks
from an invalid chain to test if latestValidHash (LVH) works correctly.

We set up the following chain where P1 ... Pn and P1” are valid while
P1' is invalid.
We expect
(1) The LVH to point to the current inserted payload if it was valid.
(2) The LVH to point to the valid parent on an invalid payload (if the parent is available).
(3) If the parent is unavailable, the LVH should not be set.

	CommonAncestor◄─▲── P1 ◄── P2  ◄─ P3  ◄─ ... ◄─ Pn
	                │
	                └── P1' ◄─ P2' ◄─ P3' ◄─ ... ◄─ Pn'
	                │
	                └── P1''
*/
func TestNewPayloadOnInvalidChain(t *testing.T) {
	genesis, blocks := generateChain(10)
	n, zondservice := startZondService(t, genesis, blocks)
	defer n.Close()

	var (
		api    = NewConsensusAPI(zondservice)
		parent = zondservice.BlockChain().CurrentBlock()
		signer = types.LatestSigner(zondservice.BlockChain().Config())
		// This ZVM code generates a log when the contract is created.
		logCode = common.Hex2Bytes("60606040525b7f24ec1d3ff24c2f6ff210738839dbc339cd45a5294d85c79361016243157aae7b60405180905060405180910390a15b600a8060416000396000f360606040526008565b00")
	)
	for i := 0; i < 10; i++ {
		statedb, _ := zondservice.BlockChain().StateAt(parent.Root)
		tx := types.MustSignNewTx(testKey, signer, &types.DynamicFeeTx{
			Nonce:     statedb.GetNonce(testAddr),
			Value:     new(big.Int),
			Gas:       1000000,
			GasFeeCap: big.NewInt(2 * params.InitialBaseFee),
			GasTipCap: big.NewInt(params.GPlanck),
			Data:      logCode,
		})
		zondservice.TxPool().Add([]*types.Transaction{tx}, false, true)
		var (
			params = engine.PayloadAttributes{
				Timestamp:             parent.Time + 1,
				Random:                crypto.Keccak256Hash([]byte{byte(i)}),
				SuggestedFeeRecipient: parent.Coinbase,
				Withdrawals:           []*types.Withdrawal{},
			}
			fcState = engine.ForkchoiceStateV1{
				HeadBlockHash:      parent.Hash(),
				SafeBlockHash:      common.Hash{},
				FinalizedBlockHash: common.Hash{},
			}
			payload *engine.ExecutionPayloadEnvelope
			resp    engine.ForkChoiceResponse
			err     error
		)
		for i := 0; ; i++ {
			if resp, err = api.ForkchoiceUpdatedV2(fcState, &params); err != nil {
				t.Fatalf("error preparing payload, err=%v", err)
			}
			if resp.PayloadStatus.Status != engine.VALID {
				t.Fatalf("error preparing payload, invalid status: %v", resp.PayloadStatus.Status)
			}
			// give the payload some time to be built
			time.Sleep(50 * time.Millisecond)
			if payload, err = api.GetPayloadV2(*resp.PayloadID); err != nil {
				t.Fatalf("can't get payload: %v", err)
			}
			if len(payload.ExecutionPayload.Transactions) > 0 {
				break
			}
			// No luck this time we need to update the params and try again.
			params.Timestamp = params.Timestamp + 1
			if i > 10 {
				t.Fatalf("payload should not be empty")
			}
		}
		execResp, err := api.NewPayloadV2(*payload.ExecutionPayload)
		if err != nil {
			t.Fatalf("can't execute payload: %v", err)
		}
		if execResp.Status != engine.VALID {
			t.Fatalf("invalid status: %v", execResp.Status)
		}
		fcState = engine.ForkchoiceStateV1{
			HeadBlockHash:      payload.ExecutionPayload.BlockHash,
			SafeBlockHash:      payload.ExecutionPayload.ParentHash,
			FinalizedBlockHash: payload.ExecutionPayload.ParentHash,
		}
		if _, err := api.ForkchoiceUpdatedV2(fcState, nil); err != nil {
			t.Fatalf("Failed to insert block: %v", err)
		}
		if zondservice.BlockChain().CurrentBlock().Number.Uint64() != payload.ExecutionPayload.Number {
			t.Fatalf("Chain head should be updated")
		}
		parent = zondservice.BlockChain().CurrentBlock()
	}
}

func assembleBlock(api *ConsensusAPI, parentHash common.Hash, params *engine.PayloadAttributes) (*engine.ExecutableData, error) {
	args := &miner.BuildPayloadArgs{
		Parent:       parentHash,
		Timestamp:    params.Timestamp,
		FeeRecipient: params.SuggestedFeeRecipient,
		Random:       params.Random,
		Withdrawals:  params.Withdrawals,
	}
	payload, err := api.zond.Miner().BuildPayload(args)
	if err != nil {
		return nil, err
	}
	return payload.ResolveFull().ExecutionPayload, nil
}

func TestEmptyBlocks(t *testing.T) {
	genesis, blocks := generateChain(10)
	n, zondservice := startZondService(t, genesis, blocks)
	defer n.Close()

	commonAncestor := zondservice.BlockChain().CurrentBlock()
	api := NewConsensusAPI(zondservice)

	// Setup 10 blocks on the canonical chain
	setupBlocks(t, zondservice, 10, commonAncestor, func(parent *types.Header) {}, nil)

	// (1) check LatestValidHash by sending a normal payload (P1'')
	payload := getNewPayload(t, api, commonAncestor, nil)

	status, err := api.NewPayloadV2(*payload)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != engine.VALID {
		t.Errorf("invalid status: expected VALID got: %v", status.Status)
	}
	if !bytes.Equal(status.LatestValidHash[:], payload.BlockHash[:]) {
		t.Fatalf("invalid LVH: got %v want %v", status.LatestValidHash, payload.BlockHash)
	}

	// (2) Now send P1' which is invalid
	payload = getNewPayload(t, api, commonAncestor, nil)
	payload.GasUsed += 1
	payload = setBlockhash(payload)
	// Now latestValidHash should be the common ancestor
	status, err = api.NewPayloadV2(*payload)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != engine.INVALID {
		t.Errorf("invalid status: expected INVALID got: %v", status.Status)
	}
	// Expect parent header hash on INVALID block on top of PoS block
	expected := commonAncestor.Hash()
	if !bytes.Equal(status.LatestValidHash[:], expected[:]) {
		t.Fatalf("invalid LVH: got %v want %v", status.LatestValidHash, expected)
	}

	// (3) Now send a payload with unknown parent
	payload = getNewPayload(t, api, commonAncestor, nil)
	payload.ParentHash = common.Hash{1}
	payload = setBlockhash(payload)
	// Now latestValidHash should be the common ancestor
	status, err = api.NewPayloadV2(*payload)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != engine.SYNCING {
		t.Errorf("invalid status: expected SYNCING got: %v", status.Status)
	}
	if status.LatestValidHash != nil {
		t.Fatalf("invalid LVH: got %v wanted nil", status.LatestValidHash)
	}
}

func getNewPayload(t *testing.T, api *ConsensusAPI, parent *types.Header, withdrawals []*types.Withdrawal) *engine.ExecutableData {
	params := engine.PayloadAttributes{
		Timestamp:             parent.Time + 1,
		Random:                crypto.Keccak256Hash([]byte{byte(1)}),
		SuggestedFeeRecipient: parent.Coinbase,
		Withdrawals:           withdrawals,
	}

	payload, err := assembleBlock(api, parent.Hash(), &params)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

// setBlockhash sets the blockhash of a modified ExecutableData.
// Can be used to make modified payloads look valid.
func setBlockhash(data *engine.ExecutableData) *engine.ExecutableData {
	txs, _ := decodeTransactions(data.Transactions)
	number := big.NewInt(0)
	number.SetUint64(data.Number)
	header := &types.Header{
		ParentHash:      data.ParentHash,
		Coinbase:        data.FeeRecipient,
		Root:            data.StateRoot,
		TxHash:          types.DeriveSha(types.Transactions(txs), trie.NewStackTrie(nil)),
		ReceiptHash:     data.ReceiptsRoot,
		Bloom:           types.BytesToBloom(data.LogsBloom),
		Number:          number,
		GasLimit:        data.GasLimit,
		GasUsed:         data.GasUsed,
		Time:            data.Timestamp,
		BaseFee:         data.BaseFeePerGas,
		Extra:           data.ExtraData,
		Random:          data.Random,
		WithdrawalsHash: &types.EmptyWithdrawalsHash,
	}
	block := types.NewBlockWithHeader(header).WithBody(types.Body{Transactions: txs})
	data.BlockHash = block.Hash()
	return data
}

func decodeTransactions(enc [][]byte) ([]*types.Transaction, error) {
	var txs = make([]*types.Transaction, len(enc))
	for i, encTx := range enc {
		var tx types.Transaction
		if err := tx.UnmarshalBinary(encTx); err != nil {
			return nil, fmt.Errorf("invalid transaction %d: %v", i, err)
		}
		txs[i] = &tx
	}
	return txs, nil
}

func TestTrickRemoteBlockCache(t *testing.T) {
	// Setup two nodes
	genesis, blocks := generateChain(10)
	nodeA, zondserviceA := startZondService(t, genesis, blocks)
	nodeB, zondserviceB := startZondService(t, genesis, blocks)
	defer nodeA.Close()
	defer nodeB.Close()
	for nodeB.Server().NodeInfo().Ports.Listener == 0 {
		time.Sleep(250 * time.Millisecond)
	}
	nodeA.Server().AddPeer(nodeB.Server().Self())
	nodeB.Server().AddPeer(nodeA.Server().Self())
	apiA := NewConsensusAPI(zondserviceA)
	apiB := NewConsensusAPI(zondserviceB)

	commonAncestor := zondserviceA.BlockChain().CurrentBlock()

	// Setup 10 blocks on the canonical chain
	setupBlocks(t, zondserviceA, 10, commonAncestor, func(parent *types.Header) {}, nil)
	commonAncestor = zondserviceA.BlockChain().CurrentBlock()

	var invalidChain []*engine.ExecutableData
	// create a valid payload (P1)
	//payload1 := getNewPayload(t, apiA, commonAncestor)
	//invalidChain = append(invalidChain, payload1)

	// create an invalid payload2 (P2)
	payload2 := getNewPayload(t, apiA, commonAncestor, nil)
	//payload2.ParentHash = payload1.BlockHash
	payload2.GasUsed += 1
	payload2 = setBlockhash(payload2)
	invalidChain = append(invalidChain, payload2)

	head := payload2
	// create some valid payloads on top
	for i := 0; i < 10; i++ {
		payload := getNewPayload(t, apiA, commonAncestor, nil)
		payload.ParentHash = head.BlockHash
		payload = setBlockhash(payload)
		invalidChain = append(invalidChain, payload)
		head = payload
	}

	// feed the payloads to node B
	for _, payload := range invalidChain {
		status, err := apiB.NewPayloadV2(*payload)
		if err != nil {
			panic(err)
		}
		if status.Status == engine.VALID {
			t.Error("invalid status: VALID on an invalid chain")
		}
		// Now reorg to the head of the invalid chain
		resp, err := apiB.ForkchoiceUpdatedV2(engine.ForkchoiceStateV1{HeadBlockHash: payload.BlockHash, SafeBlockHash: payload.BlockHash, FinalizedBlockHash: payload.ParentHash}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if resp.PayloadStatus.Status == engine.VALID {
			t.Error("invalid status: VALID on an invalid chain")
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestInvalidBloom(t *testing.T) {
	genesis, blocks := generateChain(10)
	n, zondservice := startZondService(t, genesis, blocks)
	defer n.Close()

	commonAncestor := zondservice.BlockChain().CurrentBlock()
	api := NewConsensusAPI(zondservice)

	// Setup 10 blocks on the canonical chain
	setupBlocks(t, zondservice, 10, commonAncestor, func(parent *types.Header) {}, nil)

	// (1) check LatestValidHash by sending a normal payload (P1'')
	payload := getNewPayload(t, api, commonAncestor, nil)
	payload.LogsBloom = append(payload.LogsBloom, byte(1))
	status, err := api.NewPayloadV2(*payload)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != engine.INVALID {
		t.Errorf("invalid status: expected INVALID got: %v", status.Status)
	}
}

// TestSimultaneousNewBlock does several parallel inserts, both as
// newPayLoad and forkchoiceUpdate. This is to test that the api behaves
// well even of the caller is not being 'serial'.
func TestSimultaneousNewBlock(t *testing.T) {
	genesis, blocks := generateChain(10)
	n, zondservice := startZondService(t, genesis, blocks)
	defer n.Close()

	var (
		api    = NewConsensusAPI(zondservice)
		parent = blocks[len(blocks)-1]
	)
	for i := 0; i < 10; i++ {
		execData, err := assembleBlock(api, parent.Hash(), &engine.PayloadAttributes{
			Timestamp: parent.Time() + 5,
		})
		if err != nil {
			t.Fatalf("Failed to create the executable data %v", err)
		}
		// Insert it 10 times in parallel. Should be ignored.
		{
			var (
				wg      sync.WaitGroup
				testErr error
				errMu   sync.Mutex
			)
			wg.Add(10)
			for ii := 0; ii < 10; ii++ {
				go func() {
					defer wg.Done()
					if newResp, err := api.NewPayloadV2(*execData); err != nil {
						errMu.Lock()
						testErr = fmt.Errorf("failed to insert block: %w", err)
						errMu.Unlock()
					} else if newResp.Status != "VALID" {
						errMu.Lock()
						testErr = fmt.Errorf("failed to insert block: %v", newResp.Status)
						errMu.Unlock()
					}
				}()
			}
			wg.Wait()
			if testErr != nil {
				t.Fatal(testErr)
			}
		}
		block, err := engine.ExecutableDataToBlock(*execData)
		if err != nil {
			t.Fatalf("Failed to convert executable data to block %v", err)
		}
		if zondservice.BlockChain().CurrentBlock().Number.Uint64() != block.NumberU64()-1 {
			t.Fatalf("Chain head shouldn't be updated")
		}
		fcState := engine.ForkchoiceStateV1{
			HeadBlockHash:      block.Hash(),
			SafeBlockHash:      block.Hash(),
			FinalizedBlockHash: block.Hash(),
		}
		{
			var (
				wg      sync.WaitGroup
				testErr error
				errMu   sync.Mutex
			)
			wg.Add(10)
			// Do each FCU 10 times
			for ii := 0; ii < 10; ii++ {
				go func() {
					defer wg.Done()
					if _, err := api.ForkchoiceUpdatedV2(fcState, nil); err != nil {
						errMu.Lock()
						testErr = fmt.Errorf("failed to insert block: %w", err)
						errMu.Unlock()
					}
				}()
			}
			wg.Wait()
			if testErr != nil {
				t.Fatal(testErr)
			}
		}
		if have, want := zondservice.BlockChain().CurrentBlock().Number.Uint64(), block.NumberU64(); have != want {
			t.Fatalf("Chain head should be updated, have %d want %d", have, want)
		}
		parent = block
	}
}

// TestWithdrawals creates and verifies two post-Shanghai blocks. The first
// includes zero withdrawals and the second includes two.
func TestWithdrawals(t *testing.T) {
	genesis, blocks := generateChain(10)

	n, zondservice := startZondService(t, genesis, blocks)
	defer n.Close()

	api := NewConsensusAPI(zondservice)

	// 10: Build Shanghai block with no withdrawals.
	parent := zondservice.BlockChain().CurrentHeader()
	blockParams := engine.PayloadAttributes{
		Timestamp:   parent.Time + 5,
		Withdrawals: make([]*types.Withdrawal, 0),
	}
	fcState := engine.ForkchoiceStateV1{
		HeadBlockHash: parent.Hash(),
	}
	resp, err := api.ForkchoiceUpdatedV2(fcState, &blockParams)
	if err != nil {
		t.Fatalf("error preparing payload, err=%v", err)
	}
	if resp.PayloadStatus.Status != engine.VALID {
		t.Fatalf("unexpected status (got: %s, want: %s)", resp.PayloadStatus.Status, engine.VALID)
	}

	// 10: verify state root is the same as parent
	payloadID := (&miner.BuildPayloadArgs{
		Parent:       fcState.HeadBlockHash,
		Timestamp:    blockParams.Timestamp,
		FeeRecipient: blockParams.SuggestedFeeRecipient,
		Random:       blockParams.Random,
		Withdrawals:  blockParams.Withdrawals,
	}).Id()
	execData, err := api.GetPayloadV2(payloadID)
	if err != nil {
		t.Fatalf("error getting payload, err=%v", err)
	}
	if execData.ExecutionPayload.StateRoot != parent.Root {
		t.Fatalf("mismatch state roots (got: %s, want: %s)", execData.ExecutionPayload.StateRoot, blocks[8].Root())
	}

	// 10: verify locally built block
	if status, err := api.NewPayloadV2(*execData.ExecutionPayload); err != nil {
		t.Fatalf("error validating payload: %v", err)
	} else if status.Status != engine.VALID {
		t.Fatalf("invalid payload")
	}

	// 11: build shanghai block with withdrawal
	aa := common.Address{0xaa}
	bb := common.Address{0xbb}
	blockParams = engine.PayloadAttributes{
		Timestamp: execData.ExecutionPayload.Timestamp + 5,
		Withdrawals: []*types.Withdrawal{
			{
				Index:   0,
				Address: aa,
				Amount:  32,
			},
			{
				Index:   1,
				Address: bb,
				Amount:  33,
			},
		},
	}
	fcState.HeadBlockHash = execData.ExecutionPayload.BlockHash
	_, err = api.ForkchoiceUpdatedV2(fcState, &blockParams)
	if err != nil {
		t.Fatalf("error preparing payload, err=%v", err)
	}

	// 11: verify locally build block.
	payloadID = (&miner.BuildPayloadArgs{
		Parent:       fcState.HeadBlockHash,
		Timestamp:    blockParams.Timestamp,
		FeeRecipient: blockParams.SuggestedFeeRecipient,
		Random:       blockParams.Random,
		Withdrawals:  blockParams.Withdrawals,
	}).Id()
	execData, err = api.GetPayloadV2(payloadID)
	if err != nil {
		t.Fatalf("error getting payload, err=%v", err)
	}
	if status, err := api.NewPayloadV2(*execData.ExecutionPayload); err != nil {
		t.Fatalf("error validating payload: %v", err)
	} else if status.Status != engine.VALID {
		t.Fatalf("invalid payload")
	}

	// 11: set block as head.
	fcState.HeadBlockHash = execData.ExecutionPayload.BlockHash
	_, err = api.ForkchoiceUpdatedV2(fcState, nil)
	if err != nil {
		t.Fatalf("error preparing payload, err=%v", err)
	}

	// 11: verify withdrawals were processed.
	db, _, err := zondservice.APIBackend.StateAndHeaderByNumber(context.Background(), rpc.BlockNumber(execData.ExecutionPayload.Number))
	if err != nil {
		t.Fatalf("unable to load db: %v", err)
	}
	for i, w := range blockParams.Withdrawals {
		// w.Amount is in gplanck, balance in planck
		if db.GetBalance(w.Address).Uint64() != w.Amount*params.GPlanck {
			t.Fatalf("failed to process withdrawal %d", i)
		}
	}
}

func TestNilWithdrawals(t *testing.T) {
	genesis, blocks := generateChain(10)

	n, zondservice := startZondService(t, genesis, blocks)
	defer n.Close()

	api := NewConsensusAPI(zondservice)
	parent := zondservice.BlockChain().CurrentHeader()
	aa := common.Address{0xaa}

	type test struct {
		blockParams engine.PayloadAttributes
		wantErr     bool
	}
	tests := []test{
		{
			blockParams: engine.PayloadAttributes{
				Timestamp:   parent.Time + 5,
				Withdrawals: nil,
			},
			wantErr: true,
		},
		{
			blockParams: engine.PayloadAttributes{
				Timestamp:   parent.Time + 5,
				Withdrawals: make([]*types.Withdrawal, 0),
			},
			wantErr: false,
		},
		{
			blockParams: engine.PayloadAttributes{
				Timestamp: parent.Time + 5,
				Withdrawals: []*types.Withdrawal{
					{
						Index:   0,
						Address: aa,
						Amount:  32,
					},
				},
			},
			wantErr: false,
		},
	}

	fcState := engine.ForkchoiceStateV1{
		HeadBlockHash: parent.Hash(),
	}

	for _, test := range tests {
		_, err := api.ForkchoiceUpdatedV2(fcState, &test.blockParams)
		if test.wantErr {
			if err == nil {
				t.Fatal("wanted error on fcuv2 with invalid withdrawals")
			}
			continue
		}
		if err != nil {
			t.Fatalf("error preparing payload, err=%v", err)
		}

		// 11: verify locally build block.
		payloadID := (&miner.BuildPayloadArgs{
			Parent:       fcState.HeadBlockHash,
			Timestamp:    test.blockParams.Timestamp,
			FeeRecipient: test.blockParams.SuggestedFeeRecipient,
			Random:       test.blockParams.Random,
		}).Id()
		execData, err := api.GetPayloadV2(payloadID)
		if err != nil {
			t.Fatalf("error getting payload, err=%v", err)
		}
		if status, err := api.NewPayloadV2(*execData.ExecutionPayload); err != nil {
			t.Fatalf("error validating payload: %v", err.(*engine.EngineAPIError).ErrorData())
		} else if status.Status != engine.VALID {
			t.Fatalf("invalid payload")
		}
	}
}

func setupBodies(t *testing.T) (*node.Node, *zond.Zond, []*types.Block) {
	genesis, blocks := generateChain(10)
	n, zondservice := startZondService(t, genesis, blocks)

	var (
		parent = zondservice.BlockChain().CurrentBlock()
		// This ZVM code generates a log when the contract is created.
		logCode = common.Hex2Bytes("60606040525b7f24ec1d3ff24c2f6ff210738839dbc339cd45a5294d85c79361016243157aae7b60405180905060405180910390a15b600a8060416000396000f360606040526008565b00")
	)

	callback := func(parent *types.Header) {
		statedb, _ := zondservice.BlockChain().StateAt(parent.Root)
		nonce := statedb.GetNonce(testAddr)
		signer := types.LatestSigner(zondservice.BlockChain().Config())
		tx := types.NewTx(&types.DynamicFeeTx{
			Nonce: nonce,
			Value: new(big.Int),
			Gas:   1000000,
			Data:  logCode,
		})
		signedTx, _ := types.SignTx(tx, signer, testKey)
		zondservice.TxPool().Add([]*types.Transaction{signedTx}, false, false)
	}

	withdrawals := make([][]*types.Withdrawal, 10)
	withdrawals[0] = nil // should be filtered out by miner
	withdrawals[1] = make([]*types.Withdrawal, 0)
	for i := 2; i < len(withdrawals); i++ {
		addr := make([]byte, 20)
		crand.Read(addr)
		withdrawals[i] = []*types.Withdrawal{
			{Index: rand.Uint64(), Validator: rand.Uint64(), Amount: rand.Uint64(), Address: common.BytesToAddress(addr)},
		}
	}

	postShanghaiHeaders := setupBlocks(t, zondservice, 10, parent, callback, withdrawals)
	postShanghaiBlocks := make([]*types.Block, len(postShanghaiHeaders))
	for i, header := range postShanghaiHeaders {
		postShanghaiBlocks[i] = zondservice.BlockChain().GetBlock(header.Hash(), header.Number.Uint64())
	}
	return n, zondservice, append(blocks, postShanghaiBlocks...)
}

func allHashes(blocks []*types.Block) []common.Hash {
	var hashes []common.Hash
	for _, b := range blocks {
		hashes = append(hashes, b.Hash())
	}
	return hashes
}
func allBodies(blocks []*types.Block) []*types.Body {
	var bodies []*types.Body
	for _, b := range blocks {
		bodies = append(bodies, b.Body())
	}
	return bodies
}

func TestGetBlockBodiesByHash(t *testing.T) {
	node, zond, blocks := setupBodies(t)
	api := NewConsensusAPI(zond)
	defer node.Close()

	tests := []struct {
		results []*types.Body
		hashes  []common.Hash
	}{
		// First pos block
		{
			results: []*types.Body{zond.BlockChain().GetBlockByNumber(0).Body()},
			hashes:  []common.Hash{zond.BlockChain().GetBlockByNumber(0).Hash()},
		},
		// pos blocks
		{
			results: []*types.Body{blocks[0].Body(), blocks[9].Body(), blocks[14].Body()},
			hashes:  []common.Hash{blocks[0].Hash(), blocks[9].Hash(), blocks[14].Hash()},
		},
		// unavailable block
		{
			results: []*types.Body{blocks[0].Body(), nil, blocks[14].Body()},
			hashes:  []common.Hash{blocks[0].Hash(), {1, 2}, blocks[14].Hash()},
		},
		// same block multiple times
		{
			results: []*types.Body{blocks[0].Body(), nil, blocks[0].Body(), blocks[0].Body()},
			hashes:  []common.Hash{blocks[0].Hash(), {1, 2}, blocks[0].Hash(), blocks[0].Hash()},
		},
		// all blocks
		{
			results: allBodies(blocks),
			hashes:  allHashes(blocks),
		},
	}

	for k, test := range tests {
		result := api.GetPayloadBodiesByHashV1(test.hashes)
		for i, r := range result {
			if !equalBody(test.results[i], r) {
				t.Fatalf("test %v: invalid response: expected %+v got %+v", k, test.results[i], r)
			}
		}
	}
}

func TestGetBlockBodiesByRange(t *testing.T) {
	node, zond, blocks := setupBodies(t)
	api := NewConsensusAPI(zond)
	defer node.Close()

	tests := []struct {
		results []*types.Body
		start   hexutil.Uint64
		count   hexutil.Uint64
	}{
		{
			results: []*types.Body{blocks[9].Body()},
			start:   10,
			count:   1,
		},
		// Genesis
		{
			results: []*types.Body{blocks[0].Body()},
			start:   1,
			count:   1,
		},
		// First post-merge block
		{
			results: []*types.Body{blocks[9].Body()},
			start:   10,
			count:   1,
		},
		// Pre & post merge blocks
		{
			results: []*types.Body{blocks[7].Body(), blocks[8].Body(), blocks[9].Body(), blocks[10].Body()},
			start:   8,
			count:   4,
		},
		// unavailable block
		{
			results: []*types.Body{blocks[18].Body(), blocks[19].Body()},
			start:   19,
			count:   3,
		},
		// unavailable block
		{
			results: []*types.Body{blocks[19].Body()},
			start:   20,
			count:   2,
		},
		{
			results: []*types.Body{blocks[19].Body()},
			start:   20,
			count:   1,
		},
		// whole range unavailable
		{
			results: make([]*types.Body, 0),
			start:   22,
			count:   2,
		},
		// allBlocks
		{
			results: allBodies(blocks),
			start:   1,
			count:   hexutil.Uint64(len(blocks)),
		},
	}

	for k, test := range tests {
		result, err := api.GetPayloadBodiesByRangeV1(test.start, test.count)
		if err != nil {
			t.Fatal(err)
		}
		if len(result) == len(test.results) {
			for i, r := range result {
				if !equalBody(test.results[i], r) {
					t.Fatalf("test %d: invalid response: expected \n%+v\ngot\n%+v", k, test.results[i], r)
				}
			}
		} else {
			t.Fatalf("test %d: invalid length want %v got %v", k, len(test.results), len(result))
		}
	}
}

func TestGetBlockBodiesByRangeInvalidParams(t *testing.T) {
	node, zond, _ := setupBodies(t)
	api := NewConsensusAPI(zond)
	defer node.Close()
	tests := []struct {
		start hexutil.Uint64
		count hexutil.Uint64
		want  *engine.EngineAPIError
	}{
		// Genesis
		{
			start: 0,
			count: 1,
			want:  engine.InvalidParams,
		},
		// No block requested
		{
			start: 1,
			count: 0,
			want:  engine.InvalidParams,
		},
		// Genesis & no block
		{
			start: 0,
			count: 0,
			want:  engine.InvalidParams,
		},
		// More than 1024 blocks
		{
			start: 1,
			count: 1025,
			want:  engine.TooLargeRequest,
		},
	}
	for i, tc := range tests {
		result, err := api.GetPayloadBodiesByRangeV1(tc.start, tc.count)
		if err == nil {
			t.Fatalf("test %d: expected error, got %v", i, result)
		}
		if have, want := err.Error(), tc.want.Error(); have != want {
			t.Fatalf("test %d: have %s, want %s", i, have, want)
		}
	}
}

func equalBody(a *types.Body, b *engine.ExecutionPayloadBodyV1) bool {
	if a == nil && b == nil {
		return true
	} else if a == nil || b == nil {
		return false
	}
	if len(a.Transactions) != len(b.TransactionData) {
		return false
	}
	for i, tx := range a.Transactions {
		data, _ := tx.MarshalBinary()
		if !bytes.Equal(data, b.TransactionData[i]) {
			return false
		}
	}
	return reflect.DeepEqual(a.Withdrawals, b.Withdrawals)
}
