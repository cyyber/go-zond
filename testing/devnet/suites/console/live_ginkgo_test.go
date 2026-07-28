// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package console

import (
	"path/filepath"
	"testing"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/theQRL/go-qrl/testing/devnet/internal/build"
	"github.com/theQRL/go-qrl/testing/devnet/internal/network"
)

const liveSpecTimeout = 25 * time.Minute

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Console live E2E suite")
}

var _ = ginkgo.It(
	"exercises the embedded console against a live qrl-package network",
	ginkgo.Serial,
	ginkgo.Label("e2e", "live", "console", "mutates-chain"),
	func(ctx ginkgo.SpecContext) {
		live, err := network.Inspect(ctx)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		workDir := ginkgo.GinkgoT().TempDir()

		gqrlPath := filepath.Join(workDir, "gqrl")
		ginkgo.By("building the current gqrl console")
		gomega.Expect(build.Binary(ctx, "./cmd/gqrl", gqrlPath)).To(gomega.Succeed())

		jsPath := filepath.Join(workDir, "js")
		ginkgo.By("preparing the console scripts and deployment transaction")
		gomega.Expect(prepareWorkspace(ctx, jsPath, live.RPCURL)).To(gomega.Succeed())

		for _, name := range suiteNames {
			ginkgo.By("running " + name)
			gomega.Expect(runSuite(ctx, gqrlPath, jsPath, live.RPCURL, name)).To(gomega.Succeed())
		}
	},
	ginkgo.SpecTimeout(liveSpecTimeout),
)
