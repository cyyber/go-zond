// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"fmt"

	"github.com/theQRL/go-qrl/devnet/internal/kurtosis"
)

type fakeKurtosis struct {
	Enclave                        kurtosis.EnclaveRef
	ExecutionService               kurtosis.Service
	Exists                         bool
	RunLocator, RunParameters      string
	RunRef, ServiceRef, DestroyRef kurtosis.EnclaveRef
}

func (fake *fakeKurtosis) CreateEnclave(_ context.Context, name string) (kurtosis.EnclaveRef, error) {
	if fake.Enclave.Name == "" {
		fake.Enclave.Name = name
	}
	fake.Exists = true
	return fake.Enclave, nil
}

func (fake *fakeKurtosis) LookupEnclave(ctx context.Context, name string) (kurtosis.EnclaveRef, bool, error) {
	if err := ctx.Err(); err != nil {
		return kurtosis.EnclaveRef{}, false, err
	}
	if !fake.Exists || name != fake.Enclave.Name {
		return kurtosis.EnclaveRef{}, false, nil
	}
	return fake.Enclave, true, nil
}

func (fake *fakeKurtosis) RunRemotePackage(_ context.Context, ref kurtosis.EnclaveRef, locator, parameters string) error {
	fake.RunRef = ref
	fake.RunLocator, fake.RunParameters = locator, parameters
	return nil
}

func (fake *fakeKurtosis) Service(_ context.Context, ref kurtosis.EnclaveRef, name string) (kurtosis.Service, error) {
	fake.ServiceRef = ref
	if name != executionServiceName {
		return kurtosis.Service{}, fmt.Errorf("service %q not found", name)
	}
	return fake.ExecutionService, nil
}

func (fake *fakeKurtosis) DestroyEnclave(_ context.Context, ref kurtosis.EnclaveRef) error {
	fake.DestroyRef = ref
	fake.Exists = false
	return nil
}

var _ kurtosisClient = (*fakeKurtosis)(nil)
