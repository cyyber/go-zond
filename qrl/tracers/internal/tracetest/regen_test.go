// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.
//
// The go-qrl library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package tracetest

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/common/math"
	"github.com/theQRL/go-qrl/core"
	"github.com/theQRL/go-qrl/core/rawdb"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/core/vm"
	"github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
	"github.com/theQRL/go-qrl/params"
	"github.com/theQRL/go-qrl/qrl/tracers"
	"github.com/theQRL/go-qrl/tests"
)

// fixtureScenario describes a single fresh fixture to regenerate. Each scenario
// runs a minimal signed tx from a funded sender to a pre-seeded target contract,
// and captures the resulting tracer output under the current 48-byte address /
// 512-bit VM layout.
type fixtureScenario struct {
	name       string // output JSON filename (without .json)
	targetCode []byte // optional contract bytecode at the recipient address
	txData     []byte // optional calldata
	txValue    *big.Int
}

// TestRegenerateFixtures regenerates JSON fixtures under testdata/ for each
// tracer exercised by TestCallTracerNative, TestCallTracerNativeWithLog,
// TestFlatCallTracerNative, TestPrestateTracer, and TestPrestateWithDiffModeTracer.
// It is gated by the WRITE_FIXTURES environment variable so it only runs when
// explicitly requested (e.g. WRITE_FIXTURES=1 go test -run TestRegenerateFixtures).
//
// The scenarios are deliberately small and self-contained so they exercise each
// tracer's core code paths without depending on external state snapshots.
func TestRegenerateFixtures(t *testing.T) {
	if os.Getenv("WRITE_FIXTURES") == "" {
		t.Skip("set WRITE_FIXTURES=1 to regenerate tracetest JSON fixtures")
	}

	// Deterministic-looking sender/contract addresses keep the fixtures stable
	// across regenerator invocations (only the ML-DSA-87 signature varies).
	senderWallet, err := wallet.Generate(wallet.ML_DSA_87)
	if err != nil {
		t.Fatalf("wallet.Generate: %v", err)
	}
	sender := senderWallet.GetAddress()
	contractAddr := common.BytesToAddress(common.FromHex(
		"c0decafec0decafec0decafec0decafec0decafec0decafec0decafec0decafec0decafec0decafec0decafec0decafe"))

	scenarios := []fixtureScenario{
		{
			name:    "simple_transfer",
			txValue: big.NewInt(1000),
			// no target code → plain transfer path
		},
		{
			name: "contract_call_stop",
			targetCode: []byte{
				byte(vm.STOP),
			},
			txValue: big.NewInt(0),
		},
		{
			name: "contract_call_revert",
			targetCode: []byte{
				byte(vm.PUSH1), 0x00,
				byte(vm.PUSH1), 0x00,
				byte(vm.REVERT),
			},
			txValue: big.NewInt(0),
		},
	}

	chainID := params.TestChainConfig.ChainID
	signer := types.NewZondSigner(chainID)
	miner := common.BytesToAddress(common.FromHex(
		"c0a1beef00000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0"))

	// The context is shared by all fixtures. Number/timestamp/etc are kept
	// fixed so the regenerated JSON stays stable between runs.
	ctx := &callContext{
		Number:   math.HexOrDecimal64(2),
		Time:     math.HexOrDecimal64(1700000000),
		GasLimit: 30_000_000,
		Miner:    miner,
		BaseFee:  math.NewHexOrDecimal256(1),
	}

	for _, sc := range scenarios {
		to := contractAddr
		alloc := core.GenesisAlloc{
			sender: {Balance: new(big.Int).Mul(big.NewInt(10), big.NewInt(params.Quanta))},
		}
		if len(sc.targetCode) > 0 {
			alloc[contractAddr] = core.GenesisAccount{
				Balance: new(big.Int),
				Code:    sc.targetCode,
			}
		}

		gas := uint64(100_000)
		tx := types.NewTx(&types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     0,
			GasTipCap: big.NewInt(0),
			GasFeeCap: big.NewInt(1),
			Gas:       gas,
			To:        &to,
			Value:     sc.txValue,
			Data:      sc.txData,
		})
		signedTx, err := types.SignTx(tx, signer, senderWallet)
		if err != nil {
			t.Fatalf("SignTx(%s): %v", sc.name, err)
		}
		txBytes, err := signedTx.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary(%s): %v", sc.name, err)
		}
		txHex := hexutil.Encode(txBytes)

		genesis := &core.Genesis{
			Config:  params.TestChainConfig,
			Alloc:   alloc,
			BaseFee: big.NewInt(1),
		}

		type target struct {
			dir          string
			tracerName   string
			tracerCfg    json.RawMessage
			fixtureKind  string // "call" | "flat" | "prestate"
		}
		targets := []target{
			{dir: "call_tracer", tracerName: "callTracer", fixtureKind: "call"},
			{dir: "call_tracer_withLog", tracerName: "callTracer", tracerCfg: json.RawMessage(`{"withLog":true}`), fixtureKind: "call"},
			{dir: "call_tracer_flat", tracerName: "flatCallTracer", fixtureKind: "flat"},
			{dir: "prestate_tracer", tracerName: "prestateTracer", fixtureKind: "prestate"},
			{dir: "prestate_tracer_with_diff_mode", tracerName: "prestateTracer", tracerCfg: json.RawMessage(`{"diffMode":true}`), fixtureKind: "prestate"},
		}

		for _, tgt := range targets {
			res := runTraceForFixture(t, sc.name, genesis, ctx, signedTx, signer, tgt.tracerName, tgt.tracerCfg)

			payload := map[string]any{
				"genesis": genesis,
				"context": ctx,
				"input":   txHex,
			}
			if tgt.tracerCfg != nil {
				payload["tracerConfig"] = tgt.tracerCfg
			}
			payload["result"] = json.RawMessage(res)

			out, err := json.MarshalIndent(payload, "", "  ")
			if err != nil {
				t.Fatalf("marshal %s/%s: %v", tgt.dir, sc.name, err)
			}
			path := filepath.Join("testdata", tgt.dir, sc.name+".json")
			if err := os.WriteFile(path, out, 0644); err != nil {
				t.Fatalf("write %s: %v", path, err)
			}
			t.Logf("wrote %s (%d bytes)", path, len(out))
		}
	}
}

