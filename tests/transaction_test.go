// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package tests

import (
	"testing"

	"github.com/theQRL/go-zond/common"
)

func TestTransaction(t *testing.T) {
	t.Parallel()

	txt := new(testMatcher)
	txt.walk(t, transactionTestDir, func(t *testing.T, name string, test *TransactionTest) {
		if err := txt.checkFailure(t, test.Run()); err != nil {
			t.Error(err)
		}
	})
}
func TestExecutionSpecTransaction(t *testing.T) {
	if !common.FileExist(executionSpecStateTestDir) {
		t.Skipf("directory %s does not exist", executionSpecStateTestDir)
	}
	st := new(testMatcher)

	st.walk(t, executionSpecTransactionTestDir, func(t *testing.T, name string, test *TransactionTest) {
		if err := st.checkFailure(t, test.Run()); err != nil {
			t.Error(err)
		}
	})
}
