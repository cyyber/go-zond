// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"strings"
	"testing"
)

func TestOwnershipRequiresExactUUIDForDestruction(t *testing.T) {
	record := OwnershipRecord{Name: "e2e"}
	if _, err := record.Enclave(); err == nil || !strings.Contains(err.Error(), "exact enclave UUID") {
		t.Fatalf("incomplete ownership error = %v", err)
	}
	record.UUID = strings.Repeat("a", 32)
	enclave, err := record.Enclave()
	if err != nil {
		t.Fatal(err)
	}
	if enclave.Name != record.Name || enclave.UUID != record.UUID {
		t.Fatalf("owned enclave = %+v", enclave)
	}
}
