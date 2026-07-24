// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.
//
// The go-qrl library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package bind

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

const reservedIdentifierABI = `[
	{"inputs":[{"name":"auth","type":"uint512"},{"name":"backend","type":"address"},{"name":"parsed","type":"bytes"}],"stateMutability":"nonpayable","type":"constructor"},
	{"inputs":[
		{"name":"opts","type":"uint512"},
		{"name":"out","type":"bytes"},
		{"name":"err","type":"bool"},
		{"name":"append","type":"address"},
		{"name":"_ReservedIdentifiers","type":"bytes64"},
		{"components":[{"name":"data","type":"bytes"}],"internalType":"struct ReservedIdentifiers","name":"record","type":"tuple"}
	],"name":"set","outputs":[{"name":"out","type":"uint512"},{"name":"err","type":"bool"}],"stateMutability":"nonpayable","type":"function"},
	{"anonymous":true,"inputs":[
		{"indexed":true,"name":"opts","type":"uint512"},
		{"indexed":true,"name":"sink","type":"address"},
		{"indexed":true,"name":"foo","type":"uint512"},
		{"indexed":true,"name":"fooRule","type":"uint512"},
		{"indexed":false,"name":"raw","type":"bytes64"}
	],"name":"Observed","type":"event"},
	{"anonymous":false,"inputs":[],"name":"ObservedIterator","type":"event"},
	{"anonymous":false,"inputs":[],"name":"Raw","type":"event"}
]`

const bodyScopeShadowABI = `[
	{"inputs":[
		{"name":"bind","type":"uint512"},
		{"name":"common","type":"uint512"},
		{"name":"errors","type":"uint512"},
		{"name":"strings","type":"uint512"},
		{"name":"BodyScopeShadow","type":"uint512"},
		{"name":"BodyScopeShadowMetaData","type":"uint512"},
		{"name":"BodyScopeShadowBin","type":"uint512"},
		{"name":"BodyScopeShadowCaller","type":"uint512"},
		{"name":"BodyScopeShadowTransactor","type":"uint512"},
		{"name":"BodyScopeShadowFilterer","type":"uint512"},
		{"name":"DeployShadowLib","type":"uint512"},
		{"name":"shadowLibAddr","type":"uint512"},
		{"name":"BodyScopeShadowABI","type":"uint512"},
		{"name":"BodyScopeShadowSession","type":"uint512"}
	],"stateMutability":"nonpayable","type":"constructor"},
	{"inputs":[{"name":"abi","type":"uint512"}],"name":"inspect","outputs":[
		{"name":"abi","type":"uint512"},
		{"name":"value","type":"uint512"}
	],"stateMutability":"view","type":"function"},
	{"inputs":[{"name":"abi","type":"uint512"}],"name":"ping","outputs":[],"stateMutability":"view","type":"function"},
	{"inputs":[{"name":"abi","type":"uint512"}],"name":"write","outputs":[
		{"name":"abi","type":"uint512"}
	],"stateMutability":"nonpayable","type":"function"}
]`

func TestBindResolvesGeneratedIdentifierCollisions(t *testing.T) {
	code, err := Bind(
		[]string{"ReservedIdentifiers"},
		[]string{reservedIdentifierABI},
		[]string{"00"},
		nil,
		"bindtest",
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("Bind() rejected valid reserved-name ABI: %v", err)
	}
	normalized := strings.Join(strings.Fields(code), " ")
	for _, want := range []string{
		"type ReservedIdentifiers0 struct",
		"func DeployReservedIdentifiers(auth *bind.TransactOpts, backend bind.ContractBackend, auth0 *big.Int, backend0 common.Address, parsed0 []byte)",
		"func (_ReservedIdentifiers *ReservedIdentifiersTransactor) Set(opts *bind.TransactOpts, opts0 *big.Int, out2 []byte, err0 bool, arg3 common.Address, _ReservedIdentifiers0 [64]byte, record ReservedIdentifiers0)",
		`Opts0 *big.Int "abi:\"opts\""`,
		`Sink0 common.Address "abi:\"sink\""`,
		`Raw0 [64]byte "abi:\"raw\""`,
		"Raw types.Log",
		"type ReservedIdentifiersObservedIterator0 struct",
		"type ReservedIdentifiersRaw0 struct",
	} {
		if !strings.Contains(normalized, want) {
			t.Errorf("generated binding does not contain %q\n%s", want, code)
		}
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "reserved_identifiers.go", code, parser.AllErrors|parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("generated reserved-name binding does not parse: %v", err)
	}
	if diagnostics := generatedDeclarationErrors(fset, file); len(diagnostics) != 0 {
		t.Fatalf("generated reserved-name binding has semantic declaration conflicts: %v", diagnostics)
	}
}

