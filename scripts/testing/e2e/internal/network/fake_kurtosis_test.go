// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"fmt"
	"slices"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

type fakeKurtosis struct {
	Enclave           kurtosis.EnclaveRef
	Runs              []kurtosis.PackageRun
	ServiceList       []kurtosis.Service
	Calls             []string
	CreateError       error
	CreateAfterError  bool
	RunError          error
	DestroyError      error
	DestroyAfterError bool
	Destroyed         bool
}

func (fake *fakeKurtosis) CreateEnclave(_ context.Context, name string) (kurtosis.EnclaveRef, error) {
	fake.Calls = append(fake.Calls, "create:"+name)
	if fake.CreateError != nil {
		if fake.CreateAfterError && fake.Enclave.Name == "" {
			fake.Enclave.Name = name
		}
		return kurtosis.EnclaveRef{}, fake.CreateError
	}
	if fake.Enclave.Name == "" {
		fake.Enclave.Name = name
	}
	return fake.Enclave, nil
}

func (fake *fakeKurtosis) GetEnclave(_ context.Context, identifier string) (kurtosis.EnclaveRef, error) {
	fake.Calls = append(fake.Calls, "get:"+identifier)
	if identifier != fake.Enclave.Name && identifier != fake.Enclave.UUID {
		return kurtosis.EnclaveRef{}, fmt.Errorf("enclave %q not found", identifier)
	}
	return fake.Enclave, nil
}

func (fake *fakeKurtosis) RunRemotePackage(
	_ context.Context,
	ref kurtosis.EnclaveRef,
	run kurtosis.PackageRun,
) error {
	fake.Calls = append(fake.Calls, "run:"+ref.UUID)
	fake.Runs = append(fake.Runs, run)
	return fake.RunError
}

func (fake *fakeKurtosis) EnclaveExists(_ context.Context, uuid string) (bool, error) {
	fake.Calls = append(fake.Calls, "exists:"+uuid)
	return uuid == fake.Enclave.UUID && !fake.Destroyed, nil
}

func (fake *fakeKurtosis) Services(
	_ context.Context,
	ref kurtosis.EnclaveRef,
) ([]kurtosis.Service, error) {
	fake.Calls = append(fake.Calls, "services:"+ref.UUID)
	return slices.Clone(fake.ServiceList), nil
}

func (fake *fakeKurtosis) DestroyEnclave(_ context.Context, ref kurtosis.EnclaveRef) error {
	fake.Calls = append(fake.Calls, "destroy:"+ref.UUID)
	if fake.DestroyError != nil {
		if fake.DestroyAfterError {
			fake.Destroyed = true
		}
		return fake.DestroyError
	}
	fake.Destroyed = true
	return nil
}

var _ kurtosisClient = (*fakeKurtosis)(nil)
