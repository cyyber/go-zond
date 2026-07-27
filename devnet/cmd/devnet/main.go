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
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/theQRL/go-qrl/devnet/internal/network"
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
	if len(arguments) == 0 ||
		len(arguments) == 1 && (arguments[0] == "-h" || arguments[0] == "--help") {
		_, err := fmt.Fprintln(stdout, "Usage: devnet <start|status|stop> [options]")
		return err
	}

	command := arguments[0]
	flags := flag.NewFlagSet("devnet "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintf(stderr, "Usage: devnet %s [options]\n", command)
		flags.PrintDefaults()
	}
	enclaveName := flags.String(
		"enclave-name",
		network.DefaultEnclaveName,
		"Kurtosis enclave name",
	)
	var executionImage, paramsFile string
	var timeout time.Duration
	switch command {
	case "start":
		flags.StringVar(&executionImage, "execution-image", "", "execution image reference")
		flags.StringVar(
			&paramsFile,
			"params-file",
			"",
			"complete JSON qrl-package parameters; omit for the built-in single-node profile",
		)
		flags.DurationVar(&timeout, "timeout", 30*time.Minute, "network start budget")
	case "status", "stop":
	default:
		return fmt.Errorf("unknown command %q", command)
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

	message := "network ready"
	switch command {
	case "start":
		if executionImage == "" {
			return errors.New("--execution-image is required")
		}
		var parameters []byte
		if paramsFile != "" {
			var err error
			parameters, err = os.ReadFile(paramsFile)
			if err != nil {
				return fmt.Errorf("read parameters file: %w", err)
			}
		}
		startCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		if err := networks.Start(startCtx, network.StartOptions{
			EnclaveName:    *enclaveName,
			ExecutionImage: executionImage,
			Parameters:     parameters,
		}); err != nil {
			return err
		}
	case "status":
		if _, err := networks.Inspect(ctx, *enclaveName); err != nil {
			return err
		}
	case "stop":
		if err := networks.Stop(ctx, *enclaveName); err != nil {
			return err
		}
		message = "network stopped"
	}
	_, err := fmt.Fprintln(stdout, message)
	return err
}
