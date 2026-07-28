// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package console

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/theQRL/go-qrl/common/hexutil"
)

const resultPrefix = "CONSOLE_E2E_PASS "

var suiteNames = []string{"api", "contract"}

//go:embed testdata/console/*.js testdata/contracts/EventEmitter.abi testdata/contracts/EventEmitter.bin
var fixtures embed.FS

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

func buildGQRL(ctx context.Context, output string) error {
	root, err := repositoryRoot()
	if err != nil {
		return err
	}
	command := exec.CommandContext(ctx, "go", "build", "-o", output, "./cmd/gqrl")
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("build gqrl: %w\n%s", err, output)
	}
	return nil
}

func prepareWorkspace(ctx context.Context, destination, rpcURL string) error {
	consoleFixtures, err := fs.Sub(fixtures, "testdata/console")
	if err != nil {
		return fmt.Errorf("open console fixtures: %w", err)
	}
	if err := os.CopyFS(destination, consoleFixtures); err != nil {
		return fmt.Errorf("copy console fixtures: %w", err)
	}

	abiJSON, err := fixtures.ReadFile("testdata/contracts/EventEmitter.abi")
	if err != nil {
		return fmt.Errorf("read EventEmitter ABI: %w", err)
	}
	if !json.Valid(abiJSON) {
		return fmt.Errorf("EventEmitter ABI is invalid JSON")
	}
	bytecodeHex, err := fixtures.ReadFile("testdata/contracts/EventEmitter.bin")
	if err != nil {
		return fmt.Errorf("read EventEmitter bytecode: %w", err)
	}
	bytecode, err := hexutil.Decode("0x" + strings.TrimPrefix(strings.TrimSpace(string(bytecodeHex)), "0x"))
	if err != nil {
		return fmt.Errorf("decode EventEmitter bytecode: %w", err)
	}

	params, err := deploymentParameters(ctx, rpcURL, abiJSON, bytecode)
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

func repositoryRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("locate console suite source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "..")), nil
}
