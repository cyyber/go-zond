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
	"math/big"
	"net/http"
	"slices"
	"strings"
	"time"

	qrlaccounts "github.com/theQRL/go-qrl/accounts"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
	"github.com/theQRL/go-qrl/rpc"
	signercore "github.com/theQRL/go-qrl/signer/core"
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

type clefSession struct {
	process         *clefProcess
	client          *rpc.Client
	clefPath        string
	workspace       string
	masterPassword  string
	accountPassword string
	account         common.Address
	chainID         *big.Int
	expectedWallet  wallet.Wallet
}

func newClefSession(
	ctx context.Context,
	processContext context.Context,
	clefPath,
	workspace string,
	chainID *big.Int,
	expectedWallet wallet.Wallet,
) (*clefSession, error) {
	if clefPath == "" {
		return nil, errors.New("Clef executable path is required")
	}
	seed, err := expectedWallet.GetSeed()
	if err != nil {
		return nil, fmt.Errorf("read expected wallet seed: %w", err)
	}

	masterPassword, err := randomSecret()
	if err != nil {
		return nil, err
	}
	accountPassword, err := randomSecret()
	if err != nil {
		return nil, err
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
		return nil, err
	}
	if account != common.Address(expectedWallet.GetAddress()) {
		return nil, fmt.Errorf(
			"imported account %s does not match seed address %s",
			account.Hex(),
			common.Address(expectedWallet.GetAddress()).Hex(),
		)
	}

	process, endpoint, err := startClef(
		processContext,
		clefPath,
		workspace,
		masterPassword,
		accountPassword,
		chainID,
	)
	if err != nil {
		return nil, err
	}
	client, err := rpc.DialOptions(
		ctx,
		endpoint,
		rpc.WithHTTPClient(&http.Client{Timeout: requestTimeout}),
	)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("connect to Clef: %w", err), process.stop())
	}

	if err := waitForClef(ctx, client, process); err != nil {
		client.Close()
		return nil, errors.Join(err, process.stop())
	}
	return &clefSession{
		process:         process,
		client:          client,
		clefPath:        clefPath,
		workspace:       workspace,
		masterPassword:  masterPassword,
		accountPassword: accountPassword,
		account:         account,
		chainID:         new(big.Int).Set(chainID),
		expectedWallet:  expectedWallet,
	}, nil
}

func (session *clefSession) close() error {
	if session.client != nil {
		session.client.Close()
	}
	if session.process != nil {
		return session.process.stop()
	}
	return nil
}

func (session *clefSession) restart(ctx, processContext context.Context) error {
	if err := session.close(); err != nil {
		return fmt.Errorf("stop Clef for restart: %w", err)
	}
	session.client = nil
	session.process = nil

	process, endpoint, err := startClef(
		processContext,
		session.clefPath,
		session.workspace,
		session.masterPassword,
		session.accountPassword,
		session.chainID,
	)
	if err != nil {
		return err
	}
	client, err := rpc.DialOptions(
		ctx,
		endpoint,
		rpc.WithHTTPClient(&http.Client{Timeout: requestTimeout}),
	)
	if err != nil {
		return errors.Join(fmt.Errorf("connect to restarted Clef: %w", err), process.stop())
	}
	if err := waitForClef(ctx, client, process); err != nil {
		client.Close()
		return errors.Join(err, process.stop())
	}
	session.process = process
	session.client = client
	return nil
}

func verifyAccountListing(
	ctx context.Context,
	client *rpc.Client,
	account common.Address,
) error {
	var listedAccounts []common.Address
	if err := callRPC(ctx, client, &listedAccounts, "account_list"); err != nil {
		return err
	}
	if len(listedAccounts) != 1 || listedAccounts[0] != account {
		return fmt.Errorf("account_list returned %v, want [%s]", listedAccounts, account.Hex())
	}
	return nil
}

func verifyVersion(ctx context.Context, client *rpc.Client) error {
	var version string
	if err := callRPC(ctx, client, &version, "account_version"); err != nil {
		return err
	}
	if version != signercore.ExternalAPIVersion {
		return fmt.Errorf(
			"account_version returned %q, want %q",
			version,
			signercore.ExternalAPIVersion,
		)
	}
	return nil
}

