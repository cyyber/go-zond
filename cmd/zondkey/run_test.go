// Copyright 2018 The go-ethereum Authors
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

package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/docker/docker/pkg/reexec"
	"github.com/theQRL/go-zond/internal/cmdtest"
)

type testZondkey struct {
	*cmdtest.TestCmd
}

// spawns zondkey with the given command line args.
func runZondkey(t *testing.T, args ...string) *testZondkey {
	tt := new(testZondkey)
	tt.TestCmd = cmdtest.NewTestCmd(t, tt)
	tt.Run("zondkey-test", args...)
	return tt
}

func TestMain(m *testing.M) {
	// Run the app if we've been exec'd as "zondkey-test" in runZondkey.
	reexec.Register("zondkey-test", func() {
		if err := app.Run(os.Args); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	})
	// check if we have been reexec'd
	if reexec.Init() {
		return
	}
	os.Exit(m.Run())
}
