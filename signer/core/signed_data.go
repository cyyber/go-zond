// Copyright 2019 The go-ethereum Authors
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

package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"

	"github.com/theQRL/go-zond/accounts"
	"github.com/theQRL/go-zond/common"
	"github.com/theQRL/go-zond/common/hexutil"
	"github.com/theQRL/go-zond/crypto"
	"github.com/theQRL/go-zond/signer/core/apitypes"
)

// sign receives a request and produces a signature
func (api *SignerAPI) sign(req *SignDataRequest) (hexutil.Bytes, error) {
	// We make the request prior to looking up if we actually have the account, to prevent
	// account-enumeration via the API
	res, err := api.UI.ApproveSignData(req)
	if err != nil {
		return nil, err
	}
	if !res.Approved {
		return nil, ErrRequestDenied
	}
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: req.Address.Address()}
	wallet, err := api.am.Find(account)
	if err != nil {
		return nil, err
	}
	pw, err := api.lookupOrQueryPassword(account.Address,
		"Password for signing",
		fmt.Sprintf("Please enter password for signing data with account %s", account.Address.Hex()))
	if err != nil {
		return nil, err
	}
	// Sign the data with the wallet
	signature, err := wallet.SignDataWithPassphrase(account, pw, req.ContentType, req.Rawdata)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

// SignData signs the hash of the provided data, but does so differently
// depending on the content-type specified.
//
// Different types of validation occur.
func (api *SignerAPI) SignData(ctx context.Context, contentType string, addr common.MixedcaseAddress, data interface{}) (hexutil.Bytes, error) {
	var req, err = api.determineSignatureFormat(ctx, contentType, addr, data)
	if err != nil {
		return nil, err
	}
	signature, err := api.sign(req)
	if err != nil {
		api.UI.ShowError(err.Error())
		return nil, err
	}
	return signature, nil
}

// determineSignatureFormat determines which signature method should be used based upon the mime type
// In the cases where it matters ensure that the charset is handled. The charset
// resides in the 'params' returned as the second returnvalue from mime.ParseMediaType
// charset, ok := params["charset"]
// As it is now, we accept any charset and just treat it as 'raw'.
// This method returns the mimetype for signing along with the request
func (api *SignerAPI) determineSignatureFormat(ctx context.Context, contentType string, addr common.MixedcaseAddress, data interface{}) (*SignDataRequest, error) {
	var req *SignDataRequest

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, err
	}

	switch mediaType {
	case apitypes.IntendedValidator.Mime:
		// Data with an intended validator
		validatorData, err := UnmarshalValidatorData(data)
		if err != nil {
			return nil, err
		}
		sighash, msg := SignTextValidator(validatorData)
		messages := []*apitypes.NameValueType{
			{
				Name:  "This is a request to sign data intended for a particular validator (see EIP 191 version 0)",
				Typ:   "description",
				Value: "",
			},
			{
				Name:  "Intended validator address",
				Typ:   "address",
				Value: validatorData.Address.String(),
			},
			{
				Name:  "Application-specific data",
				Typ:   "hexdata",
				Value: validatorData.Message,
			},
			{
				Name:  "Full message for signing",
				Typ:   "hexdata",
				Value: fmt.Sprintf("%#x", msg),
			},
		}
		req = &SignDataRequest{ContentType: mediaType, Rawdata: []byte(msg), Messages: messages, Hash: sighash}
	case apitypes.DataTyped.Mime:
		// EIP-712 conformant typed data
		var err error
		req, err = typedDataRequest(data)
		if err != nil {
			return nil, err
		}
	default: // also case TextPlain.Mime:
		// Calculates a Zond Dilithium signature for:
		// hash = keccak256("\x19Zond Signed Message:\n${message length}${message}")
		// We expect input to be a hex-encoded string
		textData, err := fromHex(data)
		if err != nil {
			return nil, err
		}
		sighash, msg := accounts.TextAndHash(textData)
		messages := []*apitypes.NameValueType{
			{
				Name:  "message",
				Typ:   accounts.MimetypeTextPlain,
				Value: msg,
			},
		}
		req = &SignDataRequest{ContentType: mediaType, Rawdata: []byte(msg), Messages: messages, Hash: sighash}
	}
	req.Address = addr
	req.Meta = MetadataFromContext(ctx)
	return req, nil
}

