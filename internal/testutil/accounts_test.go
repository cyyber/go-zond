// Copyright 2026 The go-qrl Authors
// This file is part of go-qrl.
//
// go-qrl is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-qrl is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

package testutil

import "testing"

// TestLoadAccountDerivesAddress confirms the JSON address matches the address
// derived from the JSON seed. If this fails, testdata/addresses.json is stale.
func TestLoadAccountDerivesAddress(t *testing.T) {
	t.Parallel()
	acc := LoadAccount(t, "alice")
	w := acc.Wallet(t)
	got := w.GetAddress()
	want := acc.AddressBytes(t)
	if got != want {
		t.Fatalf("address mismatch for %q:\n  derived: %x\n  fixture: %x", acc.Label, got, want)
	}
}

// TestLoadAccountUnknownFails sanity-checks the "no silent fallback"
// guarantee: an unknown label must fail the test loudly.
func TestLoadAccountUnknownFails(t *testing.T) {
	t.Parallel()
	fake := &fakeTB{TB: t}
	_ = LoadAccount(fake, "does-not-exist")
	if !fake.failed {
		t.Fatal("expected LoadAccount on unknown label to call t.Fatal")
	}
}

type fakeTB struct {
	testing.TB
	failed bool
}

func (f *fakeTB) Fatalf(format string, args ...any) { f.failed = true }
func (f *fakeTB) Fatal(args ...any)                 { f.failed = true }
func (f *fakeTB) Helper()                           {}
