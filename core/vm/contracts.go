// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha3"
	"encoding/binary"
	gomath "math"
	"math/big"

	pkgerrors "github.com/pkg/errors"
	ssz "github.com/prysmaticlabs/fastssz"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/math"
	"github.com/theQRL/go-qrl/crypto/pqcrypto"
	"github.com/theQRL/go-qrl/params"
	cryptomldsa87 "github.com/theQRL/go-qrllib/crypto/ml_dsa_87"
)

// PrecompiledContract is the basic interface for native Go contracts. The implementation
// requires a deterministic gas count based on the input size of the Run method of the
// contract.
type PrecompiledContract interface {
	RequiredGas(input []byte) uint64  // RequiredPrice calculates the contract gas use
	Run(input []byte) ([]byte, error) // Run runs the precompiled contract
}

// trueWord is the template returned when a precompile verification succeeds.
var trueWord = common.LeftPadBytes([]byte{1}, WordBytes)

// PrecompiledContractsZond contains the default set of pre-compiled QRL
// contracts used in the Zond release.
var PrecompiledContractsZond = map[common.Address]PrecompiledContract{
	common.BytesToAddress([]byte{1}): &depositroot{},
	common.BytesToAddress([]byte{2}): &sha256hash{},
	common.BytesToAddress([]byte{3}): &mldsa87VerifyLegacy32{},
	common.BytesToAddress([]byte{4}): &dataCopy{},
	common.BytesToAddress([]byte{5}): &bigModExp{},
}

// PrecompiledContractsQRL2PQ contains the QRL 2.0 post-quantum precompile set.
// It updates the slot 3 message representative to 64 bytes and adds SHAKE256 at
// slot 6.
var PrecompiledContractsQRL2PQ = map[common.Address]PrecompiledContract{
	common.BytesToAddress([]byte{1}): &depositroot{},
	common.BytesToAddress([]byte{2}): &sha256hash{},
	common.BytesToAddress([]byte{3}): &mldsa87Verify{},
	common.BytesToAddress([]byte{4}): &dataCopy{},
	common.BytesToAddress([]byte{5}): &bigModExp{},
	common.BytesToAddress([]byte{6}): &shake256hash{},
}

var (
	PrecompiledAddressesZond = []common.Address{
		common.BytesToAddress([]byte{1}),
		common.BytesToAddress([]byte{2}),
		common.BytesToAddress([]byte{3}),
		common.BytesToAddress([]byte{4}),
		common.BytesToAddress([]byte{5}),
	}
	PrecompiledAddressesQRL2PQ = []common.Address{
		common.BytesToAddress([]byte{1}),
		common.BytesToAddress([]byte{2}),
		common.BytesToAddress([]byte{3}),
		common.BytesToAddress([]byte{4}),
		common.BytesToAddress([]byte{5}),
		common.BytesToAddress([]byte{6}),
	}
)

// ActivePrecompiles returns the precompiles enabled by the supplied chain rules.
func ActivePrecompiles(rules params.Rules) []common.Address {
	if rules.IsQRL2PQPrecompiles {
		return PrecompiledAddressesQRL2PQ
	}
	return PrecompiledAddressesZond
}

// RunPrecompiledContract runs and evaluates the output of a precompiled contract.
// It returns
// - the returned bytes,
// - the _remaining_ gas,
// - any error that occurred
func RunPrecompiledContract(p PrecompiledContract, input []byte, suppliedGas uint64) (ret []byte, remainingGas uint64, err error) {
	gasCost := p.RequiredGas(input)
	if suppliedGas < gasCost {
		return nil, 0, ErrOutOfGas
	}
	suppliedGas -= gasCost
	output, err := p.Run(input)
	return output, suppliedGas, err
}

type depositroot struct{}

