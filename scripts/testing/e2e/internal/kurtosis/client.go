// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package kurtosis provides the narrow Kurtosis API used by the E2E network
// controller. SDK types deliberately do not escape this package.
package kurtosis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

type EnclaveRef struct {
	Name  string `json:"name"`
	UUID  string `json:"uuid"`
	Owned bool   `json:"owned"`
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

type PackageRun struct {
	Locator          string
	SerializedParams string
}

type Port struct {
	Number uint16
}

type ServiceStatus string

const (
	ServiceStatusRunning ServiceStatus = "RUNNING"
	ServiceStatusStopped ServiceStatus = "STOPPED"
	ServiceStatusUnknown ServiceStatus = "UNKNOWN"
)

type Service struct {
	Name        string
	UUID        string
	Status      ServiceStatus
	PublicIP    string
	PublicPorts map[string]Port
}

func (service Service) PublicEndpoint(portID, scheme string) (string, bool) {
	port, ok := service.PublicPorts[portID]
	if !ok || service.PublicIP == "" || port.Number == 0 {
		return "", false
	}
	return scheme + "://" + net.JoinHostPort(service.PublicIP, strconv.Itoa(int(port.Number))), true
}

type Client interface {
	CreateEnclave(context.Context, string) (EnclaveRef, error)
	GetEnclave(context.Context, string) (EnclaveRef, error)
	EnclaveExists(context.Context, string) (bool, error)
	RunRemotePackage(context.Context, EnclaveRef, PackageRun) error
	Services(context.Context, EnclaveRef) ([]Service, error)
	DestroyEnclave(context.Context, EnclaveRef) error
}
