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
	startDir      string
	startImage    string
	startDeadline time.Time
	inspectDir    string
	stopDir       string
	inspectErr    error
}

func (networks *recordingNetworks) Start(ctx context.Context, directory, image string) error {
	networks.startDir = directory
	networks.startImage = image
	networks.startDeadline, _ = ctx.Deadline()
	return nil
}

func (networks *recordingNetworks) Inspect(_ context.Context, directory string) (network.Environment, error) {
	networks.inspectDir = directory
	return network.Environment{}, networks.inspectErr
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
	if networks.startDir != networkDir ||
		networks.startImage != "local/go-qrl:test" {
		t.Fatalf("start directory/image = %q/%q", networks.startDir, networks.startImage)
	}
	if remaining := time.Until(networks.startDeadline); remaining <= 16*time.Minute || remaining > 17*time.Minute {
		t.Fatalf("start deadline remaining = %s", remaining)
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
	if networks.inspectDir != networkDir || networks.stopDir != networkDir {
		t.Fatalf("status = %q, stop = %q", networks.inspectDir, networks.stopDir)
	}
	err := run(t.Context(), []string{"network", "status"}, io.Discard, io.Discard, networks)
	if err == nil {
		t.Fatal("invalid command succeeded")
	}
}

func TestRunRequiresNetworkDirectoryAndDoesNotPrintFailedStatus(t *testing.T) {
	networks := new(recordingNetworks)
	var stdout bytes.Buffer
	err := run(t.Context(), []string{"status"}, &stdout, io.Discard, networks)
	if err == nil || !strings.Contains(err.Error(), "--network-dir is required") {
		t.Fatalf("missing directory error = %v", err)
	}
	err = run(
		t.Context(),
		[]string{"start", "--network-dir", t.TempDir()},
		&stdout,
		io.Discard,
		networks,
	)
	if err == nil || !strings.Contains(err.Error(), "--execution-image is required") {
		t.Fatalf("missing execution image error = %v", err)
	}

	networks.inspectErr = errors.New("network is not running")
	err = run(t.Context(), []string{"status", "--network-dir", t.TempDir()}, &stdout, io.Discard, networks)
	if !errors.Is(err, networks.inspectErr) || stdout.Len() != 0 {
		t.Fatalf("failed status error = %v, output = %q", err, stdout.String())
	}
}
