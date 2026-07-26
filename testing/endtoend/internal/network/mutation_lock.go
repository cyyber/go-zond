// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

// AcquireMutationLease takes the non-blocking cross-process lock shared by
// network lifecycle commands and live suites.
func AcquireMutationLease(networkDir string) (*flock.Flock, error) {
	canonical, err := canonicalNetworkDirectory(networkDir)
	if err != nil {
		return nil, err
	}
	lock := flock.New(filepath.Join(canonical, "mutation.lock"))
	locked, err := lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("acquire network mutation lease: %w", err)
	}
	if !locked {
		_ = lock.Close()
		return nil, errors.New("network mutation is already in progress")
	}
	return lock, nil
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
	return canonicalNetworkDirectory(path)
}
