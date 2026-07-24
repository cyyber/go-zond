// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/gofrs/flock"
)

// MutationLease is the non-blocking cross-process lock shared by network
// lifecycle commands and live suites.
type MutationLease struct {
	networkDir string
	lock       *flock.Flock
}

func AcquireMutationLease(networkDir string) (*MutationLease, error) {
	canonical, err := canonicalExistingDirectory(networkDir, "network directory")
	if err != nil {
		return nil, err
	}
	return acquireMutationLease(canonical)
}

func acquireMutationLease(networkDir string) (*MutationLease, error) {
	if err := validatePrivateDirectory(networkDir); err != nil {
		return nil, err
	}
	lock := flock.New(filepath.Join(privatePath(networkDir), "mutation.lock"))
	locked, err := lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("acquire network mutation lease: %w", err)
	}
	if !locked {
		_ = lock.Close()
		return nil, errors.New("network mutation is already in progress")
	}
	return &MutationLease{networkDir: networkDir, lock: lock}, nil
}

func (lease *MutationLease) NetworkDir() string {
	if lease == nil {
		return ""
	}
	return lease.networkDir
}

func (lease *MutationLease) Close() error {
	if lease == nil || lease.lock == nil {
		return nil
	}
	return lease.lock.Close()
}
