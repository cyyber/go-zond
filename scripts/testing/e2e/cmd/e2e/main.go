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

const (
	usage = `Usage:
  e2e <start|status|stop> [options]

Manage the separately running E2E network. Tests run independently through
make live-test.
`
)

var errUsage = errors.New("invalid arguments")

type controller interface {
	Start(context.Context, network.StartRequest) error
	Status(context.Context, string) error
	Stop(context.Context, string) error
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	manager := network.NewManager()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr, manager); err != nil {
		fmt.Fprintln(os.Stderr, "e2e:", err)
		os.Exit(exitCode(err))
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
	switch arguments[0] {
	case "start":
		return start(ctx, arguments[1:], stdout, stderr, networks)
	case "status", "stop":
		return operate(ctx, arguments[0], arguments[1:], stdout, stderr, networks)
	default:
		return usageError("unknown command %q", arguments[0])
	}
}

func start(
	ctx context.Context,
	arguments []string,
	stdout,
	stderr io.Writer,
	networks controller,
) error {
	var (
		networkDir     string
		executionImage = network.DefaultExecutionImage
		timeout        = 150 * time.Minute
	)
	flags := flag.NewFlagSet("e2e start", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&networkDir, "network-dir", networkDir, "E2E network directory")
	flags.StringVar(&executionImage, "execution-image", executionImage, "execution image reference")
	flags.DurationVar(&timeout, "timeout", timeout, "network start budget")
	if err := parse(flags, arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if networkDir == "" {
		return usageError("--network-dir is required")
	}
	startCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := networks.Start(startCtx, network.StartRequest{
		NetworkDir:     networkDir,
		ExecutionImage: executionImage,
	}); err != nil {
		return err
	}
	_, err := fmt.Fprintln(stdout, "network ready")
	return err
}

func operate(
	ctx context.Context,
	operation string,
	arguments []string,
	stdout,
	stderr io.Writer,
	networks controller,
) error {
	var networkDir string
	flags := flag.NewFlagSet("e2e "+operation, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&networkDir, "network-dir", networkDir, "E2E network directory")
	if err := parse(flags, arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if networkDir == "" {
		return usageError("--network-dir is required")
	}
	if operation == "status" {
		if err := networks.Status(ctx, networkDir); err != nil {
			return err
		}
		_, err := fmt.Fprintln(stdout, "network ready")
		return err
	}
	if err := networks.Stop(ctx, networkDir); err != nil {
		return err
	}
	_, err := fmt.Fprintln(stdout, "network stopped")
	return err
}

func parse(flags *flag.FlagSet, arguments []string) error {
	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return flag.ErrHelp
		}
		return fmt.Errorf("%w: %v", errUsage, err)
	}
	if flags.NArg() != 0 {
		return usageError("unexpected positional arguments: %v", flags.Args())
	}
	return nil
}

func isHelp(argument string) bool {
	return argument == "-h" || argument == "--help" || argument == "help"
}

func usageError(format string, arguments ...any) error {
	return fmt.Errorf("%w: %s", errUsage, fmt.Sprintf(format, arguments...))
}

func exitCode(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, errUsage):
		return 2
	case errors.Is(err, context.Canceled):
		return 130
	default:
		return 1
	}
}