const (
	depositPublicKeyLength           = pqcrypto.MLDSA87PublicKeyLength
	depositWithdrawalRecipientLength = common.AddressLength
	depositAmountLength              = 8
	depositSignatureLength           = pqcrypto.MLDSA87SignatureLength
	depositPublicKeyOffset           = 0
	depositWithdrawalRecipientOffset = depositPublicKeyOffset + depositPublicKeyLength
	depositAmountOffset              = depositWithdrawalRecipientOffset + depositWithdrawalRecipientLength
	depositSignatureOffset           = depositAmountOffset + depositAmountLength
)

func (c *depositroot) RequiredGas(input []byte) uint64 {
	return params.DepositrootGas
}

func (c *depositroot) Run(input []byte) ([]byte, error) {
	var (
		pkBytes                  = getData(input, depositPublicKeyOffset, depositPublicKeyLength)
		withdrawalRecipientBytes = getData(input, depositWithdrawalRecipientOffset, depositWithdrawalRecipientLength)
		amountBytes              = getData(input, depositAmountOffset, depositAmountLength)
		sigBytes                 = getData(input, depositSignatureOffset, depositSignatureLength)
	)

	var amountUint uint64
	buf := bytes.NewReader(amountBytes)
	err := binary.Read(buf, binary.LittleEndian, &amountUint)
	if err != nil {
		return nil, err
	}

	data := &depositdata{
		PublicKey:           pkBytes,
		WithdrawalRecipient: withdrawalRecipientBytes,
		Amount:              amountUint,
		Signature:           sigBytes,
	}
	h, err := data.HashTreeRoot()
	if err != nil {
		return nil, pkgerrors.Wrap(err, "could not hash tree root deposit data item")
	}

	return h[:], nil
}

const (
	// mldsa87VerifyDigestLength is the width of the message representative:
	// one QRVM word, matching the SHAKE256 precompile output. Earlier builds
	// read common.HashLength (32 bytes); the 64-byte width is the interface
	// ratified for the next testnet release.
	mldsa87VerifyDigestLength        = WordBytes
	mldsa87VerifyDigestOffset        = 0
	mldsa87VerifyPublicKeyOffset     = mldsa87VerifyDigestOffset + mldsa87VerifyDigestLength
	mldsa87VerifySignatureOffset     = mldsa87VerifyPublicKeyOffset + cryptomldsa87.CRYPTO_PUBLIC_KEY_BYTES
	mldsa87VerifyContextLengthOffset = mldsa87VerifySignatureOffset + cryptomldsa87.CRYPTO_BYTES
	mldsa87VerifyContextOffset       = mldsa87VerifyContextLengthOffset + 1
	mldsa87VerifyMinInputLength      = mldsa87VerifyContextOffset
	mldsa87VerifyMaxContextLength    = 255
)

// mldsa87Verify verifies an ML-DSA-87 signature over a fixed 64-byte message
// representative using the supplied public key and context.
type mldsa87Verify struct{}

// mldsa87VerifyLegacy32 preserves the 32-byte slot 3 interface used before the
// QRL 2.0 post-quantum precompile activation.
type mldsa87VerifyLegacy32 struct{}

func (*mldsa87Verify) RequiredGas([]byte) uint64 {
	return params.MLDSA87VerifyGas
}

func (*mldsa87VerifyLegacy32) RequiredGas([]byte) uint64 {
	return params.MLDSA87VerifyGas
}

func (*mldsa87Verify) Run(input []byte) ([]byte, error) {
	return runMLDSA87Verification(input, mldsa87VerifyDigestLength)
}

func (*mldsa87VerifyLegacy32) Run(input []byte) ([]byte, error) {
	return runMLDSA87Verification(input, common.HashLength)
}

