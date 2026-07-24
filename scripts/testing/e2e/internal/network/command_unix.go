//go:build !windows

// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package network

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func configureNetworkCommandGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killNetworkCommandGroup(pid int) error {
	if pid <= 0 {
		return os.ErrProcessDone
	}
	err := syscall.Kill(-pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return os.ErrProcessDone
	}
	return err
}
