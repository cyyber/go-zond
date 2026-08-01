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

	qrl "github.com/theQRL/go-qrl"
	"github.com/theQRL/go-qrl/accounts/abi"
	"github.com/theQRL/go-qrl/accounts/abi/bind"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	qrvm "github.com/theQRL/go-qrl/core/vm"
	endtoendlive "github.com/theQRL/go-qrl/testing/endtoend/internal/live"
)

const (
	storeValueDecimal = "6703903964971298549787012499102923063739682910296196688861780721860882015036773488400937149083451713845015929093243025426876941405973284973216824503046708"
	storeLabel        = "indexed dynamic label"

	constructorABIJSON  = `[{"inputs":[{"name":"amount","type":"uint512"},{"name":"recipient","type":"address"},{"name":"tag","type":"bytes33"},{"name":"payload","type":"bytes"}],"stateMutability":"nonpayable","type":"constructor"}]`
	indexedEventABIJSON = `[{"anonymous":false,"inputs":[{"indexed":true,"name":"flag","type":"bool"},{"indexed":true,"name":"delta","type":"int512"},{"indexed":true,"name":"amount","type":"uint512"}],"name":"IndexedNumbers","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"code","type":"bytes33"}],"name":"IndexedBytes","type":"event"}]`
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
	constructorABI, err := abi.JSON(bytes.NewBufferString(constructorABIJSON))
	if err != nil {
		return nil, fmt.Errorf("parse constructor coverage ABI: %w", err)
	}
	constructorTag := [33]byte{}
	for index := range constructorTag {
		constructorTag[index] = byte(index + 1)
	}
	constructorPayload := make([]byte, 65)
	for index := range constructorPayload {
		constructorPayload[index] = byte(0xff - index)
	}
	constructorSuffix, err := constructorABI.Constructor.Inputs.Pack(
		storeValue,
		auth.From,
		constructorTag,
		constructorPayload,
	)
	if err != nil {
		return nil, fmt.Errorf("pack constructor coverage data: %w", err)
	}
	constructorInput := append(bytes.Clone(bytecode), constructorSuffix...)
	constructorGas, err := session.Client.EstimateGas(ctx, qrl.CallMsg{
		From: auth.From,
		Data: constructorInput,
	})
	if err != nil {
		return nil, fmt.Errorf("estimate constructor coverage gas: %w", err)
	}

	indexedABI, err := abi.JSON(bytes.NewBufferString(indexedEventABIJSON))
	if err != nil {
		return nil, fmt.Errorf("parse indexed-event coverage ABI: %w", err)
	}
	indexedDelta := big.NewInt(-1)
	indexedAmount := new(big.Int).Lsh(big.NewInt(1), 400)
	indexedAmount.Add(indexedAmount, big.NewInt(17))
	indexedCode := [33]byte{}
	for index := range indexedCode {
		indexedCode[index] = byte(0xa0 + index)
	}
	numberEvent := indexedABI.Events["IndexedNumbers"]
	flagTopic, err := packEventTopic(numberEvent.Inputs[0], true)
	if err != nil {
		return nil, err
	}
	deltaTopic, err := packEventTopic(numberEvent.Inputs[1], indexedDelta)
	if err != nil {
		return nil, err
	}
	amountTopic, err := packEventTopic(numberEvent.Inputs[2], indexedAmount)
	if err != nil {
		return nil, err
	}
	bytesEvent := indexedABI.Events["IndexedBytes"]
	codeTopic, err := packEventTopic(bytesEvent.Inputs[0], indexedCode)
	if err != nil {
		return nil, err
	}
	numberTopics := []common.LogTopic{
		common.HashToLogTopic(numberEvent.ID),
		flagTopic,
		deltaTopic,
		amountTopic,
	}
	bytesTopics := []common.LogTopic{common.HashToLogTopic(bytesEvent.ID), codeTopic}

	auth.Nonce = new(big.Int).SetUint64(tx.Nonce() + 1)
	auth.GasLimit = 500_000
	_, indexedTx, _, err := bind.DeployContract(
		auth,
		indexedABI,
		indexedEventInitCode(numberTopics, bytesTopics),
		session.Client,
	)
	if err != nil {
		return nil, fmt.Errorf("prepare indexed-event transaction: %w", err)
	}
	indexedRaw, err := indexedTx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("encode indexed-event transaction: %w", err)
	}

	auth.Nonce = new(big.Int).SetUint64(tx.Nonce() + 2)
	storeTx, err := contract.Transact(auth, "store", storeValue, storeLabel, storePayload)
	if err != nil {
		return nil, fmt.Errorf("prepare store transaction: %w", err)
	}
	storeRaw, err := storeTx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("encode store transaction: %w", err)
	}
	nestedValues := [][]*big.Int{{big.NewInt(1), big.NewInt(2)}, {big.NewInt(3)}}
	nestedData, err := contractABI.Pack("roundTrip", nestedValues)
	if err != nil {
		return nil, fmt.Errorf("pack nested array call: %w", err)
	}

	return json.Marshal(struct {
		Address            string          `json:"address"`
		Bytecode           string          `json:"bytecode"`
		TxHash             string          `json:"txHash"`
		RawTransaction     string          `json:"rawTransaction"`
		StoreTxHash        string          `json:"storeTxHash"`
		StoreRaw           string          `json:"storeRawTransaction"`
		StoreData          string          `json:"storeData"`
		StoreValue         string          `json:"storeValue"`
		StoreLabel         string          `json:"storeLabel"`
		StorePayload       string          `json:"storePayload"`
		NestedData         string          `json:"nestedData"`
		ABI                json.RawMessage `json:"abi"`
		ConstructorABI     json.RawMessage `json:"constructorABI"`
		ConstructorInput   string          `json:"constructorInput"`
		ConstructorGas     uint64          `json:"constructorGas"`
		ConstructorTag     string          `json:"constructorTag"`
		ConstructorPayload string          `json:"constructorPayload"`
		IndexedABI         json.RawMessage `json:"indexedABI"`
		IndexedTxHash      string          `json:"indexedTxHash"`
		IndexedRaw         string          `json:"indexedRawTransaction"`
		IndexedDelta       string          `json:"indexedDelta"`
		IndexedAmount      string          `json:"indexedAmount"`
		IndexedCode        string          `json:"indexedCode"`
		NumberTopics       []string        `json:"numberTopics"`
		BytesTopics        []string        `json:"bytesTopics"`
	}{
		Address:            auth.From.Hex(),
		Bytecode:           hexutil.Encode(bytecode),
		TxHash:             tx.Hash().Hex(),
		RawTransaction:     hexutil.Encode(deploymentRaw),
		StoreTxHash:        storeTx.Hash().Hex(),
		StoreRaw:           hexutil.Encode(storeRaw),
		StoreData:          hexutil.Encode(storeData),
		StoreValue:         storeValueDecimal,
		StoreLabel:         storeLabel,
		StorePayload:       hexutil.Encode(storePayload),
		NestedData:         hexutil.Encode(nestedData),
		ABI:                abiJSON,
		ConstructorABI:     json.RawMessage(constructorABIJSON),
		ConstructorInput:   hexutil.Encode(constructorInput),
		ConstructorGas:     constructorGas,
		ConstructorTag:     hexutil.Encode(constructorTag[:]),
		ConstructorPayload: hexutil.Encode(constructorPayload),
		IndexedABI:         json.RawMessage(indexedEventABIJSON),
		IndexedTxHash:      indexedTx.Hash().Hex(),
		IndexedRaw:         hexutil.Encode(indexedRaw),
		IndexedDelta:       indexedDelta.String(),
		IndexedAmount:      indexedAmount.String(),
		IndexedCode:        hexutil.Encode(indexedCode[:]),
		NumberTopics:       topicStrings(numberTopics),
		BytesTopics:        topicStrings(bytesTopics),
	})
}

