// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package clef

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/theQRL/go-qrl/common"
)

type clefProcess struct {
	cancel context.CancelFunc
	done   chan struct{}
	err    error
	log    *os.File
}

func initializeClef(
	ctx context.Context,
	clefPath string,
	workspace string,
	seed string,
	masterPassword string,
	accountPassword string,
) (common.Address, error) {
	configDir := filepath.Join(workspace, "config")
	keyStore := filepath.Join(workspace, "keystore")
	seedPath := filepath.Join(workspace, "seed.hex")
	passwordPath := filepath.Join(workspace, "account-password.txt")
	rulesPath := filepath.Join(workspace, "rules.js")

	for path, contents := range map[string]string{
		seedPath:     seed + "\n",
		passwordPath: accountPassword + "\n",
		rulesPath:    rulesSource,
	} {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			return common.Address{}, fmt.Errorf("write Clef input %s: %w", filepath.Base(path), err)
		}
	}

	baseArgs := []string{
		"--suppress-bootwarn",
		"--lightkdf",
		"--configdir", configDir,
		"--keystore", keyStore,
	}
	if _, err := runClefCommand(
		ctx,
		clefPath,
		append(append([]string(nil), baseArgs...), "init"),
		masterPassword+"\n"+masterPassword+"\n",
	); err != nil {
		return common.Address{}, fmt.Errorf("initialize Clef: %w", err)
	}
	output, err := runClefCommand(
		ctx,
		clefPath,
		append(append([]string(nil), baseArgs...), "importraw", "--password", passwordPath, seedPath),
		"",
	)
	if err != nil {
		return common.Address{}, fmt.Errorf("import Clef account: %w", err)
	}
	account, err := parseImportedAccount(output)
	if err != nil {
		return common.Address{}, err
	}

	rulesHash := sha256.Sum256([]byte(rulesSource))
	if _, err := runClefCommand(
		ctx,
		clefPath,
		append(append([]string(nil), baseArgs...), "attest", hex.EncodeToString(rulesHash[:])),
		masterPassword+"\n",
	); err != nil {
		return common.Address{}, fmt.Errorf("attest Clef rules: %w", err)
	}
	if _, err := runClefCommand(
		ctx,
		clefPath,
		append(append([]string(nil), baseArgs...), "setpw", account.Hex()),
		accountPassword+"\n"+accountPassword+"\n"+masterPassword+"\n",
	); err != nil {
		return common.Address{}, fmt.Errorf("set Clef account password: %w", err)
	}
	return account, nil
}

func runClefCommand(ctx context.Context, path string, args []string, stdin string) ([]byte, error) {
	command := exec.CommandContext(ctx, path, args...)
	command.Stdin = strings.NewReader(stdin)
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w\n%s", err, output)
	}
	return output, nil
}

func parseImportedAccount(output []byte) (common.Address, error) {
	for _, line := range strings.Split(string(output), "\n") {
		if !strings.HasPrefix(line, "  Address ") {
			continue
		}
		return common.NewAddressFromString(strings.TrimSpace(strings.TrimPrefix(line, "  Address ")))
	}
	return common.Address{}, errors.New("could not parse imported Clef account")
}

func startClef(
	ctx context.Context,
	clefPath string,
	workspace string,
	masterPassword string,
) (*clefProcess, string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", fmt.Errorf("reserve Clef HTTP port: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		return nil, "", fmt.Errorf("release Clef HTTP port: %w", err)
	}

	logFile, err := os.OpenFile(
		filepath.Join(workspace, "clef.log"),
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		0o600,
	)
	if err != nil {
		return nil, "", fmt.Errorf("create Clef log: %w", err)
	}
	processCtx, cancel := context.WithCancel(ctx)
	command := exec.CommandContext(processCtx, clefPath,
		"--suppress-bootwarn",
		"--lightkdf",
		"--advanced",
		"--configdir", filepath.Join(workspace, "config"),
		"--keystore", filepath.Join(workspace, "keystore"),
		"--chainid", strconv.FormatInt(defaultChainID, 10),
		"--rules", filepath.Join(workspace, "rules.js"),
		"--http",
		"--http.addr", "127.0.0.1",
		"--http.port", strconv.Itoa(port),
		"--http.vhosts", "*",
		"--ipcdisable",
		"--auditlog", "",
	)
	command.Stdin = strings.NewReader(masterPassword + "\n")
	command.Stdout = logFile
	command.Stderr = logFile
	if err := command.Start(); err != nil {
		cancel()
		_ = logFile.Close()
		return nil, "", fmt.Errorf("start Clef: %w", err)
	}
	process := &clefProcess{cancel: cancel, done: make(chan struct{}), log: logFile}
	go func() {
		process.err = command.Wait()
		close(process.done)
	}()
	return process, "http://127.0.0.1:" + strconv.Itoa(port), nil
}

func (process *clefProcess) stop() error {
	process.cancel()
	<-process.done
	return process.log.Close()
}

func randomSecret() (string, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("generate Clef password: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(secret), nil
}
