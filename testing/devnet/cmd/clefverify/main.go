// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Command clefverify runs the standalone Clef end-to-end scenario.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/theQRL/go-qrl/testing/devnet/suites/clef"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "clefverify:", err)
		os.Exit(1)
	}
}

func run(arguments []string) error {
	flags := flag.NewFlagSet("clefverify", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	clefPath := flags.String("clef", "", "path to the Clef executable")
	seedEnvironment := flags.String(
		"seed-env",
		"DEPLOYER_SEED",
		"environment variable containing the public development seed",
	)
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	seed := os.Getenv(*seedEnvironment)
	if seed == "" {
		return fmt.Errorf("seed environment variable %s is empty or unset", *seedEnvironment)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	result, err := clef.Run(ctx, clef.Config{
		ClefPath: *clefPath,
		Seed:     seed,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Clef %s verified account %s\n", result.Version, result.Account.Hex())
	return nil
}
