// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package kurtosis provides the narrow Kurtosis API used by the development
// network controller. Raw SDK types deliberately do not escape this package.
package kurtosis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
)

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

const destroyConfirmationTimeout = 15 * time.Second

func NewSDKClient() (*SDKClient, error) {
	ctx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return nil, err
	}
	return &SDKClient{context: ctx}, nil
}

func (client *SDKClient) EnclaveExists(ctx context.Context, name string) (bool, error) {
	running, err := client.context.GetEnclaves(ctx)
	if err != nil {
		return false, fmt.Errorf("list running Kurtosis enclaves: %w", err)
	}
	_, found := running.GetEnclavesByName()[name]
	return found, nil
}

func (client *SDKClient) CreateAndRunRemotePackage(
	ctx context.Context,
	name string,
	locator,
	serializedParams string,
) error {
	enclave, err := client.context.CreateEnclave(ctx, name)
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

func (client *SDKClient) Service(ctx context.Context, enclaveName, serviceName string) (Service, error) {
	enclave, err := client.context.GetEnclaveContext(ctx, enclaveName)
	if err != nil {
		return Service{}, err
	}
	serviceContext, err := enclave.GetServiceContext(serviceName)
	if err != nil {
		return Service{}, err
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

func (client *SDKClient) DestroyEnclave(ctx context.Context, name string) error {
	destroyErr := client.context.DestroyEnclave(ctx, name)
	confirmCtx, cancel := context.WithTimeout(context.Background(), destroyConfirmationTimeout)
	defer cancel()
	enclaves, inspectErr := client.context.GetEnclaves(confirmCtx)
	if inspectErr != nil {
		return errors.Join(destroyErr, fmt.Errorf("confirm enclave destruction: %w", inspectErr))
	}
	_, exists := enclaves.GetEnclavesByName()[name]
	if exists {
		return errors.Join(destroyErr, errors.New("deterministic enclave slot remains occupied"))
	}
	return nil
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
