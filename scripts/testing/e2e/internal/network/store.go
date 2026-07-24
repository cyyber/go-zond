// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/renameio"
)

func statePath(networkDir string) string   { return filepath.Join(networkDir, "network.json") }
func privatePath(networkDir string) string { return filepath.Join(networkDir, "private") }
func ownershipPath(networkDir string) string {
	return filepath.Join(privatePath(networkDir), "ownership.json")
}

func loadOwnership(networkDir string) (OwnershipRecord, error) {
	if err := validatePrivateDirectory(networkDir); err != nil {
		return OwnershipRecord{}, err
	}
	record, err := loadJSON[OwnershipRecord](ownershipPath(networkDir), "ownership")
	if err != nil {
		return OwnershipRecord{}, err
	}
	if record.NetworkDir != networkDir {
		return OwnershipRecord{}, errors.New("ownership belongs to another network directory")
	}
	return record, record.Validate()
}

func createOwnership(record OwnershipRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}
	if record.UUID != "" {
		return errors.New("ownership must begin as a creation intent")
	}
	return writeJSONExclusive(ownershipPath(record.NetworkDir), record)
}

func captureOwnership(record OwnershipRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}
	if record.UUID == "" {
		return errors.New("captured ownership has no exact enclave UUID")
	}
	return writeJSONAtomic(ownershipPath(record.NetworkDir), record)
}

func removeOwnership(record OwnershipRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}
	if err := os.Remove(ownershipPath(record.NetworkDir)); err != nil {
		return err
	}
	return syncDirectory(privatePath(record.NetworkDir))
}

func loadState(networkDir string) (State, error) {
	state, err := loadJSON[State](statePath(networkDir), "network state")
	if errors.Is(err, os.ErrNotExist) {
		return State{}, fmt.Errorf("start the independent E2E network first: %w", err)
	}
	if err != nil {
		return State{}, err
	}
	return state, state.Validate()
}

func writeState(networkDir string, state State) error {
	if err := state.Validate(); err != nil {
		return err
	}
	return writeJSONAtomic(statePath(networkDir), state)
}

func removeState(networkDir string) error {
	if err := os.Remove(statePath(networkDir)); err != nil {
		return err
	}
	return syncDirectory(networkDir)
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

func writeJSONAtomic(path string, value any) error {
	payload, err := jsonPayload(value)
	if err != nil {
		return err
	}
	if err := renameio.WriteFile(path, payload, 0o600); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(path))
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
