// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"strings"
	"testing"
)

func TestMutationLeaseContendsAndReopens(t *testing.T) {
	networkDir, err := ensureNetworkDirectory(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	first, err := AcquireMutationLease(networkDir)
	if err != nil {
		t.Fatal(err)
	}
	if first.NetworkDir() != networkDir {
		t.Fatalf("lease directory = %q, want %q", first.NetworkDir(), networkDir)
	}
	if second, err := AcquireMutationLease(networkDir); err == nil {
		_ = second.Close()
		t.Fatal("concurrent lease was acquired")
	} else if !strings.Contains(err.Error(), "already in progress") {
		t.Fatalf("contention error = %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := AcquireMutationLease(networkDir)
	if err != nil {
		t.Fatalf("reopen lease: %v", err)
	}
	if err := reopened.Close(); err != nil {
		t.Fatal(err)
	}
}
