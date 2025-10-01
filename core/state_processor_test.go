// Copyright 2020 The go-ethereum Authors
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

package core

import (
	"math"
	"math/big"
	"testing"

	walletmldsa87 "github.com/theQRL/go-qrllib/wallet/ml_dsa_87"
	"github.com/theQRL/go-zond/common"
	"github.com/theQRL/go-zond/consensus"
	"github.com/theQRL/go-zond/consensus/beacon"
	"github.com/theQRL/go-zond/consensus/misc/eip1559"
	"github.com/theQRL/go-zond/core/rawdb"
	"github.com/theQRL/go-zond/core/types"
	"github.com/theQRL/go-zond/core/vm"
	"github.com/theQRL/go-zond/crypto/pqcrypto"
	"github.com/theQRL/go-zond/params"
	"github.com/theQRL/go-zond/trie"
	"golang.org/x/crypto/sha3"
)

// TestStateProcessorErrors tests the output from the 'core' errors
// as defined in core/error.go. These errors are generated when the
// blockchain imports bad blocks, meaning blocks which have valid headers but
// contain invalid transactions
func TestStateProcessorErrors(t *testing.T) {
	var (
		config = &params.ChainConfig{
			ChainID: big.NewInt(1),
		}
		signer  = types.LatestSigner(config)
		key1, _ = pqcrypto.HexToWallet("f29f58aff0b00de2844f7e20bd9eeaacc379150043beeb328335817512b29fbb7184da84a092f842b2a06d72a24a5d28")
		key2, _ = pqcrypto.HexToWallet("a7b1a3005d9e110009c48d45deb43f0a0e31846ed2c5aaefb6d4238040ad4c08794ffe65585c13eb6948c2faf6db90c2")
	)

	var mkDynamicTx = func(key *walletmldsa87.Wallet, nonce uint64, to common.Address, value *big.Int, gasLimit uint64, gasTipCap, gasFeeCap *big.Int) *types.Transaction {
		tx, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
			Nonce:     nonce,
			GasTipCap: gasTipCap,
			GasFeeCap: gasFeeCap,
			Gas:       gasLimit,
			To:        &to,
			Value:     value,
		}), signer, key)
		return tx
	}
	var mkDynamicCreationTx = func(nonce uint64, gasLimit uint64, gasTipCap, gasFeeCap *big.Int, data []byte) *types.Transaction {
		tx, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
			Nonce:     nonce,
			GasTipCap: gasTipCap,
			GasFeeCap: gasFeeCap,
			Gas:       gasLimit,
			Value:     big.NewInt(0),
			Data:      data,
		}), signer, key1)
		return tx
	}

	{ // Tests against a 'recent' chain definition
		var (
			address0, _ = common.NewAddressFromString("Q16B330c3ceF65A832a656DfDc7387E607593feD0")
			address1, _ = common.NewAddressFromString("Q69bd8f71A47ca646043AD210A2C41f2b73B85e15")
			db          = rawdb.NewMemoryDatabase()
			gspec       = &Genesis{
				Config: config,
				Alloc: GenesisAlloc{
					address0: GenesisAccount{
						Balance: big.NewInt(1000000000000000000), // 1 quanta
						Nonce:   0,
					},
					address1: GenesisAccount{
						Balance: big.NewInt(1000000000000000000), // 1 quanta
						Nonce:   math.MaxUint64,
					},
				},
			}
			blockchain, _  = NewBlockChain(db, nil, gspec, beacon.New(), vm.Config{}, nil)
			tooBigInitCode = [params.MaxInitCodeSize + 1]byte{}
		)

		defer blockchain.Stop()
		bigNumber := new(big.Int).SetBytes(common.MaxHash.Bytes())
		tooBigNumber := new(big.Int).Set(bigNumber)
		tooBigNumber.Add(tooBigNumber, common.Big1)
		for i, tt := range []struct {
			txs  []*types.Transaction
			want string
		}{

			{ // ErrNonceTooLow
				txs: []*types.Transaction{
					mkDynamicTx(key1, 0, common.Address{}, big.NewInt(0), params.TxGas, big.NewInt(0), big.NewInt(875000000)),
					mkDynamicTx(key1, 0, common.Address{}, big.NewInt(0), params.TxGas, big.NewInt(0), big.NewInt(875000000)),
				},
				want: "could not apply tx 1 [0xd1da9ab92c871016a1c362efa71402795c2a099c5e30edb06e7c65c640ff9d06]: nonce too low: address Q16B330c3ceF65A832a656DfDc7387E607593feD0, tx: 0 state: 1",
			},
			{ // ErrNonceTooHigh
				txs: []*types.Transaction{
					mkDynamicTx(key1, 100, common.Address{}, big.NewInt(0), params.TxGas, big.NewInt(0), big.NewInt(875000000)),
				},
				want: "could not apply tx 0 [0x0fba17cfab18390a73262231fe29c295b7a201a19119235ad5542657cb0e6f78]: nonce too high: address Q16B330c3ceF65A832a656DfDc7387E607593feD0, tx: 100 state: 0",
			},
			{ // ErrNonceMax
				txs: []*types.Transaction{
					mkDynamicTx(key2, math.MaxUint64, common.Address{}, big.NewInt(0), params.TxGas, big.NewInt(0), big.NewInt(875000000)),
				},
				want: "could not apply tx 0 [0xb3a2dac00db12cf1a0247629103b58e34e1505f08d35f4ac2d2cc4587e0959ba]: nonce has max value: address Q69bd8f71A47ca646043AD210A2C41f2b73B85e15, nonce: 18446744073709551615",
			},
			{ // ErrGasLimitReached
				txs: []*types.Transaction{
					mkDynamicTx(key1, 0, common.Address{}, big.NewInt(0), 21000000, big.NewInt(0), big.NewInt(875000000)),
				},
				want: "could not apply tx 0 [0x3e1148b890ed3bceed5ac19dfd1067cd6e563a2758ddde6b5c6e48133c3178e7]: gas limit reached",
			},
			{ // ErrInsufficientFundsForTransfer
				txs: []*types.Transaction{
					mkDynamicTx(key1, 0, common.Address{}, big.NewInt(1000000000000000000), params.TxGas, big.NewInt(0), big.NewInt(875000000)),
				},
				want: "could not apply tx 0 [0x7a03b598bb4653215502cdb007e9baf0b0cb59d9571d8cf1def0be2d71ae3c52]: insufficient funds for gas * price + value: address Q16B330c3ceF65A832a656DfDc7387E607593feD0 have 1000000000000000000 want 1000018375000000000",
			},
			{ // ErrInsufficientFunds
				txs: []*types.Transaction{
					mkDynamicTx(key1, 0, common.Address{}, big.NewInt(0), params.TxGas, big.NewInt(0), big.NewInt(900000000000000000)),
				},
				want: "could not apply tx 0 [0x190be0947227c5b6a3e95e601a51acc26a51808b444128e5ce808bd4740a24e3]: insufficient funds for gas * price + value: address Q16B330c3ceF65A832a656DfDc7387E607593feD0 have 1000000000000000000 want 18900000000000000000000",
			},
			// ErrGasUintOverflow
			// One missing 'core' error is ErrGasUintOverflow: "gas uint64 overflow",
			// In order to trigger that one, we'd have to allocate a _huge_ chunk of data, such that the
			// multiplication len(data) +gas_per_byte overflows uint64. Not testable at the moment
			{ // ErrIntrinsicGas
				txs: []*types.Transaction{
					mkDynamicTx(key1, 0, common.Address{}, big.NewInt(0), params.TxGas-1000, big.NewInt(0), big.NewInt(875000000)),
				},
				want: "could not apply tx 0 [0x43aaa3b6a81e6acfa99bd635e947a47ce75d1e45f7de2742d990a567ed2e311c]: intrinsic gas too low: have 20000, want 21000",
			},
			{ // ErrGasLimitReached
				txs: []*types.Transaction{
					mkDynamicTx(key1, 0, common.Address{}, big.NewInt(0), params.TxGas*1000, big.NewInt(0), big.NewInt(875000000)),
				},
				want: "could not apply tx 0 [0x3e1148b890ed3bceed5ac19dfd1067cd6e563a2758ddde6b5c6e48133c3178e7]: gas limit reached",
			},
			{ // ErrFeeCapTooLow
				txs: []*types.Transaction{
					mkDynamicTx(key1, 0, common.Address{}, big.NewInt(0), params.TxGas, big.NewInt(0), big.NewInt(0)),
				},
				want: "could not apply tx 0 [0x6eb29e67e3853f0b94593583554313659ff02da359d18fea2bed7dac14a1de8f]: max fee per gas less than block base fee: address Q16B330c3ceF65A832a656DfDc7387E607593feD0, maxFeePerGas: 0 baseFee: 875000000",
			},
			{ // ErrTipVeryHigh
				txs: []*types.Transaction{
					mkDynamicTx(key1, 0, common.Address{}, big.NewInt(0), params.TxGas, tooBigNumber, big.NewInt(1)),
				},
				want: "could not apply tx 0 [0x44e8ef9b956fc61bbd1cfa448629ffc2321f9cc9131abe4536275ecd0b52cc95]: max priority fee per gas higher than 2^256-1: address Q16B330c3ceF65A832a656DfDc7387E607593feD0, maxPriorityFeePerGas bit length: 257",
			},
			{ // ErrFeeCapVeryHigh
				txs: []*types.Transaction{
					mkDynamicTx(key1, 0, common.Address{}, big.NewInt(0), params.TxGas, big.NewInt(1), tooBigNumber),
				},
				want: "could not apply tx 0 [0xceae06bb6a823cd3d6065c3479f7c43f217c5004c1b6c79a6880d6006c7a9a25]: max fee per gas higher than 2^256-1: address Q16B330c3ceF65A832a656DfDc7387E607593feD0, maxFeePerGas bit length: 257",
			},
			{ // ErrTipAboveFeeCap
				txs: []*types.Transaction{
					mkDynamicTx(key1, 0, common.Address{}, big.NewInt(0), params.TxGas, big.NewInt(2), big.NewInt(1)),
				},
				want: "could not apply tx 0 [0x6a7f795c4fd270a8b6de527d0acf5276c86ca4305d2585e5db400288c0eb0003]: max priority fee per gas higher than max fee per gas: address Q16B330c3ceF65A832a656DfDc7387E607593feD0, maxPriorityFeePerGas: 2, maxFeePerGas: 1",
			},
			{ // ErrInsufficientFunds
				// Available balance:           1000000000000000000
				// Effective cost:                   18375000021000
				// FeeCap * gas:                1050000000000000000
				// This test is designed to have the effective cost be covered by the balance, but
				// the extended requirement on FeeCap*gas < balance to fail
				txs: []*types.Transaction{
					mkDynamicTx(key1, 0, common.Address{}, big.NewInt(0), params.TxGas, big.NewInt(1), big.NewInt(50000000000000)),
				},
				want: "could not apply tx 0 [0x895794ee2627dd76aaddf2a3dddd79c6416149718df42bd8b368b7a4804e37f0]: insufficient funds for gas * price + value: address Q16B330c3ceF65A832a656DfDc7387E607593feD0 have 1000000000000000000 want 1050000000000000000",
			},
			{ // Another ErrInsufficientFunds, this one to ensure that feecap/tip of max u256 is allowed
				txs: []*types.Transaction{
					mkDynamicTx(key1, 0, common.Address{}, big.NewInt(0), params.TxGas, bigNumber, bigNumber),
				},
				want: "could not apply tx 0 [0x35a82734c4f442111d1b2a96ff0a883b171a032bc226779d23de887b9061a976]: insufficient funds for gas * price + value: address Q16B330c3ceF65A832a656DfDc7387E607593feD0 have 1000000000000000000 want 2431633873983640103894990685182446064918669677978451844828609264166175722438635000",
			},
			{ // ErrMaxInitCodeSizeExceeded
				txs: []*types.Transaction{
					mkDynamicCreationTx(0, 500000, common.Big0, big.NewInt(params.InitialBaseFee), tooBigInitCode[:]),
				},
				want: "could not apply tx 0 [0xffcf101418251b253c0bb141234a4c3f625d59248442f6ec22b7d2253ff8fe20]: max initcode size exceeded: code size 49153 limit 49152",
			},
			{ // ErrIntrinsicGas: Not enough gas to cover init code
				txs: []*types.Transaction{
					mkDynamicCreationTx(0, 54299, common.Big0, big.NewInt(params.InitialBaseFee), make([]byte, 320)),
				},
				want: "could not apply tx 0 [0x4b5be50e26e6aab4c31f58c6ed1a4acfca9178da77e2785b56aae0049fb4ed3c]: intrinsic gas too low: have 54299, want 54300",
			},
		} {
			block := GenerateBadBlock(gspec.ToBlock(), beacon.New(), tt.txs, gspec.Config)
			_, err := blockchain.InsertChain(types.Blocks{block})
			if err == nil {
				t.Fatal("block imported without errors")
			}
			if have, want := err.Error(), tt.want; have != want {
				t.Errorf("test %d:\nhave \"%v\"\nwant \"%v\"\n", i, have, want)
			}
		}
	}

	// NOTE(rgeraldes24): test not valid for now
	/*
		// ErrTxTypeNotSupported, For this, we need an older chain
		{
			var (
				db    = rawdb.NewMemoryDatabase()
				gspec = &Genesis{
					Config: &params.ChainConfig{
						ChainID: big.NewInt(1),
					},
					Alloc: GenesisAlloc{
						common.HexToAddress("Q71562b71999873DB5b286dF957af199Ec94617F7"): GenesisAccount{
							Balance: big.NewInt(1000000000000000000), // 1 quanta
							Nonce:   0,
						},
					},
				}
				blockchain, _ = NewBlockChain(db, nil, gspec, beacon.NewFaker(), vm.Config{}, nil)
			)
			defer blockchain.Stop()
			for i, tt := range []struct {
				txs  []*types.Transaction
				want string
			}{
				{ // ErrTxTypeNotSupported
					txs: []*types.Transaction{
						mkDynamicTx(0, common.Address{}, params.TxGas-1000, big.NewInt(0), big.NewInt(0)),
					},
					want: "could not apply tx 0 [0x88626ac0d53cb65308f2416103c62bb1f18b805573d4f96a3640bbbfff13c14f]: transaction type not supported",
				},
			} {
				block := GenerateBadBlock(gspec.ToBlock(), beacon.NewFaker(), tt.txs, gspec.Config)
				_, err := blockchain.InsertChain(types.Blocks{block})
				if err == nil {
					t.Fatal("block imported without errors")
				}
				if have, want := err.Error(), tt.want; have != want {
					t.Errorf("test %d:\nhave \"%v\"\nwant \"%v\"\n", i, have, want)
				}
			}
		}
	*/

	// ErrSenderNoEOA, for this we need the sender to have contract code
	{
		var (
			address, _ = common.NewAddressFromString("Q16B330c3ceF65A832a656DfDc7387E607593feD0")
			db         = rawdb.NewMemoryDatabase()
			gspec      = &Genesis{
				Config: config,
				Alloc: GenesisAlloc{
					address: GenesisAccount{
						Balance: big.NewInt(1000000000000000000), // 1 quanta
						Nonce:   0,
						Code:    common.FromHex("0xB0B0FACE"),
					},
				},
			}
			blockchain, _ = NewBlockChain(db, nil, gspec, beacon.New(), vm.Config{}, nil)
		)
		defer blockchain.Stop()
		for i, tt := range []struct {
			txs  []*types.Transaction
			want string
		}{
			{ // ErrSenderNoEOA
				txs: []*types.Transaction{
					mkDynamicTx(key1, 0, common.Address{}, big.NewInt(0), params.TxGas-1000, big.NewInt(875000000), big.NewInt(0)),
				},
				want: "could not apply tx 0 [0xcad58a2acc436668d68cb4369c71baf342ee728526dbe39a597e1dfabb7f26a0]: sender not an eoa: address Q16B330c3ceF65A832a656DfDc7387E607593feD0, codehash: 0x9280914443471259d4570a8661015ae4a5b80186dbc619658fb494bebc3da3d1",
			},
		} {
			block := GenerateBadBlock(gspec.ToBlock(), beacon.New(), tt.txs, gspec.Config)
			_, err := blockchain.InsertChain(types.Blocks{block})
			if err == nil {
				t.Fatal("block imported without errors")
			}
			if have, want := err.Error(), tt.want; have != want {
				t.Errorf("test %d:\nhave \"%v\"\nwant \"%v\"\n", i, have, want)
			}
		}
	}
}

