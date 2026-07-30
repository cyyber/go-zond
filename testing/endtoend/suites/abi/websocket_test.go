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
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/theQRL/go-qrl/accounts/abi/bind"
	"github.com/theQRL/go-qrl/crypto"
)

func (fixture *liveFixture) assertWebSocketWatcher(ctx context.Context) {
	ginkgo.GinkgoHelper()

	// Hyperion:
	// event IndexedScalars(bool indexed flag, bytes5 indexed code, int16 indexed delta);
	// function emitIndexedScalars(bool flag, bytes5 code, int16 delta) external {
	//     emit IndexedScalars(flag, code, delta);
	// }
	// Goal: the generated WebSocket watcher ignores an event that fails one
	// indexed-topic rule, then delivers and decodes the event matching all rules.
	ginkgo.By("watching a filtered event through the generated WebSocket binding")
	auth := fixture.transactOpts(ctx)
	watched, err := NewEventEmitter(fixture.address, fixture.wsClient)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	events := make(chan *EventEmitterIndexedScalars, 1)
	code, delta := [5]byte{1, 2, 3, 4, 5}, int16(-777)
	subscription, err := watched.WatchIndexedScalars(
		&bind.WatchOpts{Context: ctx},
		events,
		[]bool{false},
		[][5]byte{code},
		[]int16{delta},
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(subscription.Unsubscribe)

	nonMatchingTx, err := fixture.binding.EmitIndexedScalars(
		auth,
		true,
		code,
		delta,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	waitSuccessfulTransaction(ctx, fixture.client, nonMatchingTx)
	matchingTx, err := fixture.binding.EmitIndexedScalars(auth, false, code, delta)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	receipt := waitSuccessfulTransaction(ctx, fixture.client, matchingTx)

	select {
	case event, open := <-events:
		gomega.Expect(open).To(gomega.BeTrue(), "generated IndexedScalars event channel closed")
		gomega.Expect(event).NotTo(gomega.BeNil())
		gomega.Expect(event.Raw.TxHash).To(gomega.Equal(receipt.TxHash))
		gomega.Expect(event.Raw.Address).To(gomega.Equal(fixture.address))
		gomega.Expect(event.Flag).To(gomega.BeFalse())
		gomega.Expect(event.Code).To(gomega.Equal(code))
		gomega.Expect(event.Delta).To(gomega.Equal(delta))
	case err, open := <-subscription.Err():
		gomega.Expect(open).To(gomega.BeTrue(), "generated IndexedScalars subscription closed")
		gomega.Expect(err).NotTo(gomega.BeNil(), "generated IndexedScalars subscription closed without an error")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	case <-time.After(90 * time.Second):
		ginkgo.Fail("timed out waiting for generated filtered IndexedScalars event")
	case <-ctx.Done():
		gomega.Expect(ctx.Err()).NotTo(gomega.HaveOccurred())
	}

	// Hyperion:
	// event Dynamic(bytes indexed payload, string indexed note, uint512 amount);
	// function store(...) external { emit Dynamic(payload, note, amount); }
	// Goal: the generated WebSocket watcher hashes the original dynamic filter
	// values, rejects a non-matching event, and decodes the matching hashes.
	ginkgo.By("watching indexed dynamic values through the generated WebSocket binding")
	dynamicEvents := make(chan *EventEmitterDynamic, 1)
	dynamicSubscription, err := watched.WatchDynamic(
		&bind.WatchOpts{Context: ctx},
		dynamicEvents,
		[][]byte{fixture.inputs.payload},
		[]string{fixture.inputs.note},
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(dynamicSubscription.Unsubscribe)

	nonMatchingPayload := []byte("not the watched payload")
	nonMatchingDynamicTx, err := fixture.binding.Store(
		fixture.transactOpts(ctx),
		fixture.inputs.amount,
		fixture.inputs.delta,
		fixture.inputs.tag,
		fixture.from,
		nonMatchingPayload,
		fixture.inputs.note,
		true,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	waitSuccessfulTransaction(ctx, fixture.client, nonMatchingDynamicTx)
	matchingDynamicTx, err := fixture.binding.Store(
		fixture.transactOpts(ctx),
		fixture.inputs.amount,
		fixture.inputs.delta,
		fixture.inputs.tag,
		fixture.from,
		fixture.inputs.payload,
		fixture.inputs.note,
		true,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	dynamicReceipt := waitSuccessfulTransaction(ctx, fixture.client, matchingDynamicTx)
	payloadHash := crypto.Keccak256Hash(fixture.inputs.payload)
	noteHash := crypto.Keccak256Hash([]byte(fixture.inputs.note))

	select {
	case event, open := <-dynamicEvents:
		gomega.Expect(open).To(gomega.BeTrue(), "generated Dynamic event channel closed")
		gomega.Expect(event).NotTo(gomega.BeNil())
		gomega.Expect(event.Raw.TxHash).To(gomega.Equal(dynamicReceipt.TxHash))
		gomega.Expect(event.Raw.Address).To(gomega.Equal(fixture.address))
		gomega.Expect(event.Payload).To(gomega.Equal(payloadHash))
		gomega.Expect(event.Note).To(gomega.Equal(noteHash))
		gomega.Expect(event.Amount).To(gomega.Equal(fixture.inputs.amount))
	case err, open := <-dynamicSubscription.Err():
		gomega.Expect(open).To(gomega.BeTrue(), "generated Dynamic subscription closed")
		gomega.Expect(err).NotTo(gomega.BeNil(), "generated Dynamic subscription closed without an error")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	case <-time.After(90 * time.Second):
		ginkgo.Fail("timed out waiting for generated filtered Dynamic event")
	case <-ctx.Done():
		gomega.Expect(ctx.Err()).NotTo(gomega.HaveOccurred())
	}
}
