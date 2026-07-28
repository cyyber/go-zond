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
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/theQRL/go-qrl/testing/devnet/internal/network"
	"github.com/urfave/cli/v2"
)

type controller interface {
	Start(context.Context, network.StartOptions) error
	Stop(context.Context, string) error
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := newApp(network.NewManager()).RunContext(ctx, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "devnet:", err)
		os.Exit(1)
	}
}

func newApp(networks controller) *cli.App {
	enclaveName := &cli.StringFlag{
		Name:  "enclave-name",
		Usage: "Kurtosis enclave name",
		Value: network.DefaultEnclaveName,
	}
	return &cli.App{
		Name:            "devnet",
		Usage:           "control the separately managed development network",
		HideHelpCommand: true,
		Action:          rootAction,
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "start the development network and wait for readiness",
				Flags: []cli.Flag{
					enclaveName,
					&cli.StringFlag{
						Name:     "execution-image",
						Usage:    "execution image reference",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "params-file",
						Usage: "complete JSON qrl-package parameters; omit for the built-in single-node profile",
					},
					&cli.DurationFlag{
						Name:  "timeout",
						Usage: "network start budget",
						Value: network.DefaultStartTimeout,
					},
				},
				Action: func(command *cli.Context) error {
					if err := rejectPositional(command); err != nil {
						return err
					}
					var parameters []byte
					if file := command.String("params-file"); file != "" {
						var err error
						parameters, err = os.ReadFile(file)
						if err != nil {
							return fmt.Errorf("read parameters file: %w", err)
						}
					}
					ctx, cancel := context.WithTimeout(command.Context, command.Duration("timeout"))
					defer cancel()
					if err := networks.Start(ctx, network.StartOptions{
						EnclaveName:    command.String("enclave-name"),
						ExecutionImage: command.String("execution-image"),
						Parameters:     parameters,
					}); err != nil {
						return err
					}
					_, err := fmt.Fprintln(command.App.Writer, "network ready")
					return err
				},
			},
			{
				Name:  "stop",
				Usage: "stop the development network",
				Flags: []cli.Flag{enclaveName},
				Action: func(command *cli.Context) error {
					if err := rejectPositional(command); err != nil {
						return err
					}
					if err := networks.Stop(command.Context, command.String("enclave-name")); err != nil {
						return err
					}
					_, err := fmt.Fprintln(command.App.Writer, "network stopped")
					return err
				},
			},
		},
	}
}

// rootAction runs when no subcommand matched: bare invocations get the usage
// text, unknown commands an error.
func rootAction(command *cli.Context) error {
	if command.Args().Present() {
		return fmt.Errorf("unknown command %q", command.Args().First())
	}
	return cli.ShowAppHelp(command)
}

func rejectPositional(command *cli.Context) error {
	if command.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", command.Args().Slice())
	}
	return nil
}
