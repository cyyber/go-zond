// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package bind_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	qrl "github.com/theQRL/go-qrl"
	"github.com/theQRL/go-qrl/accounts/abi"
	"github.com/theQRL/go-qrl/accounts/abi/bind"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/crypto"
	"github.com/theQRL/go-qrl/event"
)

func mockSign(addr common.Address, tx *types.Transaction) (*types.Transaction, error) { return tx, nil }

type mockTransactor struct {
	baseFee                *big.Int
	gasTipCap              *big.Int
	gasPrice               *big.Int
	suggestGasTipCapCalled bool
	suggestGasPriceCalled  bool
}

func (mt *mockTransactor) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return &types.Header{BaseFee: mt.baseFee}, nil
}

func (mt *mockTransactor) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	return []byte{1}, nil
}

func (mt *mockTransactor) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return 0, nil
}

func (mt *mockTransactor) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	mt.suggestGasPriceCalled = true
	return mt.gasPrice, nil
}

func (mt *mockTransactor) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	mt.suggestGasTipCapCalled = true
	return mt.gasTipCap, nil
}

func (mt *mockTransactor) EstimateGas(ctx context.Context, call qrl.CallMsg) (gas uint64, err error) {
	return 0, nil
}

func (mt *mockTransactor) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return nil
}

type mockCaller struct {
	codeAtBlockNumber       *big.Int
	callContractBlockNumber *big.Int
	callContractBytes       []byte
	callContractErr         error
	codeAtBytes             []byte
	codeAtErr               error
}

type recordingFilterer struct {
	filtered []qrl.FilterQuery
	watched  []qrl.FilterQuery
}

func (filterer *recordingFilterer) FilterLogs(_ context.Context, query qrl.FilterQuery) ([]types.Log, error) {
	filterer.filtered = append(filterer.filtered, query)
	return nil, nil
}

func (filterer *recordingFilterer) SubscribeFilterLogs(_ context.Context, query qrl.FilterQuery, _ chan<- types.Log) (qrl.Subscription, error) {
	filterer.watched = append(filterer.watched, query)
	return event.NewSubscription(func(quit <-chan struct{}) error {
		<-quit
		return nil
	}), nil
}

func (mc *mockCaller) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	mc.codeAtBlockNumber = blockNumber
	return mc.codeAtBytes, mc.codeAtErr
}

func (mc *mockCaller) CallContract(ctx context.Context, call qrl.CallMsg, blockNumber *big.Int) ([]byte, error) {
	mc.callContractBlockNumber = blockNumber
	return mc.callContractBytes, mc.callContractErr
}

type mockPendingCaller struct {
	*mockCaller
	pendingCodeAtBytes        []byte
	pendingCodeAtErr          error
	pendingCodeAtCalled       bool
	pendingCallContractCalled bool
	pendingCallContractBytes  []byte
	pendingCallContractErr    error
}

func (mc *mockPendingCaller) PendingCodeAt(ctx context.Context, contract common.Address) ([]byte, error) {
	mc.pendingCodeAtCalled = true
	return mc.pendingCodeAtBytes, mc.pendingCodeAtErr
}

func (mc *mockPendingCaller) PendingCallContract(ctx context.Context, call qrl.CallMsg) ([]byte, error) {
	mc.pendingCallContractCalled = true
	return mc.pendingCallContractBytes, mc.pendingCallContractErr
}

func TestPassingBlockNumber(t *testing.T) {
	mc := &mockPendingCaller{
		mockCaller: &mockCaller{
			codeAtBytes: []byte{1, 2, 3},
		},
	}
	contractAddr := common.MustParseAddress("Q00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	bc := bind.NewBoundContract(contractAddr, abi.ABI{
		Methods: map[string]abi.Method{
			"something": {
				Name:    "something",
				Outputs: abi.Arguments{},
			},
		},
	}, mc, nil, nil)

	blockNumber := big.NewInt(42)

	bc.Call(&bind.CallOpts{BlockNumber: blockNumber}, nil, "something")

	if mc.callContractBlockNumber != blockNumber {
		t.Fatalf("CallContract() was not passed the block number")
	}

	if mc.codeAtBlockNumber != blockNumber {
		t.Fatalf("CodeAt() was not passed the block number")
	}

	bc.Call(&bind.CallOpts{}, nil, "something")

	if mc.callContractBlockNumber != nil {
		t.Fatalf("CallContract() was passed a block number when it should not have been")
	}

	if mc.codeAtBlockNumber != nil {
		t.Fatalf("CodeAt() was passed a block number when it should not have been")
	}

	bc.Call(&bind.CallOpts{BlockNumber: blockNumber, Pending: true}, nil, "something")

	if !mc.pendingCallContractCalled {
		t.Fatalf("CallContract() was not passed the block number")
	}

	if !mc.pendingCodeAtCalled {
		t.Fatalf("CodeAt() was not passed the block number")
	}
}

// mockSender is the 64-byte address encoded into hexData (the non-indexed
// portion of every mock "received(...)" event used in these tests). It also
// replaces the 20-byte hex literals that used to be parsed via
// NewAddressFromString.
var mockSender = func() common.Address {
	raw, _ := hex.DecodeString("978271565f56deb45495afa69e59c16ab2376c47978271565f56deb45495afa69e59c16ab2112233445566778899aabb")
	return common.BytesToAddress(raw)
}()

// hexData is the ABI-encoded non-indexed payload (address, uint256, bytes) =
// (mockSender, 1, [88]) as produced by abi.Pack with a 64-byte slot width.
var hexData = func() []byte {
	const spec = `[{"name":"f","type":"function","inputs":[{"type":"address"},{"type":"uint256"},{"type":"bytes"}]}]`
	a, _ := abi.JSON(strings.NewReader(spec))
	packed, _ := a.Pack("f", mockSender, big.NewInt(1), []byte{88})
	return packed[4:]
}()

func TestUnpackIndexedStringTyLogIntoMap(t *testing.T) {
	hash := crypto.Keccak256Hash([]byte("testName"))
	topics := []common.LogTopic{
		common.HashToLogTopic(crypto.Keccak256Hash([]byte("received(string,address,uint256,bytes)"))),
		common.HashToLogTopic(hash),
	}
	mockLog := newMockLog(topics, common.HexToHash("0x0"))

	abiString := `[{"anonymous":false,"inputs":[{"indexed":true,"name":"name","type":"string"},{"indexed":false,"name":"sender","type":"address"},{"indexed":false,"name":"amount","type":"uint256"},{"indexed":false,"name":"memo","type":"bytes"}],"name":"received","type":"event"}]`
	parsedAbi, _ := abi.JSON(strings.NewReader(abiString))
	bc := bind.NewBoundContract(common.Address{}, parsedAbi, nil, nil, nil)

	expectedReceivedMap := map[string]any{
		"name":   hash,
		"sender": mockSender,
		"amount": big.NewInt(1),
		"memo":   []byte{88},
	}
	unpackAndCheck(t, bc, expectedReceivedMap, mockLog)
}

