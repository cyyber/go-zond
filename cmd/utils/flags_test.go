// Copyright 2019 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

// Package utils contains internal helper functions for go-ethereum commands.
package utils

import (
	"flag"
	"reflect"
	"testing"

	"github.com/theQRL/go-qrl/p2p"
	"github.com/theQRL/go-qrl/p2p/qnode"
	"github.com/theQRL/go-qrl/params"
	"github.com/urfave/cli/v2"
)

func Test_SplitTagsFlag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args string
		want map[string]string
	}{
		{
			"2 tags case",
			"host=localhost,bzzkey=123",
			map[string]string{
				"host":   "localhost",
				"bzzkey": "123",
			},
		},
		{
			"1 tag case",
			"host=localhost123",
			map[string]string{
				"host": "localhost123",
			},
		},
		{
			"empty case",
			"",
			map[string]string{},
		},
		{
			"garbage",
			"smth=smthelse=123",
			map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := SplitTagsFlag(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitTagsFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetBootstrapNodesPrecedence(t *testing.T) {
	newContext := func(args ...string) *cli.Context {
		set := flag.NewFlagSet("test", flag.ContinueOnError)
		if err := BootnodesFlag.Apply(set); err != nil {
			t.Fatalf("apply flag: %v", err)
		}
		if err := set.Parse(args); err != nil {
			t.Fatalf("parse flags: %v", err)
		}
		return cli.NewContext(nil, set, nil)
	}
	configured := qnode.MustParse(params.TestnetBootnodes[0])
	replacement := qnode.MustParse(params.TestnetBootnodes[1])

	cfg := &p2p.Config{BootstrapNodes: []*qnode.Node{configured}}
	setBootstrapNodes(newContext(), cfg)
	if len(cfg.BootstrapNodes) != 1 || cfg.BootstrapNodes[0].String() != configured.String() {
		t.Fatalf("config bootnode was not preserved: %v", cfg.BootstrapNodes)
	}

	cfg = &p2p.Config{BootstrapNodes: []*qnode.Node{configured}}
	setBootstrapNodes(newContext("--bootnodes", params.TestnetBootnodes[1]), cfg)
	if len(cfg.BootstrapNodes) != 1 || cfg.BootstrapNodes[0].String() != replacement.String() {
		t.Fatalf("CLI bootnode did not override config: %v", cfg.BootstrapNodes)
	}
}
