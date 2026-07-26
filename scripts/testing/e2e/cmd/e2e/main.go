// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Command e2e controls the separately managed E2E network.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/theQRL/go-qrl/scripts/testing/e2e/internal/network"
)

const usage = `Usage:
  e2e <start|status|stop> [options]

Manage the separately running E2E network. Tests run independently through
make live-test.
`

type controller interface {
	Start(context.Context, string, string) error
	Inspect(context.Context, string) (network.Environment, error)
	Stop(context.Context, string) error
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	manager := network.NewManager()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr, manager); err != nil {
		fmt.Fprintln(os.Stderr, "e2e:", err)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	arguments []string,
	stdout,
	stderr io.Writer,
	networks controller,
) error {
	if len(arguments) == 0 || isHelp(arguments[0]) {
		_, err := fmt.Fprint(stdout, usage)
		return err
	}
	command := arguments[0]
	if command != "start" && command != "status" && command != "stop" {
		return fmt.Errorf("unknown command %q", command)
	}
	var (
		networkDir     string
		executionImage string
		timeout        = 150 * time.Minute
	)
	flags := flag.NewFlagSet("e2e "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&networkDir, "network-dir", networkDir, "E2E network directory")
	if command == "start" {
		flags.StringVar(&executionImage, "execution-image", executionImage, "execution image reference")
		flags.DurationVar(&timeout, "timeout", timeout, "network start budget")
	}
	if err := flags.Parse(arguments[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	if networkDir == "" {
		return errors.New("--network-dir is required")
	}
	if command == "start" && executionImage == "" {
		return errors.New("--execution-image is required")
	}

	message := "network ready"
	switch command {
	case "start":
		startCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		if err := networks.Start(startCtx, networkDir, executionImage); err != nil {
			return err
		}
	case "status":
		if _, err := networks.Inspect(ctx, networkDir); err != nil {
			return err
		}
	case "stop":
		if err := networks.Stop(ctx, networkDir); err != nil {
			return err
		}
		message = "network stopped"
	}
	_, err := fmt.Fprintln(stdout, message)
	return err
}

func isHelp(argument string) bool {
	return argument == "-h" || argument == "--help" || argument == "help"
}
