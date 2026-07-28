// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package console

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/theQRL/go-qrl/accounts/abi"
	"github.com/theQRL/go-qrl/accounts/abi/bind"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	qrlwallet "github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
	"github.com/theQRL/go-qrl/qrlclient"
	"github.com/theQRL/go-qrl/testing/devnet/internal/network"
)

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

	contractABI, err := abi.JSON(bytes.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("parse contract ABI: %w", err)
	}
	tx, err := signDeployment(ctx, client, wallet, contractABI, bytecode)
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
	}{
		Address:        from.Hex(),
		TxHash:         tx.Hash().Hex(),
		RawTransaction: hexutil.Encode(raw),
		ABI:            abiJSON,
	})
}

func signDeployment(
	ctx context.Context,
	client *qrlclient.Client,
	wallet qrlwallet.Wallet,
	contractABI abi.ABI,
	bytecode []byte,
) (*types.Transaction, error) {
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("chain ID: %w", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(wallet, chainID)
	if err != nil {
		return nil, fmt.Errorf("create deployment transactor: %w", err)
	}
	auth.Context = ctx
	auth.NoSend = true

	_, tx, _, err := bind.DeployContract(auth, contractABI, bytecode, client)
	if err != nil {
		return nil, fmt.Errorf("prepare deployment transaction: %w", err)
	}
	return tx, nil
}
