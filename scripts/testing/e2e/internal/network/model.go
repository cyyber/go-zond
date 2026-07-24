// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package network owns the lifecycle and immutable identity of separately
// managed E2E networks. Suite execution receives only Authenticator.
package network

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

type Environment struct {
	RPCURL       string
	GraphQLURL   string
	WebSocketURL string
	SeedFile     string
}

type StartRequest struct {
	NetworkDir     string
	ExecutionImage string
	StartTimeout   time.Duration
}

// OwnershipRecord is the sole private lifecycle record. It always contains the
// exact enclave identity required for safe destruction.
type OwnershipRecord struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

func (record OwnershipRecord) Validate() error {
	if strings.TrimSpace(record.Name) == "" {
		return errors.New("ownership enclave name is empty")
	}
	if record.UUID == "" {
		return errors.New("ownership exact enclave UUID is empty")
	}
	enclave := kurtosis.EnclaveRef{Name: record.Name, UUID: record.UUID}
	if enclave.Validate() != nil {
		return errors.New("ownership enclave identity is invalid")
	}
	return nil
}

func (record OwnershipRecord) Enclave() (kurtosis.EnclaveRef, error) {
	if err := record.Validate(); err != nil {
		return kurtosis.EnclaveRef{}, err
	}
	return kurtosis.EnclaveRef{Name: record.Name, UUID: record.UUID}, nil
}

type Authenticator interface {
	Authenticate(context.Context, string) (Environment, error)
}

type Controller interface {
	Start(context.Context, StartRequest) error
	Status(context.Context, string) error
	Stop(context.Context, string) error
}
