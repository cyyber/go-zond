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
	payload, err := effectiveParameters(address, executionImage, nil)
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

func TestCustomParametersReplaceOnlyExactJSONTokens(t *testing.T) {
	address := "Q" + strings.Repeat("b", 128)
	custom := []byte(`{
		"participants":[{"el_image":"__DEVNET_EXECUTION_IMAGE__","custom":9007199254740993}],
		"network_params":{
			"prefunded_accounts":{"__DEVNET_WALLET_ADDRESS__":{"balance":"1QRL"}},
			"withdrawal_address":"__DEVNET_WALLET_ADDRESS__"
		},
		"untouched":"prefix-__DEVNET_EXECUTION_IMAGE__"
	}`)
	rendered, err := effectiveParameters(address, "registry.example/qrl:test", custom)
	require.NoError(t, err)
	require.Contains(t, rendered, `"custom":9007199254740993`)
	require.Contains(t, rendered, `"el_image":"registry.example/qrl:test"`)
	require.Contains(t, rendered, `"`+address+`":{"balance":"1QRL"}`)
	require.Contains(t, rendered, `"withdrawal_address":"`+address+`"`)
	require.Contains(t, rendered, `"untouched":"prefix-__DEVNET_EXECUTION_IMAGE__"`)
}

func TestCustomParametersValidateContract(t *testing.T) {
	address := "Q" + strings.Repeat("c", 128)
	for name, custom := range map[string][]byte{
		"malformed":       []byte(`{`),
		"missing image":   []byte(`{"participants":[{"el_image":"image"}],"network_params":{"prefunded_accounts":{"__DEVNET_WALLET_ADDRESS__":{}}}}`),
		"missing wallet":  []byte(`{"participants":[{"el_image":"__DEVNET_EXECUTION_IMAGE__"}],"network_params":{"prefunded_accounts":{}}}`),
		"top-level array": []byte(`[]`),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := effectiveParameters(address, "image", custom)
			require.Error(t, err)
		})
	}
}
