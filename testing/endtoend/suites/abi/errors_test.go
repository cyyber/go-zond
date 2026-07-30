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
	"errors"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/theQRL/go-qrl/accounts/abi"
	"github.com/theQRL/go-qrl/accounts/abi/bind"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/rpc"
)

func (fixture *liveFixture) callRevertData(
	ctx context.Context,
	method string,
	args ...any,
) []byte {
	ginkgo.GinkgoHelper()

	var output []any
	callErr := fixture.contract.Call(
		&bind.CallOpts{Context: ctx, BlockNumber: fixture.deploymentBlock},
		&output,
		method,
		args...,
	)
	gomega.Expect(callErr).To(gomega.HaveOccurred(), "%s unexpectedly succeeded", method)
	var dataError rpc.DataError
	gomega.Expect(errors.As(callErr, &dataError)).To(
		gomega.BeTrue(),
		"%s returned %T, want rpc.DataError",
		method,
		callErr,
	)
	encoded, ok := dataError.ErrorData().(string)
	gomega.Expect(ok).To(
		gomega.BeTrue(),
		"%s revert data has type %T, want hex string",
		method,
		dataError.ErrorData(),
	)
	revertData, err := hexutil.Decode(encoded)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "decode %s revert data %q", method, encoded)
	return revertData
}

func (fixture *liveFixture) assertErrors(ctx context.Context) {
	ginkgo.GinkgoHelper()

	inputs := fixture.inputs
	record := EventEmitterRecord{
		Amount:    inputs.amount,
		Recipient: fixture.from,
		Tag:       inputs.tag,
	}
	complexArguments := []any{
		inputs.amount,
		"unique custom error across a VM word: " + inputs.note,
		inputs.payload,
		record,
		[][]uint16{{}, {1, 0xffff}, {0x1234}},
	}

	// Hyperion:
	// error ComplexFailure(uint512, string, bytes, Record, uint16[][]);
	// function failComplex(...) external pure {
	//     revert ComplexFailure(code, reason, payload, record, nested);
	// }
	// Goal: RPC revert data equals the four-byte selector followed by the
	// canonical VM encoding of every custom-error argument; selector lookup,
	// decoding, and re-encoding then reproduce the same payload.
	ginkgo.By("round-tripping a complex custom error through raw revert data and selector-based ABI decoding")
	definition, ok := fixture.contractABI.Errors["ComplexFailure"]
	gomega.Expect(ok).To(gomega.BeTrue(), "ABI has no ComplexFailure error")
	signature := definition.Sig
	revertData := fixture.callRevertData(ctx, "failComplex", complexArguments...)
	encodedArguments, err := definition.Inputs.Pack(complexArguments...)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "pack %s", signature)
	wantRevertData := append([]byte{}, definition.ID[:4]...)
	wantRevertData = append(wantRevertData, encodedArguments...)
	gomega.Expect(revertData).To(gomega.Equal(wantRevertData), "%s compiler revert", signature)

	var errorSelector [4]byte
	copy(errorSelector[:], revertData)
	resolvedError, err := fixture.contractABI.ErrorByID(errorSelector)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "ErrorByID(%s)", signature)
	gomega.Expect(resolvedError.Sig).To(gomega.Equal(signature))
	decoded, err := resolvedError.Unpack(revertData)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "decode %s", signature)
	decodedArguments, ok := decoded.([]any)
	gomega.Expect(ok).To(gomega.BeTrue(), "decoded %s has type %T", signature, decoded)
	repackedArguments, err := resolvedError.Inputs.Pack(decodedArguments...)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "repack %s", signature)
	gomega.Expect(repackedArguments).To(
		gomega.Equal(encodedArguments),
		"decoded %s did not round-trip",
		signature,
	)

	// Hyperion:
	// function failReason() external pure { revert("VM standard revert reason"); }
	// function failPanic() external pure { assert(false); }
	// Goal: standard Error(string) and Panic(uint256) payloads decode to their
	// human-readable reason.
	ginkgo.By("decoding standard Error(string) and Panic(uint256) revert payloads")
	for _, standardError := range []struct {
		method string
		want   string
	}{
		{method: "failReason", want: "VM standard revert reason"},
		{method: "failPanic", want: "assert(false)"},
	} {
		reason, err := abi.UnpackRevert(fixture.callRevertData(ctx, standardError.method))
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "decode %s", standardError.method)
		gomega.Expect(reason).To(gomega.Equal(standardError.want), standardError.method)
	}

	// Hyperion:
	// function failReason() external pure { revert("VM standard revert reason"); }
	// Goal: submitting the reverting call as a transaction produces a mined
	// receipt whose status is explicitly failed.
	ginkgo.By("requiring a failed receipt for a reverting transaction")
	auth := fixture.transactOpts(ctx)
	auth.GasLimit = 1_000_000
	failedTx, err := fixture.contract.Transact(auth, "failReason")
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "submit reverting transaction")
	failedReceipt := fixture.waitTransaction(ctx, failedTx)
	gomega.Expect(failedReceipt.Status).To(
		gomega.Equal(types.ReceiptStatusFailed),
		"reverting transaction %s status",
		failedTx.Hash(),
	)
}
