// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package clef

import (
	"math/big"
	"slices"
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
	requestedTyped    apitypes.TypedData
	requestedTx       apitypes.SendTxArgs
}

func (api *testAccountAPI) List() []common.Address {
	return []common.Address{api.account}
}

func (api *testAccountAPI) SignData(string, string, string) hexutil.Bytes {
	return api.dataSignature
}

func (api *testAccountAPI) SignTypedData(_ string, typed apitypes.TypedData) hexutil.Bytes {
	api.requestedTyped = typed
	return api.typedSignature
}

func (api *testAccountAPI) SignTransaction(tx apitypes.SendTxArgs) signTransactionResult {
	api.requestedTx = tx
	return api.signedTransaction
}

func TestClefServerArgsUseChainID(t *testing.T) {
	chainID := big.NewInt(424_242)
	args := clefServerArgs(t.TempDir(), 12345, chainID)
	index := slices.Index(args, "--chainid")
	if index == -1 || index+1 == len(args) {
		t.Fatal("Clef arguments have no --chainid value")
	}
	if got := args[index+1]; got != chainID.String() {
		t.Fatalf("Clef --chainid = %s, want %s", got, chainID)
	}
}

func TestSigningScenarios(t *testing.T) {
	expectedWallet, err := network.UnsafeDevelopmentWallet()
	if err != nil {
		t.Fatal(err)
	}
	account := common.Address(expectedWallet.GetAddress())
	chainID := big.NewInt(424_242)

	dataSignature, err := expectedWallet.Sign(qrlaccounts.TextHash([]byte(expectedText)))
	if err != nil {
		t.Fatal(err)
	}

	typedData := expectedTypedData(account, chainID)
	typedDigest, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		t.Fatal(err)
	}
	typedSignature, err := expectedWallet.Sign(typedDigest)
	if err != nil {
		t.Fatal(err)
	}

	request := expectedTransaction(account, chainID)
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

	api := &testAccountAPI{
		account:           account,
		dataSignature:     hexutil.Bytes(dataSignature),
		typedSignature:    hexutil.Bytes(typedSignature),
		signedTransaction: signTransactionResult{Raw: hexutil.Bytes(raw), Tx: transaction},
	}
	server := rpc.NewServer()
	if err := server.RegisterName("account", api); err != nil {
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
		if err := verifyTypedDataSigning(
			t.Context(),
			client,
			account,
			chainID,
			expectedWallet,
		); err != nil {
			t.Fatal(err)
		}
		if api.requestedTyped.Domain.ChainId == nil {
			t.Fatal("typed-data request has no chain ID")
		}
		if got := (*big.Int)(api.requestedTyped.Domain.ChainId); got.Cmp(chainID) != 0 {
			t.Fatalf("typed-data chain ID = %s, want %s", got, chainID)
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
		if api.requestedTx.ChainID == nil {
			t.Fatal("transaction request has no chain ID")
		}
		if got := (*big.Int)(api.requestedTx.ChainID); got.Cmp(chainID) != 0 {
			t.Fatalf("transaction chain ID = %s, want %s", got, chainID)
		}
	})
}
