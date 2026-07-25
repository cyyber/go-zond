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

func privatePath(networkDir string) string { return filepath.Join(networkDir, "private") }
func ownershipPath(networkDir string) string {
	return filepath.Join(privatePath(networkDir), "ownership.json")
}

func loadOwnership(networkDir string) (kurtosis.EnclaveRef, error) {
	if err := validatePrivateDirectory(networkDir); err != nil {
		return kurtosis.EnclaveRef{}, err
	}
	enclave, err := loadJSON[kurtosis.EnclaveRef](ownershipPath(networkDir), "ownership")
	if err != nil {
		return kurtosis.EnclaveRef{}, err
	}
	return enclave, validateOwnershipDirectory(networkDir, enclave)
}

func createOwnership(networkDir string, enclave kurtosis.EnclaveRef) error {
	if err := validateOwnershipDirectory(networkDir, enclave); err != nil {
		return err
	}
	return writeJSONExclusive(ownershipPath(networkDir), enclave)
}

func validateOwnershipDirectory(networkDir string, enclave kurtosis.EnclaveRef) error {
	if enclave.Name == "" {
		return errors.New("ownership enclave name is empty")
	}
	if enclave.UUID == "" {
		return errors.New("ownership exact enclave UUID is empty")
	}
	if err := enclave.Validate(); err != nil {
		return fmt.Errorf("ownership enclave identity is invalid: %w", err)
	}
	if !strings.HasPrefix(enclave.Name, enclaveNamePrefix(networkDir)) {
		return errors.New("ownership enclave name does not belong to this network directory")
	}
	return nil
}

func removeOwnership(networkDir string) error {
	if err := os.Remove(ownershipPath(networkDir)); err != nil {
		return err
	}
	return syncDirectory(privatePath(networkDir))
}

func loadJSON[T any](path, description string) (T, error) {
	var value T
	payload, err := os.ReadFile(path)
	if err != nil {
		return value, err
	}
	if err := json.Unmarshal(payload, &value); err != nil {
		return value, fmt.Errorf("decode %s: %w", description, err)
	}
	return value, nil
}

func writeJSONExclusive(path string, value any) error {
	payload, err := jsonPayload(value)
	if err != nil {
		return err
	}
	return writeExclusive(path, payload)
}

func jsonPayload(value any) ([]byte, error) {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

func writeExclusive(path string, payload []byte) error {
	directory := filepath.Dir(path)
	file, err := os.CreateTemp(directory, "."+filepath.Base(path)+"-*")
	if err != nil {
		return err
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return err
	}
	if _, err := file.Write(payload); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Link(temporary, path); err != nil {
		return err
	}
	return syncDirectory(directory)
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func canonicalExistingDirectory(path, description string) (string, error) {
	if path == "" || !filepath.IsAbs(path) {
		return "", fmt.Errorf("%s must be an absolute path", description)
	}
	resolved, err := filepath.EvalSymlinks(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s must be a directory", description)
	}
	return filepath.Clean(resolved), nil
}

func ensureNetworkDirectory(path string) (string, error) {
	if path == "" || !filepath.IsAbs(path) {
		return "", errors.New("network directory must be an absolute path")
	}
	if err := os.MkdirAll(filepath.Clean(path), 0o700); err != nil {
		return "", err
	}
	networkDir, err := canonicalExistingDirectory(path, "network directory")
	if err != nil {
		return "", err
	}
	if err := os.Chmod(networkDir, 0o700); err != nil {
		return "", err
	}
	privateDir := privatePath(networkDir)
	if info, err := os.Lstat(privateDir); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return "", errors.New("private network state must be a non-symlink directory")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	} else if err := os.Mkdir(privateDir, 0o700); err != nil {
		return "", err
	}
	if err := os.Chmod(privateDir, 0o700); err != nil {
		return "", err
	}
	return networkDir, nil
}

func validatePrivateDirectory(networkDir string) error {
	info, err := os.Lstat(privatePath(networkDir))
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || info.Mode().Perm() != 0o700 {
		return errors.New("private network state must be a non-symlink 0700 directory")
	}
	return nil
}
