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
	"net/http"

	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

//go:embed testdata/api.graphql
var apiGraphQLQuery string

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type graphQLTransaction struct {
	Hash string `json:"hash"`
	To   *struct {
		Address string `json:"address"`
	} `json:"to"`
	AccessList []struct {
		Address     string   `json:"address"`
		StorageKeys []string `json:"storageKeys"`
	} `json:"accessList"`
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
	gomega.Expect(got.Data).To(
		gomega.Equal(hexutil.Encode(fixture.value[:])),
		"GraphQL log data",
	)
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
