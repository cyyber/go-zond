// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"bytes"
	"context"
	"runtime"
	runtimedebug "runtime/debug"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func (suite *liveSuite) assertRuntimeDiagnostics(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()

	var memStats runtime.MemStats
	gomega.Expect(raw.CallContext(ctx, &memStats, "debug_memStats")).To(gomega.Succeed())
	gomega.Expect(memStats.Sys).To(gomega.BeNumerically(">", 0))

	var gcStats runtimedebug.GCStats
	gomega.Expect(raw.CallContext(ctx, &gcStats, "debug_gcStats")).To(gomega.Succeed())

	var stacks string
	gomega.Expect(raw.CallContext(ctx, &stacks, "debug_stacks")).To(gomega.Succeed())
	gomega.Expect(bytes.Contains([]byte(stacks), []byte("goroutine"))).To(gomega.BeTrue())
}