func runMLDSA87Verification(input []byte, digestLength int) ([]byte, error) {
	publicKeyOffset := digestLength
	signatureOffset := publicKeyOffset + cryptomldsa87.CRYPTO_PUBLIC_KEY_BYTES
	contextLengthOffset := signatureOffset + cryptomldsa87.CRYPTO_BYTES
	contextOffset := contextLengthOffset + 1
	if len(input) < contextOffset {
		return nil, nil
	}

	context := input[contextOffset:]
	if len(context) != int(input[contextLengthOffset]) {
		return nil, nil
	}

	digest := input[:publicKeyOffset]
	publicKeyBytes := input[publicKeyOffset:signatureOffset]
	publicKey := (*[cryptomldsa87.CRYPTO_PUBLIC_KEY_BYTES]byte)(publicKeyBytes)
	signatureBytes := input[signatureOffset:contextLengthOffset]
	signature := [cryptomldsa87.CRYPTO_BYTES]byte(signatureBytes)

	if !cryptomldsa87.Verify(context, digest, signature, publicKey) {
		return nil, nil
	}

	return common.CopyBytes(trueWord), nil
}

type depositdata struct {
	PublicKey           []byte
	WithdrawalRecipient []byte
	Amount              uint64
	Signature           []byte
}

// HashTreeRoot ssz hashes the Deposit_Data object
func (d *depositdata) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(d)
}

// HashTreeRootWith ssz hashes the Deposit_Data object with a hasher
func (d *depositdata) HashTreeRootWith(hh *ssz.Hasher) (err error) {
	indx := hh.Index()

	// Field (0) 'Pubkey'
	if size := len(d.PublicKey); size != depositPublicKeyLength {
		err = ssz.ErrBytesLengthFn("--.Pubkey", size, depositPublicKeyLength)
		return
	}
	hh.PutBytes(d.PublicKey)

	// Field (1) 'WithdrawalRecipient'
	if size := len(d.WithdrawalRecipient); size != depositWithdrawalRecipientLength {
		err = ssz.ErrBytesLengthFn("--.WithdrawalRecipient", size, depositWithdrawalRecipientLength)
		return
	}
	hh.PutBytes(d.WithdrawalRecipient)

	// Field (2) 'Amount'
	hh.PutUint64(d.Amount)

	// Field (3) 'Signature'
	if size := len(d.Signature); size != depositSignatureLength {
		err = ssz.ErrBytesLengthFn("--.Signature", size, depositSignatureLength)
		return
	}
	hh.PutBytes(d.Signature)

	if ssz.EnableVectorizedHTR {
		hh.MerkleizeVectorizedHTR(indx)
	} else {
		hh.Merkleize(indx)
	}
	return
}

// SHA256 implemented as a native contract.
type sha256hash struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
//
// This method does not require any overflow checking as the input size gas costs
// required for anything significant is so high it's impossible to pay for.
func (c *sha256hash) RequiredGas(input []byte) uint64 {
	return toWordSize(uint64(len(input)))*params.Sha256PerWordGas + params.Sha256BaseGas
}
func (c *sha256hash) Run(input []byte) ([]byte, error) {
	h := sha256.Sum256(input)
	return h[:], nil
}

// shake256hash implements SHAKE256 with a fixed 512-bit output.
type shake256hash struct{}

// RequiredGas returns the gas required to execute the precompiled contract.
func (*shake256hash) RequiredGas(input []byte) uint64 {
	return shake256Gas(uint64(len(input)))
}

func shake256Gas(inputLength uint64) uint64 {
	words := toWordSize(inputLength)
	if words > (gomath.MaxUint64-params.Shake256BaseGas)/params.Shake256PerWordGas {
		return gomath.MaxUint64
	}
	return words*params.Shake256PerWordGas + params.Shake256BaseGas
}

func (*shake256hash) Run(input []byte) ([]byte, error) {
	return sha3.SumSHAKE256(input, 64), nil
}

// data copy implemented as a native contract.
type dataCopy struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
//
// This method does not require any overflow checking as the input size gas costs
// required for anything significant is so high it's impossible to pay for.
func (c *dataCopy) RequiredGas(input []byte) uint64 {
	return toWordSize(uint64(len(input)))*params.IdentityPerWordGas + params.IdentityBaseGas
}
func (c *dataCopy) Run(in []byte) ([]byte, error) {
	return common.CopyBytes(in), nil
}

// bigModExp implements a native big integer exponential modular operation.
type bigModExp struct{}

