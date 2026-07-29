// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package clef

import (
	"testing"

	qrlaccounts "github.com/theQRL/go-qrl/accounts"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/rpc"
	"github.com/theQRL/go-qrl/signer/core/apitypes"
	"github.com/theQRL/go-qrl/testing/devnet/internal/network"
)

type testAccountAPI struct {
	account           common.Address
	dataSignature     hexutil.Bytes
	typedSignature    hexutil.Bytes
	signedTransaction signTransactionResult
}

func (api *testAccountAPI) List() []common.Address {
	return []common.Address{api.account}
}

func (api *testAccountAPI) SignData(string, string, string) hexutil.Bytes {
	return api.dataSignature
}

func (api *testAccountAPI) SignTypedData(string, apitypes.TypedData) hexutil.Bytes {
	return api.typedSignature
}

func (api *testAccountAPI) SignTransaction(apitypes.SendTxArgs) signTransactionResult {
	return api.signedTransaction
}

func TestSigningScenarios(t *testing.T) {
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

	server := rpc.NewServer()
	if err := server.RegisterName("account", &testAccountAPI{
		account:           account,
		dataSignature:     hexutil.Bytes(dataSignature),
		typedSignature:    hexutil.Bytes(typedSignature),
		signedTransaction: signTransactionResult{Raw: hexutil.Bytes(raw), Tx: transaction},
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(server.Stop)

	client := rpc.DialInProc(server)
	t.Cleanup(client.Close)

	t.Run("account listing", func(t *testing.T) {
		if err := verifyAccountListing(t.Context(), client, account); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("data signing", func(t *testing.T) {
		if err := verifyDataSigning(t.Context(), client, account, expectedWallet); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("typed-data signing", func(t *testing.T) {
		if err := verifyTypedDataSigning(t.Context(), client, account, expectedWallet); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("transaction signing", func(t *testing.T) {
		signed, err := signTransaction(t.Context(), client, request)
		if err != nil {
			t.Fatal(err)
		}
		if err := verifyTransaction(signed, request, account, expectedWallet); err != nil {
			t.Fatal(err)
		}
	})
}
