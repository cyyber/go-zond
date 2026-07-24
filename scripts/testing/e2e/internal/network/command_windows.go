//go:build windows

// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"errors"
	"os"
	"os/exec"
)

func configureNetworkCommandGroup(_ *exec.Cmd) {}

func killNetworkCommandGroup(pid int) error {
	if pid <= 0 {
		return os.ErrProcessDone
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	err = process.Kill()
	if errors.Is(err, os.ErrProcessDone) {
		return os.ErrProcessDone
	}
	return err
}
