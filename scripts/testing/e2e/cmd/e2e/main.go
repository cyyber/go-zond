// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Command e2e controls the separately managed E2E network.
package main

import (
	"context"
	"encoding/json"
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
	defaultNetworkDir = "/tmp/go-qrl-e2e-network"
	usage             = `Usage:
  e2e <start|status|stop> [options]

Manage the separately running E2E network. Tests run independently through
make live-test.
`
)

var errUsage = errors.New("invalid arguments")

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	manager := network.NewManager()
	manager.Stdout, manager.Stderr = os.Stdout, os.Stderr
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
	networks network.Controller,
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
	networks network.Controller,
) error {
	var (
		root       string
		networkDir = defaultNetworkDir
		dockerBin  = "docker"
		timeout    = 150 * time.Minute
	)
	flags := flag.NewFlagSet("e2e start", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&root, "repo-root", root, "absolute go-qrl checkout root")
	flags.StringVar(&networkDir, "network-dir", networkDir, "E2E network directory")
	flags.StringVar(&dockerBin, "docker-bin", dockerBin, "Docker command path")
	flags.DurationVar(&timeout, "timeout", timeout, "network start budget")
	if err := parse(flags, arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if root == "" {
		return usageError("--repo-root is required")
	}
	result, runErr := networks.Start(ctx, network.StartRequest{
		RepoRoot:     root,
		NetworkDir:   networkDir,
		DockerBin:    dockerBin,
		StartTimeout: timeout,
	})
	return errors.Join(runErr, writeJSON(stdout, result))
}

func operate(
	ctx context.Context,
	operation string,
	arguments []string,
	stdout,
	stderr io.Writer,
	networks network.Controller,
) error {
	networkDir := defaultNetworkDir
	flags := flag.NewFlagSet("e2e "+operation, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&networkDir, "network-dir", networkDir, "E2E network directory")
	if err := parse(flags, arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	var (
		result network.Result
		err    error
	)
	if operation == "status" {
		result, err = networks.Status(ctx, networkDir)
	} else {
		result, err = networks.Stop(ctx, networkDir)
	}
	return errors.Join(err, writeJSON(stdout, result))
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

func writeJSON(destination io.Writer, result network.Result) error {
	encoder := json.NewEncoder(destination)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
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
