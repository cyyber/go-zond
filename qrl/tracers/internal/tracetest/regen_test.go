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
	"path"
	"path/filepath"
	"strings"
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
// and captures the resulting tracer output under the current 64-byte address /
// 512-bit VM layout.
type fixtureScenario struct {
	path       string // output JSON path relative to testdata/
	targetCode []byte // optional contract bytecode at the recipient address
	txData     []byte // optional calldata
	txValue    *big.Int
	create     bool
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

	const fixtureSeedHex = "01000041f6e321b31e72173f8ff2e292359e1862f24fba42fe6f97efaf641980eff29862f24fba42fe6f97efaf641980eff298"
	senderWallet, err := wallet.RestoreFromSeedHex(fixtureSeedHex)
	if err != nil {
		t.Fatalf("wallet.RestoreFromSeedHex: %v", err)
	}
	sender := senderWallet.GetAddress()
	contractAddr := common.BytesToAddress(common.FromHex(
		"c0decafec0decafec0decafec0decafec0decafec0decafec0decafec0decafec0decafec0decafec0decafec0decafe"))

	scenarios, err := collectFixtureScenarios("testdata")
	if err != nil {
		t.Fatalf("collect fixture scenarios: %v", err)
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
		var to *common.Address
		if !sc.create {
			to = &contractAddr
		}
		alloc := core.GenesisAlloc{
			sender: {Balance: new(big.Int).Mul(big.NewInt(10), big.NewInt(params.Quanta))},
		}
		if !sc.create && len(sc.targetCode) > 0 {
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
			To:        to,
			Value:     sc.txValue,
			Data:      sc.txData,
		})
		signedTx, err := types.SignTx(tx, signer, senderWallet)
		if err != nil {
			t.Fatalf("SignTx(%s): %v", sc.path, err)
		}
		txBytes, err := signedTx.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary(%s): %v", sc.path, err)
		}
		txHex := hexutil.Encode(txBytes)

		genesis := &core.Genesis{
			Config:  params.TestChainConfig,
			Alloc:   alloc,
			BaseFee: big.NewInt(1),
		}

		tracerName, tracerCfg, err := tracerForFixture(sc.path)
		if err != nil {
			t.Fatalf("%s: %v", sc.path, err)
		}
		res := runTraceForFixture(t, sc.path, genesis, ctx, signedTx, signer, tracerName, tracerCfg)

		payload := map[string]any{
			"genesis": genesis,
			"context": ctx,
			"input":   txHex,
		}
		if tracerCfg != nil {
			payload["tracerConfig"] = tracerCfg
		}
		payload["result"] = json.RawMessage(res)

		out, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			t.Fatalf("marshal %s: %v", sc.path, err)
		}
		outPath := filepath.Join("testdata", filepath.FromSlash(sc.path))
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(outPath), err)
		}
		if err := os.WriteFile(outPath, out, 0644); err != nil {
			t.Fatalf("write %s: %v", outPath, err)
		}
		t.Logf("wrote %s (%d bytes)", outPath, len(out))
	}
}

func collectFixtureScenarios(root string) ([]fixtureScenario, error) {
	var scenarios []fixtureScenario
	for _, dir := range []string{
		"call_tracer",
		"call_tracer_withLog",
		"call_tracer_flat",
		"prestate_tracer",
		"prestate_tracer_with_diff_mode",
	} {
		base := filepath.Join(root, dir)
		if err := filepath.WalkDir(base, func(file string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
				return nil
			}
			rel, err := filepath.Rel(root, file)
			if err != nil {
				return err
			}
			scenarios = append(scenarios, scenarioForFixture(filepath.ToSlash(rel)))
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return scenarios, nil
}

func scenarioForFixture(rel string) fixtureScenario {
	name := strings.TrimSuffix(path.Base(rel), ".json")
	sc := fixtureScenario{
		path:    rel,
		txValue: big.NewInt(0),
		targetCode: []byte{
			byte(vm.STOP),
		},
	}
	switch {
	case strings.Contains(name, "transfer"):
		sc.txValue = big.NewInt(1000)
		sc.targetCode = nil
	case strings.Contains(name, "create"):
		sc.create = true
		sc.txData = []byte{
			byte(vm.PUSH1), 0x00,
			byte(vm.PUSH1), 0x00,
			byte(vm.RETURN),
		}
	case strings.Contains(name, "delegatecall"):
		sc.targetCode = []byte{
			byte(vm.PUSH1), 0x00,
			byte(vm.PUSH1), 0x00,
			byte(vm.PUSH1), 0x00,
			byte(vm.PUSH1), 0x00,
			byte(vm.PUSH1), 0xcc,
			byte(vm.GAS),
			byte(vm.DELEGATECALL),
		}
	case strings.Contains(name, "revert") || strings.Contains(name, "throw") || strings.Contains(name, "failed"):
		sc.targetCode = []byte{
			byte(vm.PUSH1), 0x00,
			byte(vm.PUSH1), 0x00,
			byte(vm.REVERT),
		}
	case strings.Contains(name, "log") || strings.Contains(rel, "withLog") || strings.Contains(name, "topic"):
		sc.targetCode = []byte{
			byte(vm.PUSH1), 0x42,
			byte(vm.PUSH1), 0x00,
			byte(vm.PUSH1), 0x00,
			byte(vm.LOG1),
		}
	case strings.Contains(name, "deep") || strings.Contains(name, "inner") || strings.Contains(name, "nested"):
		sc.targetCode = []byte{
			byte(vm.PUSH1), 0x00,
			byte(vm.PUSH1), 0x00,
			byte(vm.PUSH1), 0x00,
			byte(vm.PUSH1), 0x00,
			byte(vm.PUSH1), 0x00,
			byte(vm.PUSH1), 0xbb,
			byte(vm.GAS),
			byte(vm.CALL),
		}
	}
	return sc
}

func tracerForFixture(rel string) (string, json.RawMessage, error) {
	switch {
	case strings.HasPrefix(rel, "call_tracer_withLog/"):
		return "callTracer", json.RawMessage(`{"withLog":true}`), nil
	case strings.HasPrefix(rel, "call_tracer/"):
		return "callTracer", nil, nil
	case strings.HasPrefix(rel, "call_tracer_flat/"):
		return "flatCallTracer", nil, nil
	case strings.HasPrefix(rel, "prestate_tracer_with_diff_mode/"):
		return "prestateTracer", json.RawMessage(`{"diffMode":true}`), nil
	case strings.HasPrefix(rel, "prestate_tracer/"):
		return "prestateTracer", nil, nil
	default:
		return "", nil, fmt.Errorf("unknown tracer fixture directory")
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
