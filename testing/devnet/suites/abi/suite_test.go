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

//go:build e2e

package abi

import (
	"testing"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "ABI E2E suite")
}

var suite *liveSuite

var _ = ginkgo.BeforeSuite(func(ctx ginkgo.SpecContext) {
	suite = setupLiveSuite(ctx)
})

var _ = ginkgo.Describe(
	"ABI against a live qrl-package network",
	ginkgo.Serial,
	ginkgo.Ordered,
	ginkgo.ContinueOnFailure,
	ginkgo.Label("e2e", "abi"),
	func() {
		// Each spec creates fresh transaction options. They share one deployment,
		// and no spec depends on state written by an earlier spec.
		var fixture *liveFixture

		ginkgo.BeforeAll(func(ctx ginkgo.SpecContext) {
			fixture = suite.deployEventEmitter(ctx)
		})

		ginkgo.It("round-trips ABI calls through generic decoding, generated bindings, and raw RPC", func(ctx ginkgo.SpecContext) {
			fixture.assertCallRoundTrips(ctx)
		})

		ginkgo.It("decodes custom and standard errors and requires a failed receipt", func(ctx ginkgo.SpecContext) {
			fixture.assertErrors(ctx)
		})

		ginkgo.It("validates a generated transaction, Stored event encoding, and generated and raw filters", func(ctx ginkgo.SpecContext) {
			fixture.assertEventsAndFilters(ctx)
		})

		ginkgo.It("round-trips function values and containers through generic ABI and raw RPC", func(ctx ginkgo.SpecContext) {
			fixture.assertFunctionValues(ctx)
		})

		ginkgo.It("sends value through the generated receive entrypoint", func(ctx ginkgo.SpecContext) {
			fixture.assertPayableEntrypoints(ctx)
		})

		ginkgo.It("observes an indexed generated event filter over WebSocket", func(ctx ginkgo.SpecContext) {
			fixture.assertWebSocketWatcher(ctx)
		})
	},
)