// SignTextValidator signs the given message which can be further recovered
// with the given validator.
// hash = keccak256("\x19\x00"${address}${data}).
func SignTextValidator(validatorData apitypes.ValidatorData) (hexutil.Bytes, string) {
	msg := fmt.Sprintf("\x19\x00%s%s", string(validatorData.Address.Bytes()), string(validatorData.Message))
	return crypto.Keccak256([]byte(msg)), msg
}

// SignTypedData signs EIP-712 conformant typed data
// hash = keccak256("\x19${byteVersion}${domainSeparator}${hashStruct(message)}")
// It returns
// - the signature,
// - and/or any error
func (api *SignerAPI) SignTypedData(ctx context.Context, addr common.MixedcaseAddress, typedData apitypes.TypedData) (hexutil.Bytes, error) {
	signature, _, err := api.signTypedData(ctx, addr, typedData, nil)
	return signature, err
}

// signTypedData is identical to the capitalized version, except that it also returns the hash (preimage)
// - the signature preimage (hash)
func (api *SignerAPI) signTypedData(ctx context.Context, addr common.MixedcaseAddress,
	typedData apitypes.TypedData, validationMessages *apitypes.ValidationMessages) (hexutil.Bytes, hexutil.Bytes, error) {
	req, err := typedDataRequest(typedData)
	if err != nil {
		return nil, nil, err
	}
	req.Address = addr
	req.Meta = MetadataFromContext(ctx)
	if validationMessages != nil {
		req.Callinfo = validationMessages.Messages
	}
	signature, err := api.sign(req)
	if err != nil {
		api.UI.ShowError(err.Error())
		return nil, nil, err
	}
	return signature, req.Hash, nil
}

// fromHex tries to interpret the data as type string, and convert from
// hexadecimal to []byte
func fromHex(data any) ([]byte, error) {
	if stringData, ok := data.(string); ok {
		binary, err := hexutil.Decode(stringData)
		return binary, err
	}
	return nil, fmt.Errorf("wrong type %T", data)
}

// typeDataRequest tries to convert the data into a SignDataRequest.
func typedDataRequest(data any) (*SignDataRequest, error) {
	var typedData apitypes.TypedData
	if td, ok := data.(apitypes.TypedData); ok {
		typedData = td
	} else { // Hex-encoded data
		jsonData, err := fromHex(data)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(jsonData, &typedData); err != nil {
			return nil, err
		}
	}
	messages, err := typedData.Format()
	if err != nil {
		return nil, err
	}
	sighash, rawData, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, err
	}
	return &SignDataRequest{
		ContentType: apitypes.DataTyped.Mime,
		Rawdata:     []byte(rawData),
		Messages:    messages,
		Hash:        sighash}, nil
}

// UnmarshalValidatorData converts the bytes input to typed data
func UnmarshalValidatorData(data interface{}) (apitypes.ValidatorData, error) {
	raw, ok := data.(map[string]interface{})
	if !ok {
		return apitypes.ValidatorData{}, errors.New("validator input is not a map[string]interface{}")
	}
	addrBytes, err := fromHex(raw["address"])
	if err != nil {
		return apitypes.ValidatorData{}, fmt.Errorf("validator address error: %w", err)
	}
	if len(addrBytes) == 0 {
		return apitypes.ValidatorData{}, errors.New("validator address is undefined")
	}
	messageBytes, err := fromHex(raw["message"])
	if err != nil {
		return apitypes.ValidatorData{}, fmt.Errorf("message error: %w", err)
	}
	if len(messageBytes) == 0 {
		return apitypes.ValidatorData{}, errors.New("message is undefined")
	}
	return apitypes.ValidatorData{
		Address: common.BytesToAddress(addrBytes),
		Message: messageBytes,
	}, nil
}
