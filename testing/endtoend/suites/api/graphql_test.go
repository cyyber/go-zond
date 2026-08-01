// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"math/big"
	"net/http"

	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

//go:embed testdata/api.graphql
var apiGraphQLQuery string

type graphQLAccount struct {
	Address string `json:"address"`
}

type graphQLBlock struct {
	Number string `json:"number"`
	Hash   string `json:"hash"`
}

type graphQLTransactionRef struct {
	Hash string `json:"hash"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type graphQLTransaction struct {
	Hash       string          `json:"hash"`
	Nonce      string          `json:"nonce"`
	Index      string          `json:"index"`
	From       *graphQLAccount `json:"from"`
	To         *graphQLAccount `json:"to"`
	Value      string          `json:"value"`
	FeeCap     string          `json:"maxFeePerGas"`
	TipCap     string          `json:"maxPriorityFeePerGas"`
	Tip        string          `json:"effectiveTip"`
	Gas        string          `json:"gas"`
	Input      string          `json:"inputData"`
	Block      *graphQLBlock   `json:"block"`
	Status     string          `json:"status"`
	GasUsed    string          `json:"gasUsed"`
	TotalGas   string          `json:"cumulativeGasUsed"`
	GasPrice   string          `json:"effectiveGasPrice"`
	Created    *graphQLAccount `json:"createdContract"`
	Logs       []graphQLLog    `json:"logs"`
	AccessList []struct {
		Address     string   `json:"address"`
		StorageKeys []string `json:"storageKeys"`
	} `json:"accessList"`
	Descriptor  string `json:"descriptor"`
	ExtraParams string `json:"extraParams"`
	Signature   string `json:"signature"`
	PublicKey   string `json:"publicKey"`
	Type        string `json:"type"`
	Raw         string `json:"raw"`
	RawReceipt  string `json:"rawReceipt"`
}

type graphQLLog struct {
	Index       string                `json:"index"`
	Account     graphQLAccount        `json:"account"`
	Topics      []string              `json:"topics"`
	Data        string                `json:"data"`
	Transaction graphQLTransactionRef `json:"transaction"`
}

func assertGraphQLTransaction(
	got graphQLTransaction,
	tx *types.Transaction,
	receipt *types.Receipt,
	block *types.Block,
	chainID *big.Int,
) {
	ginkgo.GinkgoHelper()

	from, err := types.Sender(types.LatestSignerForChainID(chainID), tx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	tip, err := tx.EffectiveGasTip(block.BaseFee())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	gomega.Expect(got.Nonce).To(gomega.Equal(hexutil.EncodeUint64(tx.Nonce())))
	gomega.Expect(got.Index).To(
		gomega.Equal(hexutil.EncodeUint64(uint64(receipt.TransactionIndex))),
	)
	gomega.Expect(got.From).NotTo(gomega.BeNil())
	gomega.Expect(got.From.Address).To(gomega.Equal(from.Hex()))
	if tx.To() == nil {
		gomega.Expect(got.To).To(gomega.BeNil())
	} else {
		gomega.Expect(got.To).NotTo(gomega.BeNil())
		gomega.Expect(got.To.Address).To(gomega.Equal(tx.To().Hex()))
	}
	gomega.Expect(got.Value).To(gomega.Equal(hexutil.EncodeBig(tx.Value())), "GraphQL transaction value")
	gomega.Expect(got.FeeCap).To(gomega.Equal(hexutil.EncodeBig(tx.GasFeeCap())), "GraphQL max fee per gas")
	gomega.Expect(got.TipCap).To(gomega.Equal(hexutil.EncodeBig(tx.GasTipCap())), "GraphQL max priority fee per gas")
	gomega.Expect(got.Tip).To(gomega.Equal(hexutil.EncodeBig(tip)), "GraphQL effective tip")
	gomega.Expect(got.Gas).To(gomega.Equal(hexutil.EncodeUint64(tx.Gas())), "GraphQL transaction gas")
	gomega.Expect(got.Input).To(gomega.Equal(hexutil.Encode(tx.Data())))
	gomega.Expect(got.Block).NotTo(gomega.BeNil())
	gomega.Expect(got.Block.Number).To(gomega.Equal(hexutil.EncodeBig(receipt.BlockNumber)))
	gomega.Expect(got.Block.Hash).To(gomega.Equal(receipt.BlockHash.Hex()))
	gomega.Expect(got.Status).To(gomega.Equal(hexutil.EncodeUint64(receipt.Status)))
	gomega.Expect(got.GasUsed).To(gomega.Equal(hexutil.EncodeUint64(receipt.GasUsed)))
	gomega.Expect(got.TotalGas).To(
		gomega.Equal(hexutil.EncodeUint64(receipt.CumulativeGasUsed)),
	)
	gomega.Expect(receipt.EffectiveGasPrice).NotTo(gomega.BeNil())
	gomega.Expect(got.GasPrice).To(
		gomega.Equal(hexutil.EncodeBig(receipt.EffectiveGasPrice)),
		"GraphQL effective gas price",
	)
	gomega.Expect(got.Created).NotTo(gomega.BeNil())
	gomega.Expect(got.Created.Address).To(gomega.Equal(receipt.ContractAddress.Hex()))
	gomega.Expect(got.Type).To(gomega.Equal(hexutil.EncodeUint64(uint64(tx.Type()))))
	gomega.Expect(got.AccessList).To(gomega.HaveLen(len(tx.AccessList())))
	for index, tuple := range tx.AccessList() {
		gomega.Expect(got.AccessList[index].Address).To(gomega.Equal(tuple.Address.Hex()))
		storageKeys := make([]string, len(tuple.StorageKeys))
		for keyIndex, key := range tuple.StorageKeys {
			storageKeys[keyIndex] = key.Hex()
		}
		gomega.Expect(got.AccessList[index].StorageKeys).To(gomega.Equal(storageKeys))
	}
	gomega.Expect(got.Logs).To(gomega.HaveLen(len(receipt.Logs)))
	for index := range receipt.Logs {
		assertGraphQLLog(got.Logs[index], receipt.Logs[index])
	}
	assertGraphQLTransactionEnvelope(got, tx, receipt)
}

func assertGraphQLTransactionEnvelope(
	got graphQLTransaction,
	tx *types.Transaction,
	receipt *types.Receipt,
) {
	ginkgo.GinkgoHelper()

	rawTransaction, err := tx.MarshalBinary()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	rawReceipt, err := receipt.MarshalBinary()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	gomega.Expect(got.Hash).To(gomega.Equal(tx.Hash().Hex()))
	gomega.Expect(got.Descriptor).To(gomega.Equal(hexutil.Encode(tx.Descriptor())))
	gomega.Expect(got.ExtraParams).To(gomega.Equal(hexutil.Encode(tx.ExtraParams())))
	gomega.Expect(got.Signature).To(gomega.Equal(hexutil.Encode(tx.RawSignatureValue())))
	gomega.Expect(got.PublicKey).To(gomega.Equal(hexutil.Encode(tx.RawPublicKeyValue())))
	gomega.Expect(got.Raw).To(gomega.Equal(hexutil.Encode(rawTransaction)))
	gomega.Expect(got.RawReceipt).To(gomega.Equal(hexutil.Encode(rawReceipt)))
}

func assertGraphQLLog(got graphQLLog, want *types.Log) {
	ginkgo.GinkgoHelper()

	topics := make([]string, len(want.Topics))
	for index, topic := range want.Topics {
		topics[index] = topic.Hex()
	}
	gomega.Expect(got.Index).To(gomega.Equal(hexutil.EncodeUint64(uint64(want.Index))))
	gomega.Expect(got.Account.Address).To(gomega.Equal(want.Address.Hex()))
	gomega.Expect(got.Topics).To(gomega.Equal(topics))
	gomega.Expect(got.Data).To(gomega.Equal(hexutil.Encode(want.Data)), "GraphQL log data")
	gomega.Expect(got.Transaction.Hash).To(gomega.Equal(want.TxHash.Hex()))
}

func (suite *liveSuite) queryGraphQL(
	ctx context.Context,
	query string,
	variables map[string]any,
) json.RawMessage {
	ginkgo.GinkgoHelper()

	payload, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		suite.graphQLURL,
		bytes.NewReader(payload),
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(response.StatusCode).To(
		gomega.Equal(http.StatusOK),
		"GraphQL response: %s",
		body,
	)

	var decoded graphQLResponse
	gomega.Expect(json.Unmarshal(body, &decoded)).To(gomega.Succeed())
	gomega.Expect(decoded.Errors).To(gomega.BeEmpty(), "GraphQL response: %s", body)
	gomega.Expect(decoded.Data).NotTo(gomega.BeEmpty())
	return decoded.Data
}