// GenerateBadBlock constructs a "block" which contains the transactions. The transactions are not expected to be
// valid, and no proper post-state can be made. But from the perspective of the blockchain, the block is sufficiently
// valid to be considered for import:
// - valid pow (fake), ancestry, difficulty, gaslimit etc
func GenerateBadBlock(parent *types.Block, engine consensus.Engine, txs types.Transactions, config *params.ChainConfig) *types.Block {
	header := &types.Header{
		ParentHash: parent.Hash(),
		Coinbase:   parent.Coinbase(),
		GasLimit:   parent.GasLimit(),
		Number:     new(big.Int).Add(parent.Number(), common.Big1),
		Time:       parent.Time() + 10,
	}
	header.BaseFee = eip1559.CalcBaseFee(config, parent.Header())
	header.WithdrawalsHash = &types.EmptyWithdrawalsHash
	var receipts []*types.Receipt
	// The post-state result doesn't need to be correct (this is a bad block), but we do need something there
	// Preferably something unique. So let's use a combo of blocknum + txhash
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(header.Number.Bytes())
	var cumulativeGas uint64
	for _, tx := range txs {
		txh := tx.Hash()
		hasher.Write(txh[:])
		receipt := &types.Receipt{
			Type:              types.DynamicFeeTxType,
			PostState:         common.CopyBytes(nil),
			CumulativeGasUsed: cumulativeGas + tx.Gas(),
			Status:            types.ReceiptStatusSuccessful,
		}
		receipt.TxHash = tx.Hash()
		receipt.GasUsed = tx.Gas()
		receipts = append(receipts, receipt)
		cumulativeGas += tx.Gas()
	}
	header.Root = common.BytesToHash(hasher.Sum(nil))

	// Assemble and return the final block for sealing
	body := &types.Body{Transactions: txs, Withdrawals: []*types.Withdrawal{}}
	return types.NewBlock(header, body, receipts, trie.NewStackTrie(nil))
}
