// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package network owns the lifecycle and immutable identity of separately
// managed E2E networks. Suite execution receives only Authenticator.
package network

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

// State is the sanitized public ready marker. Runtime details remain derived
// from the exact owned enclave.
type State struct {
	Ready bool `json:"ready"`
}

func (state State) Validate() error {
	if !state.Ready {
		return errors.New("network is not ready")
	}
	return nil
}

type Result struct {
	topology runtimeTopology
	Ready    bool   `json:"ready"`
	Message  string `json:"message,omitempty"`
}

type Environment struct {
	RPCURL       string
	GraphQLURL   string
	WebSocketURL string
	SeedFile     string
}

type StartRequest struct {
	RepoRoot     string
	NetworkDir   string
	DockerBin    string
	StartTimeout time.Duration
}

// OwnershipRecord is the sole private lifecycle record. A name-only creation
// intent prevents replay if Kurtosis loses the create response; the exact UUID
// is captured as soon as creation returns and retained until destruction is
// confirmed.
type OwnershipRecord struct {
	NetworkDir string `json:"network_dir"`
	Name       string `json:"name"`
	UUID       string `json:"uuid,omitempty"`
}

func (record OwnershipRecord) Validate() error {
	if !filepath.IsAbs(record.NetworkDir) || filepath.Clean(record.NetworkDir) != record.NetworkDir {
		return errors.New("invalid ownership directory")
	}
	if strings.TrimSpace(record.Name) == "" {
		return errors.New("ownership enclave name is empty")
	}
	if record.UUID == "" {
		return nil
	}
	enclave := kurtosis.EnclaveRef{Name: record.Name, UUID: record.UUID, Owned: true}
	if enclave.Validate() != nil {
		return errors.New("ownership enclave identity is invalid")
	}
	return nil
}

func (record OwnershipRecord) OwnedEnclave() (kurtosis.EnclaveRef, error) {
	if err := record.Validate(); err != nil {
		return kurtosis.EnclaveRef{}, err
	}
	if record.UUID == "" {
		return kurtosis.EnclaveRef{}, fmt.Errorf(
			"enclave creation outcome for %q is ambiguous: exact UUID was not captured",
			record.Name,
		)
	}
	return kurtosis.EnclaveRef{Name: record.Name, UUID: record.UUID, Owned: true}, nil
}

type Authenticator interface {
	Authenticate(context.Context, string) (Environment, error)
}

type Controller interface {
	Start(context.Context, StartRequest) (Result, error)
	Status(context.Context, string) (Result, error)
	Stop(context.Context, string) (Result, error)
}
