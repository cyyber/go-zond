// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestBuiltInNetworkParameters(t *testing.T) {
	address := "Q" + strings.Repeat("a", 128)
	refs := networkImageRefs(t.TempDir())
	payload, err := effectiveParameters(address, refs)
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
	if participant["el_image"] != refs["execution"] ||
		participant["cl_image"] != refs["consensus"] ||
		participant["vc_image"] != refs["validator"] ||
		network["network_id"] != "1337" ||
		network["withdrawal_address"] != address ||
		prefund["balance"] != prefundBalance {
		t.Fatalf("parameters = %#v", parameters)
	}
}

type recordingRunner struct{ command command }

func (runner *recordingRunner) Run(_ context.Context, command command) error {
	runner.command = command
	return nil
}

func TestPrepareNetworkBuildsFixedImages(t *testing.T) {
	runner := new(recordingRunner)
	address := "Q" + strings.Repeat("a", 128)
	networkDir := t.TempDir()
	refs := networkImageRefs(networkDir)
	if refs["execution"] == networkImageRefs(t.TempDir())["execution"] {
		t.Fatal("different network directories share an execution image tag")
	}
	prepared, err := prepareNetwork(context.Background(), runner, StartRequest{
		RepoRoot: t.TempDir(), NetworkDir: networkDir,
		DockerBin: "/usr/local/bin/docker",
	}, address, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Params == "" ||
		runner.command.Path != runner.command.Dir+"/scripts/local_testnet/build_network_images.sh" {
		t.Fatalf("prepared=%+v command=%+v", prepared, runner.command)
	}
	environment := strings.Join(runner.command.Env, "\n")
	for _, expected := range []string{
		"E2E_LOCAL_EL_IMAGE=" + refs["execution"],
		"E2E_LOCAL_CL_IMAGE=" + refs["consensus"],
		"E2E_LOCAL_VC_IMAGE=" + refs["validator"],
		"E2E_LOCAL_GENESIS_IMAGE=" + refs["genesis"],
		"E2E_PINNED_QRYSM_GIT_COMMIT=" + qrysmCommit,
		"E2E_PINNED_GENERATOR_GIT_COMMIT=" + genesisCommit,
		"E2E_DOCKER_BIN=/usr/local/bin/docker",
	} {
		if !strings.Contains(environment, expected) {
			t.Fatalf("build environment missing %q: %s", expected, environment)
		}
	}
}

func TestEffectiveParametersRejectsInvalidAddress(t *testing.T) {
	if _, err := effectiveParameters("0x01", networkImageRefs(t.TempDir())); err == nil {
		t.Fatal("invalid wallet address was accepted")
	}
}
