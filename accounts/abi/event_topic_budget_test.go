// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.
//
// The go-qrl library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-qrl library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-qrl library. If not, see <http://www.gnu.org/licenses/>.

package abi

import (
	"strings"
	"testing"
)

func TestJSONEventTopicBudget(t *testing.T) {
	tests := []struct {
		name       string
		definition string
		wantErr    bool
	}{
		{
			name:       "regular zero indexed",
			definition: `[{"anonymous":false,"inputs":[],"name":"Budget","type":"event"}]`,
		},
		{
			name: "regular three indexed",
			definition: `[{"anonymous":false,"inputs":[
				{"indexed":true,"name":"a","type":"uint8"},
				{"indexed":true,"name":"b","type":"address"},
				{"indexed":true,"name":"c","type":"bytes64"}
			],"name":"Budget","type":"event"}]`,
		},
		{
			name:       "anonymous zero indexed",
			definition: `[{"anonymous":true,"inputs":[],"name":"Budget","type":"event"}]`,
		},
		{
			name: "anonymous four indexed",
			definition: `[{"anonymous":true,"inputs":[
				{"indexed":true,"name":"a","type":"uint8"},
				{"indexed":true,"name":"b","type":"address"},
				{"indexed":true,"name":"c","type":"bytes64"},
				{"indexed":true,"name":"d","type":"bool"}
			],"name":"Budget","type":"event"}]`,
		},
		{
			name: "regular four indexed",
			definition: `[{"anonymous":false,"inputs":[
				{"indexed":true,"name":"a","type":"uint8"},
				{"indexed":true,"name":"b","type":"uint8"},
				{"indexed":true,"name":"c","type":"uint8"},
				{"indexed":true,"name":"d","type":"uint8"}
			],"name":"Budget","type":"event"}]`,
			wantErr: true,
		},
		{
			name: "anonymous five indexed",
			definition: `[{"anonymous":true,"inputs":[
				{"indexed":true,"name":"a","type":"uint8"},
				{"indexed":true,"name":"b","type":"uint8"},
				{"indexed":true,"name":"c","type":"uint8"},
				{"indexed":true,"name":"d","type":"uint8"},
				{"indexed":true,"name":"e","type":"uint8"}
			],"name":"Budget","type":"event"}]`,
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := JSON(strings.NewReader(test.definition))
			if test.wantErr {
				if err == nil {
					t.Fatal("JSON accepted an event exceeding the VM LOG4 topic budget")
				}
				if !strings.Contains(err.Error(), "indexed arguments, maximum is") {
					t.Fatalf("JSON error = %q, want indexed-topic budget error", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("JSON rejected a valid event topic budget: %v", err)
			}
			event, ok := parsed.Events["Budget"]
			if !ok {
				t.Fatal("parsed ABI is missing Budget event")
			}
			if err := event.Validate(); err != nil {
				t.Fatalf("parsed event failed validation: %v", err)
			}
		})
	}
}

func TestNewEventTopicBudgetValidation(t *testing.T) {
	uint8Type, err := NewType("uint8", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs := make(Arguments, 4)
	for i := range inputs {
		inputs[i] = Argument{Name: string(rune('a' + i)), Type: uint8Type, Indexed: true}
	}
	event := NewEvent("Budget", "Budget", false, inputs)
	if err := event.Validate(); err == nil || !strings.Contains(err.Error(), "maximum is 3") {
		t.Fatalf("four-indexed regular event validation error = %v, want maximum-three error", err)
	}
}
