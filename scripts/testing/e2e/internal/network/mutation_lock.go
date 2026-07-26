// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"errors"
	"fmt"
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
