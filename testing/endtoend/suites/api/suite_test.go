// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"testing"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "API live E2E suite")
}

var suite *liveSuite

var _ = ginkgo.BeforeSuite(func(ctx ginkgo.SpecContext) {
	suite = setupLiveSuite(ctx)
})

func liveIt(scenario string, assertion func(*liveSuite, context.Context)) {
	ginkgo.It(scenarioDescriptions[scenario], func(ctx ginkgo.SpecContext) {
		assertion(suite, ctx)
	}, ginkgo.Label(scenario))
}

var _ = ginkgo.Describe(
	"APIs against a live qrl-package network",
	ginkgo.Serial,
	ginkgo.Ordered,
	ginkgo.ContinueOnFailure,
	ginkgo.Label("e2e", "live", "api", "mutates-chain"),
	func() {
		liveIt(scenarioNodeMetadata, (*liveSuite).assertNodeMetadata)
		liveIt(scenarioChainState, (*liveSuite).assertChainState)
		liveIt(scenarioTransactions, (*liveSuite).assertTransactions)
		liveIt(scenarioTxPool, (*liveSuite).assertTxPool)
		liveIt(scenarioRuntimeDiagnostics, (*liveSuite).assertRuntimeDiagnostics)
		liveIt(scenarioHistoricalLogs, (*liveSuite).assertHistoricalLogs)
		liveIt(scenarioBlockFilter, (*liveSuite).assertBlockFilter)
		liveIt(scenarioPendingFilter, (*liveSuite).assertPendingFilter)
		liveIt(scenarioSubscriptionEvents, (*liveSuite).assertSubscriptionEvents)
		liveIt(scenarioSubscriptionRegistration, (*liveSuite).assertSubscriptionRegistration)
		liveIt(scenarioRawDebug, (*liveSuite).assertRawDebug)
		liveIt(scenarioDebugState, (*liveSuite).assertDebugState)
		liveIt(scenarioDebugTracing, (*liveSuite).assertDebugTracing)
		liveIt(scenarioDebugErrorPaths, (*liveSuite).assertDebugErrorPaths)
		liveIt(scenarioGraphQLSchema, (*liveSuite).assertGraphQLSchema)
		liveIt(scenarioGraphQLQueries, (*liveSuite).assertGraphQLQueries)
		liveIt(scenarioGraphQLMutation, (*liveSuite).assertGraphQLMutation)
		liveIt(scenarioGraphQLPending, (*liveSuite).assertGraphQLPending)
	},
)
