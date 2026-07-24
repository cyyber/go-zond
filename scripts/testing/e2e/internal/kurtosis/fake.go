// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package kurtosis

import (
	"context"
	"errors"
	"fmt"
	"slices"
)

// Fake is a deterministic in-memory client used by network-controller tests.
// Calls makes it possible to prove that authentication never mutates network
// state.
type Fake struct {
	Enclave           EnclaveRef
	Runs              []PackageRun
	ServiceList       []Service
	Calls             []string
	CreateError       error
	CreateAfterError  bool
	GetError          error
	RunError          error
	DestroyError      error
	DestroyAfterError bool
	Destroyed         bool
}

func (fake *Fake) CreateEnclave(_ context.Context, name string) (EnclaveRef, error) {
	fake.Calls = append(fake.Calls, "create:"+name)
	if fake.CreateError != nil {
		if fake.CreateAfterError && fake.Enclave.Name == "" {
			uuid := fake.Enclave.UUID
			if uuid == "" {
				uuid = "0123456789abcdef0123456789abcdef"
			}
			fake.Enclave = EnclaveRef{Name: name, UUID: uuid, Owned: true}
		}
		return EnclaveRef{}, fake.CreateError
	}
	if fake.Enclave.Name == "" {
		uuid := fake.Enclave.UUID
		if uuid == "" {
			uuid = "0123456789abcdef0123456789abcdef"
		}
		fake.Enclave = EnclaveRef{Name: name, UUID: uuid, Owned: true}
	}
	return fake.Enclave, nil
}

func (fake *Fake) GetEnclave(_ context.Context, identifier string) (EnclaveRef, error) {
	fake.Calls = append(fake.Calls, "get:"+identifier)
	if fake.GetError != nil {
		return EnclaveRef{}, fake.GetError
	}
	if identifier != fake.Enclave.Name && identifier != fake.Enclave.UUID {
		return EnclaveRef{}, fmt.Errorf("enclave %q not found", identifier)
	}
	return fake.Enclave, nil
}

func (fake *Fake) RunRemotePackage(_ context.Context, ref EnclaveRef, run PackageRun) error {
	fake.Calls = append(fake.Calls, "run:"+ref.UUID)
	fake.Runs = append(fake.Runs, run)
	if fake.RunError != nil {
		return fake.RunError
	}
	return nil
}

func (fake *Fake) EnclaveExists(_ context.Context, uuid string) (bool, error) {
	fake.Calls = append(fake.Calls, "exists:"+uuid)
	return uuid == fake.Enclave.UUID && !fake.Destroyed, nil
}

func (fake *Fake) Services(_ context.Context, ref EnclaveRef) ([]Service, error) {
	fake.Calls = append(fake.Calls, "services:"+ref.UUID)
	return slices.Clone(fake.ServiceList), nil
}

func (fake *Fake) DestroyEnclave(_ context.Context, ref EnclaveRef) error {
	fake.Calls = append(fake.Calls, "destroy:"+ref.UUID)
	if !ref.Owned {
		return errors.New("refusing to destroy an enclave not marked as owned")
	}
	if fake.DestroyError != nil {
		if fake.DestroyAfterError {
			fake.Destroyed = true
		}
		return fake.DestroyError
	}
	fake.Destroyed = true
	return nil
}

var _ Client = (*Fake)(nil)
