// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

//go:build e2e

package vm

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"math/big"
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
	address   common.Address
	input     []byte
	want      []byte
	gas       uint64
	emptyWant []byte
	emptyGas  uint64
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
				intrinsic, err := core.IntrinsicGas(vector.input, nil, false)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				output, err := session.Client.CallContract(ctx, qrl.CallMsg{
					From: session.Address,
					To:   &vector.address,
					Gas:  intrinsic + vector.gas,
					Data: vector.input,
				}, nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(output).To(gomega.Equal(vector.want))
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("preserves each precompile's defined empty-input behavior", func(ctx ginkgo.SpecContext) {
			for _, vector := range vectors {
				intrinsic, err := core.IntrinsicGas(nil, nil, false)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				output, err := session.Client.CallContract(ctx, qrl.CallMsg{
					From: session.Address,
					To:   &vector.address,
					Gas:  intrinsic + vector.emptyGas,
				}, nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				if vector.emptyWant == nil {
					gomega.Expect(output).To(gomega.BeEmpty())
				} else {
					gomega.Expect(output).To(gomega.Equal(vector.emptyWant))
				}
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("prices SHA-256 and identity inputs in 64-byte words", func(ctx ginkgo.SpecContext) {
			for _, size := range []int{63, 64, 65} {
				input := bytes.Repeat([]byte{byte(size)}, size)
				shaOutput := sha256.Sum256(input)
				words := uint64((size + qrvm.WordBytes - 1) / qrvm.WordBytes)
				for _, test := range []struct {
					address common.Address
					want    []byte
					gas     uint64
				}{
					{
						address: common.BytesToAddress([]byte{2}),
						want:    shaOutput[:],
						gas:     60 + 12*words,
					},
					{
						address: common.BytesToAddress([]byte{4}),
						want:    input,
						gas:     15 + 3*words,
					},
				} {
					intrinsic, err := core.IntrinsicGas(input, nil, false)
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					output, err := session.Client.CallContract(ctx, qrl.CallMsg{
						From: session.Address,
						To:   &test.address,
						Gas:  intrinsic + test.gas,
						Data: input,
					}, nil)
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					gomega.Expect(output).To(gomega.Equal(test.want))

					_, err = session.Client.CallContract(ctx, qrl.CallMsg{
						From: session.Address,
						To:   &test.address,
						Gas:  intrinsic + test.gas - 1,
						Data: input,
					}, nil)
					gomega.Expect(err).To(gomega.HaveOccurred())
				}
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("executes a precompile through STATICCALL with exact forwarded gas", func(ctx ginkgo.SpecContext) {
			suite := &liveSuite{
				session: session,
				target:  common.BytesToAddress([]byte{0xee}),
			}
			input := []byte("abc")
			wantHash := sha256.Sum256(input)

			output := suite.callCodeWithInput(
				ctx,
				staticCallPrecompileCode(2, 72),
				input,
				nil,
			)
			gomega.Expect(output).To(gomega.HaveLen(2 * qrvm.WordBytes))
			gomega.Expect(output[:sha256.Size]).To(gomega.Equal(wantHash[:]))
			gomega.Expect(output[sha256.Size:qrvm.WordBytes]).To(
				gomega.Equal(make([]byte, qrvm.WordBytes-sha256.Size)),
			)
			gomega.Expect(new(big.Int).SetBytes(output[qrvm.WordBytes:])).To(
				gomega.Equal(big.NewInt(1)),
			)

			output = suite.callCodeWithInput(
				ctx,
				staticCallPrecompileCode(2, 71),
				input,
				nil,
			)
			gomega.Expect(new(big.Int).SetBytes(output[qrvm.WordBytes:]).Sign()).To(
				gomega.BeZero(),
			)
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

		ginkgo.It("rejects a full-length invalid ML-DSA-87 signature", func(ctx ginkgo.SpecContext) {
			address := common.BytesToAddress([]byte{3})
			input := mldsaInput()
			signatureOffset := common.HashLength + cryptomldsa87.CRYPTO_PUBLIC_KEY_BYTES
			input[signatureOffset] ^= 0x01
			output, err := session.Client.CallContract(ctx, qrl.CallMsg{
				From: session.Address,
				To:   &address,
				Data: input,
			}, nil)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(output).To(gomega.BeEmpty())
		}, ginkgo.SpecTimeout(liveSpecTimeout))

		ginkgo.It("rejects insufficient gas for every precompile", func(ctx ginkgo.SpecContext) {
			for _, vector := range vectors {
				intrinsic, err := core.IntrinsicGas(vector.input, nil, false)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				_, err = session.Client.CallContract(ctx, qrl.CallMsg{
					From: session.Address,
					To:   &vector.address,
					Gas:  intrinsic + vector.gas - 1,
					Data: vector.input,
				}, nil)
				gomega.Expect(err).To(gomega.HaveOccurred())
			}
		}, ginkgo.SpecTimeout(liveSpecTimeout))
	},
)

func precompileVectors() []precompileVector {
	shaOutput := sha256.Sum256([]byte("abc"))
	emptySHAOutput := sha256.Sum256(nil)
	depositRoot := hexutil.MustDecode("0x0149b6fc5da25356dbc4de56f7976a2693c9c4bdeb5ddfcf3a983ef26b6da6a7")
	emptyDepositRoot := hexutil.MustDecode("0x474d096b9dd154f74b552328a38dee860e0a25bf3af8f0f0371266dddc9ab676")
	vectors := []precompileVector{
		{
			address:   common.BytesToAddress([]byte{1}),
			input:     depositInput(),
			want:      depositRoot,
			gas:       18_000,
			emptyWant: emptyDepositRoot,
			emptyGas:  18_000,
		},
		{
			address:   common.BytesToAddress([]byte{2}),
			input:     []byte("abc"),
			want:      shaOutput[:],
			gas:       72,
			emptyWant: emptySHAOutput[:],
			emptyGas:  60,
		},
		{
			address:  common.BytesToAddress([]byte{3}),
			input:    mldsaInput(),
			want:     common.LeftPadBytes([]byte{1}, qrvm.WordBytes),
			gas:      125_000,
			emptyGas: 125_000,
		},
		{
			address:  common.BytesToAddress([]byte{4}),
			input:    []byte("identity"),
			want:     []byte("identity"),
			gas:      18,
			emptyGas: 15,
		},
		{
			address:  common.BytesToAddress([]byte{5}),
			input:    modExpInput(),
			want:     []byte{6},
			gas:      200,
			emptyGas: 200,
		},
	}
	sort.Slice(vectors, func(i, j int) bool {
		return bytes.Compare(vectors[i].address[:], vectors[j].address[:]) < 0
	})
	return vectors
}

func depositInput() []byte {
	input := make([]byte, pqcrypto.MLDSA87PublicKeyLength+common.AddressLength+8+pqcrypto.MLDSA87SignatureLength)
	for index := range input {
		input[index] = byte(index%251 + 1)
	}
	binary.LittleEndian.PutUint64(
		input[pqcrypto.MLDSA87PublicKeyLength+common.AddressLength:],
		32_000_000_000,
	)
	return input
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
