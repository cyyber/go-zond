// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"fmt"
	"slices"
	"time"

	qrl "github.com/theQRL/go-qrl"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/rpc"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func (suite *liveSuite) assertHistoricalLogs(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()
	fixture := suite.fixture
	block := fixture.receipt.BlockNumber
	query := qrl.FilterQuery{
		FromBlock: block,
		ToBlock:   block,
		Addresses: []common.Address{fixture.address},
		Topics:    [][]common.LogTopic{{fixture.topic}},
	}

	logs, err := suite.client.FilterLogs(ctx, query)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(logs).To(gomega.HaveLen(1))
	gomega.Expect(logs[0].TxHash).To(gomega.Equal(fixture.tx.Hash()))
	gomega.Expect(logs[0].Topics).To(gomega.Equal([]common.LogTopic{fixture.topic}))
	gomega.Expect(logs[0].Data).To(gomega.Equal(fixture.value[:]))

	criteria := map[string]any{
		"fromBlock": hexutil.EncodeBig(block),
		"toBlock":   hexutil.EncodeBig(block),
		"address":   []common.Address{fixture.address},
		"topics":    [][]common.LogTopic{{fixture.topic}},
	}
	var logFilter rpc.ID
	gomega.Expect(raw.CallContext(ctx, &logFilter, "qrl_newFilter", criteria)).To(gomega.Succeed())
	gomega.Expect(logFilter).NotTo(gomega.BeEmpty())

	var filterLogs []types.Log
	gomega.Expect(raw.CallContext(
		ctx,
		&filterLogs,
		"qrl_getFilterLogs",
		logFilter,
	)).To(gomega.Succeed())
	gomega.Expect(filterLogs).To(gomega.HaveLen(1))
	gomega.Expect(filterLogs[0].TxHash).To(gomega.Equal(fixture.tx.Hash()))

	var logChanges []types.Log
	gomega.Expect(raw.CallContext(
		ctx,
		&logChanges,
		"qrl_getFilterChanges",
		logFilter,
	)).To(gomega.Succeed())

	var removed bool
	gomega.Expect(raw.CallContext(
		ctx,
		&removed,
		"qrl_uninstallFilter",
		logFilter,
	)).To(gomega.Succeed())
	gomega.Expect(removed).To(gomega.BeTrue())
}

func (suite *liveSuite) assertBlockFilter(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()

	var blockFilter rpc.ID
	gomega.Expect(raw.CallContext(ctx, &blockFilter, "qrl_newBlockFilter")).To(gomega.Succeed())
	gomega.Expect(blockFilter).NotTo(gomega.BeEmpty())

	tx, err := suite.signTransaction(ctx, &suite.from, nil, nil)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	receipt := suite.submitAndWait(ctx, tx)

	var blockHashes []common.Hash
	gomega.Eventually(func() error {
		if err := raw.CallContext(ctx, &blockHashes, "qrl_getFilterChanges", blockFilter); err != nil {
			return err
		}
		if slices.Contains(blockHashes, receipt.BlockHash) {
			return nil
		}
		return fmt.Errorf("block filter has not returned %s: %v", receipt.BlockHash, blockHashes)
	}).WithTimeout(30 * time.Second).WithPolling(time.Second).Should(gomega.Succeed())

	var removed bool
	gomega.Expect(raw.CallContext(
		ctx,
		&removed,
		"qrl_uninstallFilter",
		blockFilter,
	)).To(gomega.Succeed())
	gomega.Expect(removed).To(gomega.BeTrue())
}

func (suite *liveSuite) assertPendingFilter(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()

	var pendingFilter rpc.ID
	fullTransactions := false
	gomega.Expect(raw.CallContext(
		ctx,
		&pendingFilter,
		"qrl_newPendingTransactionFilter",
		&fullTransactions,
	)).To(gomega.Succeed())
	gomega.Expect(pendingFilter).NotTo(gomega.BeEmpty())

	pendingTx, err := suite.signTransaction(ctx, &suite.from, nil, nil)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(suite.client.SendTransaction(ctx, pendingTx)).To(gomega.Succeed())

	var pendingHashes []common.Hash
	gomega.Eventually(func() error {
		if err := raw.CallContext(
			ctx,
			&pendingHashes,
			"qrl_getFilterChanges",
			pendingFilter,
		); err != nil {
			return err
		}
		if slices.Contains(pendingHashes, pendingTx.Hash()) {
			return nil
		}
		return fmt.Errorf("pending filter has not returned %s: %v", pendingTx.Hash(), pendingHashes)
	}).WithTimeout(30 * time.Second).WithPolling(250 * time.Millisecond).Should(gomega.Succeed())

	var removed bool
	gomega.Expect(raw.CallContext(
		ctx,
		&removed,
		"qrl_uninstallFilter",
		pendingFilter,
	)).To(gomega.Succeed())
	gomega.Expect(removed).To(gomega.BeTrue())
	gomega.Expect(suite.waitReceipt(ctx, pendingTx.Hash()).Status).To(
		gomega.Equal(types.ReceiptStatusSuccessful),
	)
}
