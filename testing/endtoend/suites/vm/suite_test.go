// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package vm

import (
	"testing"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "VM and precompile live E2E suite")
}
