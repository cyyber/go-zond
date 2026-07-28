// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package clef exercises a standalone Clef signer and verifies its QRL
// signatures and signed transactions.
package clef

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	qrlaccounts "github.com/theQRL/go-qrl/accounts"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/common/math"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
	"github.com/theQRL/go-qrl/signer/core/apitypes"
)

const (
	defaultChainID   = int64(1337)
	readinessTimeout = 30 * time.Second
	requestTimeout   = 15 * time.Second
	pollInterval     = 500 * time.Millisecond

	expectedText          = "Clef VM64 signData"
	expectedRecipient     = "Qd5812f6cf4a0f645aa620cd57319a0ed649dd8f5519a9dde7770ae5b0e49e547985f35eb972a2a07041561aa39c65a3991478f9b1e6749e05277dcf58a9a8b72"
	expectedTypedName     = "Local Testnet VM64"
	expectedTypedVersion  = "1"
	expectedTypedContents = "Clef VM64 typed data"
	expectedTypedValue    = "340282366920938463463374607431768211457"
	expectedTxInputHex    = "0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f"
	expectedNonce         = uint64(9)
	expectedGas           = uint64(40000)
	expectedTip           = int64(7)
	expectedFeeCap        = int64(1_000_000_000)
	expectedValue         = int64(42)

	rulesSource = `function ApproveListing(req) { return 'Approve'; }
function ApproveSignData(req) { return 'Approve'; }
function ApproveTx(req) { return 'Approve'; }
`
)

type Config struct {
	ClefPath string
	Seed     string
}

