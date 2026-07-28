// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package clef

import (
	"path/filepath"
	"testing"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/theQRL/go-qrl/testing/devnet/internal/build"
	"github.com/theQRL/go-qrl/testing/devnet/internal/network"
)

const liveSpecTimeout = 10 * time.Minute

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Clef live E2E suite")
}

var _ = ginkgo.It(
	"signs QRL data and transactions through a standalone Clef process",
	ginkgo.Serial,
	ginkgo.Label("e2e", "live", "clef"),
	func(ctx ginkgo.SpecContext) {
		workDir := ginkgo.GinkgoT().TempDir()
		clefPath := filepath.Join(workDir, "clef")
		ginkgo.By("building the current Clef binary")
		gomega.Expect(build.Binary(ctx, "./cmd/clef", clefPath)).To(gomega.Succeed())

		developmentWallet, err := network.UnsafeDevelopmentWallet()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("running and verifying the standalone Clef signing scenario")
		gomega.Expect(run(
			ctx,
			clefPath,
			workDir,
			developmentWallet,
		)).To(gomega.Succeed())
	},
	ginkgo.SpecTimeout(liveSpecTimeout),
)
