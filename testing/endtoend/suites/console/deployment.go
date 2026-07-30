// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package console

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/theQRL/go-qrl/accounts/abi"
	"github.com/theQRL/go-qrl/accounts/abi/bind"
	"github.com/theQRL/go-qrl/common/hexutil"
	endtoendlive "github.com/theQRL/go-qrl/testing/endtoend/internal/live"
)

const (
	storeValueDecimal = "6703903964971298549787012499102923063739682910296196688861780721860882015036773488400937149083451713845015929093243025426876941405973284973216824503046708"
	storeLabel        = "indexed dynamic label"
)

func deploymentParameters(
	ctx context.Context,
	session *endtoendlive.Session,
	abiJSON, bytecode []byte,
) ([]byte, error) {
	contractABI, err := abi.JSON(bytes.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("parse contract ABI: %w", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(session.Wallet, session.ChainID)
	if err != nil {
		return nil, fmt.Errorf("create deployment transactor: %w", err)
	}
	auth.Context = ctx
	auth.NoSend = true

	_, tx, contract, err := bind.DeployContract(auth, contractABI, bytecode, session.Client)
	if err != nil {
		return nil, fmt.Errorf("prepare deployment transaction: %w", err)
	}
	deploymentRaw, err := tx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("encode deployment transaction: %w", err)
	}
	storeValue, ok := new(big.Int).SetString(storeValueDecimal, 10)
	if !ok {
		return nil, errors.New("parse store value")
	}
	storePayload := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	storeData, err := contractABI.Pack("store", storeValue, storeLabel, storePayload)
	if err != nil {
		return nil, fmt.Errorf("pack store call: %w", err)
	}
	auth.Nonce = new(big.Int).SetUint64(tx.Nonce() + 1)
	auth.GasLimit = 500_000
	storeTx, err := contract.Transact(auth, "store", storeValue, storeLabel, storePayload)
	if err != nil {
		return nil, fmt.Errorf("prepare store transaction: %w", err)
	}
	storeRaw, err := storeTx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("encode store transaction: %w", err)
	}

	return json.Marshal(struct {
		Address        string          `json:"address"`
		TxHash         string          `json:"txHash"`
		RawTransaction string          `json:"rawTransaction"`
		StoreTxHash    string          `json:"storeTxHash"`
		StoreRaw       string          `json:"storeRawTransaction"`
		StoreData      string          `json:"storeData"`
		StoreValue     string          `json:"storeValue"`
		StoreLabel     string          `json:"storeLabel"`
		StorePayload   string          `json:"storePayload"`
		ABI            json.RawMessage `json:"abi"`
	}{
		Address:        auth.From.Hex(),
		TxHash:         tx.Hash().Hex(),
		RawTransaction: hexutil.Encode(deploymentRaw),
		StoreTxHash:    storeTx.Hash().Hex(),
		StoreRaw:       hexutil.Encode(storeRaw),
		StoreData:      hexutil.Encode(storeData),
		StoreValue:     storeValueDecimal,
		StoreLabel:     storeLabel,
		StorePayload:   hexutil.Encode(storePayload),
		ABI:            abiJSON,
	})
}
