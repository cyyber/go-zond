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
)

func TestEventByIDUsesDeclarationOrderForDuplicateSignatures(t *testing.T) {
	const definition = `[
		{"type":"event","name":"Repeated","anonymous":false,"inputs":[{"name":"first","type":"uint512","indexed":true}]},
		{"type":"event","name":"Repeated","anonymous":false,"inputs":[{"name":"second","type":"uint512","indexed":false}]}
	]`
	parsed, err := JSON(strings.NewReader(definition))
	if err != nil {
		t.Fatal(err)
	}
	first := parsed.Events["Repeated"]
	second := parsed.Events["Repeated0"]
	if first.ID != second.ID {
		t.Fatalf("duplicate signature IDs differ: %s and %s", first.ID, second.ID)
	}
	for attempt := 0; attempt < 100; attempt++ {
		resolved, err := parsed.EventByID(first.ID)
		if err != nil {
			t.Fatal(err)
		}
		if resolved.Name != "Repeated" || resolved.Inputs[0].Name != "first" || !resolved.Inputs[0].Indexed {
			t.Fatalf("lookup attempt %d resolved %+v, want first declaration", attempt, resolved)
		}
	}
}

func TestEventByIDUsesLexicalFallbackForHandBuiltABI(t *testing.T) {
	event := NewEvent("Repeated", "Repeated", false, nil)
	duplicate := NewEvent("Repeated0", "Repeated", false, nil)
	parsed := ABI{Events: map[string]Event{
		"z": duplicate,
		"a": event,
	}}
	for attempt := 0; attempt < 100; attempt++ {
		resolved, err := parsed.EventByID(event.ID)
		if err != nil {
			t.Fatal(err)
		}
		if resolved.Name != "Repeated" {
			t.Fatalf("lookup attempt %d resolved %q, want lexical map entry", attempt, resolved.Name)
		}
	}
}
