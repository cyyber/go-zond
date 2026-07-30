// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type graphQLTransaction struct {
	Hash        string `json:"hash"`
	Descriptor  string `json:"descriptor"`
	ExtraParams string `json:"extraParams"`
	Signature   string `json:"signature"`
	PublicKey   string `json:"publicKey"`
	Raw         string `json:"raw"`
	RawReceipt  string `json:"rawReceipt"`
}

type graphQLLog struct {
	Topics []string `json:"topics"`
	Data   string   `json:"data"`
}

func (suite *liveSuite) assertGraphQLSurface(ctx context.Context) {
	ginkgo.GinkgoHelper()

	ginkgo.By("checking every GraphQL root field remains exposed")
	introspection := suite.graphQL(ctx, `{
		queryType: __type(name: "Query") { fields { name } }
		mutationType: __type(name: "Mutation") { fields { name } }
	}`)
	var schema struct {
		QueryType struct {
			Fields []struct {
				Name string `json:"name"`
			} `json:"fields"`
		} `json:"queryType"`
		MutationType struct {
			Fields []struct {
				Name string `json:"name"`
			} `json:"fields"`
		} `json:"mutationType"`
	}
	gomega.Expect(json.Unmarshal(introspection, &schema)).To(gomega.Succeed())
	expectGraphQLFields(schema.QueryType.Fields, []string{
		"block",
		"blocks",
		"pending",
		"transaction",
		"logs",
		"gasPrice",
		"maxPriorityFeePerGas",
		"syncing",
		"chainID",
	})
	expectGraphQLFields(schema.MutationType.Fields, []string{"sendRawTransaction"})

	fixture := suite.fixture
	block := hexutil.EncodeBig(fixture.receipt.BlockNumber)
	index := hexutil.EncodeUint64(uint64(fixture.receipt.TransactionIndex))
	slot := "0x" + fmt.Sprintf("%064x", 0)

	ginkgo.By("querying all root operations and nested API fields")
	query := `
	query APISurface(
		$block: Long!,
		$hash: Bytes32!,
		$txHash: Bytes32!,
		$address: Address!,
		$sender: Address!,
		$slot: Bytes32!,
		$topic: Bytes64!,
		$index: Long!
	) {
		block(number: $block) {
			number
			hash
			parent { number hash }
			transactionsRoot
			transactionCount
			stateRoot
			receiptsRoot
			miner { address balance transactionCount code storage(slot: $slot) }
			extraData
			gasLimit
			gasUsed
			baseFeePerGas
			nextBaseFeePerGas
			timestamp
			logsBloom
			random
			transactions {
				hash
				nonce
				index
				from { address }
				to { address }
				value
				maxFeePerGas
				maxPriorityFeePerGas
				effectiveTip
				gas
				inputData
				block { number hash }
				status
				gasUsed
				cumulativeGasUsed
				effectiveGasPrice
				createdContract { address }
				logs {
					index
					account { address }
					topics
					data
					transaction { hash }
				}
				descriptor
				extraParams
				signature
				publicKey
				type
				accessList { address storageKeys }
				raw
				rawReceipt
			}
			transactionAt(index: $index) { hash }
			logs(filter: { addresses: [$address], topics: [[$topic]] }) {
				index
				account { address }
				topics
				data
				transaction { hash }
			}
			account(address: $address) {
				address
				balance
				transactionCount
				code
				storage(slot: $slot)
			}
			call(data: { from: $sender, to: $address, data: "0x" }) {
				data
				gasUsed
				status
			}
			estimateGas(data: { from: $sender, to: $address, data: "0x" })
			rawHeader
			raw
			withdrawalsRoot
			withdrawals { index validator address amount }
		}
		blockByHash: block(hash: $hash) { number hash }
		blocks(from: $block, to: $block) { number hash }
		pending {
			transactionCount
			transactions { hash }
			account(address: $sender) { address balance transactionCount code }
			call(data: { from: $sender, to: $address, data: "0x" }) {
				data
				gasUsed
				status
			}
			estimateGas(data: { from: $sender, to: $address, data: "0x" })
		}
		transaction(hash: $txHash) {
			hash
			signature
			publicKey
			descriptor
			extraParams
			raw
			rawReceipt
		}
		logs(filter: {
			fromBlock: $block,
			toBlock: $block,
			addresses: [$address],
			topics: [[$topic]]
		}) {
			account { address }
			topics
			data
			transaction { hash }
		}
		gasPrice
		maxPriorityFeePerGas
		syncing { startingBlock currentBlock highestBlock }
		chainID
	}`
	data := suite.graphQLVariables(ctx, query, map[string]any{
		"block":   block,
		"hash":    fixture.block.Hash().Hex(),
		"txHash":  fixture.tx.Hash().Hex(),
		"address": fixture.address.Hex(),
		"sender":  suite.from.Hex(),
		"slot":    slot,
		"topic":   fixture.topic.Hex(),
		"index":   index,
	})
	var root struct {
		Block struct {
			Hash         string               `json:"hash"`
			Transactions []graphQLTransaction `json:"transactions"`
			Logs         []graphQLLog         `json:"logs"`
			Account      struct {
				Address string `json:"address"`
				Storage string `json:"storage"`
			} `json:"account"`
			Call struct {
				Data   string `json:"data"`
				Status string `json:"status"`
			} `json:"call"`
		} `json:"block"`
		BlockByHash struct {
			Hash string `json:"hash"`
		} `json:"blockByHash"`
		Blocks      []json.RawMessage  `json:"blocks"`
		Pending     json.RawMessage    `json:"pending"`
		Transaction graphQLTransaction `json:"transaction"`
		Logs        []graphQLLog       `json:"logs"`
		GasPrice    string             `json:"gasPrice"`
		PriorityFee string             `json:"maxPriorityFeePerGas"`
		Syncing     json.RawMessage    `json:"syncing"`
		ChainID     string             `json:"chainID"`
	}
	gomega.Expect(json.Unmarshal(data, &root)).To(gomega.Succeed())
	gomega.Expect(root.Block.Hash).To(gomega.Equal(fixture.block.Hash().Hex()))
	gomega.Expect(root.BlockByHash.Hash).To(gomega.Equal(fixture.block.Hash().Hex()))
	gomega.Expect(root.Blocks).To(gomega.HaveLen(1))
	gomega.Expect(root.Pending).NotTo(gomega.BeEmpty())
	gomega.Expect(root.Block.Account.Address).To(gomega.Equal(fixture.address.Hex()))
	gomega.Expect(root.Block.Account.Storage).To(gomega.Equal(fixture.value.Hex()))
	gomega.Expect(root.Block.Call.Data).To(gomega.Equal(fixture.value.Hex()))
	gomega.Expect(root.Block.Call.Status).To(gomega.Equal("0x1"))
	gomega.Expect(root.GasPrice).NotTo(gomega.BeEmpty())
	gomega.Expect(root.PriorityFee).NotTo(gomega.BeEmpty())
	gomega.Expect(root.ChainID).To(gomega.Equal(hexutil.EncodeBig(suite.chainID)))

	var blockTransaction *graphQLTransaction
	for index := range root.Block.Transactions {
		if root.Block.Transactions[index].Hash == fixture.tx.Hash().Hex() {
			blockTransaction = &root.Block.Transactions[index]
			break
		}
	}
	gomega.Expect(blockTransaction).NotTo(gomega.BeNil())
	assertGraphQLTransaction(*blockTransaction, fixture.tx, fixture.receipt)
	assertGraphQLTransaction(root.Transaction, fixture.tx, fixture.receipt)

	gomega.Expect(root.Block.Logs).To(gomega.HaveLen(1))
	assertGraphQLLog(root.Block.Logs[0], fixture)
	gomega.Expect(root.Logs).To(gomega.HaveLen(1))
	assertGraphQLLog(root.Logs[0], fixture)

	ginkgo.By("submitting a signed transaction through the GraphQL mutation")
	tx, err := suite.signTransaction(ctx, &suite.from, nil, nil)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	encoded, err := tx.MarshalBinary()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	mutationData := suite.graphQLVariables(ctx, `
		mutation Send($raw: Bytes!) {
			sendRawTransaction(data: $raw)
		}
	`, map[string]any{"raw": hexutil.Encode(encoded)})
	var mutation struct {
		Hash string `json:"sendRawTransaction"`
	}
	gomega.Expect(json.Unmarshal(mutationData, &mutation)).To(gomega.Succeed())
	gomega.Expect(mutation.Hash).To(gomega.Equal(tx.Hash().Hex()))
	gomega.Expect(suite.submitExistingAndWait(ctx, tx).Status).To(
		gomega.Equal(types.ReceiptStatusSuccessful),
	)
}

func assertGraphQLTransaction(
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

func assertGraphQLLog(got graphQLLog, fixture *liveFixture) {
	ginkgo.GinkgoHelper()

	gomega.Expect(got.Topics).To(gomega.Equal([]string{fixture.topic.Hex()}))
	gomega.Expect(got.Data).To(gomega.Equal(hexutil.Encode(fixture.value[:])))
}

func (suite *liveSuite) graphQL(ctx context.Context, query string) json.RawMessage {
	ginkgo.GinkgoHelper()
	return suite.graphQLVariables(ctx, query, nil)
}

func (suite *liveSuite) graphQLVariables(
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
		suite.environment.GraphQLURL,
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

func expectGraphQLFields(
	fields []struct {
		Name string `json:"name"`
	},
	want []string,
) {
	ginkgo.GinkgoHelper()

	names := make([]string, len(fields))
	for index, field := range fields {
		names[index] = field.Name
	}
	slices.Sort(names)
	slices.Sort(want)
	gomega.Expect(names).To(gomega.Equal(want))
}
