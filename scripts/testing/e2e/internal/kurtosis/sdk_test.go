// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package kurtosis

import (
	"strings"
	"testing"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	kurtosisservices "github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
)

func TestConvertServiceContextPreservesIdentityAndPublicEndpoints(t *testing.T) {
	serviceContext := kurtosisservices.NewServiceContext(
		nil,
		"execution",
		"11111111111111111111111111111111",
		"10.0.0.1",
		nil,
		"127.0.0.1",
		map[string]*kurtosisservices.PortSpec{
			"rpc": kurtosisservices.NewPortSpec(18545, kurtosisservices.TransportProtocol_TCP, ""),
		},
		nil,
		false,
		nil,
		false,
	)
	service, err := convertServiceContext(serviceContext)
	if err != nil {
		t.Fatal(err)
	}
	if service.Name != "execution" ||
		service.UUID != "11111111111111111111111111111111" ||
		service.PublicIP != "127.0.0.1" ||
		service.PublicPorts["rpc"].Number != 18545 {
		t.Fatalf("service = %+v", service)
	}
}

func TestConvertServiceContextRejectsNil(t *testing.T) {
	if _, err := convertServiceContext(nil); err == nil {
		t.Fatal("nil service context was accepted")
	}
}

func TestConsumeStarlarkCompletionSuppressesSecretBearingTranscript(t *testing.T) {
	const secret = "seed-that-must-never-reach-errors"
	stream := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, 2)
	stream <- binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(secret, 1, 2)
	stream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
	close(stream)
	err := consumeStarlarkCompletion(stream)
	if err == nil || strings.Contains(err.Error(), secret) {
		t.Fatalf("secret-bearing error = %v", err)
	}

	incomplete := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, 1)
	incomplete <- binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(secret, 1, 2)
	close(incomplete)
	err = consumeStarlarkCompletion(incomplete)
	if err == nil || strings.Contains(err.Error(), secret) || !strings.Contains(err.Error(), "without a terminal event") {
		t.Fatalf("incomplete error = %v", err)
	}

	success := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, 1)
	success <- binding_constructors.NewStarlarkRunResponseLineFromRunSuccessEvent(secret, time.Second)
	close(success)
	if err := consumeStarlarkCompletion(success); err != nil {
		t.Fatal(err)
	}
}
