// Copyright 2026 The go-qrl Authors
// This file is part of go-qrl.
//
// go-qrl is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-qrl is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-qrl. If not, see <http://www.gnu.org/licenses/>.

// Command devnet controls the separately managed development network.
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

	"github.com/theQRL/go-qrl/devnet/internal/network"
	"github.com/urfave/cli/v2"
)

type controller interface {
	Start(context.Context, network.StartOptions) error
	Inspect(context.Context, string) (network.Environment, error)
	Stop(context.Context, string) error
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	manager := network.NewManager()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr, manager); err != nil {
		fmt.Fprintln(os.Stderr, "devnet:", err)
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
			var parameters []byte
			if paramsFile := command.String("params-file"); paramsFile != "" {
				var err error
				parameters, err = os.ReadFile(paramsFile)
				if err != nil {
					return fmt.Errorf("read parameters file: %w", err)
				}
			}
			startCtx, cancel := context.WithTimeout(command.Context, command.Duration("timeout"))
			defer cancel()
			if err := networks.Start(startCtx, network.StartOptions{
				Directory:      networkDir,
				ExecutionImage: executionImage,
				Parameters:     parameters,
			}); err != nil {
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
		Name:        "devnet",
		Usage:       "Manage a separately running QRL development network",
		Description: "Network lifecycle is independent of end-to-end suite execution.",
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
		{"start", "Start the standalone development network"},
		{"status", "Check whether the development network is ready"},
		{"stop", "Stop the standalone development network"},
	} {
		flags := []cli.Flag{&cli.StringFlag{Name: "network-dir", Usage: "development network directory"}}
		if command.name == "start" {
			flags = append(flags,
				&cli.StringFlag{Name: "execution-image", Usage: "execution image reference"},
				&cli.PathFlag{
					Name:  "params-file",
					Usage: "complete JSON qrl-package parameters; omit for the built-in single-node profile",
				},
				&cli.DurationFlag{Name: "timeout", Usage: "network start budget", Value: 30 * time.Minute},
			)
		}
		app.Commands = append(app.Commands, &cli.Command{
			Name: command.name, Usage: command.usage, Flags: flags, Action: execute,
		})
	}
	return app.RunContext(ctx, append([]string{app.Name}, arguments...))
}
