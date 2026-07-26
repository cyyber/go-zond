// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEffectiveParametersUseRequestedExecutionAndFixedSupportImages(t *testing.T) {
	address := "Q" + strings.Repeat("a", 128)
	const executionImage = "local/go-qrl:test"
	payload, err := effectiveParameters(address, executionImage)
	require.NoError(t, err)

	var parameters map[string]any
	require.NoError(t, json.Unmarshal([]byte(payload), &parameters))

	participant := parameters["participants"].([]any)[0].(map[string]any)
	network := parameters["network_params"].(map[string]any)
	prefund := network["prefunded_accounts"].(map[string]any)[address].(map[string]any)
	require.Equal(t, executionImage, participant["el_image"])
	require.Equal(t, consensusImage, participant["cl_image"])
	require.Equal(t, validatorImage, participant["vc_image"])
	require.Equal(t, genesisImage, parameters["qrl_genesis_generator_params"].(map[string]any)["image"])
	require.Equal(t, "1337", network["network_id"])
	require.Equal(t, address, network["withdrawal_address"])
	require.Equal(t, prefundBalance, prefund["balance"])
	require.Regexp(t, `^github\.com/rgeraldes24/qrl-package@[0-9a-f]{40}$`, packageLocator)

	for _, key := range []string{
		"el_type",
		"cl_type",
		"vc_type",
		"count",
		"use_remote_signer",
	} {
		require.NotContains(t, participant, key)
	}
	for _, key := range []string{"preset", "slots_per_epoch"} {
		require.NotContains(t, network, key)
	}
}

func TestEffectiveParametersRejectInvalidInputs(t *testing.T) {
	address := "Q" + strings.Repeat("a", 128)
	_, err := effectiveParameters("0x01", testExecutionImage)
	require.Error(t, err)
	_, err = effectiveParameters(address, "")
	require.Error(t, err)
}
