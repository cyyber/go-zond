// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

var errOwnershipAbsent = errors.New("ownership record is absent")

func ownershipPath(networkDir string) string {
	return filepath.Join(networkDir, "ownership.json")
}

func loadOwnership(networkDir string) (kurtosis.EnclaveRef, error) {
	path := ownershipPath(networkDir)
	if _, err := os.Lstat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return kurtosis.EnclaveRef{}, errOwnershipAbsent
		}
		return kurtosis.EnclaveRef{}, err
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		return kurtosis.EnclaveRef{}, fmt.Errorf("read existing ownership: %w", err)
	}
	var enclave kurtosis.EnclaveRef
	if err := json.Unmarshal(payload, &enclave); err != nil {
		return kurtosis.EnclaveRef{}, fmt.Errorf("decode ownership: %w", err)
	}
	return enclave, validateOwnershipDirectory(networkDir, enclave)
}

func createOwnership(networkDir string, enclave kurtosis.EnclaveRef) error {
	if err := validateOwnershipDirectory(networkDir, enclave); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(enclave, "", "  ")
	if err != nil {
		return err
	}
	return writeExclusive(ownershipPath(networkDir), append(payload, '\n'))
}

func validateOwnershipDirectory(networkDir string, enclave kurtosis.EnclaveRef) error {
	if err := enclave.Validate(); err != nil {
		return fmt.Errorf("ownership enclave identity is invalid: %w", err)
	}
	if !strings.HasPrefix(enclave.Name, enclaveNamePrefix(networkDir)) {
		return errors.New("ownership enclave name does not belong to this network directory")
	}
	return nil
}

func removeOwnership(networkDir string) error {
	return os.Remove(ownershipPath(networkDir))
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

func canonicalNetworkDirectory(path string) (string, error) {
	if path == "" || !filepath.IsAbs(path) {
		return "", errors.New("network directory must be an absolute path")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("network directory must be a directory")
	}
	if info.Mode().Perm() != 0o700 {
		return "", errors.New("network directory must have 0700 permissions")
	}
	return resolved, nil
}

func ensureNetworkDirectory(path string) (string, error) {
	if path == "" || !filepath.IsAbs(path) {
		return "", errors.New("network directory must be an absolute path")
	}
	if err := os.MkdirAll(path, 0o700); err != nil {
		return "", err
	}
	if err := os.Chmod(path, 0o700); err != nil {
		return "", err
	}
	return canonicalNetworkDirectory(path)
}
