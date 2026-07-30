// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"testing"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

const liveSpecTimeout = 25 * time.Minute

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "API live E2E suite")
}

var suite *liveSuite

var _ = ginkgo.BeforeSuite(func(ctx ginkgo.SpecContext) {
	suite = setupLiveSuite(ctx)
})

var _ = ginkgo.Describe(
	"APIs against a live qrl-package network",
	ginkgo.Serial,
	ginkgo.Ordered,
	ginkgo.ContinueOnFailure,
	ginkgo.Label("e2e", "live", "api", "mutates-chain"),
	func() {
		ginkgo.It("covers node, chain, state, transaction, and txpool JSON-RPC methods", func(ctx ginkgo.SpecContext) {
			suite.assertRPCSurface(ctx)
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("covers historical log and polling filter methods", func(ctx ginkgo.SpecContext) {
			suite.assertFilters(ctx)
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("covers WebSocket subscriptions", func(ctx ginkgo.SpecContext) {
			suite.assertSubscriptions(ctx)
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("covers read-only debug and tracing methods", func(ctx ginkgo.SpecContext) {
			suite.assertDebugSurface(ctx)
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("covers GraphQL queries, nested fields, and mutation", func(ctx ginkgo.SpecContext) {
			suite.assertGraphQLSurface(ctx)
		}, ginkgo.SpecTimeout(liveSpecTimeout))
	},
)
