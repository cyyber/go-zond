// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/theQRL/go-qrl/common"
)

const (
	packageLocator  = "github.com/rgeraldes24/qrl-package@3892c3d2596403c080424d9e8fc99ff172483fe0"
	expectedChainID = 1337
	prefundBalance  = "2000000QRL"

	consensusImage = "qrledger/qrysm:beacon-chain-8b80fa0c3f5a"
	genesisImage   = "qrledger/qrysm:qrl-genesis-generator-360410c72353-8b80fa0c3f5a"

	executionServiceName = "el-1-gqrl-qrysm"
	rpcPortID            = "rpc"
	webSocketPortID      = "ws"
	graphQLPath          = "/graphql"
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
			"el_image":        executionImage,
			"el_extra_params": []any{"--graphql", "--graphql.vhosts=*"},
			"cl_image":        consensusImage,
			"cl_extra_params": []any{"--min-sync-peers=0", "--minimum-peers-per-subnet=0"},
		}},
		"network_params": map[string]any{
			"network_id": fmt.Sprint(expectedChainID), "seconds_per_slot": 5,
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