type Result struct {
	Account common.Address
	Version string
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
	ID      int             `json:"id"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type signTransactionResult struct {
	Raw hexutil.Bytes      `json:"raw"`
	Tx  *types.Transaction `json:"tx"`
}

type clefProcess struct {
	cancel context.CancelFunc
	done   chan struct{}
	err    error
	log    *os.File
}

func Run(ctx context.Context, config Config) (Result, error) {
	if config.ClefPath == "" {
		return Result{}, errors.New("Clef executable path is required")
	}
	config.Seed = strings.TrimPrefix(strings.TrimSpace(config.Seed), "0x")
	if config.Seed == "" {
		return Result{}, errors.New("Clef seed is required")
	}
	expectedWallet, err := wallet.RestoreFromSeedHex(config.Seed)
	if err != nil {
		return Result{}, fmt.Errorf("restore expected wallet: %w", err)
	}
	workspace, err := os.MkdirTemp("", "go-qrl-clef-e2e-")
	if err != nil {
		return Result{}, fmt.Errorf("create Clef workspace: %w", err)
	}
	defer os.RemoveAll(workspace)

	masterPassword, err := randomSecret()
	if err != nil {
		return Result{}, err
	}
	accountPassword, err := randomSecret()
	if err != nil {
		return Result{}, err
	}
	account, err := initializeClef(
		ctx,
		config.ClefPath,
		workspace,
		config.Seed,
		masterPassword,
		accountPassword,
	)
	if err != nil {
		return Result{}, err
	}
	if account != common.Address(expectedWallet.GetAddress()) {
		return Result{}, fmt.Errorf(
			"imported account %s does not match seed address %s",
			account.Hex(),
			common.Address(expectedWallet.GetAddress()).Hex(),
		)
	}

	process, endpoint, err := startClef(
		ctx,
		config.ClefPath,
		workspace,
		masterPassword,
	)
	if err != nil {
		return Result{}, err
	}
	result, runErr := exercise(
		ctx,
		&http.Client{Timeout: requestTimeout},
		endpoint,
		process,
		account,
		expectedWallet,
	)
	stopErr := process.stop()
	if runErr != nil {
		return Result{}, errors.Join(runErr, stopErr)
	}
	return result, stopErr
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

func exercise(
	ctx context.Context,
	client *http.Client,
	endpoint string,
	process *clefProcess,
	account common.Address,
	expectedWallet wallet.Wallet,
) (Result, error) {
	version, err := waitForClef(ctx, client, endpoint, process)
	if err != nil {
		return Result{}, err
	}

	var listedAccounts []common.Address
	if err := callRPC(ctx, client, endpoint, "account_list", []any{}, 2, &listedAccounts); err != nil {
		return Result{}, err
	}
	if len(listedAccounts) != 1 || listedAccounts[0] != account {
		return Result{}, fmt.Errorf("account_list returned %v, want [%s]", listedAccounts, account.Hex())
	}

	var dataSignature hexutil.Bytes
	if err := callRPC(ctx, client, endpoint, "account_signData", []any{
		qrlaccounts.MimetypeTextPlain,
		account.Hex(),
		hexutil.Encode([]byte(expectedText)),
	}, 3, &dataSignature); err != nil {
		return Result{}, err
	}
	if err := verifySignature(
		"account_signData",
		dataSignature,
		qrlaccounts.TextHash([]byte(expectedText)),
		expectedWallet,
	); err != nil {
		return Result{}, err
	}

	typedData := expectedTypedData(account)
	var typedSignature hexutil.Bytes
	if err := callRPC(ctx, client, endpoint, "account_signTypedData", []any{
		account.Hex(),
		typedData,
	}, 4, &typedSignature); err != nil {
		return Result{}, err
	}
	typedDigest, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return Result{}, fmt.Errorf("hash typed data: %w", err)
	}
	if err := verifySignature(
		"account_signTypedData",
		typedSignature,
		typedDigest,
		expectedWallet,
	); err != nil {
		return Result{}, err
	}

	transaction, err := expectedTransaction(account)
	if err != nil {
		return Result{}, err
	}
	var signed signTransactionResult
	if err := callRPC(
		ctx,
		client,
		endpoint,
		"account_signTransaction",
		[]any{transaction},
		5,
		&signed,
	); err != nil {
		return Result{}, err
	}
	if err := verifyTransaction(signed, transaction, account, expectedWallet); err != nil {
		return Result{}, err
	}

	return Result{Account: account, Version: version}, nil
}

func waitForClef(
	ctx context.Context,
	client *http.Client,
	endpoint string,
	process *clefProcess,
) (string, error) {
	readyCtx, cancel := context.WithTimeout(ctx, readinessTimeout)
	defer cancel()
	var lastErr error
	for {
		var version string
		if err := callRPC(readyCtx, client, endpoint, "account_version", []any{}, 1, &version); err == nil {
			return version, nil
		} else {
			lastErr = err
		}
		select {
		case <-process.done:
			if process.err != nil {
				return "", fmt.Errorf("Clef exited before readiness: %w", process.err)
			}
			return "", errors.New("Clef exited before readiness")
		case <-readyCtx.Done():
			return "", fmt.Errorf("wait for Clef: %w", errors.Join(readyCtx.Err(), lastErr))
		case <-time.After(pollInterval):
		}
	}
}

func callRPC(
	ctx context.Context,
	client *http.Client,
	endpoint string,
	method string,
	params []any,
	id int,
	result any,
) error {
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      id,
	})
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%s: HTTP %s", method, response.Status)
	}
	var envelope rpcResponse
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		return fmt.Errorf("%s: decode response: %w", method, err)
	}
	if envelope.JSONRPC != "2.0" || envelope.ID != id {
		return fmt.Errorf("%s: unexpected JSON-RPC response", method)
	}
	if envelope.Error != nil {
		return fmt.Errorf("%s: RPC error %d: %s", method, envelope.Error.Code, envelope.Error.Message)
	}
	if len(envelope.Result) == 0 || bytes.Equal(envelope.Result, []byte("null")) {
		return fmt.Errorf("%s: missing result", method)
	}
	if err := json.Unmarshal(envelope.Result, result); err != nil {
		return fmt.Errorf("%s: decode result: %w", method, err)
	}
	return nil
}

func expectedTypedData(account common.Address) apitypes.TypedData {
	chainID := math.HexOrDecimal256(*big.NewInt(defaultChainID))
	return apitypes.TypedData{
		Types: apitypes.Types{
			"QRLTypedDataDomain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Message": {
				{Name: "sender", Type: "address"},
				{Name: "contents", Type: "string"},
				{Name: "value", Type: "uint256"},
			},
		},
		PrimaryType: "Message",
		Domain: apitypes.TypedDataDomain{
			Name:              expectedTypedName,
			Version:           expectedTypedVersion,
			ChainId:           &chainID,
			VerifyingContract: account.Hex(),
		},
		Message: apitypes.TypedDataMessage{
			"sender":   account.Hex(),
			"contents": expectedTypedContents,
			"value":    expectedTypedValue,
		},
	}
}

func expectedTransaction(account common.Address) (apitypes.SendTxArgs, error) {
	recipient, err := common.NewAddressFromString(expectedRecipient)
	if err != nil {
		return apitypes.SendTxArgs{}, err
	}
	input, err := hexutil.Decode(expectedTxInputHex)
	if err != nil {
		return apitypes.SendTxArgs{}, err
	}
	from := common.NewMixedcaseAddress(account)
	to := common.NewMixedcaseAddress(recipient)
	tip := hexutil.Big(*big.NewInt(expectedTip))
	feeCap := hexutil.Big(*big.NewInt(expectedFeeCap))
	value := hexutil.Big(*big.NewInt(expectedValue))
	chainID := hexutil.Big(*big.NewInt(defaultChainID))
	data := hexutil.Bytes(input)
	accessList := types.AccessList{}
	return apitypes.SendTxArgs{
		From:                 from,
		To:                   &to,
		Gas:                  hexutil.Uint64(expectedGas),
		MaxFeePerGas:         &feeCap,
		MaxPriorityFeePerGas: &tip,
		Value:                value,
		Nonce:                hexutil.Uint64(expectedNonce),
		Input:                &data,
		AccessList:           &accessList,
		ChainID:              &chainID,
	}, nil
}

func randomSecret() (string, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("generate Clef password: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(secret), nil
}

func buildClef(ctx context.Context, output string) error {
	root, err := repositoryRoot()
	if err != nil {
		return err
	}
	command := exec.CommandContext(ctx, "go", "build", "-o", output, "./cmd/clef")
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("build Clef: %w\n%s", err, output)
	}
	return nil
}

func repositoryRoot() (string, error) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("locate Clef suite source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(source), "..", "..", "..", "..")), nil
}
