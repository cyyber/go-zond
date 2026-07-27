// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/theQRL/go-qrl/common"
)

const (
	packageLocator = "github.com/rgeraldes24/qrl-package@3892c3d2596403c080424d9e8fc99ff172483fe0"
	defaultChainID = 1337
	prefundBalance = "2000000QRL"

	executionImagePlaceholder = "__DEVNET_EXECUTION_IMAGE__"
	walletAddressPlaceholder  = "__DEVNET_WALLET_ADDRESS__"

	consensusImage = "qrledger/qrysm:beacon-chain-8b80fa0c3f5a"
	validatorImage = "qrledger/qrysm:validator-8b80fa0c3f5a"
	genesisImage   = "qrledger/qrysm:qrl-genesis-generator-360410c72353-8b80fa0c3f5a"

	executionServiceName = "el-1-gqrl-qrysm"
	rpcPortID            = "rpc"
	webSocketPortID      = "ws"
	graphQLPath          = "/graphql"
)

func effectiveParameters(address, executionImage string, custom []byte) (string, error) {
	if _, err := common.NewAddressFromString(address); err != nil {
		return "", errors.New("wallet address is invalid")
	}
	if strings.TrimSpace(executionImage) == "" {
		return "", errors.New("execution image is empty")
	}
	if custom != nil {
		return renderCustomParameters(custom, address, executionImage)
	}
	parameters := map[string]any{
		"participants": []any{map[string]any{
			"el_image":        executionImage,
			"el_extra_params": []any{"--graphql", "--graphql.vhosts=*"},
			"cl_image":        consensusImage,
			"cl_extra_params": []any{"--min-sync-peers=0", "--minimum-peers-per-subnet=0"},
			"vc_image":        validatorImage,
		}},
		"network_params": map[string]any{
			"network_id": fmt.Sprint(defaultChainID), "seconds_per_slot": 5,
			"execution_follow_distance": 8,
			"prefunded_accounts":        map[string]any{address: map[string]any{"balance": prefundBalance}},
			"withdrawal_address":        address,
			"light_kdf_enabled":         true,
		},
		"qrl_genesis_generator_params": map[string]any{"image": genesisImage},
	}
	payload, err := json.Marshal(parameters)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func renderCustomParameters(payload []byte, address, executionImage string) (string, error) {
	var shape struct {
		Participants []struct {
			ExecutionImage string `json:"el_image"`
		} `json:"participants"`
		Network struct {
			PrefundedAccounts map[string]json.RawMessage `json:"prefunded_accounts"`
		} `json:"network_params"`
	}
	if err := json.Unmarshal(payload, &shape); err != nil {
		return "", errors.New("parameters file must contain one JSON object")
	}
	if len(shape.Participants) == 0 || shape.Participants[0].ExecutionImage != executionImagePlaceholder {
		return "", fmt.Errorf(
			"first participant el_image must be %q",
			executionImagePlaceholder,
		)
	}
	if _, ok := shape.Network.PrefundedAccounts[walletAddressPlaceholder]; !ok {
		return "", fmt.Errorf(
			"network_params.prefunded_accounts must contain %q",
			walletAddressPlaceholder,
		)
	}

	rendered := bytes.Clone(payload)
	for placeholder, value := range map[string]string{
		executionImagePlaceholder: executionImage,
		walletAddressPlaceholder:  address,
	} {
		encodedPlaceholder, _ := json.Marshal(placeholder)
		encodedValue, _ := json.Marshal(value)
		rendered = bytes.ReplaceAll(rendered, encodedPlaceholder, encodedValue)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(rendered, &object); err != nil || object == nil {
		return "", errors.New("rendered parameters must contain one JSON object")
	}
	return string(rendered), nil
}
