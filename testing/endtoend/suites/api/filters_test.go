// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	qrl "github.com/theQRL/go-qrl"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/p2p"
	"github.com/theQRL/go-qrl/rpc"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func (suite *liveSuite) assertFilters(ctx context.Context) {
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

	ginkgo.By("reading the fixture log directly and through a polling filter")
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

	ginkgo.By("observing a newly mined block through a block filter")
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
		for _, hash := range blockHashes {
			if hash == receipt.BlockHash {
				return nil
			}
		}
		return fmt.Errorf("block filter has not returned %s: %v", receipt.BlockHash, blockHashes)
	}).WithTimeout(30 * time.Second).WithPolling(time.Second).Should(gomega.Succeed())

	gomega.Expect(raw.CallContext(
		ctx,
		&removed,
		"qrl_uninstallFilter",
		blockFilter,
	)).To(gomega.Succeed())
	gomega.Expect(removed).To(gomega.BeTrue())

	ginkgo.By("observing a submitted transaction through a pending filter")
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
		for _, hash := range pendingHashes {
			if hash == pendingTx.Hash() {
				return nil
			}
		}
		return fmt.Errorf("pending filter has not returned %s: %v", pendingTx.Hash(), pendingHashes)
	}).WithTimeout(30 * time.Second).WithPolling(250 * time.Millisecond).Should(gomega.Succeed())

	gomega.Expect(raw.CallContext(
		ctx,
		&removed,
		"qrl_uninstallFilter",
		pendingFilter,
	)).To(gomega.Succeed())
	gomega.Expect(removed).To(gomega.BeTrue())
	gomega.Expect(suite.submitExistingAndWait(ctx, pendingTx).Status).To(
		gomega.Equal(types.ReceiptStatusSuccessful),
	)
}

func (suite *liveSuite) assertSubscriptions(ctx context.Context) {
	ginkgo.GinkgoHelper()

	ws := suite.wsClient.Client()
	headers := make(chan *types.Header, 8)
	headSub, err := ws.Subscribe(ctx, "qrl", headers, "newHeads")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer headSub.Unsubscribe()

	logs := make(chan types.Log, 8)
	logSub, err := ws.Subscribe(ctx, "qrl", logs, "logs", map[string]any{
		"topics": [][]common.LogTopic{{suite.fixture.topic}},
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer logSub.Unsubscribe()

	pending := make(chan common.Hash, 8)
	pendingSub, err := ws.Subscribe(ctx, "qrl", pending, "newPendingTransactions", false)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer pendingSub.Unsubscribe()

	syncEvents := make(chan json.RawMessage, 1)
	syncSub, err := ws.Subscribe(ctx, "qrl", syncEvents, "syncing")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer syncSub.Unsubscribe()

	peerEvents := make(chan *p2p.PeerEvent, 1)
	peerSub, err := ws.Subscribe(ctx, "admin", peerEvents, "peerEvents")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer peerSub.Unsubscribe()

	ginkgo.By("submitting a log-emitting deployment after subscriptions are active")
	var value = suite.fixture.value
	tx, err := suite.signTransaction(
		ctx,
		nil,
		nil,
		apiContractCode(value, suite.fixture.topic),
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(suite.client.SendTransaction(ctx, tx)).To(gomega.Succeed())
	receipt := suite.submitExistingAndWait(ctx, tx)
	gomega.Expect(receipt.Status).To(gomega.Equal(types.ReceiptStatusSuccessful))

	deadline := time.NewTimer(90 * time.Second)
	defer deadline.Stop()
	var gotHead, gotLog, gotPending bool
	for !gotHead || !gotLog || !gotPending {
		select {
		case header := <-headers:
			if header != nil && header.Number != nil &&
				header.Number.Cmp(receipt.BlockNumber) >= 0 {
				gotHead = true
			}
		case log := <-logs:
			if log.TxHash == tx.Hash() &&
				len(log.Topics) == 1 &&
				log.Topics[0] == suite.fixture.topic {
				gotLog = true
			}
		case hash := <-pending:
			if hash == tx.Hash() {
				gotPending = true
			}
		case err := <-headSub.Err():
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		case err := <-logSub.Err():
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		case err := <-pendingSub.Err():
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		case err := <-syncSub.Err():
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		case err := <-peerSub.Err():
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		case <-deadline.C:
			ginkgo.Fail(fmt.Sprintf(
				"timed out waiting for subscriptions: head=%t log=%t pending=%t",
				gotHead,
				gotLog,
				gotPending,
			))
		case <-ctx.Done():
			ginkgo.Fail("subscription context ended: " + ctx.Err().Error())
		}
	}
}
