// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/network"
)

type recordingNetworks struct {
	startRequest network.StartRequest
	statusDir    string
	stopDir      string
	statusErr    error
}

func (networks *recordingNetworks) Start(_ context.Context, request network.StartRequest) error {
	networks.startRequest = request
	return nil
}

func (networks *recordingNetworks) Status(_ context.Context, directory string) error {
	networks.statusDir = directory
	return networks.statusErr
}

func (networks *recordingNetworks) Stop(_ context.Context, directory string) error {
	networks.stopDir = directory
	return nil
}

func TestRunNetworkCommands(t *testing.T) {
	networkDir := filepath.Join(t.TempDir(), "network")
	networks := new(recordingNetworks)
	var stdout, stderr bytes.Buffer
	if err := run(t.Context(), []string{
		"start",
		"--network-dir", networkDir,
		"--execution-image", "local/go-qrl:test",
		"--timeout", "17m",
	}, &stdout, &stderr, networks); err != nil {
		t.Fatal(err)
	}
	request := networks.startRequest
	if request.NetworkDir != networkDir ||
		request.ExecutionImage != "local/go-qrl:test" ||
		request.StartTimeout != 17*time.Minute {
		t.Fatalf("start request = %+v", request)
	}
	if stdout.String() != "network ready\n" {
		t.Fatalf("start output = %q", stdout.String())
	}
	stdout.Reset()
	if err := run(t.Context(), []string{"status", "--network-dir", networkDir}, &stdout, &stderr, networks); err != nil {
		t.Fatal(err)
	}
	if stdout.String() != "network ready\n" {
		t.Fatalf("status output = %q", stdout.String())
	}
	stdout.Reset()
	if err := run(t.Context(), []string{"stop", "--network-dir", networkDir}, &stdout, &stderr, networks); err != nil {
		t.Fatal(err)
	}
	if stdout.String() != "network stopped\n" {
		t.Fatalf("stop output = %q", stdout.String())
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

func TestRunRequiresNetworkDirectoryAndDoesNotPrintFailedStatus(t *testing.T) {
	networks := new(recordingNetworks)
	var stdout bytes.Buffer
	err := run(t.Context(), []string{"status"}, &stdout, io.Discard, networks)
	if exitCode(err) != 2 || !strings.Contains(err.Error(), "--network-dir is required") {
		t.Fatalf("missing directory error = %v, exit code = %d", err, exitCode(err))
	}

	networks.statusErr = errors.New("network is not running")
	err = run(t.Context(), []string{"status", "--network-dir", t.TempDir()}, &stdout, io.Discard, networks)
	if !errors.Is(err, networks.statusErr) || stdout.Len() != 0 {
		t.Fatalf("failed status error = %v, output = %q", err, stdout.String())
	}
}