func verifyNewAccount(
	ctx context.Context,
	session *clefSession,
) (common.Address, error) {
	var account common.Address
	if err := callRPC(ctx, session.client, &account, "account_new"); err != nil {
		return common.Address{}, err
	}
	if account == (common.Address{}) || account == session.account {
		return common.Address{}, fmt.Errorf("account_new returned invalid address %s", account.Hex())
	}
	if err := verifyAccountPresent(ctx, session.client, account); err != nil {
		return common.Address{}, err
	}
	return account, nil
}

func verifyAccountPresent(ctx context.Context, client *rpc.Client, account common.Address) error {
	var listedAccounts []common.Address
	if err := callRPC(ctx, client, &listedAccounts, "account_list"); err != nil {
		return err
	}
	if !slices.Contains(listedAccounts, account) {
		return fmt.Errorf("account_list does not contain new account %s", account.Hex())
	}
	return nil
}

func verifyDataSigning(
	ctx context.Context,
	client *rpc.Client,
	account common.Address,
	expectedWallet wallet.Wallet,
) error {
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
	return nil
}

func verifyDataRejection(ctx context.Context, client *rpc.Client, account common.Address) error {
	var signature hexutil.Bytes
	err := callRPC(ctx, client, &signature, "account_signData",
		qrlaccounts.MimetypeTextPlain,
		account.Hex(),
		hexutil.Encode([]byte(rejectedText)),
	)
	if err == nil {
		return errors.New("account_signData unexpectedly approved rejected data")
	}
	if !strings.Contains(err.Error(), signercore.ErrRequestDenied.Error()) {
		return fmt.Errorf("account_signData returned unexpected rejection: %w", err)
	}
	return nil
}

func verifyValidatorDataSigning(
	ctx context.Context,
	client *rpc.Client,
	account common.Address,
	expectedWallet wallet.Wallet,
) error {
	validator := common.MustParseAddress(expectedRecipient)
	message := []byte(expectedValidatorText)
	var signature hexutil.Bytes
	if err := callRPC(
		ctx,
		client,
		&signature,
		"account_signData",
		qrlaccounts.MimetypeDataWithValidator,
		account.Hex(),
		map[string]any{
			"address": hexutil.Encode(validator.Bytes()),
			"message": hexutil.Encode(message),
		},
	); err != nil {
		return err
	}
	digest, _ := signercore.SignTextValidator(apitypes.ValidatorData{
		Address: validator,
		Message: message,
	})
	return verifySignature(
		"account_signData data/validator",
		signature,
		digest,
		expectedWallet,
	)
}

func verifyTypedDataSigning(
	ctx context.Context,
	client *rpc.Client,
	account common.Address,
	chainID *big.Int,
	expectedWallet wallet.Wallet,
) error {
	typedSignature, typedDigest, err := signTypedData(ctx, client, account, chainID)
	if err != nil {
		return err
	}
	return verifySignature(
		"account_signTypedData",
		typedSignature,
		typedDigest,
		expectedWallet,
	)
}

func signTypedData(
	ctx context.Context,
	client *rpc.Client,
	account common.Address,
	chainID *big.Int,
) (hexutil.Bytes, []byte, error) {
	typedData := expectedTypedData(account, chainID)
	var typedSignature hexutil.Bytes
	if err := callRPC(ctx, client, &typedSignature, "account_signTypedData",
		account.Hex(),
		typedData,
	); err != nil {
		return nil, nil, err
	}
	typedDigest, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, nil, fmt.Errorf("hash typed data: %w", err)
	}
	return typedSignature, typedDigest, nil
}

func signTransaction(
	ctx context.Context,
	client *rpc.Client,
	transaction apitypes.SendTxArgs,
) (signTransactionResult, error) {
	var signed signTransactionResult
	if err := callRPC(
		ctx,
		client,
		&signed,
		"account_signTransaction",
		transaction,
	); err != nil {
		return signTransactionResult{}, err
	}
	return signed, nil
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
