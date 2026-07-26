// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/theQRL/go-qrl/common"
	qrlwallet "github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
)

func walletSeedPath(networkDir string) string {
	return filepath.Join(networkDir, "wallet.seed")
}

func ensureWallet(networkDir string) (string, error) {
	seedPath := walletSeedPath(networkDir)
	if _, err := os.Lstat(seedPath); err == nil {
		return validateWalletSeed(seedPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	wallet, err := qrlwallet.Generate(qrlwallet.ML_DSA_87)
	if err != nil {
		return "", fmt.Errorf("generate ML-DSA wallet: %w", err)
	}
	seed, err := wallet.GetSeed()
	if err != nil {
		return "", fmt.Errorf("read wallet seed: %w", err)
	}
	address := common.Address(wallet.GetAddress()).Hex()
	if err := writeExclusive(seedPath, []byte(hex.EncodeToString(seed.ToBytes())+"\n")); err != nil {
		if errors.Is(err, os.ErrExist) {
			return validateWalletSeed(seedPath)
		}
		return "", err
	}
	return address, nil
}

func validateWalletSeed(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		return "", fmt.Errorf("%s must be a non-symlink 0600 regular file", path)
	}
	seed, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	wallet, err := qrlwallet.RestoreFromSeedHex(strings.TrimSpace(string(seed)))
	if err != nil {
		return "", fmt.Errorf("restore existing wallet: %w", err)
	}
	address := common.Address(wallet.GetAddress()).Hex()
	return address, nil
}
