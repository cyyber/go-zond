// Copyright 2026 The go-qrl Authors
// This file is part of the go-qrl library.

package clef

import (
	"math/big"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/go-qrl/common/math"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/signer/core/apitypes"
)

const (
	expectedText          = "Clef VM64 signData"
	rejectedText          = "Clef rule rejection"
	expectedValidatorText = "Clef validator-bound data"
	expectedRecipient     = "Qd5812f6cf4a0f645aa620cd57319a0ed649dd8f5519a9dde7770ae5b0e49e547985f35eb972a2a07041561aa39c65a3991478f9b1e6749e05277dcf58a9a8b72"
	expectedTypedName     = "Local Testnet VM64"
	expectedTypedVersion  = "1"
	expectedTypedContents = "Clef VM64 typed data"
	expectedTypedValue    = "340282366920938463463374607431768211457"
	expectedTxInputHex    = "0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f"
	expectedGas           = uint64(40000)
	expectedValue         = int64(42)
	rejectedValue         = int64(43)

	rulesSource = `function big(value) {
    if (value.slice(0, 2) == '0x') {
        return new BigNumber(value.slice(2), 16);
    }
    return new BigNumber(value);
}
function ApproveListing(req) { return 'Approve'; }
function ApproveSignData(req) {
    if (req.messages[0].value.indexOf('Clef rule rejection') >= 0) {
        return 'Reject';
    }
    return 'Approve';
}
function ApproveTx(req) {
    if (big(req.transaction.value).eq(43)) {
        return 'Reject';
    }
    return 'Approve';
}
`
)

func expectedTypedData(account common.Address, chainID *big.Int) apitypes.TypedData {
	typedChainID := math.HexOrDecimal256(*new(big.Int).Set(chainID))
	return apitypes.TypedData{
		Types: apitypes.Types{
			"QRLTypedDataDomain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Message": {
				{Name: "sender", Type: "address"},
				{Name: "contents", Type: "string"},
				{Name: "value", Type: "uint256"},
			},
		},
		PrimaryType: "Message",
		Domain: apitypes.TypedDataDomain{
			Name:              expectedTypedName,
			Version:           expectedTypedVersion,
			ChainId:           &typedChainID,
			VerifyingContract: account.Hex(),
		},
		Message: apitypes.TypedDataMessage{
			"sender":   account.Hex(),
			"contents": expectedTypedContents,
			"value":    expectedTypedValue,
		},
	}
}

func transactionArgs(
	account common.Address,
	chainID *big.Int,
	nonce uint64,
	tip *big.Int,
	feeCap *big.Int,
) apitypes.SendTxArgs {
	recipient := common.MustParseAddress(expectedRecipient)
	input := hexutil.MustDecode(expectedTxInputHex)
	from := common.NewMixedcaseAddress(account)
	to := common.NewMixedcaseAddress(recipient)
	tipValue := hexutil.Big(*new(big.Int).Set(tip))
	feeCapValue := hexutil.Big(*new(big.Int).Set(feeCap))
	value := hexutil.Big(*big.NewInt(expectedValue))
	chainIDValue := hexutil.Big(*new(big.Int).Set(chainID))
	data := hexutil.Bytes(input)
	accessList := types.AccessList{}
	return apitypes.SendTxArgs{
		From:                 from,
		To:                   &to,
		Gas:                  hexutil.Uint64(expectedGas),
		MaxFeePerGas:         &feeCapValue,
		MaxPriorityFeePerGas: &tipValue,
		Value:                value,
		Nonce:                hexutil.Uint64(nonce),
		Input:                &data,
		AccessList:           &accessList,
		ChainID:              &chainIDValue,
	}
}
