// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package clef

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/theQRL/go-qrl/common"
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
		workDir, err := os.MkdirTemp("", "go-qrl-clef-suite-")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		ginkgo.DeferCleanup(os.RemoveAll, workDir)

		clefPath := filepath.Join(workDir, "clef")
		ginkgo.By("building the current Clef binary")
		gomega.Expect(buildClef(ctx, clefPath)).To(gomega.Succeed())

		developmentWallet, err := network.UnsafeDevelopmentWallet()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		seed, err := developmentWallet.GetSeed()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ginkgo.By("running and verifying the standalone Clef signing scenario")
		result, err := Run(ctx, Config{
			ClefPath: clefPath,
			Seed:     hex.EncodeToString(seed.ToBytes()),
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(result.Account).To(gomega.Equal(
			common.Address(developmentWallet.GetAddress()),
		))
		gomega.Expect(result.Version).NotTo(gomega.BeEmpty())
	},
	ginkgo.SpecTimeout(liveSpecTimeout),
)

func buildClef(ctx context.Context, output string) error {
	root, err := repositoryRoot()
	if err != nil {
		return err
	}
	command := exec.CommandContext(ctx, "go", "build", "-o", output, "./cmd/clef")
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("build Clef: %w\n%s", err, output)
	}
	return nil
}

func repositoryRoot() (string, error) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("locate Clef suite source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(source), "..", "..", "..", "..")), nil
}
