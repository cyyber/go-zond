// Copyright 2016 The go-ethereum Authors
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

package types

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/crypto/pqcrypto"
	"github.com/theQRL/go-qrl/crypto/pqcrypto/wallet"
)

func TestSignerChainID(t *testing.T) {
	wallet, _ := wallet.Generate(wallet.ML_DSA_87)
	addr := common.Address(wallet.GetAddress())

	signer := NewZondSigner(big.NewInt(18))
	tx, err := SignTx(NewTx(&DynamicFeeTx{Nonce: 0, To: &addr, Value: new(big.Int), Gas: 0, GasFeeCap: new(big.Int), Data: nil}), signer, wallet)
	if err != nil {
		t.Fatal(err)
	}

	if tx.ChainId().Cmp(signer.ChainID()) != 0 {
		t.Error("expected chainId to be", signer.ChainID(), "got", tx.ChainId())
	}
}

func TestZondSigner_Sender(t *testing.T) {
	mkTx := func(chainID *big.Int) *Transaction {
		return NewTx(&DynamicFeeTx{
			ChainID:    new(big.Int).Set(chainID),
			Nonce:      0,
			GasTipCap:  common.Big1,
			GasFeeCap:  common.Big1,
			Gas:        21000,
			To:         nil,
			Value:      common.Big0,
			Data:       nil,
			AccessList: nil,
		})
	}
	extraParams := []byte{}

	signTx := func(t *testing.T, signer Signer, tx *Transaction, wallet wallet.Wallet) (*Transaction, []byte, []byte, []byte) {
		t.Helper()

		desc := wallet.GetDescriptor().ToBytes()
		h := signer.Hash(tx, desc, extraParams)
		sigArr, err := wallet.Sign(h.Bytes())
		if err != nil {
			t.Fatalf("sign: %v", err)
		}

		pkArr := wallet.GetPK()

		signed, err := tx.WithAuthValues(signer, sigArr[:], pkArr[:], desc, extraParams)
		if err != nil {
			t.Fatalf("WithAuthValues: %v", err)
		}

		return signed, sigArr[:], pkArr[:], desc
	}

	t.Run("ok/recovers-sender", func(t *testing.T) {
		t.Parallel()

		wallet, err := wallet.Generate(wallet.ML_DSA_87)
		if err != nil {
			t.Fatalf("wallet: %v", err)
		}
		chainID := big.NewInt(31337)
		signer := NewZondSigner(chainID)

		tx := mkTx(chainID)
		signed, _, _, _ := signTx(t, signer, tx, wallet)

		got, err := signer.Sender(signed)
		if err != nil {
			t.Fatalf("Sender error: %v", err)
		}
		if got != wallet.GetAddress() {
			t.Fatalf("sender mismatch: got %x want %x", got.Bytes(), wallet.GetAddress())
		}
	})
	t.Run("error/invalid-chain-id", func(t *testing.T) {
		t.Parallel()

		wallet, err := wallet.Generate(wallet.ML_DSA_87)
		if err != nil {
			t.Fatalf("wallet: %v", err)
		}
		s1 := NewZondSigner(common.Big1)
		tx := mkTx(common.Big1)
		signed, _, _, _ := signTx(t, s1, tx, wallet)

		s2 := NewZondSigner(common.Big2)
		_, err = s2.Sender(signed)
		if !errors.Is(err, ErrInvalidChainId) && err == nil {
			t.Fatalf("expected chain id error; got %v", err)
		}
	})

	t.Run("error/descriptor-mismatch", func(t *testing.T) {
		t.Parallel()

		wallet, err := wallet.Generate(wallet.ML_DSA_87)
		if err != nil {
			t.Fatalf("wallet: %v", err)
		}
		signer := NewZondSigner(big.NewInt(7))
		tx := mkTx(big.NewInt(7))
		signed, sig, pk, desc := signTx(t, signer, tx, wallet)

		// Flip one bit in the descriptor and re-wrap.
		desc[len(desc)-1] ^= 0x01
		tampered, err := signed.WithAuthValues(signer, sig, pk, desc, extraParams)
		if err != nil {
			t.Fatalf("re-wrap with bad descriptor: %v", err)
		}
		_, err = signer.Sender(tampered)
		if !errors.Is(err, pqcrypto.ErrBadSignature) && err == nil {
			t.Fatalf("expected bad signature error; got %v", err)
		}
	})
	t.Run("error/mutated-signature", func(t *testing.T) {
		t.Parallel()

		wallet, err := wallet.Generate(wallet.ML_DSA_87)
		if err != nil {
			t.Fatalf("wallet: %v", err)
		}
		signer := NewZondSigner(big.NewInt(7))
		tx := mkTx(big.NewInt(7))
		signed, sig, pk, desc := signTx(t, signer, tx, wallet)

		// Tweak the signature bytes and re-wrap.
		sig[len(sig)-1] ^= 0x80
		tampered, err := signed.WithAuthValues(signer, sig, pk, desc, extraParams)
		if err != nil {
			t.Fatalf("re-wrap with bad signature: %v", err)
		}
		_, err = signer.Sender(tampered)
		if !errors.Is(err, pqcrypto.ErrBadSignature) && err == nil {
			t.Fatalf("expected bad signature error; got %v", err)
		}
	})
	t.Run("error/non-empty-extra-params", func(t *testing.T) {
		t.Parallel()

		wallet, err := wallet.Generate(wallet.ML_DSA_87)
		if err != nil {
			t.Fatalf("wallet: %v", err)
		}
		signer := NewZondSigner(big.NewInt(7))
		tx := mkTx(big.NewInt(7))
		signed, sig, pk, desc := signTx(t, signer, tx, wallet)

		tampered, err := signed.WithAuthValues(signer, sig, pk, desc, []byte{0x01})
		if err != nil {
			t.Fatalf("re-wrap with extra params: %v", err)
		}
		_, err = signer.Sender(tampered)
		if err == nil {
			t.Fatal("expected non-empty extraParams error, got nil")
		}
		if got := err.Error(); got != "non-empty extraParams not supported" {
			t.Fatalf("unexpected error: got %q", got)
		}
	})
	t.Run("error/rejects-malformed-auth-lengths", func(t *testing.T) {
		t.Parallel()

		wallet, err := wallet.Generate(wallet.ML_DSA_87)
		if err != nil {
			t.Fatalf("wallet: %v", err)
		}
		signer := NewZondSigner(big.NewInt(7))
		tx := mkTx(big.NewInt(7))
		_, sig, pk, desc := signTx(t, signer, tx, wallet)

		tests := []struct {
			name string
			sig  []byte
			pk   []byte
			desc []byte
			want string
		}{
			{
				name: "signature",
				sig:  sig[:len(sig)-1],
				pk:   pk,
				desc: desc,
				want: "wrong size for ml-dsa-87 signature",
			},
			{
				name: "public-key",
				sig:  sig,
				pk:   pk[:len(pk)-1],
				desc: desc,
				want: "wrong size for ml-dsa-87 publickey",
			},
			{
				name: "descriptor",
				sig:  sig,
				pk:   pk,
				desc: desc[:len(desc)-1],
				want: "wrong size for descriptor",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				_, err := tx.WithAuthValues(signer, tt.sig, tt.pk, tt.desc, extraParams)
				if err == nil {
					t.Fatal("expected malformed auth error, got nil")
				}
				if got := err.Error(); !strings.HasPrefix(got, tt.want) {
					t.Fatalf("unexpected error: got %q want prefix %q", got, tt.want)
				}
			})
		}
	})
}
