// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/theQRL/go-qrl/common"
)

const (
	packageLocator  = "github.com/rgeraldes24/qrl-package@1f31cd03dbe2061225701ea79d956cfeceaf91db"
	expectedChainID = "0x539"
	prefundBalance  = "2000000QRL"

	consensusImage = "local/qrysm-beacon:8b80fa0c3f5a"
	validatorImage = "local/qrysm-validator:8b80fa0c3f5a"
	genesisImage   = "local/qrl-genesis-generator:360410c72353-8b80fa0c3f5a"

	executionServiceName = "el-1-gqrl-qrysm"
	consensusServiceName = "cl-1-qrysm-gqrl"
	validatorServiceName = "vc-1-gqrl-qrysm"
	rpcPortID            = "rpc"
	webSocketPortID      = "ws"
	graphQLPath          = "/graphql"
)

const DefaultExecutionImage = "local/go-qrl:e2e"

var (
	requiredServices = map[string]string{
		"execution": executionServiceName,
		"consensus": consensusServiceName,
		"validator": validatorServiceName,
	}
)

func effectiveParameters(address, executionImage string) (string, error) {
	if _, err := common.NewAddressFromString(address); err != nil {
		return "", errors.New("wallet address is invalid")
	}
	if strings.TrimSpace(executionImage) == "" {
		return "", errors.New("execution image is empty")
	}
	parameters := map[string]any{
		"participants": []any{map[string]any{
			"el_type": "gqrl", "el_image": executionImage,
			"el_extra_params": []any{"--graphql", "--graphql.vhosts=*"},
			"cl_type":         "qrysm", "cl_image": consensusImage,
			"cl_extra_params":   []any{"--min-sync-peers=0", "--minimum-peers-per-subnet=0"},
			"vc_type":           "qrysm",
			"vc_image":          validatorImage,
			"count":             1,
			"use_remote_signer": false,
		}},
		"network_params": map[string]any{
			"preset": "mainnet", "network_id": "1337",
			"seconds_per_slot": 5, "slots_per_epoch": 128,
			"execution_follow_distance": 8,
			"prefunded_accounts":        map[string]any{address: map[string]any{"balance": prefundBalance}},
			"withdrawal_address":        address,
			"light_kdf_enabled":         true,
		},
		"qrl_genesis_generator_params": map[string]any{"image": genesisImage},
		"additional_services":          []any{},
	}
	payload, err := json.Marshal(parameters)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func networkID(networkDir string) string {
	digest := sha256.Sum256([]byte(networkDir))
	return fmt.Sprintf("%x", digest[:6])
}