func TestBindResolvesFunctionBodyIdentifierCollisions(t *testing.T) {
	code, err := Bind(
		[]string{"BodyScopeShadow", "ShadowLib"},
		[]string{bodyScopeShadowABI, "[]"},
		[]string{"__$body_scope_shadow_lib$__", "00"},
		nil,
		"bindtest",
		map[string]string{"body_scope_shadow_lib": "ShadowLib"},
		nil,
	)
	if err != nil {
		t.Fatalf("Bind() rejected valid body-scope shadowing ABI: %v", err)
	}
	normalized := strings.Join(strings.Fields(code), " ")
	for _, want := range []string{
		"bind0 *big.Int",
		"common0 *big.Int",
		"errors0 *big.Int",
		"strings0 *big.Int",
		"BodyScopeShadow0 *big.Int",
		"BodyScopeShadowMetaData0 *big.Int",
		"BodyScopeShadowBin0 *big.Int",
		"BodyScopeShadowCaller0 *big.Int",
		"BodyScopeShadowTransactor0 *big.Int",
		"BodyScopeShadowFilterer0 *big.Int",
		"DeployShadowLib0 *big.Int",
		"shadowLibAddr0 *big.Int",
		"BodyScopeShadowABI *big.Int",
		"BodyScopeShadowSession *big.Int",
		"Inspect(opts *bind.CallOpts, abi0 *big.Int)",
		"Ping(opts *bind.CallOpts, abi *big.Int)",
		"Write(opts *bind.TransactOpts, abi *big.Int)",
		"struct { Abi *big.Int Value *big.Int }",
	} {
		if !strings.Contains(normalized, want) {
			t.Errorf("generated binding does not contain %q\n%s", want, code)
		}
	}
}

func TestBindTupleIdentityIncludesOriginalComponentNamesRecursively(t *testing.T) {
	const contractABI = `[
		{"inputs":[{"components":[{"name":"data","type":"bytes"}],"internalType":"struct Shared.Record","name":"value","type":"tuple"}],"name":"first","outputs":[],"stateMutability":"nonpayable","type":"function"},
		{"inputs":[{"components":[{"name":"_data","type":"bytes"}],"internalType":"struct Shared.Record","name":"value","type":"tuple"}],"name":"second","outputs":[],"stateMutability":"nonpayable","type":"function"},
		{"inputs":[{"components":[{"components":[{"name":"data","type":"bytes"}],"internalType":"struct Shared.Child","name":"child","type":"tuple"}],"internalType":"struct Shared.Outer","name":"value","type":"tuple"}],"name":"nestedFirst","outputs":[],"stateMutability":"nonpayable","type":"function"},
		{"inputs":[{"components":[{"components":[{"name":"_data","type":"bytes"}],"internalType":"struct Shared.Child","name":"child","type":"tuple"}],"internalType":"struct Shared.Outer","name":"value","type":"tuple"}],"name":"nestedSecond","outputs":[],"stateMutability":"nonpayable","type":"function"}
	]`
	code, err := Bind([]string{"TupleTags"}, []string{contractABI}, []string{""}, nil, "bindtest", nil, nil)
	if err != nil {
		t.Fatalf("Bind() failed: %v", err)
	}
	normalized := strings.Join(strings.Fields(code), " ")
	for _, want := range []string{
		`type SharedRecord struct { Data []byte "abi:\"data\"" }`,
		`type SharedRecord0 struct { Data []byte "abi:\"_data\"" }`,
		`type SharedChild struct { Data []byte "abi:\"data\"" }`,
		`type SharedChild0 struct { Data []byte "abi:\"_data\"" }`,
		`type SharedOuter struct { Child SharedChild "abi:\"child\"" }`,
		`type SharedOuter0 struct { Child SharedChild0 "abi:\"child\"" }`,
	} {
		if !strings.Contains(normalized, want) {
			t.Errorf("generated binding does not preserve tuple identity %q\n%s", want, code)
		}
	}
}

