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

package main

import (
	"encoding/json"
	"fmt"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
)

// testAccount is the on-disk fixture shape: a stable label, the ExtendedSeed
// used to derive the wallet, and the resulting 64-byte QRL address.
type testAccount struct {
	Label   string `json:"label"`
	Seed    string `json:"seed"`
	Address string `json:"address"`
}

// addressLabels is the canonical list of fixture accounts that tests can pull
// by name. Add new labels here — never inline new hard-coded seeds in tests.
var addressLabels = []string{
	"alice",
	"bob",
	"carol",
	"dave",
	"eve",
	"relayer",
	"miner-1",
	"miner-2",
	"contract-deployer",
	"notifier",
}

// generateAddresses derives a deterministic wallet for each label and writes
// the resulting table to testdata/addresses.json relative to outDir.
func generateAddresses(outDir string) error {
	accounts := make([]testAccount, 0, len(addressLabels))
	for _, label := range addressLabels {
		seed := deterministicSeed(label)
		w, err := wallet.RestoreFromSeedHex(seed)
		if err != nil {
			return fmt.Errorf("generate %q: restore seed: %w", label, err)
		}
		addr := common.Address(w.GetAddress())
		accounts = append(accounts, testAccount{
			Label:   label,
			Seed:    seed,
			Address: addr.String(),
		})
	}
	payload, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	const rel = "testdata/addresses.json"
	if err := writeFile(outDir, rel, payload); err != nil {
		return err
	}
	reportGenerated("addresses", rel)
	return nil
}
