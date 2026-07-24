// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const networkCommandWaitDelay = 2 * time.Second

type command struct {
	Path, Dir string
	Args      []string
	Env       []string
	Stdout    io.Writer
	Stderr    io.Writer
}

type commandRunner interface {
	Run(context.Context, command) error
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, specification command) error {
	cmd, err := newNetworkCommand(ctx, specification)
	if err != nil {
		return err
	}
	cmd.Stdout, cmd.Stderr = specification.Stdout, specification.Stderr
	return runNetworkCommand(ctx, cmd)
}

func newNetworkCommand(ctx context.Context, specification command) (*exec.Cmd, error) {
	if ctx == nil {
		return nil, errors.New("command context is nil")
	}
	if specification.Path == "" {
		return nil, errors.New("command path is required")
	}
	cmd := exec.CommandContext(ctx, specification.Path, specification.Args...)
	cmd.Dir = specification.Dir
	cmd.Env = networkCommandEnvironment(specification.Env)
	configureNetworkCommandGroup(cmd)
	cmd.Cancel = func() error { return killNetworkCommandGroup(cmd.Process.Pid) }
	cmd.WaitDelay = networkCommandWaitDelay
	return cmd, nil
}

func runNetworkCommand(ctx context.Context, cmd *exec.Cmd) error {
	err := cmd.Run()
	if context.Cause(ctx) != nil {
		return context.Cause(ctx)
	}
	if err != nil {
		return fmt.Errorf("command %s: %w", filepath.Base(cmd.Path), err)
	}
	return nil
}

func networkCommandEnvironment(explicit []string) []string {
	environment := make([]string, 0, len(os.Environ())+len(explicit))
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		if !strings.HasPrefix(name, "E2E_") {
			environment = append(environment, entry)
		}
	}
	return append(environment, explicit...)
}