func TestBindRejectsSessionIdentifierCollisions(t *testing.T) {
	tests := []struct {
		name         string
		contractABI  string
		aliases      map[string]string
		identifier   string
		fixedAliases map[string]string
	}{
		{
			name: "ordinary fallback and special fallback",
			contractABI: `[
				{"inputs":[],"name":"fallback","outputs":[],"stateMutability":"view","type":"function"},
				{"stateMutability":"payable","type":"fallback"}
			]`,
			identifier:   "Fallback",
			fixedAliases: map[string]string{"fallback": "OrdinaryFallback"},
		},
		{
			name: "ordinary receive and special receive",
			contractABI: `[
				{"inputs":[],"name":"receive","outputs":[],"stateMutability":"nonpayable","type":"function"},
				{"stateMutability":"payable","type":"receive"}
			]`,
			identifier:   "Receive",
			fixedAliases: map[string]string{"receive": "OrdinaryReceive"},
		},
		{
			name: "call and transact alias",
			contractABI: `[
				{"inputs":[],"name":"foo","outputs":[],"stateMutability":"view","type":"function"},
				{"inputs":[],"name":"bar","outputs":[],"stateMutability":"nonpayable","type":"function"}
			]`,
			aliases:      map[string]string{"foo": "Shared", "bar": "Shared"},
			identifier:   "Shared",
			fixedAliases: map[string]string{"foo": "Shared", "bar": "WriteShared"},
		},
		{
			name: "session contract field",
			contractABI: `[
				{"inputs":[],"name":"Contract","outputs":[],"stateMutability":"view","type":"function"}
			]`,
			identifier:   "Contract",
			fixedAliases: map[string]string{"Contract": "ContractMethod"},
		},
		{
			name: "caller session options field",
			contractABI: `[
				{"inputs":[],"name":"CallOpts","outputs":[],"stateMutability":"view","type":"function"}
			]`,
			identifier:   "CallOpts",
			fixedAliases: map[string]string{"CallOpts": "ReadOptions"},
		},
		{
			name: "transactor session options field",
			contractABI: `[
				{"inputs":[],"name":"TransactOpts","outputs":[],"stateMutability":"nonpayable","type":"function"}
			]`,
			identifier:   "TransactOpts",
			fixedAliases: map[string]string{"TransactOpts": "WriteOptions"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Bind([]string{"SessionNames"}, []string{tt.contractABI}, []string{""}, nil, "bindtest", nil, tt.aliases)
			if err == nil {
				t.Fatalf("Bind() accepted duplicate session identifier %q", tt.identifier)
			}
			if !strings.Contains(err.Error(), tt.identifier) || !strings.Contains(err.Error(), "--alias") {
				t.Fatalf("Bind() returned unactionable collision error: %v", err)
			}

			code, err := Bind([]string{"SessionNames"}, []string{tt.contractABI}, []string{""}, nil, "bindtest", nil, tt.fixedAliases)
			if err != nil {
				t.Fatalf("Bind() rejected collision-resolving alias: %v", err)
			}
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "session_names.go", code, parser.AllErrors|parser.SkipObjectResolution)
			if err != nil {
				t.Fatalf("binding generated after alias does not parse: %v", err)
			}
			if diagnostics := generatedDeclarationErrors(fset, file); len(diagnostics) != 0 {
				t.Fatalf("binding generated after alias has declaration conflicts: %v", diagnostics)
			}
		})
	}
}

func TestBindRejectsInvalidMethodIdentifier(t *testing.T) {
	const contractABI = `[
		{"inputs":[],"name":"bad-name","outputs":[],"stateMutability":"view","type":"function"}
	]`
	_, err := Bind([]string{"InvalidMethod"}, []string{contractABI}, []string{""}, nil, "bindtest", nil, nil)
	if err == nil {
		t.Fatal("Bind() accepted a method name that cannot be a Go identifier")
	}
	if !strings.Contains(err.Error(), "Bad-name") || !strings.Contains(err.Error(), "--alias") {
		t.Fatalf("Bind() returned unactionable invalid-identifier error: %v", err)
	}
	if _, err := Bind(
		[]string{"InvalidMethod"},
		[]string{contractABI},
		[]string{""},
		nil,
		"bindtest",
		nil,
		map[string]string{"bad-name": "ValidName"},
	); err != nil {
		t.Fatalf("Bind() rejected valid alias for invalid method identifier: %v", err)
	}
}
