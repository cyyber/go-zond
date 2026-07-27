// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Command e2e controls the separately managed E2E network.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/theQRL/go-qrl/testing/endtoend/internal/network"
	"github.com/urfave/cli/v2"
)

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
	execute := func(command *cli.Context) error {
		if command.NArg() != 0 {
			return fmt.Errorf("unexpected positional arguments: %v", command.Args().Slice())
		}
		networkDir := command.String("network-dir")
		if networkDir == "" {
			return errors.New("--network-dir is required")
		}
		message := "network ready"
		switch command.Command.Name {
		case "start":
			executionImage := command.String("execution-image")
			if executionImage == "" {
				return errors.New("--execution-image is required")
			}
			startCtx, cancel := context.WithTimeout(command.Context, command.Duration("timeout"))
			defer cancel()
			if err := networks.Start(startCtx, networkDir, executionImage); err != nil {
				return err
			}
		case "status":
			if _, err := networks.Inspect(command.Context, networkDir); err != nil {
				return err
			}
		case "stop":
			if err := networks.Stop(command.Context, networkDir); err != nil {
				return err
			}
			message = "network stopped"
		}
		_, err := fmt.Fprintln(command.App.Writer, message)
		return err
	}

	app := &cli.App{
		Name:        "e2e",
		Usage:       "Manage the separately running E2E network",
		Description: "Tests run independently through make live-test.",
		Writer:      stdout,
		ErrWriter:   stderr,
		ExitErrHandler: func(*cli.Context, error) {
			// Return all errors to main and tests instead of exiting in the library.
		},
		Action: func(command *cli.Context) error {
			if command.NArg() == 0 {
				return cli.ShowAppHelp(command)
			}
			return fmt.Errorf("unknown command %q", command.Args().First())
		},
	}
	for _, command := range []struct{ name, usage string }{
		{"start", "Start the standalone E2E network"},
		{"status", "Check whether the E2E network is ready"},
		{"stop", "Stop the standalone E2E network"},
	} {
		flags := []cli.Flag{&cli.StringFlag{Name: "network-dir", Usage: "E2E network directory"}}
		if command.name == "start" {
			flags = append(flags,
				&cli.StringFlag{Name: "execution-image", Usage: "execution image reference"},
				&cli.DurationFlag{Name: "timeout", Usage: "network start budget", Value: 150 * time.Minute},
			)
		}
		app.Commands = append(app.Commands, &cli.Command{
			Name: command.name, Usage: command.usage, Flags: flags, Action: execute,
		})
	}
	return app.RunContext(ctx, append([]string{app.Name}, arguments...))
}
