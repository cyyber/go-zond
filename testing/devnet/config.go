// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package devnet

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

// The qrl-package parameter schema, as far as the built-in profile uses it.
type packageParameters struct {
	Participants  []participant   `json:"participants"`
	NetworkParams networkParams   `json:"network_params"`
	GenesisParams generatorParams `json:"qrl_genesis_generator_params"`
}

type participant struct {
	ELImage       string   `json:"el_image"`
	ELExtraParams []string `json:"el_extra_params"`
	CLImage       string   `json:"cl_image"`
	CLExtraParams []string `json:"cl_extra_params"`
	VCImage       string   `json:"vc_image"`
}

type networkParams struct {
	NetworkID               string             `json:"network_id"`
	SecondsPerSlot          int                `json:"seconds_per_slot"`
	ExecutionFollowDistance int                `json:"execution_follow_distance"`
	PrefundedAccounts       map[string]account `json:"prefunded_accounts"`
	WithdrawalAddress       string             `json:"withdrawal_address"`
	LightKDFEnabled         bool               `json:"light_kdf_enabled"`
}

type account struct {
	Balance string `json:"balance"`
}

type generatorParams struct {
	Image string `json:"image"`
}

func effectiveParameters(address, executionImage string, custom []byte) (string, error) {
	if strings.TrimSpace(executionImage) == "" {
		return "", errors.New("execution image is empty")
	}
	if custom != nil {
		return renderCustomParameters(custom, address, executionImage)
	}
	payload, err := json.Marshal(packageParameters{
		Participants: []participant{{
			ELImage:       executionImage,
			ELExtraParams: []string{"--graphql", "--graphql.vhosts=*"},
			CLImage:       consensusImage,
			CLExtraParams: []string{"--min-sync-peers=0", "--minimum-peers-per-subnet=0"},
			VCImage:       validatorImage,
		}},
		NetworkParams: networkParams{
			NetworkID:               defaultNetworkID,
			SecondsPerSlot:          5,
			ExecutionFollowDistance: 8,
			PrefundedAccounts:       map[string]account{address: {Balance: prefundBalance}},
			WithdrawalAddress:       address,
			LightKDFEnabled:         true,
		},
		GenesisParams: generatorParams{Image: genesisImage},
	})
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
