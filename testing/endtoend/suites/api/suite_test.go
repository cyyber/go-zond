// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
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

func liveIt(description string, assertion func(*liveSuite, context.Context)) {
	ginkgo.It(description, func(ctx ginkgo.SpecContext) {
		assertion(suite, ctx)
	}, ginkgo.SpecTimeout(liveSpecTimeout))
}

var _ = ginkgo.Describe(
	"APIs against a live qrl-package network",
	ginkgo.Serial,
	ginkgo.Ordered,
	ginkgo.ContinueOnFailure,
	ginkgo.Label("e2e", "live", "api", "mutates-chain"),
	func() {
		liveIt("covers node and network metadata APIs", (*liveSuite).assertNodeMetadata)
		liveIt("covers chain, account, state, call, proof, and fee APIs", (*liveSuite).assertChainState)
		liveIt("covers managed account signing APIs", (*liveSuite).assertManagedSigning)
		liveIt("covers transaction lookup and raw encoding APIs", (*liveSuite).assertTransactions)
		liveIt("covers transaction-pool inspection APIs", (*liveSuite).assertTxPool)
		liveIt("covers runtime diagnostic APIs", (*liveSuite).assertRuntimeDiagnostics)
		liveIt("covers historical log filtering APIs", (*liveSuite).assertHistoricalLogs)
		liveIt("covers newly mined block filters", (*liveSuite).assertBlockFilter)
		liveIt("covers pending-transaction filters", (*liveSuite).assertPendingFilter)
		liveIt("covers emitted WebSocket events", (*liveSuite).assertSubscriptionEvents)
		liveIt("covers passive WebSocket subscription registration", (*liveSuite).assertSubscriptionRegistration)
		liveIt("covers raw debug chain APIs", (*liveSuite).assertRawDebug)
		liveIt("covers debug state diagnostics", (*liveSuite).assertDebugState)
		liveIt("covers debug tracing APIs", (*liveSuite).assertDebugTracing)
		liveIt("covers registered debug and node-control error paths", (*liveSuite).assertDebugErrorPaths)
		liveIt("covers the GraphQL schema", (*liveSuite).assertGraphQLSchema)
		liveIt("covers GraphQL query fields", (*liveSuite).assertGraphQLQueries)
		liveIt("covers the GraphQL transaction mutation", (*liveSuite).assertGraphQLMutation)
	},
)
