// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package clef

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/crypto/pqcrypto"
	"github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
	"github.com/theQRL/go-qrl/signer/core/apitypes"
)

func verifySignature(label string, signature, digest []byte, expectedWallet wallet.Wallet) error {
	if len(signature) != pqcrypto.MLDSA87SignatureLength {
		return fmt.Errorf(
			"%s signature width: got %d, want %d",
			label,
			len(signature),
			pqcrypto.MLDSA87SignatureLength,
		)
	}
	ok, err := pqcrypto.MLDSA87VerifySignature(
		signature,
		digest,
		expectedWallet.GetPK(),
		expectedWallet.GetDescriptor(),
	)
	if err != nil {
		return fmt.Errorf("verify %s signature: %w", label, err)
	}
	if !ok {
		return fmt.Errorf("verify %s signature: verification failed", label)
	}
	return nil
}

func verifyTransaction(
	result signTransactionResult,
	request apitypes.SendTxArgs,
	account common.Address,
	expectedWallet wallet.Wallet,
) error {
	if len(result.Raw) == 0 || result.Tx == nil {
		return errors.New("account_signTransaction must return raw and tx")
	}
	var decoded types.Transaction
	if err := decoded.UnmarshalBinary(result.Raw); err != nil {
		return fmt.Errorf("decode signed transaction: %w", err)
	}
	jsonTransaction, err := result.Tx.MarshalBinary()
	if err != nil {
		return fmt.Errorf("encode signed transaction result: %w", err)
	}
	if !bytes.Equal(jsonTransaction, result.Raw) {
		return errors.New("account_signTransaction raw and tx results differ")
	}

	if decoded.Type() != types.DynamicFeeTxType {
		return fmt.Errorf("signed transaction type: got %d, want %d", decoded.Type(), types.DynamicFeeTxType)
	}
	if request.ChainID == nil || decoded.ChainId().Cmp((*big.Int)(request.ChainID)) != 0 {
		return fmt.Errorf("signed transaction chain ID: got %s, want %v", decoded.ChainId(), request.ChainID)
	}
	if decoded.Nonce() != uint64(request.Nonce) || decoded.Gas() != uint64(request.Gas) {
		return errors.New("signed transaction nonce or gas does not match the request")
	}
	if request.MaxPriorityFeePerGas == nil ||
		decoded.GasTipCap().Cmp((*big.Int)(request.MaxPriorityFeePerGas)) != 0 {
		return errors.New("signed transaction priority fee does not match the request")
	}
	if request.MaxFeePerGas == nil ||
		decoded.GasFeeCap().Cmp((*big.Int)(request.MaxFeePerGas)) != 0 {
		return errors.New("signed transaction fee cap does not match the request")
	}
	if decoded.Value().Cmp((*big.Int)(&request.Value)) != 0 {
		return errors.New("signed transaction value does not match the request")
	}
	if request.Input == nil || !bytes.Equal(decoded.Data(), *request.Input) {
		return errors.New("signed transaction input does not match the request")
	}
	if request.To == nil || decoded.To() == nil || *decoded.To() != request.To.Address() {
		return errors.New("signed transaction recipient does not match the request")
	}
	if len(decoded.AccessList()) != 0 {
		return errors.New("signed transaction access list is not empty")
	}

	if len(decoded.RawSignatureValue()) != pqcrypto.MLDSA87SignatureLength {
		return fmt.Errorf(
			"signed transaction signature width: got %d, want %d",
			len(decoded.RawSignatureValue()),
			pqcrypto.MLDSA87SignatureLength,
		)
	}
	if len(decoded.RawPublicKeyValue()) != pqcrypto.MLDSA87PublicKeyLength {
		return fmt.Errorf(
			"signed transaction public-key width: got %d, want %d",
			len(decoded.RawPublicKeyValue()),
			pqcrypto.MLDSA87PublicKeyLength,
		)
	}
	if !bytes.Equal(decoded.RawPublicKeyValue(), expectedWallet.GetPK()) ||
		!bytes.Equal(decoded.Descriptor(), expectedWallet.GetDescriptor().ToBytes()) {
		return errors.New("signed transaction signer metadata does not match the imported seed")
	}
	signer := types.LatestSignerForChainID(decoded.ChainId())
	sender, err := types.Sender(signer, &decoded)
	if err != nil {
		return fmt.Errorf("recover signed transaction sender: %w", err)
	}
	if sender != account {
		return fmt.Errorf("signed transaction sender: got %s, want %s", sender.Hex(), account.Hex())
	}
	return nil
}
