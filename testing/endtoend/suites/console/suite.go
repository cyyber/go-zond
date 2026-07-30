// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package console

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/theQRL/go-qrl/common/hexutil"
)

const resultPrefix = "CONSOLE_E2E_PASS "

type consoleScenario struct {
	name        string
	description string
}

var consoleScenarios = []consoleScenario{
	{
		name:        "api",
		description: "validates console and RPC APIs against the live network",
	},
	{
		name:        "contract",
		description: "deploys a contract and validates VM64 ABI, receipts, events, and filters",
	},
}

// Regenerate the source-controlled Hyperion artifacts.
// The compiler must be cyyber/hyperion@2b9a0f1d.
//
//go:generate sh -c "hypc --version 2>&1 | grep -Fq commit.2b9a0f1d || { echo 'hypc from cyyber/hyperion@2b9a0f1d is required; found:' >&2; hypc --version >&2; exit 1; }"
//go:generate hypc --abi --bin --no-cbor-metadata --overwrite -o testdata/contracts testdata/contracts/EventEmitter.hyp

//go:embed testdata/console/*.js
var consoleFixtures embed.FS

//go:embed testdata/contracts/EventEmitter.abi
var eventEmitterABI []byte

//go:embed testdata/contracts/EventEmitter.bin
var eventEmitterBytecode string

func runSuite(ctx context.Context, gqrlPath, jsPath, rpcURL, name string) error {
	expression := "loadScript('harness.js');loadScript('" + name + ".js')"
	command := exec.CommandContext(
		ctx,
		gqrlPath,
		"attach",
		"--jspath",
		jsPath,
		"--exec",
		expression,
		rpcURL,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("run console suite %s: %w\n%s", name, err, output)
	}
	if err := parseSuiteResult(name, output); err != nil {
		return fmt.Errorf("%w\n%s", err, output)
	}
	return nil
}

func parseSuiteResult(name string, output []byte) error {
	marker := []byte(resultPrefix + name)
	var matches int
	for _, line := range bytes.Split(output, []byte{'\n'}) {
		if bytes.Equal(bytes.TrimSpace(line), marker) {
			matches++
		}
	}
	if matches != 1 {
		return fmt.Errorf("console suite %s emitted %d success markers", name, matches)
	}
	return nil
}

func prepareWorkspace(ctx context.Context, destination, rpcURL string) error {
	consoleScripts, err := fs.Sub(consoleFixtures, "testdata/console")
	if err != nil {
		return fmt.Errorf("open console fixtures: %w", err)
	}
	if err := os.CopyFS(destination, consoleScripts); err != nil {
		return fmt.Errorf("copy console fixtures: %w", err)
	}

	bytecode, err := hexutil.Decode("0x" + strings.TrimPrefix(strings.TrimSpace(eventEmitterBytecode), "0x"))
	if err != nil {
		return fmt.Errorf("decode EventEmitter bytecode: %w", err)
	}

	params, err := deploymentParameters(ctx, rpcURL, eventEmitterABI, bytecode)
	if err != nil {
		return err
	}
	script := append([]byte("var PARAMS = "), params...)
	script = append(script, ';', '\n')
	if err := os.WriteFile(filepath.Join(destination, ".params.js"), script, 0o600); err != nil {
		return fmt.Errorf("write console parameters: %w", err)
	}
	return nil
}
