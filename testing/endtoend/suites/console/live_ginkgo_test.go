// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package console

import (
	"path/filepath"
	"testing"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/theQRL/go-qrl/testing/endtoend/internal/build"
	endtoendlive "github.com/theQRL/go-qrl/testing/endtoend/internal/live"
)

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Console live E2E suite")
}

var _ = ginkgo.Describe(
	"embedded console against a live qrl-package network",
	ginkgo.Serial,
	ginkgo.Ordered,
	ginkgo.ContinueOnFailure,
	ginkgo.Label("e2e", "live", "console", "mutates-chain"),
	func() {
		var (
			gqrlPath string
			jsPath   string
			rpcURL   string
			session  *endtoendlive.Session
		)

		ginkgo.BeforeAll(func(ctx ginkgo.SpecContext) {
			var err error
			session, err = endtoendlive.Open(ctx, true)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			ginkgo.DeferCleanup(session.Close)
			rpcURL = session.Environment.RPCURL

			workDir := ginkgo.GinkgoT().TempDir()

			gqrlPath = filepath.Join(workDir, "gqrl")
			ginkgo.By("building the current gqrl console")
			gomega.Expect(build.Binary(ctx, "./cmd/gqrl", gqrlPath)).To(gomega.Succeed())

			jsPath = filepath.Join(workDir, "js")
			ginkgo.By("preparing the console scripts and deployment transaction")
			gomega.Expect(prepareWorkspace(ctx, jsPath, session)).To(gomega.Succeed())
		})

		for _, scenario := range consoleScenarios {
			ginkgo.It(
				scenario.description,
				func(ctx ginkgo.SpecContext) {
					if scenario.webSocket {
						gomega.Expect(runWatchedSuite(
							ctx,
							session.WebSocketClient.Client(),
							jsPath,
							scenario.name,
						)).To(gomega.Succeed())
						return
					}
					gomega.Expect(
						runSuite(ctx, gqrlPath, jsPath, rpcURL, scenario.name),
					).To(gomega.Succeed())
				},
			)
		}
	},
)
