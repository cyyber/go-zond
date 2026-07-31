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

	var blockTrace []json.RawMessage
	gomega.Expect(raw.CallContext(
		ctx,
		&blockTrace,
		"debug_traceBlockByNumber",
		blockNumber,
		map[string]any{},
	)).To(gomega.Succeed())
	gomega.Expect(blockTrace).NotTo(gomega.BeEmpty())

	var rawBlock hexutil.Bytes
	gomega.Expect(raw.CallContext(
		ctx,
		&rawBlock,
		"debug_getRawBlock",
		blockSelector,
	)).To(gomega.Succeed())
	gomega.Expect(raw.CallContext(
		ctx,
		&blockTrace,
		"debug_traceBlockByHash",
		fixture.block.Hash(),
		map[string]any{},
	)).To(gomega.Succeed())
	gomega.Expect(raw.CallContext(
		ctx,
		&blockTrace,
		"debug_traceBlock",
		rawBlock,
		map[string]any{},
	)).To(gomega.Succeed())

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
		gomega.Expect(json.Valid(result)).To(gomega.BeTrue())
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
