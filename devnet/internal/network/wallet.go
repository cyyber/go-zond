// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/theQRL/go-qrl/common"
	qrlwallet "github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
)

func ensureWallet(networkDir string) (string, error) {
	seedPath := filepath.Join(networkDir, "wallet.seed")
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
	wallet, err := qrlwallet.RestoreFromFile(path)
	if err != nil {
		return "", fmt.Errorf("restore existing wallet: %w", err)
	}
	address := common.Address(wallet.GetAddress()).Hex()
	return address, nil
}

func writeExclusive(path string, payload []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.Write(payload); err != nil {
		return errors.Join(err, file.Close(), os.Remove(path))
	}
	if err := file.Close(); err != nil {
		return errors.Join(err, os.Remove(path))
	}
	return nil
}
