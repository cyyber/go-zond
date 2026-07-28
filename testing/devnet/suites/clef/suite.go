// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

// Package clef exercises a standalone Clef signer and verifies its QRL
// signatures and signed transactions.
package clef

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	qrlaccounts "github.com/theQRL/go-qrl/accounts"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
	"github.com/theQRL/go-qrl/rpc"
	"github.com/theQRL/go-qrl/signer/core/apitypes"
)

const (
	readinessTimeout = 30 * time.Second
	requestTimeout   = 15 * time.Second
	pollInterval     = 500 * time.Millisecond
)

type signTransactionResult struct {
	Raw hexutil.Bytes      `json:"raw"`
	Tx  *types.Transaction `json:"tx"`
}

func run(ctx context.Context, clefPath, workspace string, expectedWallet wallet.Wallet) error {
	if clefPath == "" {
		return errors.New("Clef executable path is required")
	}
	seed, err := expectedWallet.GetSeed()
	if err != nil {
		return fmt.Errorf("read expected wallet seed: %w", err)
	}

	masterPassword, err := randomSecret()
	if err != nil {
		return err
	}
	accountPassword, err := randomSecret()
	if err != nil {
		return err
	}
	account, err := initializeClef(
		ctx,
		clefPath,
		workspace,
		hex.EncodeToString(seed.ToBytes()),
		masterPassword,
		accountPassword,
	)
	if err != nil {
		return err
	}
	if account != common.Address(expectedWallet.GetAddress()) {
		return fmt.Errorf(
			"imported account %s does not match seed address %s",
			account.Hex(),
			common.Address(expectedWallet.GetAddress()).Hex(),
		)
	}

	process, endpoint, err := startClef(ctx, clefPath, workspace, masterPassword)
	if err != nil {
		return err
	}
	client, err := rpc.DialOptions(
		ctx,
		endpoint,
		rpc.WithHTTPClient(&http.Client{Timeout: requestTimeout}),
	)
	if err != nil {
		return errors.Join(fmt.Errorf("connect to Clef: %w", err), process.stop())
	}

	runErr := waitForClef(ctx, client, process)
	if runErr == nil {
		runErr = exercise(ctx, client, account, expectedWallet)
	}
	client.Close()
	return errors.Join(runErr, process.stop())
}

func exercise(
	ctx context.Context,
	client *rpc.Client,
	account common.Address,
	expectedWallet wallet.Wallet,
) error {
	var listedAccounts []common.Address
	if err := callRPC(ctx, client, &listedAccounts, "account_list"); err != nil {
		return err
	}
	if len(listedAccounts) != 1 || listedAccounts[0] != account {
		return fmt.Errorf("account_list returned %v, want [%s]", listedAccounts, account.Hex())
	}

	var dataSignature hexutil.Bytes
	if err := callRPC(ctx, client, &dataSignature, "account_signData",
		qrlaccounts.MimetypeTextPlain,
		account.Hex(),
		hexutil.Encode([]byte(expectedText)),
	); err != nil {
		return err
	}
	if err := verifySignature(
		"account_signData",
		dataSignature,
		qrlaccounts.TextHash([]byte(expectedText)),
		expectedWallet,
	); err != nil {
		return err
	}

	typedData := expectedTypedData(account)
	var typedSignature hexutil.Bytes
	if err := callRPC(ctx, client, &typedSignature, "account_signTypedData",
		account.Hex(),
		typedData,
	); err != nil {
		return err
	}
	typedDigest, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return fmt.Errorf("hash typed data: %w", err)
	}
	if err := verifySignature(
		"account_signTypedData",
		typedSignature,
		typedDigest,
		expectedWallet,
	); err != nil {
		return err
	}

	transaction := expectedTransaction(account)
	var signed signTransactionResult
	if err := callRPC(
		ctx,
		client,
		&signed,
		"account_signTransaction",
		transaction,
	); err != nil {
		return err
	}
	return verifyTransaction(signed, transaction, account, expectedWallet)
}

func waitForClef(
	ctx context.Context,
	client *rpc.Client,
	process *clefProcess,
) error {
	readyCtx, cancel := context.WithTimeout(ctx, readinessTimeout)
	defer cancel()
	var lastErr error
	for {
		var version string
		if err := callRPC(readyCtx, client, &version, "account_version"); err == nil {
			if version == "" {
				return errors.New("account_version returned an empty version")
			}
			return nil
		} else {
			lastErr = err
		}
		select {
		case <-process.done:
			if process.err != nil {
				return fmt.Errorf("Clef exited before readiness: %w", process.err)
			}
			return errors.New("Clef exited before readiness")
		case <-readyCtx.Done():
			return fmt.Errorf("wait for Clef: %w", errors.Join(readyCtx.Err(), lastErr))
		case <-time.After(pollInterval):
		}
	}
}

func callRPC(
	ctx context.Context,
	client *rpc.Client,
	result any,
	method string,
	args ...any,
) error {
	if err := client.CallContext(ctx, result, method, args...); err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	return nil
}