func TestUnpackNonAnonymousLogWithoutSignatureIntoMap(t *testing.T) {
	mockLog := newMockLog(nil, common.HexToHash("0x0"))

	abiString := `[{"anonymous":false,"inputs":[{"indexed":false,"name":"amount","type":"uint256"}],"name":"received","type":"event"}]`
	parsedAbi, _ := abi.JSON(strings.NewReader(abiString))
	contractAddr := common.MustParseAddress("Q00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	bc := bind.NewBoundContract(contractAddr, parsedAbi, nil, nil, nil)

	var received map[string]any
	err := bc.UnpackLogIntoMap(received, "received", mockLog)
	if err == nil {
		t.Fatal("expected a non-anonymous event without its signature topic to fail")
	}
	if err.Error() != "no event signature" {
		t.Errorf("expected error 'no event signature', got '%s'", err)
	}
}

func TestUnpackIndexedSliceTyLogIntoMap(t *testing.T) {
	hash := crypto.Keccak256Hash([]byte("opaque indexed slice digest"))
	topics := []common.LogTopic{
		common.HashToLogTopic(crypto.Keccak256Hash([]byte("received(string[],address,uint256,bytes)"))),
		common.HashToLogTopic(hash),
	}
	mockLog := newMockLog(topics, common.HexToHash("0x0"))

	abiString := `[{"anonymous":false,"inputs":[{"indexed":true,"name":"names","type":"string[]"},{"indexed":false,"name":"sender","type":"address"},{"indexed":false,"name":"amount","type":"uint256"},{"indexed":false,"name":"memo","type":"bytes"}],"name":"received","type":"event"}]`
	parsedAbi, _ := abi.JSON(strings.NewReader(abiString))
	bc := bind.NewBoundContract(common.Address{}, parsedAbi, nil, nil, nil)

	expectedReceivedMap := map[string]any{
		"names":  hash,
		"sender": mockSender,
		"amount": big.NewInt(1),
		"memo":   []byte{88},
	}
	unpackAndCheck(t, bc, expectedReceivedMap, mockLog)
}

func TestUnpackIndexedArrayTyLogIntoMap(t *testing.T) {
	hash := crypto.Keccak256Hash([]byte("opaque indexed array digest"))
	topics := []common.LogTopic{
		common.HashToLogTopic(crypto.Keccak256Hash([]byte("received(address[2],address,uint256,bytes)"))),
		common.HashToLogTopic(hash),
	}
	mockLog := newMockLog(topics, common.HexToHash("0x0"))

	abiString := `[{"anonymous":false,"inputs":[{"indexed":true,"name":"addresses","type":"address[2]"},{"indexed":false,"name":"sender","type":"address"},{"indexed":false,"name":"amount","type":"uint256"},{"indexed":false,"name":"memo","type":"bytes"}],"name":"received","type":"event"}]`
	parsedAbi, _ := abi.JSON(strings.NewReader(abiString))
	bc := bind.NewBoundContract(common.Address{}, parsedAbi, nil, nil, nil)

	expectedReceivedMap := map[string]any{
		"addresses": hash,
		"sender":    mockSender,
		"amount":    big.NewInt(1),
		"memo":      []byte{88},
	}
	unpackAndCheck(t, bc, expectedReceivedMap, mockLog)
}

func TestUnpackIndexedFuncTyLogIntoMap(t *testing.T) {
	addrBytes := mockSender.Bytes()
	hash := crypto.Keccak256Hash([]byte("mockFunction(address,uint)"))
	functionSelector := hash[:4]
	functionTyBytes := append(addrBytes, functionSelector...)
	topics := []common.LogTopic{
		common.HashToLogTopic(crypto.Keccak256Hash([]byte("received(function,address,uint256,bytes)"))),
		// function values are bytesN-family, hence left-aligned; a placeholder
		// until the double-word representation defines the real layout.
		common.BytesToLeftAlignedLogTopic(functionTyBytes),
	}
	mockLog := newMockLog(topics, common.HexToHash("0x5c698f13940a2153440c6d19660878bc90219d9298fdcf37365aa8d88d40fc42"))
	abiString := `[{"anonymous":false,"inputs":[{"indexed":true,"name":"function","type":"function"},{"indexed":false,"name":"sender","type":"address"},{"indexed":false,"name":"amount","type":"uint256"},{"indexed":false,"name":"memo","type":"bytes"}],"name":"received","type":"event"}]`
	parsedAbi, _ := abi.JSON(strings.NewReader(abiString))
	bc := bind.NewBoundContract(common.Address{}, parsedAbi, nil, nil, nil)

	received := make(map[string]any)
	err := bc.UnpackLogIntoMap(received, "received", mockLog)
	if err == nil {
		t.Fatal("expected indexed function topic to be rejected")
	}
	if !strings.Contains(err.Error(), "function type does not fit") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnpackIndexedBytesTyLogIntoMap(t *testing.T) {
	bytes := []byte{1, 2, 3, 4, 5}
	hash := crypto.Keccak256Hash(bytes)
	topics := []common.LogTopic{
		common.HashToLogTopic(crypto.Keccak256Hash([]byte("received(bytes,address,uint256,bytes)"))),
		common.HashToLogTopic(hash),
	}
	mockLog := newMockLog(topics, common.HexToHash("0x5c698f13940a2153440c6d19660878bc90219d9298fdcf37365aa8d88d40fc42"))

	abiString := `[{"anonymous":false,"inputs":[{"indexed":true,"name":"content","type":"bytes"},{"indexed":false,"name":"sender","type":"address"},{"indexed":false,"name":"amount","type":"uint256"},{"indexed":false,"name":"memo","type":"bytes"}],"name":"received","type":"event"}]`
	parsedAbi, _ := abi.JSON(strings.NewReader(abiString))
	bc := bind.NewBoundContract(common.Address{}, parsedAbi, nil, nil, nil)

	expectedReceivedMap := map[string]any{
		"content": hash,
		"sender":  mockSender,
		"amount":  big.NewInt(1),
		"memo":    []byte{88},
	}
	unpackAndCheck(t, bc, expectedReceivedMap, mockLog)
}

func TestTransactGasFee(t *testing.T) {
	assert := assert.New(t)

	// GasTipCap and GasFeeCap
	// When opts.GasTipCap and opts.GasFeeCap are nil
	mt := &mockTransactor{baseFee: big.NewInt(100), gasTipCap: big.NewInt(5)}
	bc := bind.NewBoundContract(common.Address{}, abi.ABI{}, nil, mt, nil)
	opts := &bind.TransactOpts{Signer: mockSign}
	tx, err := bc.Transact(opts, "")
	assert.Nil(err)
	assert.Equal(big.NewInt(5), tx.GasTipCap())
	assert.Equal(big.NewInt(205), tx.GasFeeCap())
	assert.Nil(opts.GasTipCap)
	assert.Nil(opts.GasFeeCap)
	assert.True(mt.suggestGasTipCapCalled)

	// Second call to Transact should use latest suggested GasTipCap
	mt.gasTipCap = big.NewInt(6)
	mt.suggestGasTipCapCalled = false
	tx, err = bc.Transact(opts, "")
	assert.Nil(err)
	assert.Equal(big.NewInt(6), tx.GasTipCap())
	assert.Equal(big.NewInt(206), tx.GasFeeCap())
	assert.True(mt.suggestGasTipCapCalled)
}

func unpackAndCheck(t *testing.T, bc *bind.BoundContract, expected map[string]any, mockLog types.Log) {
	received := make(map[string]any)
	if err := bc.UnpackLogIntoMap(received, "received", mockLog); err != nil {
		t.Error(err)
	}

	if len(received) != len(expected) {
		t.Fatalf("unpacked map length %v not equal expected length of %v", len(received), len(expected))
	}
	for name, elem := range expected {
		if !reflect.DeepEqual(elem, received[name]) {
			t.Errorf("field %v does not match expected, want %v, got %v", name, elem, received[name])
		}
	}
}

func newMockLog(topics []common.LogTopic, txHash common.Hash) types.Log {
	return types.Log{
		Address:     common.Address{},
		Topics:      topics,
		Data:        hexData,
		BlockNumber: uint64(26),
		TxHash:      txHash,
		TxIndex:     111,
		BlockHash:   common.BytesToHash([]byte{1, 2, 3, 4, 5}),
		Index:       7,
		Removed:     false,
	}
}

func TestBoundContractAnonymousEventTopicsAndUnpack(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(`[
		{"anonymous":true,"inputs":[{"indexed":true,"name":"range","type":"address"},{"indexed":false,"name":"type","type":"bytes64"}],"name":"Anonymous","type":"event"},
		{"anonymous":true,"inputs":[{"indexed":false,"name":"value","type":"uint512"}],"name":"AnonymousNoIndex","type":"event"},
		{"anonymous":false,"inputs":[],"name":"Regular","type":"event"}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	filterer := new(recordingFilterer)
	contractAddress := common.BytesToAddress([]byte{0x99})
	contract := bind.NewBoundContract(contractAddress, parsed, nil, nil, filterer)

	// Anonymous filters must not acquire a synthetic signature topic. A nil
	// indexed rule is retained as the wildcard at topic position zero.
	_, subscription, err := contract.FilterLogs(nil, "Anonymous", nil)
	if err != nil {
		t.Fatalf("FilterLogs(anonymous) failed: %v", err)
	}
	subscription.Unsubscribe()
	if got := filterer.filtered[len(filterer.filtered)-1].Topics; len(got) != 1 || got[0] != nil {
		t.Fatalf("anonymous wildcard topics = %#v, want one nil topic rule", got)
	}
	_, subscription, err = contract.FilterLogs(nil, "AnonymousNoIndex")
	if err != nil {
		t.Fatalf("FilterLogs(anonymous no-index) failed: %v", err)
	}
	subscription.Unsubscribe()
	if got := filterer.filtered[len(filterer.filtered)-1].Topics; len(got) != 0 {
		t.Fatalf("anonymous no-index topics = %#v, want empty", got)
	}
	_, subscription, err = contract.FilterLogs(nil, "Regular")
	if err != nil {
		t.Fatalf("FilterLogs(regular) failed: %v", err)
	}
	subscription.Unsubscribe()
	regularTopics := filterer.filtered[len(filterer.filtered)-1].Topics
	if len(regularTopics) != 1 || len(regularTopics[0]) != 1 || regularTopics[0][0] != common.HashToLogTopic(parsed.Events["Regular"].ID) {
		t.Fatalf("regular event topics = %#v, want signature", regularTopics)
	}
	_, subscription, err = contract.WatchLogs(nil, "Anonymous", nil)
	if err != nil {
		t.Fatalf("WatchLogs(anonymous) failed: %v", err)
	}
	if got := filterer.watched[len(filterer.watched)-1].Topics; len(got) != 1 || got[0] != nil {
		t.Fatalf("watched anonymous topics = %#v, want one nil topic rule", got)
	}
	subscription.Unsubscribe()

	address := common.BytesToAddress([]byte{0x42})
	topicSets, err := abi.MakeTopics([]any{address})
	if err != nil {
		t.Fatal(err)
	}
	var payload [64]byte
	payload[0], payload[len(payload)-1] = 0xaa, 0x55
	data, err := parsed.Events["Anonymous"].Inputs.NonIndexed().Pack(payload)
	if err != nil {
		t.Fatal(err)
	}
	log := types.Log{Topics: []common.LogTopic{topicSets[0][0]}, Data: data}
	var decoded struct {
		Address common.Address `abi:"range"`
		Payload [64]byte       `abi:"type"`
	}
	if err := contract.UnpackLog(&decoded, "Anonymous", log); err != nil {
		t.Fatalf("UnpackLog(anonymous) failed: %v", err)
	}
	if decoded.Address != address || decoded.Payload != payload {
		t.Fatalf("decoded anonymous event = %#v, want address %v and payload %x", decoded, address, payload)
	}
	decodedMap := make(map[string]any)
	if err := contract.UnpackLogIntoMap(decodedMap, "Anonymous", log); err != nil {
		t.Fatalf("UnpackLogIntoMap(anonymous) failed: %v", err)
	}
	if decodedMap["range"] != address || decodedMap["type"] != payload {
		t.Fatalf("decoded anonymous event map = %#v", decodedMap)
	}
	noIndexData, err := parsed.Events["AnonymousNoIndex"].Inputs.NonIndexed().Pack(big.NewInt(77))
	if err != nil {
		t.Fatal(err)
	}
	var noIndexDecoded struct {
		Value *big.Int `abi:"value"`
	}
	if err := contract.UnpackLog(&noIndexDecoded, "AnonymousNoIndex", types.Log{Data: noIndexData}); err != nil {
		t.Fatalf("UnpackLog(anonymous no-index) failed: %v", err)
	}
	if noIndexDecoded.Value == nil || noIndexDecoded.Value.Int64() != 77 {
		t.Fatalf("decoded anonymous no-index value = %v, want 77", noIndexDecoded.Value)
	}
	noIndexMap := make(map[string]any)
	if err := contract.UnpackLogIntoMap(noIndexMap, "AnonymousNoIndex", types.Log{Data: noIndexData}); err != nil {
		t.Fatalf("UnpackLogIntoMap(anonymous no-index) failed: %v", err)
	}
	if got, ok := noIndexMap["value"].(*big.Int); !ok || got.Int64() != 77 {
		t.Fatalf("decoded anonymous no-index map = %#v, want value 77", noIndexMap)
	}

	for _, topics := range [][]common.LogTopic{nil, {}} {
		if err := contract.UnpackLog(&decoded, "Regular", types.Log{Topics: topics}); err == nil || err.Error() != "no event signature" {
			t.Fatalf("regular event without signature via UnpackLog topics=%#v error=%v, want no event signature", topics, err)
		}
		if err := contract.UnpackLogIntoMap(make(map[string]any), "Regular", types.Log{Topics: topics}); err == nil || err.Error() != "no event signature" {
			t.Fatalf("regular event without signature via UnpackLogIntoMap topics=%#v error=%v, want no event signature", topics, err)
		}
	}
	if err := contract.UnpackLog(&decoded, "Regular", types.Log{Topics: []common.LogTopic{{0x01}}}); err == nil || err.Error() != "event signature mismatch" {
		t.Fatalf("regular event with wrong signature error = %v, want event signature mismatch", err)
	}
	var regularDecoded struct{}
	if err := contract.UnpackLog(&regularDecoded, "Regular", types.Log{Topics: []common.LogTopic{common.HashToLogTopic(parsed.Events["Regular"].ID)}}); err != nil {
		t.Fatalf("UnpackLog(regular signature only) failed: %v", err)
	}
	if _, _, err := contract.FilterLogs(nil, "Missing"); err == nil {
		t.Fatal("FilterLogs accepted an unknown event")
	}
}

func TestBoundContractAnonymousFourIndexedTopics(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(`[
		{"anonymous":true,"inputs":[
			{"indexed":true,"name":"marker","type":"bytes64"},
			{"indexed":true,"name":"account","type":"address"},
			{"indexed":true,"name":"code","type":"uint8"},
			{"indexed":true,"name":"enabled","type":"bool"}
		],"name":"AnonymousFour","type":"event"},
		{"anonymous":false,"inputs":[],"name":"SignatureCollision","type":"event"}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	filterer := new(recordingFilterer)
	contract := bind.NewBoundContract(common.BytesToAddress([]byte{0x99}), parsed, nil, nil, filterer)

	// An anonymous event may use all four topic positions for indexed values.
	// In particular, a value which happens to equal another event's signature
	// is still data and must not be stripped as a selector.
	signatureTopic := common.HashToLogTopic(parsed.Events["SignatureCollision"].ID)
	marker := [64]byte(signatureTopic)
	account := common.BytesToAddress([]byte{0x42})
	rules := [][]any{{marker}, {account}, {uint8(0x7a)}, {true}}
	wantTopics, err := abi.MakeTopics(rules...)
	if err != nil {
		t.Fatal(err)
	}

	_, subscription, err := contract.FilterLogs(nil, "AnonymousFour", rules...)
	if err != nil {
		t.Fatalf("FilterLogs(anonymous four-topic event) failed: %v", err)
	}
	subscription.Unsubscribe()
	gotTopics := filterer.filtered[len(filterer.filtered)-1].Topics
	if !reflect.DeepEqual(gotTopics, wantTopics) {
		t.Fatalf("anonymous four-topic filter = %#v, want %#v", gotTopics, wantTopics)
	}
	if len(gotTopics) != 4 || gotTopics[0][0] != signatureTopic {
		t.Fatalf("anonymous filter topics = %#v, want four data topics beginning with signature collision %x", gotTopics, signatureTopic)
	}

	_, subscription, err = contract.WatchLogs(nil, "AnonymousFour", rules...)
	if err != nil {
		t.Fatalf("WatchLogs(anonymous four-topic event) failed: %v", err)
	}
	gotTopics = filterer.watched[len(filterer.watched)-1].Topics
	if !reflect.DeepEqual(gotTopics, wantTopics) {
		t.Fatalf("anonymous four-topic watch = %#v, want %#v", gotTopics, wantTopics)
	}
	subscription.Unsubscribe()

	log := types.Log{Topics: []common.LogTopic{
		wantTopics[0][0],
		wantTopics[1][0],
		wantTopics[2][0],
		wantTopics[3][0],
	}}
	var decoded struct {
		Marker  [64]byte       `abi:"marker"`
		Account common.Address `abi:"account"`
		Code    uint8          `abi:"code"`
		Enabled bool           `abi:"enabled"`
	}
	if err := contract.UnpackLog(&decoded, "AnonymousFour", log); err != nil {
		t.Fatalf("UnpackLog(anonymous four-topic event) failed: %v", err)
	}
	if decoded.Marker != marker || decoded.Account != account || decoded.Code != 0x7a || !decoded.Enabled {
		t.Fatalf("decoded anonymous four-topic event = %#v", decoded)
	}
	decodedMap := make(map[string]any)
	if err := contract.UnpackLogIntoMap(decodedMap, "AnonymousFour", log); err != nil {
		t.Fatalf("UnpackLogIntoMap(anonymous four-topic event) failed: %v", err)
	}
	if decodedMap["marker"] != marker || decodedMap["account"] != account || decodedMap["code"] != uint8(0x7a) || decodedMap["enabled"] != true {
		t.Fatalf("decoded anonymous four-topic map = %#v", decodedMap)
	}

	for _, test := range []struct {
		name   string
		topics []common.LogTopic
	}{
		{name: "zero", topics: nil},
		{name: "too few", topics: log.Topics[:3]},
		{name: "too many", topics: append(append([]common.LogTopic{}, log.Topics...), common.LogTopic{0x01})},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := contract.UnpackLog(&decoded, "AnonymousFour", types.Log{Topics: test.topics}); err == nil || !strings.Contains(err.Error(), "topic/field count mismatch") {
				t.Fatalf("UnpackLog topics=%d error=%v, want topic/field count mismatch", len(test.topics), err)
			}
			if err := contract.UnpackLogIntoMap(make(map[string]any), "AnonymousFour", types.Log{Topics: test.topics}); err == nil || !strings.Contains(err.Error(), "topic/field count mismatch") {
				t.Fatalf("UnpackLogIntoMap topics=%d error=%v, want topic/field count mismatch", len(test.topics), err)
			}
		})
	}
	if _, _, err := contract.FilterLogs(nil, "AnonymousFour", append(rules, []any{uint8(1)})...); err == nil {
		t.Fatal("FilterLogs accepted five rules for a four-indexed-topic anonymous event")
	}
	if _, _, err := contract.WatchLogs(nil, "AnonymousFour", append(rules, []any{uint8(1)})...); err == nil {
		t.Fatal("WatchLogs accepted five rules for a four-indexed-topic anonymous event")
	}
}

func TestBoundContractIndexedCompositeHashTopics(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(`[
		{"anonymous":true,"inputs":[
			{"indexed":true,"name":"values","type":"uint512[]"},
			{"indexed":true,"name":"tags","type":"bytes64[2]"},
			{"components":[{"name":"amount","type":"uint512"},{"name":"recipient","type":"address"}],"indexed":true,"internalType":"struct Composite.Record","name":"record","type":"tuple"},
			{"indexed":true,"name":"note","type":"string"}
		],"name":"Composite","type":"event"}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	filterer := new(recordingFilterer)
	contract := bind.NewBoundContract(common.BytesToAddress([]byte{0x98}), parsed, nil, nil, filterer)
	valuesHash := common.BytesToHash([]byte("values"))
	tagsHash := common.BytesToHash([]byte("tags"))
	recordHash := common.BytesToHash([]byte("record"))
	note := "indexed composite VM64 note"
	noteHash := crypto.Keccak256Hash([]byte(note))

	_, subscription, err := contract.FilterLogs(
		nil,
		"Composite",
		[]any{valuesHash},
		[]any{tagsHash},
		[]any{recordHash},
		[]any{note},
	)
	if err != nil {
		t.Fatalf("FilterLogs(composite): %v", err)
	}
	subscription.Unsubscribe()
	got := filterer.filtered[len(filterer.filtered)-1].Topics
	want := [][]common.LogTopic{
		{common.HashToLogTopic(valuesHash)},
		{common.HashToLogTopic(tagsHash)},
		{common.HashToLogTopic(recordHash)},
		{common.HashToLogTopic(noteHash)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("composite filter topics = %#v, want %#v", got, want)
	}
	valuesAlternative := common.BytesToHash([]byte("values-alternative"))
	_, subscription, err = contract.FilterLogs(
		nil,
		"Composite",
		[]any{valuesHash, valuesAlternative},
		nil,
		[]any{recordHash},
		[]any{note, "other note"},
	)
	if err != nil {
		t.Fatalf("FilterLogs(composite OR/wildcard): %v", err)
	}
	subscription.Unsubscribe()
	got = filterer.filtered[len(filterer.filtered)-1].Topics
	if len(got) != 4 || len(got[0]) != 2 || len(got[1]) != 0 || len(got[2]) != 1 || len(got[3]) != 2 ||
		got[0][0] != common.HashToLogTopic(valuesHash) || got[0][1] != common.HashToLogTopic(valuesAlternative) {
		t.Fatalf("composite OR/wildcard filter topics = %#v", got)
	}

	log := types.Log{Topics: []common.LogTopic{want[0][0], want[1][0], want[2][0], want[3][0]}}
	type decodedComposite struct {
		Values common.Hash `abi:"values"`
		Tags   common.Hash `abi:"tags"`
		Record common.Hash `abi:"record"`
		Note   common.Hash `abi:"note"`
	}
	var decoded decodedComposite
	if err := contract.UnpackLog(&decoded, "Composite", log); err != nil {
		t.Fatalf("UnpackLog(composite): %v", err)
	}
	if decoded.Values != valuesHash || decoded.Tags != tagsHash || decoded.Record != recordHash || decoded.Note != noteHash {
		t.Fatalf("decoded composite hashes = %+v", decoded)
	}
	mapped := make(map[string]any)
	if err := contract.UnpackLogIntoMap(mapped, "Composite", log); err != nil {
		t.Fatalf("UnpackLogIntoMap(composite): %v", err)
	}
	if mapped["values"] != valuesHash || mapped["tags"] != tagsHash || mapped["record"] != recordHash || mapped["note"] != noteHash {
		t.Fatalf("mapped composite hashes = %#v", mapped)
	}

	malformed := log
	malformed.Topics = append([]common.LogTopic{}, log.Topics...)
	malformed.Topics[0][common.HashLength] = 1
	if err := contract.UnpackLog(&decoded, "Composite", malformed); err == nil {
		t.Fatal("UnpackLog accepted a composite hash with non-zero VM64 topic padding")
	}
}

func TestBoundContractOverloadedEventFilters(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(`[
		{"anonymous":false,"inputs":[{"indexed":true,"name":"value","type":"address"}],"name":"Changed","type":"event"},
		{"anonymous":false,"inputs":[{"indexed":true,"name":"value","type":"uint512"}],"name":"Changed","type":"event"}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Events["Changed"].Sig != "Changed(address)" || parsed.Events["Changed0"].Sig != "Changed(uint512)" {
		t.Fatalf("overloaded events = %#v", parsed.Events)
	}
	filterer := new(recordingFilterer)
	contract := bind.NewBoundContract(common.BytesToAddress([]byte{0x77}), parsed, nil, nil, filterer)
	address := common.BytesToAddress(bytes.Repeat([]byte{0x5a}, common.AddressLength))
	amount := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1))
	for _, vector := range []struct {
		name  string
		value any
	}{
		{name: "Changed", value: address},
		{name: "Changed0", value: amount},
	} {
		_, subscription, err := contract.FilterLogs(nil, vector.name, []any{vector.value})
		if err != nil {
			t.Fatalf("FilterLogs(%s): %v", vector.name, err)
		}
		subscription.Unsubscribe()
		query := filterer.filtered[len(filterer.filtered)-1].Topics
		valueTopics, err := abi.MakeTopics([]any{vector.value})
		if err != nil {
			t.Fatal(err)
		}
		if len(query) != 2 || len(query[0]) != 1 || len(query[1]) != 1 || query[0][0] != common.HashToLogTopic(parsed.Events[vector.name].ID) || query[1][0] != valueTopics[0][0] {
			t.Fatalf("overloaded %s filter topics = %#v", vector.name, query)
		}
		decoded := make(map[string]any)
		log := types.Log{Topics: []common.LogTopic{query[0][0], query[1][0]}}
		if err := contract.UnpackLogIntoMap(decoded, vector.name, log); err != nil {
			t.Fatalf("UnpackLogIntoMap(%s): %v", vector.name, err)
		}
		if !reflect.DeepEqual(decoded["value"], vector.value) {
			t.Fatalf("overloaded %s decoded value = (%T) %#v, want (%T) %#v", vector.name, decoded["value"], decoded["value"], vector.value, vector.value)
		}
	}
}

func TestBoundContractValidatesDeclaredIndexedQueryTypes(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(`[
		{"anonymous":false,"inputs":[
			{"indexed":true,"name":"unsigned","type":"uint128"},
			{"indexed":true,"name":"signed","type":"int72"}
		],"name":"Numbers","type":"event"}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	filterer := new(recordingFilterer)
	contract := bind.NewBoundContract(common.Address{}, parsed, nil, nil, filterer)
	uintMax := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	uintOverflow := new(big.Int).Add(new(big.Int).Set(uintMax), big.NewInt(1))
	intLimit := new(big.Int).Lsh(big.NewInt(1), 71)
	intMax := new(big.Int).Sub(new(big.Int).Set(intLimit), big.NewInt(1))
	intMin := new(big.Int).Neg(new(big.Int).Set(intLimit))
	intOverflow := new(big.Int).Set(intLimit)
	intUnderflow := new(big.Int).Sub(new(big.Int).Set(intMin), big.NewInt(1))

	invalid := []struct {
		name  string
		query [][]any
	}{
		{name: "negative uint", query: [][]any{{big.NewInt(-1)}}},
		{name: "uint overflow", query: [][]any{{uintOverflow}}},
		{name: "signed overflow", query: [][]any{nil, {intOverflow}}},
		{name: "signed underflow", query: [][]any{nil, {intUnderflow}}},
		{name: "invalid OR alternative", query: [][]any{{big.NewInt(0), uintOverflow}, {intMin, intMax}}},
	}
	for _, path := range []struct {
		name string
		run  func([][]any) error
	}{
		{
			name: "filter",
			run: func(query [][]any) error {
				_, _, err := contract.FilterLogs(nil, "Numbers", query...)
				return err
			},
		},
		{
			name: "watch",
			run: func(query [][]any) error {
				_, _, err := contract.WatchLogs(nil, "Numbers", query...)
				return err
			},
		},
	} {
		for _, test := range invalid {
			t.Run(path.name+"/"+test.name, func(t *testing.T) {
				filtered, watched := len(filterer.filtered), len(filterer.watched)
				if err := path.run(test.query); err == nil {
					t.Fatalf("%s accepted invalid indexed query %#v", path.name, test.query)
				}
				if len(filterer.filtered) != filtered || len(filterer.watched) != watched {
					t.Fatalf("%s submitted invalid indexed query to backend", path.name)
				}
			})
		}
	}

	valid := [][]any{{big.NewInt(0), uintMax}, {intMin, intMax}}
	_, filterSub, err := contract.FilterLogs(nil, "Numbers", valid...)
	if err != nil {
		t.Fatalf("FilterLogs valid OR query: %v", err)
	}
	filterSub.Unsubscribe()
	filterTopics := filterer.filtered[len(filterer.filtered)-1].Topics
	if len(filterTopics) != 3 || len(filterTopics[0]) != 1 || len(filterTopics[1]) != 2 || len(filterTopics[2]) != 2 {
		t.Fatalf("FilterLogs valid OR topics = %#v", filterTopics)
	}
	_, watchSub, err := contract.WatchLogs(nil, "Numbers", valid...)
	if err != nil {
		t.Fatalf("WatchLogs valid OR query: %v", err)
	}
	watchSub.Unsubscribe()
	watchTopics := filterer.watched[len(filterer.watched)-1].Topics
	if !reflect.DeepEqual(watchTopics, filterTopics) {
		t.Fatalf("WatchLogs topics = %#v, want FilterLogs topics %#v", watchTopics, filterTopics)
	}
}

func TestBoundContractCompositeEventDataMatrix(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(`[
		{"anonymous":false,"inputs":[
			{"indexed":true,"name":"sender","type":"address"},
			{"indexed":false,"name":"values","type":"uint512[]"},
			{"indexed":false,"name":"tags","type":"bytes64[2]"},
			{"indexed":false,"name":"paths","type":"address[][2]"},
			{"components":[{"name":"recipient","type":"address"},{"name":"amount","type":"uint512"},{"name":"notes","type":"string[]"}],"indexed":false,"internalType":"struct Logs.Record","name":"record","type":"tuple"},
			{"components":[{"name":"recipient","type":"address"},{"name":"amount","type":"uint512"},{"name":"notes","type":"string[]"}],"indexed":false,"internalType":"struct Logs.Record[]","name":"records","type":"tuple[]"},
			{"indexed":false,"name":"payload","type":"bytes"},
			{"indexed":false,"name":"note","type":"string"},
			{"indexed":false,"name":"enabled","type":"bool"},
			{"indexed":false,"name":"delta","type":"int512"},
			{"indexed":false,"name":"small","type":"bytes1"}
		],"name":"CompositeData","type":"event"}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	event := parsed.Events["CompositeData"]
	filterer := new(recordingFilterer)
	contract := bind.NewBoundContract(common.BytesToAddress([]byte{0x76}), parsed, nil, nil, filterer)
	sender := common.BytesToAddress(bytes.Repeat([]byte{0x6b}, common.AddressLength))
	recipient := common.BytesToAddress(bytes.Repeat([]byte{0xa4}, common.AddressLength))
	maximum := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1))
	minimum := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 511))
	values := []*big.Int{new(big.Int), new(big.Int).Set(maximum)}
	var tags [2][64]byte
	for i := range tags[0] {
		tags[0][i] = byte(i*7 + 1)
		tags[1][i] = byte(i*11 + 3)
	}
	paths := [2][]common.Address{{sender, recipient}, {}}
	type record struct {
		Recipient common.Address
		Amount    *big.Int
		Notes     []string
	}
	recordValue := record{Recipient: recipient, Amount: maximum, Notes: []string{"", strings.Repeat("n", 65), "\u754c\x00"}}
	records := []record{
		recordValue,
		{Recipient: sender, Amount: new(big.Int), Notes: []string{}},
	}
	payload := bytes.Repeat([]byte{0xc7}, 65)
	note := "composite event data across a VM64 word \u754c\x00"
	small := [1]byte{0xfe}
	data, err := event.Inputs.NonIndexed().Pack(values, tags, paths, recordValue, records, payload, note, true, minimum, small)
	if err != nil {
		t.Fatalf("pack composite event data: %v", err)
	}
	senderTopics, err := abi.MakeTopics([]any{sender})
	if err != nil {
		t.Fatal(err)
	}
	log := types.Log{
		Topics: []common.LogTopic{common.HashToLogTopic(event.ID), senderTopics[0][0]},
		Data:   data,
	}

	_, subscription, err := contract.FilterLogs(nil, "CompositeData", []any{sender})
	if err != nil {
		t.Fatalf("FilterLogs(CompositeData): %v", err)
	}
	subscription.Unsubscribe()
	query := filterer.filtered[len(filterer.filtered)-1].Topics
	if len(query) != 2 || query[0][0] != log.Topics[0] || query[1][0] != log.Topics[1] {
		t.Fatalf("composite-data filter query = %#v", query)
	}

	type decodedEvent struct {
		Sender  common.Address      `abi:"sender"`
		Values  []*big.Int          `abi:"values"`
		Tags    [2][64]byte         `abi:"tags"`
		Paths   [2][]common.Address `abi:"paths"`
		Record  record              `abi:"record"`
		Records []record            `abi:"records"`
		Payload []byte              `abi:"payload"`
		Note    string              `abi:"note"`
		Enabled bool                `abi:"enabled"`
		Delta   *big.Int            `abi:"delta"`
		Small   [1]byte             `abi:"small"`
	}
	var decoded decodedEvent
	if err := contract.UnpackLog(&decoded, "CompositeData", log); err != nil {
		t.Fatalf("UnpackLog(CompositeData): %v", err)
	}
	equalBigInts := func(have, want []*big.Int) bool {
		if len(have) != len(want) {
			return false
		}
		for i := range have {
			if have[i] == nil || want[i] == nil || have[i].Cmp(want[i]) != 0 {
				return false
			}
		}
		return true
	}
	equalRecord := func(have, want record) bool {
		return have.Recipient == want.Recipient && have.Amount != nil && want.Amount != nil && have.Amount.Cmp(want.Amount) == 0 && reflect.DeepEqual(have.Notes, want.Notes)
	}
	equalRecords := len(decoded.Records) == len(records)
	for i := 0; equalRecords && i < len(records); i++ {
		equalRecords = equalRecord(decoded.Records[i], records[i])
	}
	if decoded.Sender != sender || !equalBigInts(decoded.Values, values) || decoded.Tags != tags || !reflect.DeepEqual(decoded.Paths, paths) ||
		!equalRecord(decoded.Record, recordValue) || !equalRecords || !bytes.Equal(decoded.Payload, payload) || decoded.Note != note || !decoded.Enabled ||
		decoded.Delta == nil || decoded.Delta.Cmp(minimum) != 0 || decoded.Small != small {
		t.Fatalf("decoded composite event = %+v", decoded)
	}
	mapped := make(map[string]any)
	if err := contract.UnpackLogIntoMap(mapped, "CompositeData", log); err != nil {
		t.Fatalf("UnpackLogIntoMap(CompositeData): %v", err)
	}
	mappedValues, ok := mapped["values"].([]*big.Int)
	if !ok || mapped["sender"] != sender || !equalBigInts(mappedValues, values) || mapped["tags"] != tags ||
		!bytes.Equal(mapped["payload"].([]byte), payload) || mapped["note"] != note || mapped["enabled"] != true || mapped["small"] != small {
		t.Fatalf("mapped composite event = %#v", mapped)
	}
}