var (
	big0  = big.NewInt(0)
	big1  = big.NewInt(1)
	big3  = big.NewInt(3)
	big7  = big.NewInt(7)
	big8  = big.NewInt(8)
	big32 = big.NewInt(32)
)

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (c *bigModExp) RequiredGas(input []byte) uint64 {
	var (
		baseLen = new(big.Int).SetBytes(getData(input, 0, 32))
		expLen  = new(big.Int).SetBytes(getData(input, 32, 32))
		modLen  = new(big.Int).SetBytes(getData(input, 64, 32))
	)
	if len(input) > 96 {
		input = input[96:]
	} else {
		input = input[:0]
	}
	// Retrieve the head 32 bytes of exp for the adjusted exponent length
	var expHead *big.Int
	if big.NewInt(int64(len(input))).Cmp(baseLen) <= 0 {
		expHead = new(big.Int)
	} else {
		if expLen.Cmp(big32) > 0 {
			expHead = new(big.Int).SetBytes(getData(input, baseLen.Uint64(), 32))
		} else {
			expHead = new(big.Int).SetBytes(getData(input, baseLen.Uint64(), expLen.Uint64()))
		}
	}
	// Calculate the adjusted exponent length
	var msb int
	if bitlen := expHead.BitLen(); bitlen > 0 {
		msb = bitlen - 1
	}
	adjExpLen := new(big.Int)
	if expLen.Cmp(big32) > 0 {
		adjExpLen.Sub(expLen, big32)
		adjExpLen.Mul(big8, adjExpLen)
	}
	adjExpLen.Add(adjExpLen, big.NewInt(int64(msb)))
	// Calculate the gas cost of the operation
	gas := new(big.Int).Set(math.BigMax(modLen, baseLen))

	// EIP-2565 has three changes
	// 1. Different multComplexity (inlined here)
	// in EIP-2565 (https://eips.ethereum.org/EIPS/eip-2565):
	//
	// def mult_complexity(x):
	//    ceiling(x/8)^2
	//
	//where is x is max(length_of_MODULUS, length_of_BASE)
	gas = gas.Add(gas, big7)
	gas = gas.Div(gas, big8)
	gas.Mul(gas, gas)

	gas.Mul(gas, math.BigMax(adjExpLen, big1))
	// 2. Different divisor (`GQUADDIVISOR`) (3)
	gas.Div(gas, big3)
	if gas.BitLen() > 64 {
		return gomath.MaxUint64
	}
	// 3. Minimum price of 200 gas
	if gas.Uint64() < 200 {
		return 200
	}
	return gas.Uint64()
}

func (c *bigModExp) Run(input []byte) ([]byte, error) {
	var (
		baseLen = new(big.Int).SetBytes(getData(input, 0, 32)).Uint64()
		expLen  = new(big.Int).SetBytes(getData(input, 32, 32)).Uint64()
		modLen  = new(big.Int).SetBytes(getData(input, 64, 32)).Uint64()
	)
	if len(input) > 96 {
		input = input[96:]
	} else {
		input = input[:0]
	}
	// Handle a special case when both the base and mod length is zero
	if baseLen == 0 && modLen == 0 {
		return []byte{}, nil
	}
	// Retrieve the operands and execute the exponentiation
	var (
		base = new(big.Int).SetBytes(getData(input, 0, baseLen))
		exp  = new(big.Int).SetBytes(getData(input, baseLen, expLen))
		mod  = new(big.Int).SetBytes(getData(input, baseLen+expLen, modLen))
		v    []byte
	)
	switch {
	case mod.BitLen() == 0:
		// Modulo 0 is undefined, return zero
		return common.LeftPadBytes([]byte{}, int(modLen)), nil
	case base.BitLen() == 1: // a bit length of 1 means it's 1 (or -1).
		//If base == 1, then we can just return base % mod (if mod >= 1, which it is)
		v = base.Mod(base, mod).Bytes()
	default:
		v = base.Exp(base, exp, mod).Bytes()
	}
	return common.LeftPadBytes(v, int(modLen)), nil
}
