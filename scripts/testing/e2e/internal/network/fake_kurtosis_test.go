// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"fmt"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/kurtosis"
)

type fakeKurtosis struct {
	Enclave           kurtosis.EnclaveRef
	Runs              []packageRun
	ExecutionService  kurtosis.Service
	Calls             []string
	CreateError       error
	CreateAfterError  bool
	CreateCallback    func()
	GetFailures       int
	RunError          error
	DestroyError      error
	DestroyAfterError bool
	Destroyed         bool
}

type packageRun struct {
	Locator, SerializedParams string
}

func (fake *fakeKurtosis) CreateEnclave(_ context.Context, name string) (kurtosis.EnclaveRef, error) {
	fake.Calls = append(fake.Calls, "create:"+name)
	if fake.CreateCallback != nil {
		fake.CreateCallback()
	}
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

func (fake *fakeKurtosis) GetEnclave(ctx context.Context, identifier string) (kurtosis.EnclaveRef, error) {
	fake.Calls = append(fake.Calls, "get:"+identifier)
	if err := ctx.Err(); err != nil {
		return kurtosis.EnclaveRef{}, err
	}
	if fake.GetFailures > 0 {
		fake.GetFailures--
		return kurtosis.EnclaveRef{}, fmt.Errorf("enclave %q not visible yet", identifier)
	}
	if identifier != fake.Enclave.Name && identifier != fake.Enclave.UUID {
		return kurtosis.EnclaveRef{}, fmt.Errorf("enclave %q not found", identifier)
	}
	return fake.Enclave, nil
}

func (fake *fakeKurtosis) RunRemotePackage(
	_ context.Context,
	ref kurtosis.EnclaveRef,
	locator,
	parameters string,
) error {
	fake.Calls = append(fake.Calls, "run:"+ref.UUID)
	fake.Runs = append(fake.Runs, packageRun{
		Locator: locator, SerializedParams: parameters,
	})
	return fake.RunError
}

func (fake *fakeKurtosis) Service(
	_ context.Context,
	ref kurtosis.EnclaveRef,
	name string,
) (kurtosis.Service, error) {
	fake.Calls = append(fake.Calls, "service:"+ref.UUID+":"+name)
	if name != executionServiceName {
		return kurtosis.Service{}, fmt.Errorf("service %q not found", name)
	}
	return fake.ExecutionService, nil
}

func (fake *fakeKurtosis) DestroyEnclave(_ context.Context, ref kurtosis.EnclaveRef) error {
	fake.Calls = append(fake.Calls, "destroy:"+ref.UUID)
	if fake.DestroyError != nil {
		if fake.DestroyAfterError {
			fake.Destroyed = true
		}
		if fake.Destroyed {
			return nil
		}
		return fake.DestroyError
	}
	fake.Destroyed = true
	return nil
}

var _ kurtosisClient = (*fakeKurtosis)(nil)
