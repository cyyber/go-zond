// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package build provides helpers for building repository binaries used by
// development-network suites.
package build

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Binary builds packagePath from the repository root into output.
func Binary(ctx context.Context, packagePath, output string) error {
	root, err := repositoryRoot()
	if err != nil {
		return err
	}
	command := exec.CommandContext(ctx, "go", "build", "-o", output, packagePath)
	command.Dir = root
	if commandOutput, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("build %s: %w\n%s", packagePath, err, commandOutput)
	}
	return nil
}

func repositoryRoot() (string, error) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("locate devnet build helper")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(source), "..", "..", "..", "..")), nil
}
