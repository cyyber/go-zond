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
	"sync"
	"time"

	"github.com/theQRL/go-qrl/common/hexutil"
	qrlconsole "github.com/theQRL/go-qrl/console"
	"github.com/theQRL/go-qrl/rpc"
	endtoendlive "github.com/theQRL/go-qrl/testing/endtoend/internal/live"
)

const resultPrefix = "CONSOLE_E2E_PASS "
const failurePrefix = "CONSOLE_E2E_FAIL "

type consoleScenario struct {
	name        string
	description string
	webSocket   bool
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
	{
		name:        "topics",
		description: "formats and decodes indexed VM64 scalar topics",
	},
	{
		name:        "constructor",
		description: "deploys a contract through the embedded web3 contract factory",
		webSocket:   true,
	},
	{
		name:        "events",
		description: "formats and submits a contract transaction and watches indexed events over WebSocket",
		webSocket:   true,
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

type synchronizedBuffer struct {
	lock sync.Mutex
	data bytes.Buffer
}

func (buffer *synchronizedBuffer) Write(data []byte) (int, error) {
	buffer.lock.Lock()
	defer buffer.lock.Unlock()
	return buffer.data.Write(data)
}

func (buffer *synchronizedBuffer) Bytes() []byte {
	buffer.lock.Lock()
	defer buffer.lock.Unlock()
	return bytes.Clone(buffer.data.Bytes())
}

func runWatchedSuite(ctx context.Context, client *rpc.Client, jsPath, name string) error {
	var output synchronizedBuffer
	console, err := qrlconsole.New(qrlconsole.Config{
		DataDir: filepath.Join(jsPath, "console-data"),
		DocRoot: jsPath,
		Client:  client,
		Printer: &output,
	})
	if err != nil {
		return fmt.Errorf("create console suite %s: %w", name, err)
	}
	defer console.Stop(false)

	console.Evaluate("loadScript('harness.js');loadScript('" + name + ".js')")

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		result := output.Bytes()
		if bytes.Contains(result, []byte(resultPrefix+name)) {
			return parseSuiteResult(name, result)
		}
		if bytes.Contains(result, []byte(failurePrefix+name)) ||
			bytes.Contains(result, []byte("GoError:")) {
			return fmt.Errorf("console suite %s failed\n%s", name, result)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("console suite %s: %w\n%s", name, ctx.Err(), result)
		case <-ticker.C:
		}
	}
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

func prepareWorkspace(ctx context.Context, destination string, session *endtoendlive.Session) error {
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

	params, err := deploymentParameters(ctx, session, eventEmitterABI, bytecode)
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
