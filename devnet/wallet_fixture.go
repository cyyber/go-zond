// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package devnet exposes fixtures shared by development networks and suites.
package devnet

import (
	_ "embed"
	"strings"

	qrlwallet "github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
)

//go:embed testdata/unsafe-development-wallet.seed
var unsafeDevelopmentWalletSeed string

// UnsafeDevelopmentWallet restores the public wallet used by disposable
// development networks. Never use this credential outside disposable local
// devnets.
func UnsafeDevelopmentWallet() (qrlwallet.Wallet, error) {
	return qrlwallet.RestoreFromSeedHex(strings.TrimSpace(unsafeDevelopmentWalletSeed))
}
