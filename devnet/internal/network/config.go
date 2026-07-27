// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	packageLocator   = "github.com/rgeraldes24/qrl-package@3892c3d2596403c080424d9e8fc99ff172483fe0"
	defaultNetworkID = "1337"
	prefundBalance   = "2000000QRL"

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

type parameterShape struct {
	Participants []struct {
		ExecutionImage string `json:"el_image"`
	} `json:"participants"`
	Network struct {
		PrefundedAccounts map[string]json.RawMessage `json:"prefunded_accounts"`
	} `json:"network_params"`
}

func effectiveParameters(address, executionImage string, custom []byte) (string, error) {
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
			"network_id": defaultNetworkID, "seconds_per_slot": 5,
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
	shape, err := decodeParameterShape(payload)
	if err != nil {
		return "", err
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

	encodedImagePlaceholder, _ := json.Marshal(executionImagePlaceholder)
	encodedExecutionImage, _ := json.Marshal(executionImage)
	rendered := bytes.ReplaceAll(payload, encodedImagePlaceholder, encodedExecutionImage)
	encodedWalletPlaceholder, _ := json.Marshal(walletAddressPlaceholder)
	encodedAddress, _ := json.Marshal(address)
	rendered = bytes.ReplaceAll(rendered, encodedWalletPlaceholder, encodedAddress)

	renderedShape, err := decodeParameterShape(rendered)
	if err != nil {
		return "", errors.New("rendered parameters must contain one JSON object")
	}
	if len(renderedShape.Participants) == 0 ||
		renderedShape.Participants[0].ExecutionImage != executionImage {
		return "", errors.New("execution-image token must use its literal JSON spelling")
	}
	if _, ok := renderedShape.Network.PrefundedAccounts[address]; !ok {
		return "", errors.New("wallet-address token must use its literal JSON spelling")
	}
	return string(rendered), nil
}

func decodeParameterShape(payload []byte) (parameterShape, error) {
	var shape parameterShape
	if err := json.Unmarshal(payload, &shape); err != nil {
		return parameterShape{}, errors.New("parameters file must contain one JSON object")
	}
	return shape, nil
}
