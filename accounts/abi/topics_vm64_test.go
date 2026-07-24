// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.
//
// The go-qrl library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package abi

import (
	"strings"
	"testing"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/crypto"
)

func TestParseTopicsCompositeHashVM64(t *testing.T) {
	t.Parallel()

	tupleType, err := NewType("tuple", "struct Topics.Record", []ArgumentMarshaling{
		{Name: "amount", Type: "uint512"},
		{Name: "recipient", Type: "address"},
	})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		typeName   string
		parsedType *Type
	}{
		{name: "string", typeName: "string"},
		{name: "bytes", typeName: "bytes"},
		{name: "slice", typeName: "uint256[]"},
		{name: "array", typeName: "bytes64[2]"},
		{name: "tuple", parsedType: &tupleType},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compositeType := test.parsedType
			if compositeType == nil {
				parsed, err := NewType(test.typeName, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				compositeType = &parsed
			}
			fields := Arguments{{Name: "value", Type: *compositeType, Indexed: true}}
			want := crypto.Keccak256Hash([]byte("vm64 indexed " + test.name))
			topic := common.HashToLogTopic(want)

			out := new(struct{ Value common.Hash })
			if err := ParseTopics(out, fields, []common.LogTopic{topic}); err != nil {
				t.Fatalf("parse canonical indexed hash: %v", err)
			}
			if out.Value != want {
				t.Fatalf("decoded hash = %s, want %s", out.Value, want)
			}

			mapped := make(map[string]any)
			if err := ParseTopicsIntoMap(mapped, fields, []common.LogTopic{topic}); err != nil {
				t.Fatalf("parse canonical indexed hash into map: %v", err)
			}
			if got, ok := mapped["value"].(common.Hash); !ok || got != want {
				t.Fatalf("mapped hash = %T(%v), want %s", mapped["value"], mapped["value"], want)
			}

			nonCanonical := topic
			nonCanonical[common.HashLength] = 1
			if err := ParseTopics(out, fields, []common.LogTopic{nonCanonical}); err == nil || !strings.Contains(err.Error(), "improperly encoded") {
				t.Fatalf("non-canonical indexed hash error = %v, want encoding error", err)
			}
		})
	}
}

func TestParseTopicsCompositeHashDestinationErrors(t *testing.T) {
	t.Parallel()

	stringType, err := NewType("string", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	fields := Arguments{{Name: "value", Type: stringType, Indexed: true}}
	topic := common.HashToLogTopic(crypto.Keccak256Hash([]byte("vm64 indexed topic")))

	wrongType := new(struct{ Value common.LogTopic })
	if err := ParseTopics(wrongType, fields, []common.LogTopic{topic}); err == nil || !strings.Contains(err.Error(), "cannot unmarshal indexed string") {
		t.Fatalf("destination mismatch error = %v, want typed unmarshal error", err)
	}
	if err := ParseTopics(new(struct{}), fields, []common.LogTopic{topic}); err == nil || !strings.Contains(err.Error(), "field Value not found") {
		t.Fatalf("missing field error = %v, want field error", err)
	}
	if err := ParseTopicsIntoMap(nil, fields, []common.LogTopic{topic}); err == nil || !strings.Contains(err.Error(), "output map is nil") {
		t.Fatalf("nil map error = %v, want output-map error", err)
	}
}
