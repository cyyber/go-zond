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
	"github.com/theQRL/go-qrl/rpc"
	"github.com/theQRL/go-qrl/signer/core/apitypes"
	"github.com/theQRL/go-qrl/testing/devnet/internal/network"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestExercise(t *testing.T) {
	expectedWallet, err := network.UnsafeDevelopmentWallet()
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

	request := expectedTransaction(account)
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
		"account_list":            []common.Address{account},
		"account_signData":        hexutil.Bytes(dataSignature),
		"account_signTypedData":   hexutil.Bytes(typedSignature),
		"account_signTransaction": signTransactionResult{Raw: hexutil.Bytes(raw), Tx: transaction},
	}
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
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
	client, err := rpc.DialOptions(
		context.Background(),
		"http://clef.test",
		rpc.WithHTTPClient(httpClient),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	err = exercise(
		context.Background(),
		client,
		account,
		expectedWallet,
	)
	if err != nil {
		t.Fatal(err)
	}
}
