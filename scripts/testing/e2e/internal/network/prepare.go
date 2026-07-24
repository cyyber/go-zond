// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/theQRL/go-qrl/common"
)

type preparedNetwork struct {
	Params string
}

var imageBuildVariables = map[string]string{
	"E2E_LOCAL_EL_IMAGE":      "execution",
	"E2E_LOCAL_CL_IMAGE":      "consensus",
	"E2E_LOCAL_VC_IMAGE":      "validator",
	"E2E_LOCAL_GENESIS_IMAGE": "genesis",
}

func prepareNetwork(
	ctx context.Context,
	runner commandRunner,
	request StartRequest,
	walletAddress string,
	stdout, stderr io.Writer,
) (preparedNetwork, error) {
	if request.DockerBin == "" {
		request.DockerBin = "docker"
	}
	refs := networkImageRefs(request.NetworkDir)
	params, err := effectiveParameters(walletAddress, refs)
	if err != nil {
		return preparedNetwork{}, err
	}
	pathEnv := os.Getenv("PATH")
	if filepath.IsAbs(request.DockerBin) {
		pathEnv = filepath.Dir(request.DockerBin) + string(os.PathListSeparator) + pathEnv
	}
	environment := []string{"PATH=" + pathEnv, "E2E_DOCKER_BIN=" + request.DockerBin}
	for name, role := range imageBuildVariables {
		environment = append(environment, name+"="+refs[role])
	}
	environment = append(environment, pinnedBuildEnvironment()...)
	if err := runner.Run(ctx, command{
		Path: filepath.Join(request.RepoRoot, "scripts", "local_testnet", "build_network_images.sh"),
		Dir:  request.RepoRoot, Env: environment,
		Stdout: stdout, Stderr: stderr,
	}); err != nil {
		return preparedNetwork{}, fmt.Errorf("build network images: %w", err)
	}
	return preparedNetwork{Params: params}, nil
}

func networkImageRefs(networkDir string) map[string]string {
	refs := make(map[string]string, len(localImageRefs))
	for role, ref := range localImageRefs {
		refs[role] = ref
	}
	refs["execution"] = "local/go-qrl:e2e-" + networkID(networkDir)
	return refs
}

func networkID(networkDir string) string {
	digest := sha256.Sum256([]byte(networkDir))
	return fmt.Sprintf("%x", digest[:6])
}

func effectiveParameters(address string, refs map[string]string) (string, error) {
	if _, err := common.NewAddressFromString(address); err != nil {
		return "", errors.New("wallet address is invalid")
	}
	parameters := map[string]any{
		"participants": []any{map[string]any{
			"el_type": "gqrl", "el_image": refs["execution"],
			"el_extra_params": []any{"--graphql", "--graphql.vhosts=*"},
			"cl_type":         "qrysm", "cl_image": refs["consensus"],
			"cl_extra_params":   []any{"--min-sync-peers=0", "--minimum-peers-per-subnet=0"},
			"vc_type":           "qrysm",
			"vc_image":          refs["validator"],
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
		"qrl_genesis_generator_params": map[string]any{"image": refs["genesis"]},
		"additional_services":          []any{},
	}
	payload, err := json.Marshal(parameters)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}
