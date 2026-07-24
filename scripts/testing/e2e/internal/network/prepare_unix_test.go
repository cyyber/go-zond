//go:build !windows

// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExecRunnerCancellationKillsDescendants(t *testing.T) {
	temporaryDirectory := t.TempDir()
	readyPath := filepath.Join(temporaryDirectory, "ready")
	markerPath := filepath.Join(temporaryDirectory, "descendant-marker")
	ctx, cancel := context.WithCancel(context.Background())
	finished := make(chan error, 1)
	go func() {
		finished <- (execRunner{}).Run(ctx, command{
			Path: "/bin/sh",
			Args: []string{
				"-c",
				`(sleep 1; printf late > "$E2E_MARKER") & printf ready > "$E2E_READY"; wait`,
			},
			Env: []string{"E2E_MARKER=" + markerPath, "E2E_READY=" + readyPath},
		})
	}()

	waitForFile(t, readyPath, 2*time.Second)
	cancel()
	select {
	case err := <-finished:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled command error = %v; want context cancellation", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("canceled command did not exit")
	}

	time.Sleep(1200 * time.Millisecond)
	if _, err := os.Stat(markerPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("descendant wrote delayed marker after cancellation: %v", err)
	}
}

func TestExecRunnerScrubsInheritedE2EEnvironment(t *testing.T) {
	t.Setenv("E2E_INHERITED_SECRET", "must-not-be-inherited")
	var output bytes.Buffer
	err := (execRunner{}).Run(context.Background(), command{
		Path:   "/bin/sh",
		Args:   []string{"-c", `printf '%s|%s' "$E2E_INHERITED_SECRET" "$E2E_EXPLICIT"`},
		Env:    []string{"E2E_EXPLICIT=kept"},
		Stdout: &output,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "|kept" {
		t.Fatalf("filtered command environment = %q; want %q", got, "|kept")
	}
}

func TestExecRunnerStreamsOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := (execRunner{}).Run(context.Background(), command{
		Path:   "/bin/sh",
		Args:   []string{"-c", `printf standard; printf diagnostic >&2`},
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := stdout.String(); got != "standard" {
		t.Fatalf("streamed stdout = %q; want %q", got, "standard")
	}
	if got := stderr.String(); got != "diagnostic" {
		t.Fatalf("streamed stderr = %q; want %q", got, "diagnostic")
	}
}

func waitForFile(t *testing.T, path string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("inspect readiness file: %v", err)
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for descendant readiness")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
