// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package kurtosis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const maxAPIContainerResponseBytes = 100 * 1024 * 1024

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
	ref := EnclaveRef{Name: enclave.GetEnclaveName(), UUID: string(enclave.GetEnclaveUuid()), Owned: true}
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

func (client *SDKClient) Services(ctx context.Context, ref EnclaveRef) ([]Service, error) {
	infos, err := client.serviceInfos(ctx, ref)
	if err != nil {
		return nil, err
	}
	result := make([]Service, 0, len(infos))
	for _, info := range infos {
		service, err := convertServiceInfo(info)
		if err != nil {
			return nil, err
		}
		result = append(result, service)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func (client *SDKClient) DestroyEnclave(ctx context.Context, ref EnclaveRef) error {
	if !ref.Owned {
		return errors.New("refusing to destroy an enclave not marked as owned")
	}
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

func (client *SDKClient) serviceInfos(ctx context.Context, ref EnclaveRef) (map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := ref.Validate(); err != nil {
		return nil, err
	}
	info, err := client.context.GetEnclave(ctx, ref.UUID)
	if err != nil {
		return nil, err
	}
	if info.GetEnclaveUuid() != ref.UUID || info.GetName() != ref.Name {
		return nil, fmt.Errorf("enclave identity changed: got %s/%s, want %s/%s", info.GetName(), info.GetEnclaveUuid(), ref.Name, ref.UUID)
	}
	host := info.GetApiContainerHostMachineInfo()
	if host == nil || host.GetIpOnHostMachine() == "" || host.GetGrpcPortOnHostMachine() == 0 {
		return nil, fmt.Errorf("enclave %s/%s has no API-container host endpoint", ref.Name, ref.UUID)
	}
	endpoint := net.JoinHostPort(host.GetIpOnHostMachine(), strconv.FormatUint(uint64(host.GetGrpcPortOnHostMachine()), 10))
	connection, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxAPIContainerResponseBytes)))
	if err != nil {
		return nil, fmt.Errorf("connect to Kurtosis API container %s: %w", endpoint, err)
	}
	defer connection.Close()
	response, err := kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient(connection).GetServices(ctx, binding_constructors.NewGetServicesArgs(map[string]bool{}))
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("Kurtosis GetServices returned a nil response")
	}
	return response.GetServiceInfo(), nil
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

func convertServiceInfo(info *kurtosis_core_rpc_api_bindings.ServiceInfo) (Service, error) {
	if info == nil {
		return Service{}, errors.New("Kurtosis GetServices returned nil service info")
	}
	serviceStatus, err := convertServiceStatus(info.GetServiceStatus())
	if err != nil {
		return Service{}, fmt.Errorf("service %q/%q: %w", info.GetName(), info.GetServiceUuid(), err)
	}
	return Service{
		Name: info.GetName(), UUID: info.GetServiceUuid(), Status: serviceStatus,
		PublicIP: info.GetMaybePublicIpAddr(), PublicPorts: convertAPIPorts(info.GetMaybePublicPorts()),
	}, nil
}

func convertServiceStatus(value kurtosis_core_rpc_api_bindings.ServiceStatus) (ServiceStatus, error) {
	switch value {
	case kurtosis_core_rpc_api_bindings.ServiceStatus_RUNNING:
		return ServiceStatusRunning, nil
	case kurtosis_core_rpc_api_bindings.ServiceStatus_STOPPED:
		return ServiceStatusStopped, nil
	case kurtosis_core_rpc_api_bindings.ServiceStatus_UNKNOWN:
		return ServiceStatusUnknown, nil
	default:
		return "", fmt.Errorf("Kurtosis returned unsupported service status %d", value)
	}
}

func convertAPIPorts(ports map[string]*kurtosis_core_rpc_api_bindings.Port) map[string]Port {
	result := make(map[string]Port, len(ports))
	for id, port := range ports {
		result[id] = Port{Number: uint16(port.GetNumber())}
	}
	return result
}

var _ Client = (*SDKClient)(nil)
