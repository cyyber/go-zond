// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package clef

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	qrlaccounts "github.com/theQRL/go-qrl/accounts"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
	"github.com/theQRL/go-qrl/signer/core/apitypes"
)

const testSeed = "010000f29f58aff0b00de2844f7e20bd9eeaacc379150043beeb328335817512b29fbb7184da84a092f842b2a06d72a24a5d28"

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestExercise(t *testing.T) {
	expectedWallet, err := wallet.RestoreFromSeedHex(testSeed)
	if err != nil {
		t.Fatal(err)
	}
	account := common.Address(expectedWallet.GetAddress())

	dataSignature, err := expectedWallet.Sign(qrlaccounts.TextHash([]byte(expectedText)))
	if err != nil {
		t.Fatal(err)
	}

	typedData := expectedTypedData(account)
	typedDigest, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		t.Fatal(err)
	}
	typedSignature, err := expectedWallet.Sign(typedDigest)
	if err != nil {
		t.Fatal(err)
	}

	request, err := expectedTransaction(account)
	if err != nil {
		t.Fatal(err)
	}
	unsigned := request.ToTransaction()
	transaction, err := types.SignTx(
		unsigned,
		types.LatestSignerForChainID(unsigned.ChainId()),
		expectedWallet,
	)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := transaction.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	responses := map[string]any{
		"account_version":         "1.0.0",
		"account_list":            []common.Address{account},
		"account_signData":        hexutil.Bytes(dataSignature),
		"account_signTypedData":   hexutil.Bytes(typedSignature),
		"account_signTransaction": signTransactionResult{Raw: hexutil.Bytes(raw), Tx: transaction},
	}
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		defer request.Body.Close()
		var call struct {
			Method string `json:"method"`
			ID     int    `json:"id"`
		}
		if err := json.NewDecoder(request.Body).Decode(&call); err != nil {
			return nil, err
		}
		result, ok := responses[call.Method]
		if !ok {
			t.Fatalf("unexpected method %q", call.Method)
		}
		var body bytes.Buffer
		if err := json.NewEncoder(&body).Encode(map[string]any{
			"jsonrpc": "2.0",
			"result":  result,
			"id":      call.ID,
		}); err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body.Bytes())),
			Header:     make(http.Header),
		}, nil
	})}

	result, err := exercise(
		context.Background(),
		client,
		"http://clef.test",
		&clefProcess{done: make(chan struct{})},
		account,
		expectedWallet,
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Account != account || result.Version != "1.0.0" {
		t.Fatalf("unexpected result: %+v", result)
	}
}
