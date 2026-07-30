// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.
//
// The go-qrl library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-qrl library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-qrl library. If not, see <http://www.gnu.org/licenses/>.

package abi

import (
	"context"
	"math/big"
	"strings"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/theQRL/go-qrl/common"
	qrlmath "github.com/theQRL/go-qrl/common/math"
)

func (fixture *liveFixture) assertPayableEntrypoints(ctx context.Context) {
	ginkgo.GinkgoHelper()

	// Hyperion:
	// event Received(uint256 amount);
	// receive() external payable { emit Received(msg.value); }
	// Goal: the generated receive entrypoint sends the requested value with
	// empty calldata, and its receipt plus generated parser reproduce that value.
	ginkgo.By("sending value through the generated receive entrypoint")
	amount := big.NewInt(11)
	auth := fixture.transactOpts(ctx)
	auth.Value = amount
	auth.GasLimit = 1_000_000
	tx, err := fixture.binding.Receive(auth)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "generated receive transaction")
	gomega.Expect(tx.To()).NotTo(gomega.BeNil())
	gomega.Expect(*tx.To()).To(gomega.Equal(fixture.address))
	gomega.Expect(tx.Data()).To(gomega.BeEmpty())
	gomega.Expect(tx.Value()).To(gomega.Equal(amount))
	receipt := waitSuccessfulTransaction(ctx, fixture.client, tx)
	gomega.Expect(receipt.Logs).To(gomega.HaveLen(1))
	fixture.assertEvent(ctx, eventExpectation{
		name: "Received",
		log:  *receipt.Logs[0],
		data: []any{amount},
		exactTopics: []common.LogTopic{
			common.HashToLogTopic(fixture.contractABI.Events["Received"].ID),
		},
		want: map[string]any{"amount": amount},
	})
	received, err := fixture.binding.ParseReceived(*receipt.Logs[0])
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(received.Amount).To(gomega.Equal(amount))

	// Hyperion:
	// event FallbackCalled(bytes payload, uint256 amount);
	// fallback() external payable { emit FallbackCalled(msg.data, msg.value); }
	// Goal: the generated fallback entrypoint preserves calldata larger than
	// one VM word and the transferred value in both the transaction and log.
	ginkgo.By("sending calldata and value through the generated fallback entrypoint")
	payload := []byte(strings.Repeat("\x5a", 65))
	amount = big.NewInt(13)
	auth = fixture.transactOpts(ctx)
	auth.Value = amount
	auth.GasLimit = 1_000_000
	tx, err = fixture.binding.Fallback(auth, payload)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "generated fallback transaction")
	gomega.Expect(tx.To()).NotTo(gomega.BeNil())
	gomega.Expect(tx.To()).NotTo(gomega.BeNil())
	gomega.Expect(*tx.To()).To(gomega.Equal(fixture.address))
	gomega.Expect(tx.Data()).To(gomega.Equal(payload))
	gomega.Expect(tx.Value()).To(gomega.Equal(amount))
	receipt = waitSuccessfulTransaction(ctx, fixture.client, tx)
	gomega.Expect(receipt.Logs).To(gomega.HaveLen(1))
	fixture.assertEvent(ctx, eventExpectation{
		name: "FallbackCalled",
		log:  *receipt.Logs[0],
		data: []any{payload, amount},
		exactTopics: []common.LogTopic{
			common.HashToLogTopic(fixture.contractABI.Events["FallbackCalled"].ID),
		},
		want: map[string]any{
			"payload": payload,
			"amount":  amount,
		},
	})
	fallback, err := fixture.binding.ParseFallbackCalled(*receipt.Logs[0])
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(fallback.Payload).To(gomega.Equal(payload))
	gomega.Expect(fallback.Amount).To(gomega.Equal(amount))

	// Hyperion:
	// event Paid(address indexed sender, uint16 indexed marker, uint256 amount);
	// function pay(uint16 marker) external payable {
	//     emit Paid(msg.sender, marker, msg.value);
	// }
	// Goal: a named generated payable method preserves its argument and value,
	// and its indexed event can be decoded and positively or negatively filtered.
	ginkgo.By("sending value through a named generated payable method")
	const marker = uint16(0xbabe)
	amount = big.NewInt(17)
	auth = fixture.transactOpts(ctx)
	auth.Value = amount
	payTx, err := fixture.binding.Pay(auth, marker)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "generated named payable transaction")
	payReceipt := waitSuccessfulTransaction(ctx, fixture.client, payTx)
	gomega.Expect(payReceipt.Logs).To(gomega.HaveLen(1))
	fixture.assertEvent(ctx, eventExpectation{
		name: "Paid",
		log:  *payReceipt.Logs[0],
		data: []any{amount},
		exactTopics: []common.LogTopic{
			common.HashToLogTopic(fixture.contractABI.Events["Paid"].ID),
			common.BytesToLeftAlignedLogTopic(fixture.from[:]),
			common.BytesToRightAlignedLogTopic(qrlmath.U512Bytes(new(big.Int).SetUint64(uint64(marker)))),
		},
		want: map[string]any{
			"sender": fixture.from,
			"marker": marker,
			"amount": amount,
		},
		filter: [][]any{{fixture.from}, {marker}},
		reject: [][]any{{fixture.from}, {marker + 1}},
	})
	paid, err := fixture.binding.ParsePaid(*payReceipt.Logs[0])
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(paid.Sender).To(gomega.Equal(fixture.from))
	gomega.Expect(paid.Marker).To(gomega.Equal(marker))
	gomega.Expect(paid.Amount).To(gomega.Equal(amount))
}