func packEventTopic(argument abi.Argument, value any) (common.LogTopic, error) {
	encoded, err := (abi.Arguments{argument}).Pack(value)
	if err != nil {
		return common.LogTopic{}, fmt.Errorf("pack indexed %s topic: %w", argument.Type, err)
	}
	if len(encoded) != common.LogTopicLength {
		return common.LogTopic{}, fmt.Errorf(
			"pack indexed %s topic: got %d bytes",
			argument.Type,
			len(encoded),
		)
	}
	var topic common.LogTopic
	copy(topic[:], encoded)
	return topic, nil
}

func indexedEventInitCode(events ...[]common.LogTopic) []byte {
	var code []byte
	for _, topics := range events {
		for index := len(topics) - 1; index >= 0; index-- {
			code = append(code, byte(qrvm.PUSH64))
			code = append(code, topics[index][:]...)
		}
		code = append(
			code,
			byte(qrvm.PUSH1), 0,
			byte(qrvm.PUSH1), 0,
			byte(qrvm.LOG0)+byte(len(topics)),
		)
	}
	return append(code,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.PUSH1), 0,
		byte(qrvm.RETURN),
	)
}

func topicStrings(topics []common.LogTopic) []string {
	encoded := make([]string, len(topics))
	for index, topic := range topics {
		encoded[index] = topic.Hex()
	}
	return encoded
}
