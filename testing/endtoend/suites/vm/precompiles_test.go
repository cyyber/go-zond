// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package vm

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"sort"

	qrl "github.com/theQRL/go-qrl"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/core"
	qrvm "github.com/theQRL/go-qrl/core/vm"
	"github.com/theQRL/go-qrl/crypto"
	"github.com/theQRL/go-qrl/crypto/pqcrypto"
	endtoendlive "github.com/theQRL/go-qrl/testing/endtoend/internal/live"
	cryptomldsa87 "github.com/theQRL/go-qrllib/crypto/ml_dsa_87"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

type precompileVector struct {
	address common.Address
	input   []byte
	want    []byte
}

var _ = ginkgo.Describe(
	"registered precompiles",
	ginkgo.Serial,
	ginkgo.Ordered,
	ginkgo.ContinueOnFailure,
	ginkgo.Label("e2e", "live", "precompile"),
	func() {
		var (
			session *endtoendlive.Session
			vectors []precompileVector
		)

		ginkgo.BeforeAll(func(ctx ginkgo.SpecContext) {
			var err error
			session, err = endtoendlive.Open(ctx, false)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			ginkgo.DeferCleanup(session.Close)
			vectors = precompileVectors()
			gomega.Expect(vectors).To(gomega.HaveLen(len(qrvm.PrecompiledContractsZond)))
		})

		ginkgo.It("executes one successful vector for every precompile", func(ctx ginkgo.SpecContext) {
			for _, vector := range vectors {
				output, err := session.Client.CallContract(ctx, qrl.CallMsg{
					From: session.Address,
					To:   &vector.address,
					Data: vector.input,
				}, nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(output).To(gomega.Equal(vector.want))
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("preserves each precompile's defined empty-input behavior", func(ctx ginkgo.SpecContext) {
			for _, vector := range vectors {
				contract := qrvm.PrecompiledContractsZond[vector.address]
				want, err := contract.Run(nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				output, err := session.Client.CallContract(ctx, qrl.CallMsg{
					From: session.Address,
					To:   &vector.address,
				}, nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(output).To(gomega.Equal(want))
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("rejects malformed ML-DSA-87 input", func(ctx ginkgo.SpecContext) {
			address := common.BytesToAddress([]byte{3})
			output, err := session.Client.CallContract(ctx, qrl.CallMsg{
				From: session.Address,
				To:   &address,
				Data: []byte{1, 2, 3},
			}, nil)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(output).To(gomega.BeEmpty())
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("rejects insufficient gas for every precompile", func(ctx ginkgo.SpecContext) {
			for _, vector := range vectors {
				contract := qrvm.PrecompiledContractsZond[vector.address]
				intrinsic, err := core.IntrinsicGas(vector.input, nil, false)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				required := contract.RequiredGas(vector.input)
				_, err = session.Client.CallContract(ctx, qrl.CallMsg{
					From: session.Address,
					To:   &vector.address,
					Gas:  intrinsic + required - 1,
					Data: vector.input,
				}, nil)
				gomega.Expect(err).To(gomega.HaveOccurred())
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))
	},
)

func precompileVectors() []precompileVector {
	shaOutput := sha256.Sum256([]byte("abc"))
	vectors := []precompileVector{
		{
			address: common.BytesToAddress([]byte{1}),
			input:   depositInput(),
			want:    hexutil.MustDecode("0x474d096b9dd154f74b552328a38dee860e0a25bf3af8f0f0371266dddc9ab676"),
		},
		{address: common.BytesToAddress([]byte{2}), input: []byte("abc"), want: shaOutput[:]},
		{address: common.BytesToAddress([]byte{3}), input: mldsaInput(), want: common.LeftPadBytes([]byte{1}, qrvm.WordBytes)},
		{address: common.BytesToAddress([]byte{4}), input: []byte("identity"), want: []byte("identity")},
		{address: common.BytesToAddress([]byte{5}), input: modExpInput(), want: []byte{6}},
	}
	sort.Slice(vectors, func(i, j int) bool {
		return bytes.Compare(vectors[i].address[:], vectors[j].address[:]) < 0
	})
	return vectors
}

func depositInput() []byte {
	return make([]byte, pqcrypto.MLDSA87PublicKeyLength+common.AddressLength+8+pqcrypto.MLDSA87SignatureLength)
}

func mldsaInput() []byte {
	signer, err := cryptomldsa87.New()
	if err != nil {
		panic(err)
	}
	digest := crypto.Keccak256([]byte("QRL live ML-DSA-87 precompile"))
	context := []byte("E2E")
	signature, err := signer.Sign(context, digest)
	if err != nil {
		panic(err)
	}
	publicKey := signer.GetPK()
	input := make([]byte, 0, len(digest)+len(publicKey)+len(signature)+1+len(context))
	input = append(input, digest...)
	input = append(input, publicKey[:]...)
	input = append(input, signature[:]...)
	input = append(input, byte(len(context)))
	return append(input, context...)
}

func modExpInput() []byte {
	input := make([]byte, 96, 99)
	binary.BigEndian.PutUint64(input[24:32], 1)
	binary.BigEndian.PutUint64(input[56:64], 1)
	binary.BigEndian.PutUint64(input[88:96], 1)
	return append(input, 2, 5, 13)
}
