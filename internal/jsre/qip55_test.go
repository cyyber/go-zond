// Copyright 2026 The go-qrl Authors
// This file is part of go-qrl.
//
// go-qrl is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-qrl is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

package jsre

import (
	"os"
	"strings"
	"testing"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/internal/jsre/deps"
)

func TestEmbeddedWeb3QIP55Checksum(t *testing.T) {
	t.Parallel()

	re := New("", os.Stdout)
	defer re.Stop(false)

	if err := re.Compile("bignumber.js", deps.BigNumberJS); err != nil {
		t.Fatalf("compile bignumber.js: %v", err)
	}
	if err := re.Compile("web3.js", deps.Web3JS); err != nil {
		t.Fatalf("compile web3.js: %v", err)
	}
	if _, err := re.Run("var Web3 = require('web3'); var web3 = new Web3();"); err != nil {
		t.Fatalf("init web3: %v", err)
	}

	lower := "Q" + strings.Repeat("0", 88) + "b26f2b342aab24bcf63ea218c6a9274d30ab9a15"
	addr, err := common.NewAddressFromString(lower)
	if err != nil {
		t.Fatal(err)
	}
	want := addr.Hex()

	gotValue, err := re.Run("web3.toChecksumAddress('" + lower + "')")
	if err != nil {
		t.Fatalf("toChecksumAddress: %v", err)
	}
	if got := gotValue.String(); got != want {
		t.Fatalf("toChecksumAddress mismatch: got %s want %s", got, want)
	}

	validValue, err := re.Run("web3.isChecksumAddress('" + want + "')")
	if err != nil {
		t.Fatalf("isChecksumAddress(valid): %v", err)
	}
	if !validValue.ToBoolean() {
		t.Fatal("canonical QIP-55 address was rejected")
	}

	invalidValue, err := re.Run("web3.isChecksumAddress('" + lower + "')")
	if err != nil {
		t.Fatalf("isChecksumAddress(lower): %v", err)
	}
	if invalidValue.ToBoolean() {
		t.Fatal("lowercase compatibility address passed strict checksum validation")
	}
}