// runTraceForFixture runs the given tracer against a fresh state transition
// driven by the signed tx and returns the tracer's JSON result bytes.
func runTraceForFixture(t *testing.T, scenario string, genesis *core.Genesis, ctx *callContext,
	tx *types.Transaction, signer types.Signer, tracerName string, cfg json.RawMessage) []byte {
	t.Helper()

	origin, err := signer.Sender(tx)
	if err != nil {
		t.Fatalf("recover sender for %s: %v", scenario, err)
	}
	txCtx := vm.TxContext{Origin: origin, GasPrice: tx.GasPrice()}
	blockCtx := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Coinbase:    ctx.Miner,
		BlockNumber: new(big.Int).SetUint64(uint64(ctx.Number)),
		Time:        uint64(ctx.Time),
		GasLimit:    uint64(ctx.GasLimit),
		BaseFee:     genesis.BaseFee,
	}

	triedb, _, statedb := tests.MakePreState(rawdb.NewMemoryDatabase(), genesis.Alloc, false, rawdb.HashScheme)
	defer triedb.Close()

	tracer, err := tracers.DefaultDirectory.New(tracerName, new(tracers.Context), cfg)
	if err != nil {
		t.Fatalf("new tracer %s: %v", tracerName, err)
	}
	qrvm := vm.NewQRVM(blockCtx, txCtx, statedb, genesis.Config, vm.Config{Tracer: tracer})
	msg, err := core.TransactionToMessage(tx, signer, nil)
	if err != nil {
		t.Fatalf("TransactionToMessage(%s): %v", scenario, err)
	}
	if _, err := core.ApplyMessage(qrvm, msg, new(core.GasPool).AddGas(tx.Gas())); err != nil {
		t.Fatalf("ApplyMessage(%s): %v", scenario, err)
	}
	out, err := tracer.GetResult()
	if err != nil {
		t.Fatalf("GetResult(%s): %v", scenario, err)
	}
	// Prove the payload round-trips through json so the written file is
	// canonical and free of garbage whitespace from tracer internals.
	var any any
	if err := json.Unmarshal(out, &any); err != nil {
		t.Fatalf("unmarshal tracer output for %s: %v", scenario, err)
	}
	canon, err := json.Marshal(any)
	if err != nil {
		t.Fatalf("re-marshal tracer output for %s: %v", scenario, err)
	}
	fmt.Fprintf(os.Stdout, "  [regen] %s/%s result=%d bytes\n", tracerName, scenario, len(canon))
	return canon
}
