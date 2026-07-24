// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.
//
// The go-qrl library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package bind

import (
	"strings"
	"testing"
)

func TestBindTupleComponentNameCollisions(t *testing.T) {
	contractABI := `[
		{"inputs":[{"components":[
			{"name":"data","type":"bytes"},
			{"name":"_data","type":"bytes"},
			{"name":"data","type":"bytes"}
		],"internalType":"struct Collision.Record[][2]","name":"records","type":"tuple[][2]"}],
		"name":"store","outputs":[],"stateMutability":"nonpayable","type":"function"}
	]`
	code, err := Bind([]string{"Collision"}, []string{contractABI}, []string{""}, nil, "bindtest", nil, nil)
	if err != nil {
		t.Fatalf("Bind() failed: %v", err)
	}
	normalized := strings.Join(strings.Fields(code), " ")
	for _, want := range []string{
		`type CollisionRecord struct`,
		`Data []byte "abi:\"data\""`,
		`Data0 []byte "abi:\"_data\""`,
		`Data1 []byte "abi:\"data\""`,
		`records [2][]CollisionRecord`,
	} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("generated binding does not contain %q\n%s", want, code)
		}
	}
}

func TestBindNamedTupleIdentityIncludesStructure(t *testing.T) {
	firstABI := `[
		{"inputs":[{"components":[{"name":"amount","type":"uint512"}],
		"internalType":"struct A.B","name":"value","type":"tuple"}],
		"name":"first","outputs":[],"stateMutability":"nonpayable","type":"function"}
	]`
	secondABI := `[
		{"inputs":[{"components":[{"name":"recipient","type":"address"},{"name":"memo","type":"bytes64"}],
		"internalType":"struct AB","name":"value","type":"tuple"}],
		"name":"second","outputs":[],"stateMutability":"nonpayable","type":"function"}
	]`
	code, err := Bind(
		[]string{"One", "Two"},
		[]string{firstABI, secondABI},
		[]string{"", ""},
		nil,
		"bindtest",
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("Bind() failed: %v", err)
	}
	normalized := strings.Join(strings.Fields(code), " ")
	for _, want := range []string{
		`type AB struct { Amount *big.Int "abi:\"amount\"" }`,
		`type AB0 struct { Recipient common.Address "abi:\"recipient\"" Memo [64]byte "abi:\"memo\"" }`,
		`func (_One *OneTransactor) First(opts *bind.TransactOpts, value AB)`,
		`func (_Two *TwoTransactor) Second(opts *bind.TransactOpts, value AB0)`,
	} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("generated binding does not contain %q\n%s", want, code)
		}
	}
	if strings.Count(normalized, "type AB struct") != 1 || strings.Count(normalized, "type AB0 struct") != 1 {
		t.Fatalf("expected two unique, non-duplicated tuple declarations\n%s", code)
	}
}

func TestBindIdenticalNamedTupleIsReused(t *testing.T) {
	contractABI := `[
		{"inputs":[{"components":[{"name":"amount","type":"uint512"}],"internalType":"struct Shared.Record","name":"value","type":"tuple"}],"name":"first","outputs":[],"stateMutability":"nonpayable","type":"function"},
		{"inputs":[{"components":[{"name":"amount","type":"uint512"}],"internalType":"struct Shared.Record","name":"value","type":"tuple"}],"name":"second","outputs":[],"stateMutability":"nonpayable","type":"function"}
	]`
	code, err := Bind([]string{"Shared"}, []string{contractABI}, []string{""}, nil, "bindtest", nil, nil)
	if err != nil {
		t.Fatalf("Bind() failed: %v", err)
	}
	if got := strings.Count(code, "type SharedRecord struct"); got != 1 {
		t.Fatalf("generated %d identical SharedRecord declarations, want 1\n%s", got, code)
	}
}
