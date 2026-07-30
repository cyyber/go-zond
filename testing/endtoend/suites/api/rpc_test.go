// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"runtime"
	runtimedebug "runtime/debug"

	qrl "github.com/theQRL/go-qrl"
	qrlaccounts "github.com/theQRL/go-qrl/accounts"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/crypto/pqcrypto"
	"github.com/theQRL/go-qrl/p2p"
	"github.com/theQRL/go-qrl/qrlclient/gqrlclient"
	"github.com/theQRL/go-qrl/rpc"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func (suite *liveSuite) assertNodeMetadata(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()

	var modules map[string]string
	gomega.Expect(raw.CallContext(ctx, &modules, "rpc_modules")).To(gomega.Succeed())
	for _, namespace := range []string{"admin", "debug", "net", "qrl", "txpool", "web3"} {
		gomega.Expect(modules).To(gomega.HaveKey(namespace))
	}

	var clientVersion string
	gomega.Expect(raw.CallContext(ctx, &clientVersion, "web3_clientVersion")).To(gomega.Succeed())
	gomega.Expect(clientVersion).NotTo(gomega.BeEmpty())

	var digest hexutil.Bytes
	gomega.Expect(raw.CallContext(ctx, &digest, "web3_sha3", hexutil.Bytes("api"))).To(gomega.Succeed())
	gomega.Expect(digest).To(gomega.HaveLen(common.HashLength))

	var networkVersion string
	gomega.Expect(raw.CallContext(ctx, &networkVersion, "net_version")).To(gomega.Succeed())
	gomega.Expect(networkVersion).To(gomega.Equal(suite.chainID.String()))

	var listening bool
	gomega.Expect(raw.CallContext(ctx, &listening, "net_listening")).To(gomega.Succeed())
	gomega.Expect(listening).To(gomega.BeTrue())

	var peerCount hexutil.Uint
	gomega.Expect(raw.CallContext(ctx, &peerCount, "net_peerCount")).To(gomega.Succeed())

	var nodeInfo p2p.NodeInfo
	gomega.Expect(raw.CallContext(ctx, &nodeInfo, "admin_nodeInfo")).To(gomega.Succeed())
	gomega.Expect(nodeInfo.ID).NotTo(gomega.BeEmpty())

	var peers []*p2p.PeerInfo
	gomega.Expect(raw.CallContext(ctx, &peers, "admin_peers")).To(gomega.Succeed())

	var datadir string
	gomega.Expect(raw.CallContext(ctx, &datadir, "admin_datadir")).To(gomega.Succeed())
	gomega.Expect(datadir).NotTo(gomega.BeEmpty())
}

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
	history, err := suite.client.FeeHistory(ctx, 1, blockNumber, []float64{50})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(history.GasUsedRatio).To(gomega.HaveLen(1))

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

	accessList, accessGas, accessError, err := proofClient.CreateAccessList(ctx, call)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(accessList).NotTo(gomega.BeNil())
	gomega.Expect(accessGas).To(gomega.BeNumerically(">", 0))
	gomega.Expect(accessError).To(gomega.BeEmpty())

	receiptsByNumber, err := suite.client.BlockReceipts(ctx, blockSelector)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(receiptsByNumber).NotTo(gomega.BeEmpty())
	receiptsByHash, err := suite.client.BlockReceipts(
		ctx,
		rpc.BlockNumberOrHashWithHash(fixture.block.Hash(), true),
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(receiptsByHash).To(gomega.HaveLen(len(receiptsByNumber)))
}

