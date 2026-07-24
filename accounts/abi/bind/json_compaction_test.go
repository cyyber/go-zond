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

func TestBindCompactsABIWithoutMutatingJSONStringWhitespace(t *testing.T) {
	const contractABI = `[
		{
			"type": "function",
			"name": "store",
			"stateMutability": "nonpayable",
			"inputs": [{
				"name": "value",
				"type": "tuple",
				"internalType": "struct Space.Record",
				"components": [{
					"name": "displayName",
					"type": "string",
					"internalType": "string"
				}]
			}],
			"outputs": []
		}
	]`
	code, err := Bind(
		[]string{"Whitespace"},
		[]string{contractABI},
		[]string{""},
		nil,
		"bindtest",
		nil,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(code, `struct Space.Record`) {
		t.Fatalf("generated metadata mutated whitespace inside internalType:\n%s", code)
	}
	if strings.Contains(code, `structSpace.Record`) {
		t.Fatalf("generated metadata contains lossy internalType normalization:\n%s", code)
	}
}
