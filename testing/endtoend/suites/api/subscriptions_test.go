// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/internal/qrlapi"
	"github.com/theQRL/go-qrl/p2p"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func (suite *liveSuite) assertSubscriptionEvents(ctx context.Context) {
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

	fullPending := make(chan *qrlapi.RPCTransaction, 8)
	fullPendingSub, err := ws.Subscribe(
		ctx,
		"qrl",
		fullPending,
		"newPendingTransactions",
		true,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer fullPendingSub.Unsubscribe()

	tx, err := suite.signTransaction(
		ctx,
		nil,
		nil,
		apiContractCode(suite.fixture.value, suite.fixture.topic),
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(suite.client.SendTransaction(ctx, tx)).To(gomega.Succeed())
	receipt := suite.waitReceipt(ctx, tx.Hash())
	gomega.Expect(receipt.Status).To(gomega.Equal(types.ReceiptStatusSuccessful))

	deadline := time.NewTimer(90 * time.Second)
	defer deadline.Stop()
	var gotHead, gotLog, gotPending, gotFullPending bool
	for !gotHead || !gotLog || !gotPending || !gotFullPending {
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
		case transaction := <-fullPending:
			if transaction != nil && transaction.Hash == tx.Hash() {
				gomega.Expect(transaction.From).To(gomega.Equal(suite.from))
				gomega.Expect(uint64(transaction.Nonce)).To(gomega.Equal(tx.Nonce()))
				gomega.Expect(transaction.Input).To(gomega.Equal(hexutil.Bytes(tx.Data())))
				gotFullPending = true
			}
		case err := <-headSub.Err():
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		case err := <-logSub.Err():
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		case err := <-pendingSub.Err():
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		case err := <-fullPendingSub.Err():
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		case <-deadline.C:
			ginkgo.Fail(fmt.Sprintf(
				"timed out waiting for subscriptions: head=%t log=%t pending=%t fullPending=%t",
				gotHead,
				gotLog,
				gotPending,
				gotFullPending,
			))
		case <-ctx.Done():
			ginkgo.Fail("subscription context ended: " + ctx.Err().Error())
		}
	}
}

func (suite *liveSuite) assertSubscriptionRegistration(ctx context.Context) {
	ginkgo.GinkgoHelper()

	ws := suite.wsClient.Client()
	syncEvents := make(chan json.RawMessage, 1)
	syncSub, err := ws.Subscribe(ctx, "qrl", syncEvents, "syncing")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer syncSub.Unsubscribe()

	peerEvents := make(chan *p2p.PeerEvent, 1)
	peerSub, err := ws.Subscribe(ctx, "admin", peerEvents, "peerEvents")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer peerSub.Unsubscribe()
}
