// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package console

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	qrl "github.com/theQRL/go-qrl"
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

	signed, err := types.SignNewTx(wallet, types.LatestSignerForChainID(chainID), &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gas,
		Value:     new(big.Int),
		Data:      bytecode,
	})
	if err != nil {
		return nil, fmt.Errorf("sign deployment transaction: %w", err)
	}
	return signed, nil
}