func (suite *liveSuite) assertTransactions(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()
	fixture := suite.fixture
	blockNumber := fixture.receipt.BlockNumber

	index := hexutil.Uint64(fixture.receipt.TransactionIndex)
	var countByNumber, countByHash hexutil.Uint
	gomega.Expect(raw.CallContext(
		ctx,
		&countByNumber,
		"qrl_getBlockTransactionCountByNumber",
		rpc.BlockNumber(blockNumber.Int64()),
	)).To(gomega.Succeed())
	gomega.Expect(raw.CallContext(
		ctx,
		&countByHash,
		"qrl_getBlockTransactionCountByHash",
		fixture.block.Hash(),
	)).To(gomega.Succeed())
	gomega.Expect(uint64(countByNumber)).To(gomega.BeNumerically(">", uint64(index)))
	gomega.Expect(countByHash).To(gomega.Equal(countByNumber))

	var transaction map[string]json.RawMessage
	gomega.Expect(raw.CallContext(
		ctx,
		&transaction,
		"qrl_getTransactionByBlockNumberAndIndex",
		rpc.BlockNumber(blockNumber.Int64()),
		index,
	)).To(gomega.Succeed())
	gomega.Expect(transaction).To(gomega.HaveKey("hash"))
	gomega.Expect(raw.CallContext(
		ctx,
		&transaction,
		"qrl_getTransactionByBlockHashAndIndex",
		fixture.block.Hash(),
		index,
	)).To(gomega.Succeed())

	var rawByNumber, rawByHash, rawByTransactionHash hexutil.Bytes
	gomega.Expect(raw.CallContext(
		ctx,
		&rawByNumber,
		"qrl_getRawTransactionByBlockNumberAndIndex",
		rpc.BlockNumber(blockNumber.Int64()),
		index,
	)).To(gomega.Succeed())
	gomega.Expect(raw.CallContext(
		ctx,
		&rawByHash,
		"qrl_getRawTransactionByBlockHashAndIndex",
		fixture.block.Hash(),
		index,
	)).To(gomega.Succeed())
	gomega.Expect(raw.CallContext(
		ctx,
		&rawByTransactionHash,
		"qrl_getRawTransactionByHash",
		fixture.tx.Hash(),
	)).To(gomega.Succeed())
	gomega.Expect(rawByNumber).To(gomega.Equal(rawByHash))
	gomega.Expect(rawByHash).To(gomega.Equal(rawByTransactionHash))

	wantRaw, err := fixture.tx.MarshalBinary()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(rawByTransactionHash).To(gomega.Equal(hexutil.Bytes(wantRaw)))

	found, pending, err := suite.client.TransactionByHash(ctx, fixture.tx.Hash())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(pending).To(gomega.BeFalse())
	gomega.Expect(found.Hash()).To(gomega.Equal(fixture.tx.Hash()))

	inBlock, err := suite.client.TransactionInBlock(
		ctx,
		fixture.block.Hash(),
		uint(fixture.receipt.TransactionIndex),
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(inBlock.Hash()).To(gomega.Equal(fixture.tx.Hash()))

	receipt, err := suite.client.TransactionReceipt(ctx, fixture.tx.Hash())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(receipt.TxHash).To(gomega.Equal(fixture.tx.Hash()))

	var receiptJSON map[string]json.RawMessage
	gomega.Expect(raw.CallContext(
		ctx,
		&receiptJSON,
		"qrl_getTransactionReceipt",
		fixture.tx.Hash(),
	)).To(gomega.Succeed())
	gomega.Expect(receiptJSON).To(gomega.HaveKey("logs"))

	var filled struct {
		Raw hexutil.Bytes   `json:"raw"`
		Tx  json.RawMessage `json:"tx"`
	}
	gomega.Expect(raw.CallContext(ctx, &filled, "qrl_fillTransaction", map[string]any{
		"from":  suite.from,
		"to":    suite.from,
		"value": "0x0",
	})).To(gomega.Succeed())
	gomega.Expect(filled.Raw).NotTo(gomega.BeEmpty())
	gomega.Expect(filled.Tx).NotTo(gomega.BeEmpty())

	var pendingTransactions []json.RawMessage
	gomega.Expect(raw.CallContext(
		ctx,
		&pendingTransactions,
		"qrl_pendingTransactions",
	)).To(gomega.Succeed())
}

func (suite *liveSuite) assertTxPool(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()

	var txpoolContent, txpoolInspect json.RawMessage
	gomega.Expect(raw.CallContext(ctx, &txpoolContent, "txpool_content")).To(gomega.Succeed())
	gomega.Expect(json.Valid(txpoolContent)).To(gomega.BeTrue())
	gomega.Expect(raw.CallContext(ctx, &txpoolContent, "txpool_contentFrom", suite.from)).To(gomega.Succeed())
	gomega.Expect(json.Valid(txpoolContent)).To(gomega.BeTrue())

	var txpoolStatus map[string]hexutil.Uint
	gomega.Expect(raw.CallContext(ctx, &txpoolStatus, "txpool_status")).To(gomega.Succeed())
	gomega.Expect(txpoolStatus).To(gomega.HaveKey("pending"))
	gomega.Expect(txpoolStatus).To(gomega.HaveKey("queued"))

	gomega.Expect(raw.CallContext(ctx, &txpoolInspect, "txpool_inspect")).To(gomega.Succeed())
	gomega.Expect(json.Valid(txpoolInspect)).To(gomega.BeTrue())
}

func (suite *liveSuite) assertRuntimeDiagnostics(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()

	var memStats runtime.MemStats
	gomega.Expect(raw.CallContext(ctx, &memStats, "debug_memStats")).To(gomega.Succeed())
	gomega.Expect(memStats.Sys).To(gomega.BeNumerically(">", 0))

	var gcStats runtimedebug.GCStats
	gomega.Expect(raw.CallContext(ctx, &gcStats, "debug_gcStats")).To(gomega.Succeed())

	var stacks string
	gomega.Expect(raw.CallContext(ctx, &stacks, "debug_stacks")).To(gomega.Succeed())
	gomega.Expect(bytes.Contains([]byte(stacks), []byte("goroutine"))).To(gomega.BeTrue())
}

func (suite *liveSuite) assertManagedSigning(ctx context.Context) {
	ginkgo.GinkgoHelper()

	raw := suite.client.Client()
	var accounts []common.Address
	gomega.Expect(raw.CallContext(ctx, &accounts, "qrl_accounts")).To(gomega.Succeed())
	gomega.Expect(accounts).To(gomega.ContainElement(suite.from))

	message := hexutil.Bytes("live API signing")
	var signature hexutil.Bytes
	gomega.Expect(raw.CallContext(
		ctx,
		&signature,
		"qrl_sign",
		suite.from,
		message,
	)).To(gomega.Succeed())
	valid, err := pqcrypto.MLDSA87VerifySignature(
		signature,
		qrlaccounts.TextHash(message),
		suite.wallet.GetPK(),
		suite.wallet.GetDescriptor(),
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(valid).To(gomega.BeTrue())

	nonce, err := suite.client.PendingNonceAt(ctx, suite.from)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	tip, err := suite.client.SuggestGasTipCap(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	feeCap, err := suite.client.SuggestGasPrice(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	feeCap = new(big.Int).Mul(feeCap, big.NewInt(4))
	if feeCap.Cmp(tip) < 0 {
		feeCap = new(big.Int).Set(tip)
	}
	gas, err := suite.client.EstimateGas(ctx, qrl.CallMsg{
		From: suite.from,
		To:   &suite.from,
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	args := map[string]any{
		"from":                 suite.from,
		"to":                   suite.from,
		"gas":                  hexutil.Uint64(gas),
		"nonce":                hexutil.Uint64(nonce),
		"value":                (*hexutil.Big)(new(big.Int)),
		"maxFeePerGas":         (*hexutil.Big)(feeCap),
		"maxPriorityFeePerGas": (*hexutil.Big)(tip),
	}
	var signed struct {
		Raw hexutil.Bytes      `json:"raw"`
		Tx  *types.Transaction `json:"tx"`
	}
	gomega.Expect(raw.CallContext(
		ctx,
		&signed,
		"qrl_signTransaction",
		args,
	)).To(gomega.Succeed())
	gomega.Expect(signed.Raw).NotTo(gomega.BeEmpty())
	gomega.Expect(signed.Tx).NotTo(gomega.BeNil())
	encoded, err := signed.Tx.MarshalBinary()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(signed.Raw).To(gomega.Equal(hexutil.Bytes(encoded)))
	sender, err := types.Sender(
		types.LatestSignerForChainID(suite.chainID),
		signed.Tx,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(sender).To(gomega.Equal(suite.from))

	var hash common.Hash
	gomega.Expect(raw.CallContext(
		ctx,
		&hash,
		"qrl_sendTransaction",
		args,
	)).To(gomega.Succeed())
	gomega.Expect(hash).NotTo(gomega.Equal(common.Hash{}))
	gomega.Expect(suite.waitReceipt(ctx, hash).Status).To(
		gomega.Equal(types.ReceiptStatusSuccessful),
	)
}
