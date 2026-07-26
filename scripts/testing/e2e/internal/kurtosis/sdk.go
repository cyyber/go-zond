// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package kurtosis provides the narrow Kurtosis API used by the E2E network
// controller. Raw SDK types deliberately do not escape this package.
package kurtosis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

type EnclaveRef struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

func (ref EnclaveRef) Validate() error {
	if ref.Name == "" {
		return errors.New("enclave name is empty")
	}
	if !uuidPattern.MatchString(ref.UUID) {
		return fmt.Errorf("enclave UUID %q is not a full 32-character lowercase UUID", ref.UUID)
	}
	return nil
}

type Service struct {
	PublicIP    string
	PublicPorts map[string]uint16
}

func (service Service) PublicEndpoint(portID, scheme string) (string, bool) {
	port, ok := service.PublicPorts[portID]
	if !ok || service.PublicIP == "" || port == 0 {
		return "", false
	}
	return scheme + "://" + net.JoinHostPort(service.PublicIP, strconv.Itoa(int(port))), true
}

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

func (client *SDKClient) RunRemotePackage(
	ctx context.Context,
	ref EnclaveRef,
	locator,
	serializedParams string,
) error {
	enclave, err := client.enclaveContext(ctx, ref)
	if err != nil {
		return err
	}
	configuration := starlark_run_config.NewRunStarlarkConfig(starlark_run_config.WithSerializedParams(serializedParams))
	stream, cancel, err := enclave.RunStarlarkRemotePackage(ctx, locator, configuration)
	if err != nil {
		return err
	}
	defer cancel()
	// qrl-package output can contain generated seed material. Completion is all
	// the network controller needs, so raw serialized output never escapes this
	// SDK boundary.
	return consumeStarlarkCompletion(stream)
}

func (client *SDKClient) Service(ctx context.Context, ref EnclaveRef, name string) (Service, error) {
	enclave, err := client.enclaveContext(ctx, ref)
	if err != nil {
		return Service{}, err
	}
	serviceContext, err := enclave.GetServiceContext(name)
	if err != nil {
		return Service{}, err
	}
	// Kurtosis v1.20 GetServiceContext does not accept a context. Checking
	// again prevents publishing stale results when the caller was cancelled
	// while the SDK request was in flight.
	if err := ctx.Err(); err != nil {
		return Service{}, err
	}
	if serviceContext == nil {
		return Service{}, errors.New("Kurtosis GetServiceContext returned a nil service context")
	}
	ports := make(map[string]uint16, len(serviceContext.GetPublicPorts()))
	for id, port := range serviceContext.GetPublicPorts() {
		ports[id] = port.GetNumber()
	}
	return Service{
		PublicIP:    serviceContext.GetMaybePublicIPAddress(),
		PublicPorts: ports,
	}, nil
}

func (client *SDKClient) DestroyEnclave(ctx context.Context, ref EnclaveRef) error {
	if err := ref.Validate(); err != nil {
		return err
	}
	var destroyErr error
	current, err := client.GetEnclave(ctx, ref.UUID)
	switch {
	case err != nil:
		destroyErr = err
	case current.Name != ref.Name || current.UUID != ref.UUID:
		destroyErr = fmt.Errorf(
			"enclave identity changed: got %s/%s, want %s/%s",
			current.Name,
			current.UUID,
			ref.Name,
			ref.UUID,
		)
	default:
		destroyErr = client.context.DestroyEnclave(ctx, ref.UUID)
	}
	enclaves, inspectErr := client.context.GetEnclaves(ctx)
	var exists bool
	if inspectErr == nil {
		_, exists = enclaves.GetEnclavesByUuid()[ref.UUID]
	}
	return reconcileDestroy(destroyErr, inspectErr, exists)
}

func reconcileDestroy(destroyErr, inspectErr error, exists bool) error {
	if inspectErr != nil {
		return errors.Join(destroyErr, fmt.Errorf("confirm owned enclave destruction: %w", inspectErr))
	}
	if exists {
		return errors.Join(destroyErr, errors.New("owned enclave still exists after destruction"))
	}
	return nil
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
