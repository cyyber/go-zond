// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package api

import (
	"context"
	"encoding/json"
	"slices"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

func (suite *liveSuite) assertGraphQLSchema(ctx context.Context) {
	ginkgo.GinkgoHelper()

	introspection := suite.queryGraphQL(ctx, `{
			queryType: __type(name: "Query") { fields { name } }
			mutationType: __type(name: "Mutation") { fields { name } }
		}`, nil)
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
