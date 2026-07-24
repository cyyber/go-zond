// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStateIsOnlyAReadyMarker(t *testing.T) {
	payload, err := json.Marshal(State{Ready: true})
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) != `{"ready":true}` {
		t.Fatalf("state = %s", payload)
	}
	if err := (State{}).Validate(); err == nil {
		t.Fatal("non-ready state was accepted")
	}
}

func TestOwnershipRequiresExactUUIDForDestruction(t *testing.T) {
	record := OwnershipRecord{NetworkDir: t.TempDir(), Name: "e2e"}
	if err := record.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, err := record.OwnedEnclave(); err == nil || !strings.Contains(err.Error(), "exact UUID") {
		t.Fatalf("creation intent error = %v", err)
	}
	record.UUID = strings.Repeat("a", 32)
	enclave, err := record.OwnedEnclave()
	if err != nil {
		t.Fatal(err)
	}
	if enclave.Name != record.Name || enclave.UUID != record.UUID || !enclave.Owned {
		t.Fatalf("owned enclave = %+v", enclave)
	}
}
