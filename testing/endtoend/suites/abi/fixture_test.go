// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.
//
// The go-qrl library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-qrl library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-qrl library. If not, see <http://www.gnu.org/licenses/>.

package abi

import (
	"context"
	_ "embed"
	"math/big"
	"strings"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/theQRL/go-qrl/accounts/abi"
	"github.com/theQRL/go-qrl/accounts/abi/bind"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/qrlclient"
	"github.com/theQRL/go-qrl/testing/devnet"
)

// Regenerate the source-controlled Hyperion artifacts and generated binding.
// The compiler must be cyyber/hyperion@2b9a0f1d.
//
//go:generate sh -c "hypc --version 2>&1 | grep -Fq commit.2b9a0f1d || { echo 'hypc from cyyber/hyperion@2b9a0f1d is required; found:' >&2; hypc --version >&2; exit 1; }"
//go:generate hypc --abi --bin --optimize --optimize-runs 1 --no-cbor-metadata --overwrite -o testdata testdata/EventEmitter.hyp
//go:generate go -C ../../../.. run ./cmd/abigen --abi testing/endtoend/suites/abi/testdata/EventEmitter.abi --bin testing/endtoend/suites/abi/testdata/EventEmitter.bin --pkg abi --type EventEmitter --out testing/endtoend/suites/abi/contract.go

//go:embed testdata/EventEmitter.abi
var eventEmitterABIJSON string

type liveSuite struct {
	client      *qrlclient.Client
	wsClient    *qrlclient.Client
	from        common.Address
	signer      bind.SignerFn
	contractABI abi.ABI
	inputs      scenarioInputs
}

func setupLiveSuite(ctx context.Context) *liveSuite {
	ginkgo.GinkgoHelper()

	environment, err := devnet.Inspect(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	client, err := qrlclient.DialContext(ctx, environment.RPCURL)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(client.Close)

	wsClient, err := qrlclient.DialContext(ctx, environment.WebSocketURL)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.DeferCleanup(wsClient.Close)

	wallet, err := devnet.UnsafeDevelopmentWallet()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	chainID, err := client.ChainID(ctx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	transactor, err := bind.NewKeyedTransactorWithChainID(wallet, chainID)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	inputs := scenarioInputs{
		amount: new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 511), big.NewInt(0x1234)),
		delta:  new(big.Int).Add(new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 510)), big.NewInt(42)),
		note:   "VM string crosses the 64-byte ABI word boundary: 0123456789abcdef0123456789abcdef",
	}
	for index := range inputs.tag {
		inputs.tag[index] = byte(0x80 + index)
	}

	inputs.payload = make([]byte, 129)
	for index := range inputs.payload {
		inputs.payload[index] = byte((index*29 + 7) & 0xff)
	}

	parsed, err := abi.JSON(strings.NewReader(eventEmitterABIJSON))
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return &liveSuite{
		client:      client,
		wsClient:    wsClient,
		from:        transactor.From,
		signer:      transactor.Signer,
		contractABI: parsed,
		inputs:      inputs,
	}
}

func (suite *liveSuite) deployEventEmitter(ctx context.Context) *liveFixture {
	ginkgo.GinkgoHelper()

	deploymentAuth := suite.transactOpts(ctx)
	initial := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 500), big.NewInt(1337))
	deploymentNote := "dynamic constructor value: " + suite.inputs.note
	deploymentPayload := append([]byte(nil), suite.inputs.payload...)
	deploymentRecord := EventEmitterRecord{
		Amount:    suite.inputs.amount,
		Recipient: suite.from,
		Tag:       suite.inputs.tag,
	}
	deploymentNumbers := []uint16{0, 1, 0xffff, 0x1234}
	address, tx, binding, err := DeployEventEmitter(
		deploymentAuth,
		suite.client,
		initial,
		deploymentNote,
		deploymentPayload,
		deploymentRecord,
		deploymentNumbers,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	receipt := waitSuccessfulTransaction(ctx, suite.client, tx)
	gomega.Expect(receipt.ContractAddress).To(gomega.Equal(address))
	gomega.Expect(receipt.Logs).To(gomega.HaveLen(1))

	deployed := suite.contractABI.Events["Deployed"]
	log := receipt.Logs[0]
	gomega.Expect(log.Topics).To(gomega.Equal([]common.LogTopic{
		common.HashToLogTopic(deployed.ID),
	}))

	wantDeploymentData, err := deployed.Inputs.Pack(
		initial,
		deploymentNote,
		deploymentPayload,
		deploymentRecord,
		deploymentNumbers,
	)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "pack canonical Deployed event data")
	gomega.Expect(log.Data).To(gomega.Equal(wantDeploymentData))

	deployedEvent, err := binding.ParseDeployed(*log)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(deployedEvent.Value).To(gomega.Equal(initial))
	gomega.Expect(deployedEvent.Note).To(gomega.Equal(deploymentNote))
	gomega.Expect(deployedEvent.Payload).To(gomega.Equal(deploymentPayload))
	gomega.Expect(deployedEvent.Record).To(gomega.Equal(deploymentRecord))
	gomega.Expect(deployedEvent.Numbers).To(gomega.Equal(deploymentNumbers))

	return &liveFixture{
		liveSuite:       suite,
		deploymentBlock: receipt.BlockNumber,
		address:         address,
		contract: bind.NewBoundContract(
			address,
			suite.contractABI,
			suite.client,
			suite.client,
			suite.client,
		),
		binding: binding,
		initial: initial,
	}
}

type liveFixture struct {
	*liveSuite
	deploymentBlock *big.Int
	address         common.Address
	contract        *bind.BoundContract
	binding         *EventEmitter
	initial         *big.Int
}

type scenarioInputs struct {
	// Large uint512 value with upper-half bits set.
	amount *big.Int

	// Negative int512 value exercising signed 64-byte encoding.
	delta *big.Int

	// Fully populated bytes64 value.
	tag [64]byte

	// 129-byte dynamic value spanning three ABI data words.
	payload []byte

	// Dynamic string crossing the 64-byte ABI word boundary.
	note string
}

func (suite *liveSuite) transactOpts(ctx context.Context) *bind.TransactOpts {
	return &bind.TransactOpts{
		From:    suite.from,
		Signer:  suite.signer,
		Context: ctx,
	}
}

func waitTransaction(
	ctx context.Context,
	client *qrlclient.Client,
	tx *types.Transaction,
) *types.Receipt {
	ginkgo.GinkgoHelper()

	receipt, err := bind.WaitMined(ctx, client, tx)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "wait for transaction %s", tx.Hash())
	gomega.Expect(receipt).NotTo(gomega.BeNil(), "transaction %s has no mined receipt", tx.Hash())
	gomega.Expect(receipt.BlockNumber).NotTo(gomega.BeNil(), "transaction %s has no block number", tx.Hash())
	return receipt
}

func waitSuccessfulTransaction(
	ctx context.Context,
	client *qrlclient.Client,
	tx *types.Transaction,
) *types.Receipt {
	ginkgo.GinkgoHelper()

	receipt := waitTransaction(ctx, client, tx)
	gomega.Expect(receipt.Status).To(
		gomega.Equal(types.ReceiptStatusSuccessful),
		"transaction %s status",
		tx.Hash(),
	)
	return receipt
}
