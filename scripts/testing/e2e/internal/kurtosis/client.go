// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package kurtosis provides the narrow Kurtosis API used by the E2E network
// controller. SDK types deliberately do not escape this package.
package kurtosis

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
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

type PackageRun struct {
	Locator          string
	SerializedParams string
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
