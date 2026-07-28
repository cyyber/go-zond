// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package console

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	qrl "github.com/theQRL/go-qrl"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	qrlwallet "github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
	"github.com/theQRL/go-qrl/qrlclient"
	"github.com/theQRL/go-qrl/testing/devnet/internal/network"
)

const resultPrefix = "VM64_E2E_RESULT "

var suiteNames = []string{
	"web3_sanity",
	"api_surfaces",
	"logs_topics",
	"event_roundtrip",
	"abi_vm64",
}

type suiteResult struct {
	Schema int    `json:"schema"`
	Suite  string `json:"suite"`
	Status string `json:"status"`
	Passed int    `json:"passed"`
	Failed int    `json:"failed"`
	Total  int    `json:"total"`
}

func runSuite(ctx context.Context, gqrlPath, jsPath, rpcURL, name string) error {
	expression := "loadScript('console/" + name + ".js')"
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
	var results []suiteResult
	for _, line := range bytes.Split(output, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, []byte(resultPrefix)) {
			continue
		}
		var result suiteResult
		if err := json.Unmarshal(bytes.TrimPrefix(line, []byte(resultPrefix)), &result); err != nil {
			return fmt.Errorf("decode console suite %s result: %w", name, err)
		}
		results = append(results, result)
	}
	if len(results) != 1 {
		return fmt.Errorf("console suite %s emitted %d result records", name, len(results))
	}
	result := results[0]
	if result.Schema != 1 ||
		result.Suite != name ||
		result.Status != "passed" ||
		result.Passed < 0 ||
		result.Failed != 0 ||
		result.Total != result.Passed {
		return fmt.Errorf("console suite %s failed: %+v", name, result)
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
	source, err := testdataRoot()
	if err != nil {
		return err
	}
	if err := os.CopyFS(destination, os.DirFS(source)); err != nil {
		return fmt.Errorf("copy console testdata: %w", err)
	}

	abiJSON, err := os.ReadFile(filepath.Join(destination, "contracts", "EventEmitter.abi"))
	if err != nil {
		return fmt.Errorf("read EventEmitter ABI: %w", err)
	}
	if !json.Valid(abiJSON) {
		return fmt.Errorf("EventEmitter ABI is invalid JSON")
	}
	bytecodeHex, err := os.ReadFile(filepath.Join(destination, "contracts", "EventEmitter.bin"))
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
	path := filepath.Join(destination, "console", ".params.js")
	if err := os.WriteFile(path, script, 0o600); err != nil {
		return fmt.Errorf("write console parameters: %w", err)
	}
	return nil
}

func deploymentParameters(ctx context.Context, rpcURL string, abiJSON, bytecode []byte) ([]byte, error) {
	wallet, err := network.UnsafeDevelopmentWallet()
	if err != nil {
		return nil, err
	}
	from := common.Address(wallet.GetAddress())
	client, err := qrlclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial RPC: %w", err)
	}
	defer client.Close()

	tx, err := signDeployment(ctx, client, wallet, from, bytecode)
	if err != nil {
		return nil, err
	}
	raw, err := tx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("encode deployment transaction: %w", err)
	}
	return json.Marshal(struct {
		Address        string          `json:"address"`
		TxHash         string          `json:"txHash"`
		RawTransaction string          `json:"rawTransaction"`
		ABI            json.RawMessage `json:"abi"`
		Signature      string          `json:"signature"`
		Value          uint64          `json:"value"`
	}{
		Address:        from.Hex(),
		TxHash:         tx.Hash().Hex(),
		RawTransaction: hexutil.Encode(raw),
		ABI:            abiJSON,
		Signature:      "Deployed(uint256)",
		Value:          1337,
	})
}

func signDeployment(
	ctx context.Context,
	client *qrlclient.Client,
	wallet qrlwallet.Wallet,
	from common.Address,
	bytecode []byte,
) (*types.Transaction, error) {
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("chain ID: %w", err)
	}
	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return nil, fmt.Errorf("deployment nonce: %w", err)
	}
	gasFeeCap, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("gas price: %w", err)
	}
	gasTipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, fmt.Errorf("gas tip: %w", err)
	}
	gasFeeCap = new(big.Int).Mul(gasFeeCap, big.NewInt(4))
	if gasFeeCap.Cmp(gasTipCap) < 0 {
		gasFeeCap = gasTipCap
	}
	gas, err := client.EstimateGas(ctx, qrl.CallMsg{
		From:  from,
		Value: new(big.Int),
		Data:  bytecode,
	})
	if err != nil {
		return nil, fmt.Errorf("estimate deployment gas: %w", err)
	}
	gas += gas / 5

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gas,
		Value:     new(big.Int),
		Data:      bytecode,
	})
	signed, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), wallet)
	if err != nil {
		return nil, fmt.Errorf("sign deployment transaction: %w", err)
	}
	return signed, nil
}

func repositoryRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("locate console suite source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "..")), nil
}

func testdataRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("locate console testdata")
	}
	return filepath.Join(filepath.Dir(file), "testdata"), nil
}
