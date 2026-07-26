// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEffectiveParametersUseRequestedExecutionAndFixedSupportImages(t *testing.T) {
	address := "Q" + strings.Repeat("a", 128)
	const executionImage = "local/go-qrl:test"
	payload, err := effectiveParameters(address, executionImage)
	if err != nil {
		t.Fatal(err)
	}
	var parameters map[string]any
	if err := json.Unmarshal([]byte(payload), &parameters); err != nil {
		t.Fatal(err)
	}
	participant := parameters["participants"].([]any)[0].(map[string]any)
	network := parameters["network_params"].(map[string]any)
	prefund := network["prefunded_accounts"].(map[string]any)[address].(map[string]any)
	serializedNetworkID, networkIDIsString := network["network_id"].(string)
	if participant["el_image"] != executionImage ||
		participant["cl_image"] != consensusImage ||
		participant["vc_image"] != validatorImage ||
		parameters["qrl_genesis_generator_params"].(map[string]any)["image"] != genesisImage ||
		!networkIDIsString ||
		serializedNetworkID != "1337" ||
		network["withdrawal_address"] != address ||
		prefund["balance"] != prefundBalance {
		t.Fatalf("parameters = %#v", parameters)
	}
	if _, exists := parameters["additional_services"]; exists {
		t.Fatalf("redundant additional_services was serialized: %#v", parameters)
	}
	if packageLocator != "github.com/rgeraldes24/qrl-package@3892c3d2596403c080424d9e8fc99ff172483fe0" {
		t.Fatalf("package locator = %q", packageLocator)
	}
}

func TestEffectiveParametersRejectInvalidInputs(t *testing.T) {
	address := "Q" + strings.Repeat("a", 128)
	if _, err := effectiveParameters("0x01", testExecutionImage); err == nil {
		t.Fatal("invalid wallet address was accepted")
	}
	if _, err := effectiveParameters(address, ""); err == nil {
		t.Fatal("empty execution image was accepted")
	}
}
