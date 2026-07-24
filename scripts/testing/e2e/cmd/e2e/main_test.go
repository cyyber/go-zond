// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package main

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/network"
)

type recordingNetworks struct {
	startRequest network.StartRequest
	statusDir    string
	stopDir      string
}

func (networks *recordingNetworks) Start(_ context.Context, request network.StartRequest) (network.Result, error) {
	networks.startRequest = request
	return network.Result{}, nil
}

func (networks *recordingNetworks) Status(_ context.Context, directory string) (network.Result, error) {
	networks.statusDir = directory
	return network.Result{Ready: true}, nil
}

func (networks *recordingNetworks) Stop(_ context.Context, directory string) (network.Result, error) {
	networks.stopDir = directory
	return network.Result{}, nil
}

func TestRunNetworkCommands(t *testing.T) {
	root := t.TempDir()
	networkDir := filepath.Join(t.TempDir(), "network")
	networks := new(recordingNetworks)
	var stdout, stderr bytes.Buffer
	if err := run(t.Context(), []string{
		"start",
		"--repo-root", root,
		"--network-dir", networkDir,
		"--docker-bin", "/opt/bin/docker",
		"--timeout", "17m",
	}, &stdout, &stderr, networks); err != nil {
		t.Fatal(err)
	}
	request := networks.startRequest
	if request.RepoRoot != root ||
		request.NetworkDir != networkDir ||
		request.DockerBin != "/opt/bin/docker" ||
		request.StartTimeout != 17*time.Minute {
		t.Fatalf("start request = %+v", request)
	}
	for _, operation := range []string{"status", "stop"} {
		if err := run(t.Context(), []string{operation, "--network-dir", networkDir}, &stdout, &stderr, networks); err != nil {
			t.Fatalf("%s: %v", operation, err)
		}
	}
	if networks.statusDir != networkDir || networks.stopDir != networkDir {
		t.Fatalf("status = %q, stop = %q", networks.statusDir, networks.stopDir)
	}
	err := run(t.Context(), []string{"network", "status"}, io.Discard, io.Discard, networks)
	if exitCode(err) != 2 {
		t.Fatalf("invalid command error = %v, exit code = %d", err, exitCode(err))
	}
	if exitCode(context.Canceled) != 130 {
		t.Fatal("canceled command did not use exit code 130")
	}
}