func TestBoundContractInterleavedDuplicateEventNames(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(`[
		{"anonymous":false,"inputs":[
			{"indexed":false,"name":"duplicate","type":"bytes64"},
			{"indexed":true,"name":"duplicate","type":"uint512"},
			{"indexed":false,"name":"","type":"uint64"},
			{"indexed":true,"name":"","type":"address"}
		],"name":"Interleaved","type":"event"}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	contract := bind.NewBoundContract(common.Address{}, parsed, nil, nil, nil)
	event := parsed.Events["Interleaved"]
	var data [64]byte
	var sender common.Address
	for i := range data {
		data[i] = byte(i*17 + 3)
		sender[i] = byte(i*19 + 5)
	}
	amount := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 512), big.NewInt(1))
	encoded, err := event.Inputs.NonIndexed().Pack(data, uint64(0xfedcba9876543210))
	if err != nil {
		t.Fatal(err)
	}
	indexed, err := abi.MakeTopics([]any{amount}, []any{sender})
	if err != nil {
		t.Fatal(err)
	}
	log := types.Log{
		Topics: []common.LogTopic{
			common.HashToLogTopic(event.ID),
			indexed[0][0],
			indexed[1][0],
		},
		Data: encoded,
	}
	type decodedEvent struct {
		Data   [64]byte       `abi:"duplicate"`
		Amount *big.Int       `abi:"duplicate"`
		Count  uint64         `abi:"arg2"`
		Sender common.Address `abi:"arg3"`
	}
	var decoded decodedEvent
	if err := contract.UnpackLog(&decoded, "Interleaved", log); err != nil {
		t.Fatalf("UnpackLog(interleaved duplicate names): %v", err)
	}
	if decoded.Data != data || decoded.Amount == nil || decoded.Amount.Cmp(amount) != 0 || decoded.Count != 0xfedcba9876543210 || decoded.Sender != sender {
		t.Fatalf("interleaved duplicate-name event = %+v", decoded)
	}
	mapped := map[string]any{"sentinel": true}
	if err := contract.UnpackLogIntoMap(mapped, "Interleaved", log); err == nil || !strings.Contains(err.Error(), "duplicate event argument name") {
		t.Fatalf("UnpackLogIntoMap duplicate-name error = %v", err)
	}
	if !reflect.DeepEqual(mapped, map[string]any{"sentinel": true}) {
		t.Fatalf("duplicate-name map was partially mutated: %#v", mapped)
	}
}

func TestBoundContractAllowsNormalizedUnnamedEventMapKey(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(`[
		{"anonymous":false,"inputs":[{"indexed":false,"name":"","type":"uint64"}],"name":"Unnamed","type":"event"}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	eventDef := parsed.Events["Unnamed"]
	data, err := eventDef.Inputs.NonIndexed().Pack(uint64(7))
	if err != nil {
		t.Fatal(err)
	}
	log := types.Log{Topics: []common.LogTopic{common.HashToLogTopic(eventDef.ID)}, Data: data}
	contract := bind.NewBoundContract(common.Address{}, parsed, nil, nil, nil)
	mapped := make(map[string]any)
	if err := contract.UnpackLogIntoMap(mapped, "Unnamed", log); err != nil {
		t.Fatalf("UnpackLogIntoMap unnamed input: %v", err)
	}
	if got, ok := mapped["arg0"].(uint64); !ok || got != 7 || len(mapped) != 1 {
		t.Fatalf("normalized unnamed event map = %#v, want arg0=7", mapped)
	}
}

func TestBoundContractUnpackNormalizedEventFields(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(`[
		{"anonymous":false,"inputs":[{"indexed":true,"name":"range","type":"uint512"},{"indexed":true,"name":"_msg","type":"uint512"},{"indexed":true,"name":"msg","type":"uint512"},{"indexed":false,"name":"type","type":"uint512"},{"indexed":false,"name":"_value","type":"uint512"}],"name":"Normalized","type":"event"}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	eventDef := parsed.Events["Normalized"]
	data, err := eventDef.Inputs.NonIndexed().Pack(big.NewInt(4), big.NewInt(5))
	if err != nil {
		t.Fatal(err)
	}
	topicSets, err := abi.MakeTopics(
		[]any{eventDef.ID},
		[]any{big.NewInt(1)},
		[]any{big.NewInt(2)},
		[]any{big.NewInt(3)},
	)
	if err != nil {
		t.Fatal(err)
	}
	log := types.Log{Data: data}
	for _, set := range topicSets {
		log.Topics = append(log.Topics, set[0])
	}
	var decoded struct {
		Arg0  *big.Int `abi:"range"`
		Msg   *big.Int `abi:"_msg"`
		Msg0  *big.Int `abi:"msg"`
		Arg3  *big.Int `abi:"type"`
		Value *big.Int `abi:"_value"`
	}
	contract := bind.NewBoundContract(common.Address{}, parsed, nil, nil, nil)
	if err := contract.UnpackLog(&decoded, "Normalized", log); err != nil {
		t.Fatalf("UnpackLog(normalized names) failed: %v", err)
	}
	for name, got := range map[string]*big.Int{
		"range":  decoded.Arg0,
		"_msg":   decoded.Msg,
		"msg":    decoded.Msg0,
		"type":   decoded.Arg3,
		"_value": decoded.Value,
	} {
		want := map[string]int64{"range": 1, "_msg": 2, "msg": 3, "type": 4, "_value": 5}[name]
		if got == nil || got.Int64() != want {
			t.Fatalf("decoded %s = %v, want %d", name, got, want)
		}
	}
}

func TestCall(t *testing.T) {
	contractAddr := common.MustParseAddress("Q00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	var method, methodWithArg = "something", "somethingArrrrg"
	tests := []struct {
		name, method string
		opts         *bind.CallOpts
		mc           bind.ContractCaller
		results      *[]any
		wantErr      bool
		wantErrExact error
	}{{
		name: "ok not pending",
		mc: &mockCaller{
			codeAtBytes: []byte{0},
		},
		method: method,
	}, {
		name: "ok pending",
		mc: &mockPendingCaller{
			pendingCodeAtBytes: []byte{0},
		},
		opts: &bind.CallOpts{
			Pending: true,
		},
		method: method,
	}, {
		name:    "pack error, no method",
		mc:      new(mockCaller),
		method:  "else",
		wantErr: true,
	}, {
		name: "interface error, pending but not a PendingContractCaller",
		mc:   new(mockCaller),
		opts: &bind.CallOpts{
			Pending: true,
		},
		method:       method,
		wantErrExact: bind.ErrNoPendingState,
	}, {
		name: "pending call canceled",
		mc: &mockPendingCaller{
			pendingCallContractErr: context.DeadlineExceeded,
		},
		opts: &bind.CallOpts{
			Pending: true,
		},
		method:       method,
		wantErrExact: context.DeadlineExceeded,
	}, {
		name: "pending code at error",
		mc: &mockPendingCaller{
			pendingCodeAtErr: errors.New(""),
		},
		opts: &bind.CallOpts{
			Pending: true,
		},
		method:  method,
		wantErr: true,
	}, {
		name: "no pending code at",
		mc:   new(mockPendingCaller),
		opts: &bind.CallOpts{
			Pending: true,
		},
		method:       method,
		wantErrExact: bind.ErrNoCode,
	}, {
		name: "call contract error",
		mc: &mockCaller{
			callContractErr: context.DeadlineExceeded,
		},
		method:       method,
		wantErrExact: context.DeadlineExceeded,
	}, {
		name: "code at error",
		mc: &mockCaller{
			codeAtErr: errors.New(""),
		},
		method:  method,
		wantErr: true,
	}, {
		name:         "no code at",
		mc:           new(mockCaller),
		method:       method,
		wantErrExact: bind.ErrNoCode,
	}, {
		name: "unpack error missing arg",
		mc: &mockCaller{
			codeAtBytes: []byte{0},
		},
		method:  methodWithArg,
		wantErr: true,
	}, {
		name: "interface unpack error",
		mc: &mockCaller{
			codeAtBytes: []byte{0},
		},
		method:  method,
		results: &[]any{0},
		wantErr: true,
	}}
	for _, test := range tests {
		bc := bind.NewBoundContract(contractAddr, abi.ABI{
			Methods: map[string]abi.Method{
				method: {
					Name:    method,
					Outputs: abi.Arguments{},
				},
				methodWithArg: {
					Name:    methodWithArg,
					Outputs: abi.Arguments{abi.Argument{}},
				},
			},
		}, test.mc, nil, nil)
		err := bc.Call(test.opts, test.results, test.method)
		if test.wantErr || test.wantErrExact != nil {
			if err == nil {
				t.Fatalf("%q expected error", test.name)
			}
			if test.wantErrExact != nil && !errors.Is(err, test.wantErrExact) {
				t.Fatalf("%q expected error %q but got %q", test.name, test.wantErrExact, err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%q unexpected error: %v", test.name, err)
		}
	}
}

// TestCrashers contains some strings which previously caused the abi codec to crash.
func TestCrashers(t *testing.T) {
	abi.JSON(strings.NewReader(`[{"inputs":[{"type":"tuple[]","components":[{"type":"bool","name":"_1"}]}]}]`))
	abi.JSON(strings.NewReader(`[{"inputs":[{"type":"tuple[]","components":[{"type":"bool","name":"&"}]}]}]`))
	abi.JSON(strings.NewReader(`[{"inputs":[{"type":"tuple[]","components":[{"type":"bool","name":"----"}]}]}]`))
	abi.JSON(strings.NewReader(`[{"inputs":[{"type":"tuple[]","components":[{"type":"bool","name":"foo.Bar"}]}]}]`))
}
