// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Command clefauto runs Clef with an automated UI for disposable E2E networks.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/theQRL/go-qrl/internal/qrlapi"
	"github.com/theQRL/go-qrl/rpc"
	signercore "github.com/theQRL/go-qrl/signer/core"
	"github.com/theQRL/go-qrl/testing/endtoend/internal/fixture"
)

type automatedUI struct{}

func (*automatedUI) ApproveTx(request *signercore.SignTxRequest) (signercore.SignTxResponse, error) {
	return signercore.SignTxResponse{Transaction: request.Transaction, Approved: true}, nil
}

func (*automatedUI) ApproveSignData(*signercore.SignDataRequest) (signercore.SignDataResponse, error) {
	return signercore.SignDataResponse{Approved: true}, nil
}

func (*automatedUI) ApproveListing(request *signercore.ListRequest) (signercore.ListResponse, error) {
	return signercore.ListResponse{Accounts: request.Accounts}, nil
}

func (*automatedUI) ApproveNewAccount(*signercore.NewAccountRequest) (signercore.NewAccountResponse, error) {
	return signercore.NewAccountResponse{Approved: true}, nil
}

func (*automatedUI) ShowError(signercore.Message) {}

func (*automatedUI) ShowInfo(signercore.Message) {}

func (*automatedUI) OnApprovedTx(qrlapi.SignTransactionResult) {}

func (*automatedUI) OnSignerStartup(signercore.StartupInfo) {}

func (*automatedUI) OnInputRequired(signercore.UserInputRequest) (signercore.UserInputResponse, error) {
	return signercore.UserInputResponse{Text: fixture.RemoteSignerPassword}, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	args, cleanup, err := clefArgs(ctx, os.Args[1:])
	if err != nil {
		exit(err)
	}
	defer cleanup()

	command := exec.CommandContext(ctx, "clef-bin", args...)
	command.Stderr = os.Stderr

	input, err := command.StdinPipe()
	if err != nil {
		exit(err)
	}
	output, err := command.StdoutPipe()
	if err != nil {
		exit(err)
	}
	client, err := rpc.DialIO(ctx, output, input)
	if err != nil {
		exit(err)
	}
	defer client.Close()
	if err := client.RegisterName("ui", new(automatedUI)); err != nil {
		exit(err)
	}
	if err := command.Start(); err != nil {
		exit(err)
	}
	if err := command.Wait(); err != nil && ctx.Err() == nil {
		exit(err)
	}
}

func clefArgs(ctx context.Context, args []string) ([]string, func(), error) {
	dir, err := os.MkdirTemp("", "go-qrl-clef-")
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { os.RemoveAll(dir) }

	passwordPath := filepath.Join(dir, "password")
	seedPath := filepath.Join(dir, "seed")
	keystorePath := filepath.Join(dir, "keystore")
	if err := os.WriteFile(passwordPath, []byte(fixture.RemoteSignerPassword), 0o600); err != nil {
		cleanup()
		return nil, nil, err
	}
	if err := os.WriteFile(seedPath, []byte(fixture.RemoteSignerSeed), 0o600); err != nil {
		cleanup()
		return nil, nil, err
	}

	importer := exec.CommandContext(ctx, "clef-bin",
		"--suppress-bootwarn",
		"--keystore="+keystorePath,
		"importraw",
		"--password="+passwordPath,
		seedPath,
	)
	importer.Stdout = os.Stderr
	importer.Stderr = os.Stderr
	if err := importer.Run(); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("import development account: %w", err)
	}

	configured := make([]string, 0, len(args)+2)
	keystoreSet := false
	for _, arg := range args {
		if strings.HasPrefix(arg, "--keystore=") {
			arg = "--keystore=" + keystorePath
			keystoreSet = true
		}
		configured = append(configured, arg)
	}
	if !keystoreSet {
		configured = append(configured, "--keystore="+keystorePath)
	}
	return append(configured, "--stdio-ui"), cleanup, nil
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
