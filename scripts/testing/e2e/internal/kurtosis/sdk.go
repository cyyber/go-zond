// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package kurtosis

import (
	"context"
	"errors"
	"fmt"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	kurtosisservices "github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
)

type SDKClient struct {
	context *kurtosis_context.KurtosisContext
}

func NewSDKClient() (*SDKClient, error) {
	ctx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return nil, err
	}
	return &SDKClient{context: ctx}, nil
}

func (client *SDKClient) CreateEnclave(ctx context.Context, name string) (EnclaveRef, error) {
	enclave, err := client.context.CreateEnclave(ctx, name)
	if err != nil {
		return EnclaveRef{}, err
	}
	ref := EnclaveRef{Name: enclave.GetEnclaveName(), UUID: string(enclave.GetEnclaveUuid())}
	if err := ref.Validate(); err != nil {
		return EnclaveRef{}, fmt.Errorf("Kurtosis returned invalid enclave identity: %w", err)
	}
	return ref, nil
}

func (client *SDKClient) GetEnclave(ctx context.Context, identifier string) (EnclaveRef, error) {
	info, err := client.context.GetEnclave(ctx, identifier)
	if err != nil {
		return EnclaveRef{}, err
	}
	ref := EnclaveRef{Name: info.GetName(), UUID: info.GetEnclaveUuid()}
	if err := ref.Validate(); err != nil {
		return EnclaveRef{}, fmt.Errorf("Kurtosis returned invalid enclave identity: %w", err)
	}
	return ref, nil
}

func (client *SDKClient) EnclaveExists(ctx context.Context, uuid string) (bool, error) {
	if err := (&EnclaveRef{Name: "identity-check", UUID: uuid}).Validate(); err != nil {
		return false, err
	}
	enclaves, err := client.context.GetEnclaves(ctx)
	if err != nil {
		return false, err
	}
	_, exists := enclaves.GetEnclavesByUuid()[uuid]
	return exists, nil
}

func (client *SDKClient) RunRemotePackage(ctx context.Context, ref EnclaveRef, run PackageRun) error {
	enclave, err := client.enclaveContext(ctx, ref)
	if err != nil {
		return err
	}
	configuration := starlark_run_config.NewRunStarlarkConfig(starlark_run_config.WithSerializedParams(run.SerializedParams))
	stream, cancel, err := enclave.RunStarlarkRemotePackage(ctx, run.Locator, configuration)
	if err != nil {
		return err
	}
	defer cancel()
	// qrl-package output can contain generated seed material. Completion is all
	// the network controller needs, so raw serialized output never escapes this
	// SDK boundary.
	return consumeStarlarkCompletion(stream)
}

func (client *SDKClient) Services(ctx context.Context, ref EnclaveRef) (map[string]Service, error) {
	enclave, err := client.enclaveContext(ctx, ref)
	if err != nil {
		return nil, err
	}
	contexts, err := enclave.GetServiceContexts(map[string]bool{})
	if err != nil {
		return nil, err
	}
	// Kurtosis v1.20 GetServiceContexts does not accept a context. Checking
	// again prevents publishing stale results when the caller was cancelled
	// while the SDK request was in flight.
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result := make(map[string]Service, len(contexts))
	for name, serviceContext := range contexts {
		service, err := convertServiceContext(serviceContext)
		if err != nil {
			return nil, err
		}
		result[string(name)] = service
	}
	return result, nil
}

func (client *SDKClient) DestroyEnclave(ctx context.Context, ref EnclaveRef) error {
	if err := ref.Validate(); err != nil {
		return err
	}
	current, err := client.GetEnclave(ctx, ref.UUID)
	if err != nil {
		return err
	}
	if current.Name != ref.Name || current.UUID != ref.UUID {
		return fmt.Errorf("enclave identity changed: got %s/%s, want %s/%s", current.Name, current.UUID, ref.Name, ref.UUID)
	}
	return client.context.DestroyEnclave(ctx, ref.UUID)
}

func (client *SDKClient) enclaveContext(ctx context.Context, ref EnclaveRef) (*enclaves.EnclaveContext, error) {
	if err := ref.Validate(); err != nil {
		return nil, err
	}
	current, err := client.context.GetEnclaveContext(ctx, ref.UUID)
	if err != nil {
		return nil, err
	}
	if string(current.GetEnclaveUuid()) != ref.UUID || current.GetEnclaveName() != ref.Name {
		return nil, fmt.Errorf("enclave identity changed: got %s/%s, want %s/%s", current.GetEnclaveName(), current.GetEnclaveUuid(), ref.Name, ref.UUID)
	}
	return current, nil
}

func consumeStarlarkCompletion(stream <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine) error {
	for line := range stream {
		if finished := line.GetRunFinishedEvent(); finished != nil {
			if !finished.GetIsRunSuccessful() {
				return errors.New("Kurtosis Starlark package run failed; response content suppressed")
			}
			return nil
		}
	}
	return errors.New("Kurtosis Starlark response stream closed without a terminal event; response content suppressed")
}

func convertServiceContext(serviceContext *kurtosisservices.ServiceContext) (Service, error) {
	if serviceContext == nil {
		return Service{}, errors.New("Kurtosis GetServiceContexts returned a nil service context")
	}
	return Service{
		PublicIP:    serviceContext.GetMaybePublicIPAddress(),
		PublicPorts: convertSDKPorts(serviceContext.GetPublicPorts()),
	}, nil
}

func convertSDKPorts(ports map[string]*kurtosisservices.PortSpec) map[string]uint16 {
	result := make(map[string]uint16, len(ports))
	for id, port := range ports {
		result[id] = uint16(port.GetNumber())
	}
	return result
}
