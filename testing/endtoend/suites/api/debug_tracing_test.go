// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"encoding/json"
	"math/big"
	"time"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/qrl/tracers/logger"
	"github.com/theQRL/go-qrl/rpc"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func (suite *liveSuite) assertDebugTracing(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()
	fixture := suite.fixture
	blockNumber := rpc.BlockNumber(fixture.block.NumberU64())
	blockSelector := rpc.BlockNumberOrHashWithNumber(blockNumber)

	var trace logger.ExecutionResult
	gomega.Expect(raw.CallContext(
		ctx,
		&trace,
		"debug_traceTransaction",
		fixture.tx.Hash(),
		map[string]any{},
	)).To(gomega.Succeed())
	gomega.Expect(trace.Failed).To(gomega.BeFalse())
	gomega.Expect(trace.StructLogs).NotTo(gomega.BeEmpty())

	wantOpcodes := []string{"PUSH64", "SSTORE", "MSTORE", "LOG1", "RETURN"}
	nextOpcode := 0
	firstPush64 := -1
	for index, step := range trace.StructLogs {
		if firstPush64 == -1 && step.Op == "PUSH64" {
			firstPush64 = index
		}
		if nextOpcode < len(wantOpcodes) && step.Op == wantOpcodes[nextOpcode] {
			nextOpcode++
		}
	}
	gomega.Expect(nextOpcode).To(gomega.Equal(len(wantOpcodes)))
	gomega.Expect(firstPush64).To(gomega.BeNumerically(">=", 0))
	gomega.Expect(firstPush64 + 1).To(gomega.BeNumerically("<", len(trace.StructLogs)))
	stack := trace.StructLogs[firstPush64+1].Stack
	gomega.Expect(stack).NotTo(gomega.BeNil())
	gomega.Expect(*stack).NotTo(gomega.BeEmpty())
	wantStackValue := "0x" + new(big.Int).SetBytes(fixture.value[:]).Text(16)
	gomega.Expect((*stack)[len(*stack)-1]).To(gomega.Equal(wantStackValue))

	var traceByNumber []json.RawMessage
	gomega.Expect(raw.CallContext(
		ctx,
		&traceByNumber,
		"debug_traceBlockByNumber",
		blockNumber,
		map[string]any{},
	)).To(gomega.Succeed())
	expectFixtureBlockTrace(traceByNumber, fixture)

	var rawBlock hexutil.Bytes
	gomega.Expect(raw.CallContext(
		ctx,
		&rawBlock,
		"debug_getRawBlock",
		blockSelector,
	)).To(gomega.Succeed())
	var traceByHash []json.RawMessage
	gomega.Expect(raw.CallContext(
		ctx,
		&traceByHash,
		"debug_traceBlockByHash",
		fixture.block.Hash(),
		map[string]any{},
	)).To(gomega.Succeed())
	var traceByRawBlock []json.RawMessage
	gomega.Expect(raw.CallContext(
		ctx,
		&traceByRawBlock,
		"debug_traceBlock",
		rawBlock,
		map[string]any{},
	)).To(gomega.Succeed())
	expectFixtureBlockTrace(traceByHash, fixture)
	expectFixtureBlockTrace(traceByRawBlock, fixture)
	expectEquivalentBlockTraces(traceByNumber, traceByHash)
	expectEquivalentBlockTraces(traceByNumber, traceByRawBlock)

	gomega.Expect(blockNumber).To(gomega.BeNumerically(">", 0))
	traceResults := make(chan json.RawMessage, 1)
	traceSub, err := suite.wsClient.Client().Subscribe(
		ctx,
		"debug",
		traceResults,
		"traceChain",
		blockNumber-1,
		blockNumber,
		map[string]any{},
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer traceSub.Unsubscribe()

	select {
	case result := <-traceResults:
		var chainTrace struct {
			Block  hexutil.Uint64    `json:"block"`
			Hash   common.Hash       `json:"hash"`
			Traces []json.RawMessage `json:"traces"`
		}
		gomega.Expect(json.Unmarshal(result, &chainTrace)).To(gomega.Succeed())
		gomega.Expect(uint64(chainTrace.Block)).To(gomega.Equal(fixture.block.NumberU64()))
		gomega.Expect(chainTrace.Hash).To(gomega.Equal(fixture.block.Hash()))
		expectFixtureBlockTrace(chainTrace.Traces, fixture)
		expectEquivalentBlockTraces(traceByNumber, chainTrace.Traces)
	case err := <-traceSub.Err():
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	case <-time.After(30 * time.Second):
		ginkgo.Fail("timed out waiting for debug_traceChain result")
	case <-ctx.Done():
		ginkgo.Fail("trace context ended: " + ctx.Err().Error())
	}

	var callTrace struct {
		Failed      bool          `json:"failed"`
		ReturnValue hexutil.Bytes `json:"returnValue"`
		StructLogs  []struct {
			Op string `json:"op"`
		} `json:"structLogs"`
	}
	gomega.Expect(raw.CallContext(ctx, &callTrace, "debug_traceCall", map[string]any{
		"from": suite.from,
		"to":   fixture.address,
		"data": "0x",
	}, blockSelector, map[string]any{})).To(gomega.Succeed())
	gomega.Expect(callTrace.Failed).To(gomega.BeFalse())
	gomega.Expect(callTrace.ReturnValue).To(gomega.Equal(hexutil.Bytes(fixture.value[:])))
	gomega.Expect(callTrace.StructLogs).NotTo(gomega.BeEmpty())
	gomega.Expect(callTrace.StructLogs[len(callTrace.StructLogs)-1].Op).To(gomega.Equal("RETURN"))

	var namedCallTrace struct {
		Type   string          `json:"type"`
		From   common.Address  `json:"from"`
		To     *common.Address `json:"to"`
		Input  hexutil.Bytes   `json:"input"`
		Output hexutil.Bytes   `json:"output"`
		Error  string          `json:"error"`
	}
	gomega.Expect(raw.CallContext(ctx, &namedCallTrace, "debug_traceCall", map[string]any{
		"from": suite.from,
		"to":   fixture.address,
		"data": "0x",
	}, blockSelector, map[string]any{"tracer": "callTracer"})).To(gomega.Succeed())
	gomega.Expect(namedCallTrace.Type).To(gomega.Equal("CALL"))
	gomega.Expect(namedCallTrace.From).To(gomega.Equal(suite.from))
	gomega.Expect(namedCallTrace.To).NotTo(gomega.BeNil())
	gomega.Expect(*namedCallTrace.To).To(gomega.Equal(fixture.address))
	gomega.Expect(namedCallTrace.Input).To(gomega.BeEmpty())
	gomega.Expect(namedCallTrace.Output).To(gomega.Equal(hexutil.Bytes(fixture.value[:])))
	gomega.Expect(namedCallTrace.Error).To(gomega.BeEmpty())

	type prestateAccount struct {
		Balance *hexutil.Big                          `json:"balance"`
		Code    hexutil.Bytes                         `json:"code"`
		Nonce   uint64                                `json:"nonce"`
		Storage map[common.Hash]common.StorageValue64 `json:"storage"`
	}
	var prestate map[common.Address]prestateAccount
	gomega.Expect(raw.CallContext(ctx, &prestate, "debug_traceCall", map[string]any{
		"from": suite.from,
		"to":   fixture.address,
		"data": "0x",
	}, blockSelector, map[string]any{"tracer": "prestateTracer"})).To(gomega.Succeed())
	caller, exists := prestate[suite.from]
	gomega.Expect(exists).To(gomega.BeTrue())
	gomega.Expect(caller.Balance).NotTo(gomega.BeNil())
	contract, exists := prestate[fixture.address]
	gomega.Expect(exists).To(gomega.BeTrue())
	gomega.Expect(contract.Code).NotTo(gomega.BeEmpty())
	gomega.Expect(contract.Storage).To(gomega.HaveKeyWithValue(common.Hash{}, fixture.value))

	var roots []common.Hash
	gomega.Expect(raw.CallContext(
		ctx,
		&roots,
		"debug_intermediateRoots",
		fixture.block.Hash(),
		map[string]any{},
	)).To(gomega.Succeed())
	gomega.Expect(roots).To(gomega.HaveLen(len(fixture.block.Transactions())))
}

func expectFixtureBlockTrace(traces []json.RawMessage, fixture *liveFixture) {
	ginkgo.GinkgoHelper()
	gomega.Expect(traces).To(gomega.HaveLen(len(fixture.block.Transactions())))

	var found bool
	for _, rawTrace := range traces {
		var trace struct {
			TxHash common.Hash            `json:"txHash"`
			Result logger.ExecutionResult `json:"result"`
			Error  string                 `json:"error"`
		}
		gomega.Expect(json.Unmarshal(rawTrace, &trace)).To(gomega.Succeed())
		if trace.TxHash != fixture.tx.Hash() {
			continue
		}
		found = true
		gomega.Expect(trace.Error).To(gomega.BeEmpty())
		gomega.Expect(trace.Result.Failed).To(gomega.BeFalse())
		gomega.Expect(trace.Result.StructLogs).NotTo(gomega.BeEmpty())
	}
	gomega.Expect(found).To(gomega.BeTrue())
}

func expectEquivalentBlockTraces(left, right []json.RawMessage) {
	ginkgo.GinkgoHelper()
	gomega.Expect(right).To(gomega.HaveLen(len(left)))
	for index := range left {
		gomega.Expect(right[index]).To(gomega.MatchJSON(left[index]))
	}
}
