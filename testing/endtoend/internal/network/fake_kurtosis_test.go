// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"fmt"

	"github.com/theQRL/go-qrl/testing/endtoend/internal/kurtosis"
)

type fakeKurtosis struct {
	Enclave                                          kurtosis.EnclaveRef
	ExecutionService                                 kurtosis.Service
	CreateError, LookupError, RunError, DestroyError error
	CreateAfterError, DestroyAfterError, Exists      bool
	Creates, Lookups, Runs, Destroys                 int
	RunLocator, RunParameters                        string
	RunRef, ServiceRef, DestroyRef                   kurtosis.EnclaveRef
	RunStarted, RunRelease                           chan struct{}
}

func (fake *fakeKurtosis) CreateEnclave(_ context.Context, name string) (kurtosis.EnclaveRef, error) {
	fake.Creates++
	if fake.CreateError != nil {
		if fake.CreateAfterError && fake.Enclave.Name == "" {
			fake.Enclave.Name = name
		}
		fake.Exists = fake.CreateAfterError
		return kurtosis.EnclaveRef{}, fake.CreateError
	}
	if fake.Enclave.Name == "" {
		fake.Enclave.Name = name
	}
	fake.Exists = true
	return fake.Enclave, nil
}

func (fake *fakeKurtosis) LookupEnclave(ctx context.Context, name string) (kurtosis.EnclaveRef, bool, error) {
	fake.Lookups++
	if err := ctx.Err(); err != nil {
		return kurtosis.EnclaveRef{}, false, err
	}
	if fake.LookupError != nil {
		return kurtosis.EnclaveRef{}, false, fake.LookupError
	}
	if !fake.Exists || name != fake.Enclave.Name {
		return kurtosis.EnclaveRef{}, false, nil
	}
	return fake.Enclave, true, nil
}

func (fake *fakeKurtosis) RunRemotePackage(ctx context.Context, ref kurtosis.EnclaveRef, locator, parameters string) error {
	fake.Runs++
	fake.RunRef = ref
	fake.RunLocator, fake.RunParameters = locator, parameters
	if fake.RunStarted != nil {
		close(fake.RunStarted)
	}
	if fake.RunRelease != nil {
		select {
		case <-fake.RunRelease:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return fake.RunError
}

func (fake *fakeKurtosis) Service(_ context.Context, ref kurtosis.EnclaveRef, name string) (kurtosis.Service, error) {
	fake.ServiceRef = ref
	if name != executionServiceName {
		return kurtosis.Service{}, fmt.Errorf("service %q not found", name)
	}
	return fake.ExecutionService, nil
}

func (fake *fakeKurtosis) DestroyEnclave(_ context.Context, ref kurtosis.EnclaveRef) error {
	fake.Destroys++
	fake.DestroyRef = ref
	if fake.DestroyError != nil {
		if fake.DestroyAfterError {
			fake.Exists = false
		}
		return fake.DestroyError
	}
	fake.Exists = false
	return nil
}

var _ kurtosisClient = (*fakeKurtosis)(nil)
