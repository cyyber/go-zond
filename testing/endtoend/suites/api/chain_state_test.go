// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"slices"

	qrl "github.com/theQRL/go-qrl"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/consensus/misc/eip1559"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/crypto"
	"github.com/theQRL/go-qrl/params"
	"github.com/theQRL/go-qrl/qrlclient/gqrlclient"
	"github.com/theQRL/go-qrl/qrldb/memorydb"
	"github.com/theQRL/go-qrl/rlp"
	"github.com/theQRL/go-qrl/rpc"
	"github.com/theQRL/go-qrl/trie"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func (suite *liveSuite) assertChainState(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()
	fixture := suite.fixture
	blockNumber := fixture.receipt.BlockNumber
	blockSelector := rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(blockNumber.Int64()))

	chainID, err := suite.client.ChainID(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(chainID).To(gomega.Equal(suite.chainID))

	headNumber, err := suite.client.BlockNumber(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(headNumber).To(gomega.BeNumerically(">=", fixture.block.NumberU64()))

	headerByNumber, err := suite.client.HeaderByNumber(ctx, blockNumber)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	headerByHash, err := suite.client.HeaderByHash(ctx, fixture.block.Hash())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(headerByNumber.Hash()).To(gomega.Equal(fixture.block.Hash()))
	gomega.Expect(headerByHash.Hash()).To(gomega.Equal(fixture.block.Hash()))

	blockByNumber, err := suite.client.BlockByNumber(ctx, blockNumber)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	blockByHash, err := suite.client.BlockByHash(ctx, fixture.block.Hash())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(blockByNumber.Hash()).To(gomega.Equal(fixture.block.Hash()))
	gomega.Expect(blockByHash.Hash()).To(gomega.Equal(fixture.block.Hash()))

	var header map[string]json.RawMessage
	gomega.Expect(raw.CallContext(
		ctx,
		&header,
		"qrl_getHeaderByNumber",
		rpc.BlockNumber(blockNumber.Int64()),
	)).To(gomega.Succeed())
	gomega.Expect(header).To(gomega.HaveKey("hash"))
	gomega.Expect(raw.CallContext(ctx, &header, "qrl_getHeaderByHash", fixture.block.Hash())).To(gomega.Succeed())

	balance, err := suite.client.BalanceAt(ctx, suite.from, blockNumber)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(balance.Sign()).To(gomega.BeNumerically(">", 0))

	nonce, err := suite.client.NonceAt(ctx, suite.from, blockNumber)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(nonce).To(gomega.BeNumerically(">", 0))

	code, err := suite.client.CodeAt(ctx, fixture.address, blockNumber)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(code).NotTo(gomega.BeEmpty())

	storage, err := suite.client.StorageAt(ctx, fixture.address, common.Hash{}, blockNumber)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(storage).To(gomega.Equal(fixture.value[:]))

	call := qrl.CallMsg{From: suite.from, To: &fixture.address}
	output, err := suite.client.CallContract(ctx, call, blockNumber)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(output).To(gomega.Equal(fixture.value[:]))
	output, err = suite.client.CallContractAtHash(ctx, call, fixture.block.Hash())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(output).To(gomega.Equal(fixture.value[:]))
	output, err = suite.client.PendingCallContract(ctx, call)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(output).To(gomega.Equal(fixture.value[:]))

	gas, err := suite.client.EstimateGas(ctx, call)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(gas).To(gomega.BeNumerically(">", 0))

	gasPrice, err := suite.client.SuggestGasPrice(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(gasPrice.Sign()).To(gomega.BeNumerically(">", 0))
	tip, err := suite.client.SuggestGasTipCap(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(tip.Sign()).To(gomega.BeNumerically(">=", 0))

	receiptsByNumber, err := suite.client.BlockReceipts(ctx, blockSelector)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(receiptsByNumber).To(gomega.HaveLen(len(fixture.block.Transactions())))
	receiptIndex := int(fixture.receipt.TransactionIndex)
	gomega.Expect(receiptsByNumber[receiptIndex].TxHash).To(gomega.Equal(fixture.tx.Hash()))
	receiptsByHash, err := suite.client.BlockReceipts(
		ctx,
		rpc.BlockNumberOrHashWithHash(fixture.block.Hash(), true),
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(receiptsByHash).To(gomega.HaveLen(len(receiptsByNumber)))
	gomega.Expect(receiptsByHash[receiptIndex].TxHash).To(gomega.Equal(fixture.tx.Hash()))

	history, err := suite.client.FeeHistory(ctx, 1, blockNumber, []float64{50})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(history.OldestBlock).To(gomega.Equal(blockNumber))
	gomega.Expect(history.BaseFee).To(gomega.HaveLen(2))
	gomega.Expect(history.BaseFee[0]).To(gomega.Equal(fixture.block.BaseFee()))
	gomega.Expect(history.BaseFee[1]).To(gomega.Equal(
		eip1559.CalcBaseFee(params.AllBeaconProtocolChanges, fixture.block.Header()),
	))
	gomega.Expect(history.GasUsedRatio).To(gomega.Equal([]float64{
		float64(fixture.block.GasUsed()) / float64(fixture.block.GasLimit()),
	}))
	gomega.Expect(history.Reward).To(gomega.HaveLen(1))
	gomega.Expect(history.Reward[0]).To(gomega.HaveLen(1))
	gomega.Expect(history.Reward[0][0]).To(gomega.Equal(
		feeHistoryReward(fixture.block, receiptsByNumber, 50),
	))

	syncProgress, err := suite.client.SyncProgress(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(syncProgress).To(gomega.BeNil())

	proofClient := gqrlclient.New(raw)
	proof, err := proofClient.GetProof(
		ctx,
		fixture.address,
		[]string{common.Hash{}.Hex()},
		blockNumber,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(proof.Address).To(gomega.Equal(fixture.address))
	gomega.Expect(proof.StorageProof).To(gomega.HaveLen(1))
	gomega.Expect(proof.StorageProof[0].Value).To(
		gomega.Equal(new(big.Int).SetBytes(fixture.value[:])),
	)
	accountProof, err := proofDatabase(proof.AccountProof)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	accountData, err := trie.VerifyProof(
		fixture.block.Root(),
		crypto.Keccak256(fixture.address.Bytes()),
		accountProof,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	var account types.StateAccount
	gomega.Expect(rlp.DecodeBytes(accountData, &account)).To(gomega.Succeed())
	gomega.Expect(account.Nonce).To(gomega.Equal(proof.Nonce))
	gomega.Expect(account.Balance.Cmp(proof.Balance)).To(gomega.BeZero())
	gomega.Expect(account.Root).To(gomega.Equal(proof.StorageHash))
	gomega.Expect(common.BytesToHash(account.CodeHash)).To(gomega.Equal(proof.CodeHash))

	storageProof, err := proofDatabase(proof.StorageProof[0].Proof)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	storageData, err := trie.VerifyProof(
		proof.StorageHash,
		crypto.Keccak256((common.Hash{}).Bytes()),
		storageProof,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	wantStorageData, err := rlp.EncodeToBytes(common.TrimLeftZeroes(fixture.value[:]))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(storageData).To(gomega.Equal(wantStorageData))

	accessList, accessGas, accessError, err := proofClient.CreateAccessList(ctx, call)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(accessList).NotTo(gomega.BeNil())
	gomega.Expect(*accessList).To(gomega.Equal(types.AccessList{{
		Address:     fixture.address,
		StorageKeys: []common.Hash{{}},
	}}))
	gomega.Expect(accessGas).To(gomega.BeNumerically(">", 0))
	gomega.Expect(accessError).To(gomega.BeEmpty())

}

func feeHistoryReward(block *types.Block, receipts types.Receipts, percentile float64) *big.Int {
	ginkgo.GinkgoHelper()

	type gasAndReward struct {
		gasUsed uint64
		reward  *big.Int
	}
	rewards := make([]gasAndReward, len(block.Transactions()))
	for index, transaction := range block.Transactions() {
		reward, err := transaction.EffectiveGasTip(block.BaseFee())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		rewards[index] = gasAndReward{receipts[index].GasUsed, reward}
	}
	slices.SortStableFunc(rewards, func(left, right gasAndReward) int {
		return left.reward.Cmp(right.reward)
	})

	index := 0
	gasUsed := rewards[0].gasUsed
	threshold := uint64(float64(block.GasUsed()) * percentile / 100)
	for gasUsed < threshold && index < len(rewards)-1 {
		index++
		gasUsed += rewards[index].gasUsed
	}
	return rewards[index].reward
}

func proofDatabase(nodes []string) (*memorydb.Database, error) {
	database := memorydb.New()
	for _, encoded := range nodes {
		node, err := hexutil.Decode(encoded)
		if err != nil {
			return nil, fmt.Errorf("decode proof node: %w", err)
		}
		if err := database.Put(crypto.Keccak256(node), node); err != nil {
			return nil, fmt.Errorf("store proof node: %w", err)
		}
	}
	return database, nil
}
