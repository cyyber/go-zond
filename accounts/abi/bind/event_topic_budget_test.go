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

package bind_test

import (
	"strings"
	"testing"

	"github.com/theQRL/go-qrl/accounts/abi"
	"github.com/theQRL/go-qrl/accounts/abi/bind"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/core/types"
)

func TestBoundContractAcceptedEventsStayWithinTopicBudget(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(`[
		{"anonymous":false,"inputs":[
			{"indexed":true,"name":"a","type":"uint8"},
			{"indexed":true,"name":"b","type":"uint8"},
			{"indexed":true,"name":"c","type":"uint8"}
		],"name":"Regular","type":"event"},
		{"anonymous":true,"inputs":[
			{"indexed":true,"name":"a","type":"uint8"},
			{"indexed":true,"name":"b","type":"uint8"},
			{"indexed":true,"name":"c","type":"uint8"},
			{"indexed":true,"name":"d","type":"uint8"}
		],"name":"Anonymous","type":"event"}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	filterer := new(recordingFilterer)
	contract := bind.NewBoundContract(common.Address{}, parsed, nil, nil, filterer)
	tests := []struct {
		name  string
		rules [][]any
	}{
		{
			name:  "Regular",
			rules: [][]any{{uint8(1)}, {uint8(2)}, {uint8(3)}},
		},
		{
			name:  "Anonymous",
			rules: [][]any{{uint8(1)}, {uint8(2)}, {uint8(3)}, {uint8(4)}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, subscription, err := contract.FilterLogs(nil, test.name, test.rules...)
			if err != nil {
				t.Fatalf("FilterLogs: %v", err)
			}
			subscription.Unsubscribe()
			filterQuery := filterer.filtered[len(filterer.filtered)-1]
			if got := len(filterQuery.Topics); got != 4 {
				t.Fatalf("FilterLogs produced %d topic criteria, want 4", got)
			}

			_, subscription, err = contract.WatchLogs(nil, test.name, test.rules...)
			if err != nil {
				t.Fatalf("WatchLogs: %v", err)
			}
			subscription.Unsubscribe()
			watchQuery := filterer.watched[len(filterer.watched)-1]
			if got := len(watchQuery.Topics); got != 4 {
				t.Fatalf("WatchLogs produced %d topic criteria, want 4", got)
			}
		})
	}
}

func TestBoundContractRejectsHandBuiltOverBudgetEvents(t *testing.T) {
	uint8Type, err := abi.NewType("uint8", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name      string
		anonymous bool
		indexed   int
	}{
		{name: "regular", indexed: 4},
		{name: "anonymous", anonymous: true, indexed: 5},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inputs := make(abi.Arguments, test.indexed)
			for i := range inputs {
				inputs[i] = abi.Argument{Name: string(rune('a' + i)), Type: uint8Type, Indexed: true}
			}
			event := abi.NewEvent("TooMany", "TooMany", test.anonymous, inputs)
			handBuilt := abi.ABI{Events: map[string]abi.Event{"TooMany": event}}
			filterer := new(recordingFilterer)
			contract := bind.NewBoundContract(common.Address{}, handBuilt, nil, nil, filterer)

			if _, _, err := contract.FilterLogs(nil, "TooMany"); err == nil || !strings.Contains(err.Error(), "indexed arguments, maximum is") {
				t.Fatalf("FilterLogs error = %v, want topic-budget error", err)
			}
			if _, _, err := contract.WatchLogs(nil, "TooMany"); err == nil || !strings.Contains(err.Error(), "indexed arguments, maximum is") {
				t.Fatalf("WatchLogs error = %v, want topic-budget error", err)
			}
			if len(filterer.filtered) != 0 || len(filterer.watched) != 0 {
				t.Fatal("over-budget event query reached the backend")
			}
			if err := contract.UnpackLog(new(struct{}), "TooMany", types.Log{}); err == nil || !strings.Contains(err.Error(), "indexed arguments, maximum is") {
				t.Fatalf("UnpackLog error = %v, want topic-budget error", err)
			}
			if err := contract.UnpackLogIntoMap(make(map[string]any), "TooMany", types.Log{}); err == nil || !strings.Contains(err.Error(), "indexed arguments, maximum is") {
				t.Fatalf("UnpackLogIntoMap error = %v, want topic-budget error", err)
			}
		})
	}
}
