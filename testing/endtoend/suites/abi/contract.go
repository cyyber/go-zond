// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package abi

import (
	"errors"
	"math/big"
	"strings"

	qrl "github.com/theQRL/go-qrl"
	"github.com/theQRL/go-qrl/accounts/abi"
	"github.com/theQRL/go-qrl/accounts/abi/bind"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = qrl.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// EventEmitterBoundaryEdges is an auto generated low-level Go binding around an user-defined struct.
type EventEmitterBoundaryEdges struct {
	Unsigned248  *big.Int
	Signed248    *big.Int
	Unsigned256  *big.Int
	Signed256    *big.Int
	Unsigned264  *big.Int
	Signed264    *big.Int
	Unsigned504  *big.Int
	Signed504    *big.Int
	Unsigned512  *big.Int
	Signed512    *big.Int
	Bytes31Value [31]byte
	Bytes32Value [32]byte
	Bytes33Value [33]byte
	Bytes63Value [63]byte
	Bytes64Value [64]byte
}

// EventEmitterDynamicRecord is an auto generated low-level Go binding around an user-defined struct.
type EventEmitterDynamicRecord struct {
	Amount  *big.Int
	Note    string
	Payload []byte
	Values  [][]uint16
}

// EventEmitterFunctionRecord is an auto generated low-level Go binding around an user-defined struct.
type EventEmitterFunctionRecord struct {
	Callback [68]byte
	Note     string
}

// EventEmitterNestedRecord is an auto generated low-level Go binding around an user-defined struct.
type EventEmitterNestedRecord struct {
	FixedRecord   EventEmitterRecord
	DynamicRecord EventEmitterDynamicRecord
	Extra         []byte
}

// EventEmitterRecord is an auto generated low-level Go binding around an user-defined struct.
type EventEmitterRecord struct {
	Amount    *big.Int
	Recipient common.Address
	Tag       [64]byte
}

// EventEmitterMetaData contains all meta data concerning the EventEmitter contract.
var EventEmitterMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint512\",\"name\":\"initial\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"bytes64\",\"name\":\"tag\",\"type\":\"bytes64\"}],\"internalType\":\"structEventEmitter.Record\",\"name\":\"record\",\"type\":\"tuple\"},{\"internalType\":\"uint16[]\",\"name\":\"numbers\",\"type\":\"uint16[]\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint512\",\"name\":\"code\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"bytes64\",\"name\":\"tag\",\"type\":\"bytes64\"}],\"internalType\":\"structEventEmitter.Record\",\"name\":\"record\",\"type\":\"tuple\"},{\"internalType\":\"uint16[][]\",\"name\":\"nested\",\"type\":\"uint16[][]\"}],\"name\":\"ComplexFailure\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint16[][]\",\"name\":\"values\",\"type\":\"uint16[][]\"}],\"indexed\":false,\"internalType\":\"structEventEmitter.DynamicRecord\",\"name\":\"record\",\"type\":\"tuple\"},{\"indexed\":false,\"internalType\":\"uint16[3]\",\"name\":\"fixedNumbers\",\"type\":\"uint16[3]\"},{\"indexed\":false,\"internalType\":\"string[2]\",\"name\":\"fixedStrings\",\"type\":\"string[2]\"},{\"indexed\":false,\"internalType\":\"uint16[][2]\",\"name\":\"mixed\",\"type\":\"uint16[][2]\"}],\"name\":\"Composite\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint512\",\"name\":\"value\",\"type\":\"uint512\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"bytes64\",\"name\":\"tag\",\"type\":\"bytes64\"}],\"indexed\":false,\"internalType\":\"structEventEmitter.Record\",\"name\":\"record\",\"type\":\"tuple\"},{\"indexed\":false,\"internalType\":\"uint16[]\",\"name\":\"numbers\",\"type\":\"uint16[]\"}],\"name\":\"Deployed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"}],\"name\":\"Dynamic\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"FallbackCalled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"function(uint512)pureexternalreturns(uint512)\",\"name\":\"indexedCallback\",\"type\":\"function\"},{\"indexed\":false,\"internalType\":\"function(uint512)pureexternalreturns(uint512)\",\"name\":\"callback\",\"type\":\"function\"},{\"indexed\":false,\"internalType\":\"uint512\",\"name\":\"result\",\"type\":\"uint512\"}],\"name\":\"FunctionObserved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bool\",\"name\":\"flag\",\"type\":\"bool\"},{\"indexed\":true,\"internalType\":\"bytes5\",\"name\":\"code\",\"type\":\"bytes5\"},{\"indexed\":true,\"internalType\":\"int16\",\"name\":\"delta\",\"type\":\"int16\"}],\"name\":\"IndexedScalars\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint16\",\"name\":\"marker\",\"type\":\"uint16\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Paid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Received\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"indexed\":true,\"internalType\":\"int512\",\"name\":\"delta\",\"type\":\"int512\"},{\"indexed\":false,\"internalType\":\"bytes64\",\"name\":\"tag\",\"type\":\"bytes64\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"enabled\",\"type\":\"bool\"}],\"name\":\"Stored\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"value\",\"type\":\"uint16\"}],\"name\":\"Transformed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"}],\"name\":\"Transformed\",\"type\":\"event\"},{\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"inputs\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"int512\",\"name\":\"delta\",\"type\":\"int512\"},{\"internalType\":\"bytes64\",\"name\":\"tag\",\"type\":\"bytes64\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bool\",\"name\":\"enabled\",\"type\":\"bool\"}],\"name\":\"echo\",\"outputs\":[{\"internalType\":\"uint512\",\"name\":\"\",\"type\":\"uint512\"},{\"internalType\":\"int512\",\"name\":\"\",\"type\":\"int512\"},{\"internalType\":\"bytes64\",\"name\":\"\",\"type\":\"bytes64\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"},{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"smallUnsigned\",\"type\":\"uint8\"},{\"internalType\":\"int8\",\"name\":\"smallSigned\",\"type\":\"int8\"},{\"internalType\":\"uint256\",\"name\":\"wideUnsigned\",\"type\":\"uint256\"},{\"internalType\":\"int256\",\"name\":\"wideSigned\",\"type\":\"int256\"},{\"internalType\":\"bytes5\",\"name\":\"shortBytes\",\"type\":\"bytes5\"},{\"internalType\":\"uint16[3]\",\"name\":\"fixedNumbers\",\"type\":\"uint16[3]\"},{\"internalType\":\"string[2]\",\"name\":\"fixedStrings\",\"type\":\"string[2]\"},{\"internalType\":\"uint16[][2]\",\"name\":\"mixed\",\"type\":\"uint16[][2]\"}],\"name\":\"echoBoundaries\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"},{\"internalType\":\"int8\",\"name\":\"\",\"type\":\"int8\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"int256\",\"name\":\"\",\"type\":\"int256\"},{\"internalType\":\"bytes5\",\"name\":\"\",\"type\":\"bytes5\"},{\"internalType\":\"uint16[3]\",\"name\":\"\",\"type\":\"uint16[3]\"},{\"internalType\":\"string[2]\",\"name\":\"\",\"type\":\"string[2]\"},{\"internalType\":\"uint16[][2]\",\"name\":\"\",\"type\":\"uint16[][2]\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint248\",\"name\":\"unsigned248\",\"type\":\"uint248\"},{\"internalType\":\"int248\",\"name\":\"signed248\",\"type\":\"int248\"},{\"internalType\":\"uint256\",\"name\":\"unsigned256\",\"type\":\"uint256\"},{\"internalType\":\"int256\",\"name\":\"signed256\",\"type\":\"int256\"},{\"internalType\":\"uint264\",\"name\":\"unsigned264\",\"type\":\"uint264\"},{\"internalType\":\"int264\",\"name\":\"signed264\",\"type\":\"int264\"},{\"internalType\":\"uint504\",\"name\":\"unsigned504\",\"type\":\"uint504\"},{\"internalType\":\"int504\",\"name\":\"signed504\",\"type\":\"int504\"},{\"internalType\":\"uint512\",\"name\":\"unsigned512\",\"type\":\"uint512\"},{\"internalType\":\"int512\",\"name\":\"signed512\",\"type\":\"int512\"},{\"internalType\":\"bytes31\",\"name\":\"bytes31Value\",\"type\":\"bytes31\"},{\"internalType\":\"bytes32\",\"name\":\"bytes32Value\",\"type\":\"bytes32\"},{\"internalType\":\"bytes33\",\"name\":\"bytes33Value\",\"type\":\"bytes33\"},{\"internalType\":\"bytes63\",\"name\":\"bytes63Value\",\"type\":\"bytes63\"},{\"internalType\":\"bytes64\",\"name\":\"bytes64Value\",\"type\":\"bytes64\"}],\"internalType\":\"structEventEmitter.BoundaryEdges\",\"name\":\"edges\",\"type\":\"tuple\"}],\"name\":\"echoBoundaryEdges\",\"outputs\":[{\"components\":[{\"internalType\":\"uint248\",\"name\":\"unsigned248\",\"type\":\"uint248\"},{\"internalType\":\"int248\",\"name\":\"signed248\",\"type\":\"int248\"},{\"internalType\":\"uint256\",\"name\":\"unsigned256\",\"type\":\"uint256\"},{\"internalType\":\"int256\",\"name\":\"signed256\",\"type\":\"int256\"},{\"internalType\":\"uint264\",\"name\":\"unsigned264\",\"type\":\"uint264\"},{\"internalType\":\"int264\",\"name\":\"signed264\",\"type\":\"int264\"},{\"internalType\":\"uint504\",\"name\":\"unsigned504\",\"type\":\"uint504\"},{\"internalType\":\"int504\",\"name\":\"signed504\",\"type\":\"int504\"},{\"internalType\":\"uint512\",\"name\":\"unsigned512\",\"type\":\"uint512\"},{\"internalType\":\"int512\",\"name\":\"signed512\",\"type\":\"int512\"},{\"internalType\":\"bytes31\",\"name\":\"bytes31Value\",\"type\":\"bytes31\"},{\"internalType\":\"bytes32\",\"name\":\"bytes32Value\",\"type\":\"bytes32\"},{\"internalType\":\"bytes33\",\"name\":\"bytes33Value\",\"type\":\"bytes33\"},{\"internalType\":\"bytes63\",\"name\":\"bytes63Value\",\"type\":\"bytes63\"},{\"internalType\":\"bytes64\",\"name\":\"bytes64Value\",\"type\":\"bytes64\"}],\"internalType\":\"structEventEmitter.BoundaryEdges\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16[2][2]\",\"name\":\"fixedMatrix\",\"type\":\"uint16[2][2]\"},{\"internalType\":\"uint16[2][]\",\"name\":\"rows\",\"type\":\"uint16[2][]\"},{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint16[][]\",\"name\":\"values\",\"type\":\"uint16[][]\"}],\"internalType\":\"structEventEmitter.DynamicRecord[2]\",\"name\":\"records\",\"type\":\"tuple[2]\"},{\"components\":[{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"bytes64\",\"name\":\"tag\",\"type\":\"bytes64\"}],\"internalType\":\"structEventEmitter.Record\",\"name\":\"fixedRecord\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint16[][]\",\"name\":\"values\",\"type\":\"uint16[][]\"}],\"internalType\":\"structEventEmitter.DynamicRecord\",\"name\":\"dynamicRecord\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"extra\",\"type\":\"bytes\"}],\"internalType\":\"structEventEmitter.NestedRecord\",\"name\":\"nested\",\"type\":\"tuple\"}],\"name\":\"echoCompositeContainers\",\"outputs\":[{\"internalType\":\"uint16[2][2]\",\"name\":\"\",\"type\":\"uint16[2][2]\"},{\"internalType\":\"uint16[2][]\",\"name\":\"\",\"type\":\"uint16[2][]\"},{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint16[][]\",\"name\":\"values\",\"type\":\"uint16[][]\"}],\"internalType\":\"structEventEmitter.DynamicRecord[2]\",\"name\":\"\",\"type\":\"tuple[2]\"},{\"components\":[{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"bytes64\",\"name\":\"tag\",\"type\":\"bytes64\"}],\"internalType\":\"structEventEmitter.Record\",\"name\":\"fixedRecord\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint16[][]\",\"name\":\"values\",\"type\":\"uint16[][]\"}],\"internalType\":\"structEventEmitter.DynamicRecord\",\"name\":\"dynamicRecord\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"extra\",\"type\":\"bytes\"}],\"internalType\":\"structEventEmitter.NestedRecord\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes[2]\",\"name\":\"fixedBytes\",\"type\":\"bytes[2]\"},{\"internalType\":\"bytes[]\",\"name\":\"byteSlices\",\"type\":\"bytes[]\"},{\"internalType\":\"string[]\",\"name\":\"strings\",\"type\":\"string[]\"}],\"name\":\"echoDynamicContainers\",\"outputs\":[{\"internalType\":\"bytes[2]\",\"name\":\"\",\"type\":\"bytes[2]\"},{\"internalType\":\"bytes[]\",\"name\":\"\",\"type\":\"bytes[]\"},{\"internalType\":\"string[]\",\"name\":\"\",\"type\":\"string[]\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"function(uint512)pureexternalreturns(uint512)\",\"name\":\"callback\",\"type\":\"function\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"function(uint512)pureexternalreturns(uint512)[2]\",\"name\":\"fixedCallbacks\",\"type\":\"function[2]\"},{\"internalType\":\"function(uint512)pureexternalreturns(uint512)[]\",\"name\":\"callbacks\",\"type\":\"function[]\"},{\"components\":[{\"internalType\":\"function(uint512)pureexternalreturns(uint512)\",\"name\":\"callback\",\"type\":\"function\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"}],\"internalType\":\"structEventEmitter.FunctionRecord\",\"name\":\"record\",\"type\":\"tuple\"}],\"name\":\"echoFunctions\",\"outputs\":[{\"internalType\":\"function(uint512)pureexternalreturns(uint512)\",\"name\":\"\",\"type\":\"function\"},{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"function(uint512)pureexternalreturns(uint512)[2]\",\"name\":\"\",\"type\":\"function[2]\"},{\"internalType\":\"function(uint512)pureexternalreturns(uint512)[]\",\"name\":\"\",\"type\":\"function[]\"},{\"components\":[{\"internalType\":\"function(uint512)pureexternalreturns(uint512)\",\"name\":\"callback\",\"type\":\"function\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"}],\"internalType\":\"structEventEmitter.FunctionRecord\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[2]\",\"name\":\"fixedAddresses\",\"type\":\"address[2]\"},{\"internalType\":\"address[]\",\"name\":\"addresses\",\"type\":\"address[]\"},{\"internalType\":\"bytes64[2]\",\"name\":\"fixedTags\",\"type\":\"bytes64[2]\"},{\"internalType\":\"bytes64[]\",\"name\":\"tags\",\"type\":\"bytes64[]\"}],\"name\":\"echoLeafContainers\",\"outputs\":[{\"internalType\":\"address[2]\",\"name\":\"\",\"type\":\"address[2]\"},{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"},{\"internalType\":\"bytes64[2]\",\"name\":\"\",\"type\":\"bytes64[2]\"},{\"internalType\":\"bytes64[]\",\"name\":\"\",\"type\":\"bytes64[]\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint16[][]\",\"name\":\"values\",\"type\":\"uint16[][]\"}],\"internalType\":\"structEventEmitter.DynamicRecord\",\"name\":\"record\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint16[][]\",\"name\":\"values\",\"type\":\"uint16[][]\"}],\"internalType\":\"structEventEmitter.DynamicRecord[]\",\"name\":\"records\",\"type\":\"tuple[]\"},{\"internalType\":\"uint16[][][]\",\"name\":\"cube\",\"type\":\"uint16[][][]\"}],\"name\":\"echoNested\",\"outputs\":[{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint16[][]\",\"name\":\"values\",\"type\":\"uint16[][]\"}],\"internalType\":\"structEventEmitter.DynamicRecord\",\"name\":\"\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint16[][]\",\"name\":\"values\",\"type\":\"uint16[][]\"}],\"internalType\":\"structEventEmitter.DynamicRecord[]\",\"name\":\"\",\"type\":\"tuple[]\"},{\"internalType\":\"uint16[][][]\",\"name\":\"\",\"type\":\"uint16[][][]\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint16[][]\",\"name\":\"values\",\"type\":\"uint16[][]\"}],\"internalType\":\"structEventEmitter.DynamicRecord\",\"name\":\"record\",\"type\":\"tuple\"},{\"internalType\":\"uint16[3]\",\"name\":\"fixedNumbers\",\"type\":\"uint16[3]\"},{\"internalType\":\"string[2]\",\"name\":\"fixedStrings\",\"type\":\"string[2]\"},{\"internalType\":\"uint16[][2]\",\"name\":\"mixed\",\"type\":\"uint16[][2]\"}],\"name\":\"emitComposite\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"flag\",\"type\":\"bool\"},{\"internalType\":\"bytes5\",\"name\":\"code\",\"type\":\"bytes5\"},{\"internalType\":\"int16\",\"name\":\"delta\",\"type\":\"int16\"}],\"name\":\"emitIndexedScalars\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"}],\"name\":\"emitTransformed\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"value\",\"type\":\"uint16\"}],\"name\":\"emitTransformed\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"function(uint512)pureexternalreturns(uint512)\",\"name\":\"callback\",\"type\":\"function\"},{\"internalType\":\"uint512\",\"name\":\"value\",\"type\":\"uint512\"}],\"name\":\"exerciseFunction\",\"outputs\":[{\"internalType\":\"function(uint512)pureexternalreturns(uint512)\",\"name\":\"\",\"type\":\"function\"},{\"internalType\":\"uint512\",\"name\":\"\",\"type\":\"uint512\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint512\",\"name\":\"code\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"bytes64\",\"name\":\"tag\",\"type\":\"bytes64\"}],\"internalType\":\"structEventEmitter.Record\",\"name\":\"record\",\"type\":\"tuple\"},{\"internalType\":\"uint16[][]\",\"name\":\"nested\",\"type\":\"uint16[][]\"}],\"name\":\"failComplex\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"failPanic\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"failReason\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"observe\",\"outputs\":[{\"internalType\":\"uint512\",\"name\":\"value\",\"type\":\"uint512\"},{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"marker\",\"type\":\"uint16\"}],\"name\":\"pay\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint512\",\"name\":\"value\",\"type\":\"uint512\"}],\"name\":\"plusOne\",\"outputs\":[{\"internalType\":\"uint512\",\"name\":\"\",\"type\":\"uint512\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"int512\",\"name\":\"delta\",\"type\":\"int512\"},{\"internalType\":\"bytes64\",\"name\":\"tag\",\"type\":\"bytes64\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bool\",\"name\":\"enabled\",\"type\":\"bool\"}],\"name\":\"store\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"}],\"name\":\"transform\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"value\",\"type\":\"uint16\"}],\"name\":\"transform\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x61010060805234a015610010575fa0fd5b506080516133983803a0613398a339a1016080a1b05261002fb1610254565b5fa5b0556080517e080761fb3eddcdd63161e23554f1566d1bc4623a2a4f3be0b73987716811df6101091bb061006eb0a7b0a7b0a7b0a7b0a7b0610334565b608051a0b103b0c150505050506103c7565b634e487b716101e01b5f52604160045260445ffd5b608051603fa201603f1916a1016001600160401b03a111a2a21017156100bd576100bd610080565b608052b1b050565b5f5ba3a110156100df57a1a10151a3a201526040016100c7565ba35ba1a110156100f6575fa1a501536001016100e1565b5050505050565b5f6001600160401b03a3111561011557610115610080565b610128603fa401603f1916604001610095565bb050a2a152a3a3a301111561013b575fa0fd5b610149a36040a301a46100c5565bb3b2505050565b5fa2603fa3011261015f575fa0fd5b610149a3a3516040a5016100fd565b5f60c0a2a403121561017e575fa0fd5b60805160c0a1016001600160401b03a111a2a21017156101a0576101a0610080565ba060805250a0b150a251a1526040a301516040a201526080a301516080a2015250b2b15050565b5fa2603fa301126101d6575fa0fd5ba15160406001600160401b03a211156101f1576101f1610080565ba160061b610200a2a201610095565bb2a352a4a101a201b2a2a101b0a7a51115610219575fa0fd5ba3a701b2505ba4a3101561024957a25161ffffa116a0a214610239575fa0fd5ba35250b1a301b1b0a301b061021f565bb7b650505050505050565b5fa05fa05f6101c0a6a8031215610269575fa0fd5ba5516040a70151b0b5506001600160401b03a0a21115610287575fa0fd5ba1a801b150a8603fa3011261029a575fa0fd5b6102a9a9a3516040a5016100fd565bb5506080a80151b150a0a211156102be575fa0fd5b6102caa9a3aa01610150565bb4506102d9a960c0aa0161016e565bb350610180a80151b150a0a211156102ef575fa0fd5b506102fca8a2a9016101c7565bb15050b2b550b2b5b0b350565b5fa151a0a452610320a16040a6016040a6016100c5565b603f01603f1916b2b0b201604001b2b15050565b5f6101c0a7a3526040a1a1a5015261034ea2a501a9610309565bb150a3a2036080a50152610362a2a8610309565ba65160c0a60152a1a70151610100a601526080a70151610140a60152a4a103610180a60152a551a0a252a6a301b350b0a201b05f5ba1a110156103b757a45161ffff16a352b3a301b3b1a301b1600101610397565b50b0bab950505050505050505050565b612fc4a06103d45f395ff3fe6101006080526004361061010a575f356101e01ca0630753c06a1461018357a06314fc78fc146101bb57a0631e3ed7e4146101dc57a0632fb0dbcd146101fd57a0633b0e4d671461021057a0633d0e10891461024357a0634b79d0e31461026257a0634dc96ec01461029457a06350aa10c9146102c057a06379531c40146102ee57a06399cf235f1461031b57a0639e420a8f1461033a57a063a43e73c91461036657a063af73dcd11461039457a063b0b75436146103c357a063b94d6fa6146103d757a063c66c9028146103f657a063e558a3a71461040a57a063ed928c961461043b57a063f404ae991461046d57a063f80412291461048c57a063fb144722146104ba5761014c565b3661014c577fa8142743f8f70a4c26f3691cf4ed59718381fb2f18070ec52be1f1022d8555576101001b34608051610142b1b06110cc565b608051a0b103b0c1005b7fe5b92b8ba08394dd9b027fafca0dc888f149e8f420b55893ecee14ea148aa08b6101001b5f3634608051610142b3b2b1b0611109565b34a01561018e575fa0fd5b506101a261019d36600461118c565b6104d9565b6080516101b2b4b3b2b1b061127f565b608051a0b103b0f35b34a0156101c6575fa0fd5b505f546080a051b1a252336040a30152016101b2565b34a0156101e7575fa0fd5b506101fb6101f6366004611350565b61063b565b005b6101fb61020b3660046113a4565b61067c565b34a01561021b575fa0fd5b5061022f61022a366004611424565b6106c0565b6080516101b2b8b7b6b5b4b3b2b1b06115e9565b34a01561024e575fa0fd5b506101fb61025d36600461169f565b610764565b34a01561026d575fa0fd5b5061028161027c36600461169f565b61082b565b6080516101b2b7b6b5b4b3b2b1b0611740565b34a01561029f575fa0fd5b506102b36102ae366004611799565b6108f8565b6080516101b2b1b06117b0565b34a0156102cb575fa0fd5b506102df6102da3660046118c6565b610989565b6080516101b2b3b2b1b06119ab565b34a0156102f9575fa0fd5b5061030d610308366004611a63565b6109d1565b608051b0a1526040016101b2565b34a015610326575fa0fd5b506101fb610335366004611a8b565b6109dd565b34a015610345575fa0fd5b50610359610354366004611350565b610a24565b6080516101b2b1b0611b21565b34a015610371575fa0fd5b50610385610380366004611b33565b610a6d565b6080516101b2b3b2b1b0611b72565b34a01561039f575fa0fd5b506103b36103ae366004611be0565b610b59565b6080516101b2b4b3b2b1b0611e11565b34a0156103ce575fa0fd5b506101fb610cb3565b34a0156103e2575fa0fd5b506101fb6103f1366004611eae565b610cfe565b34a015610401575fa0fd5b506101fb610d46565b34a015610415575fa0fd5b50610429610424366004611f07565b610d50565b6080516101b2b6b5b4b3b2b1b0611fff565b34a015610446575fa0fd5b5061045a6104553660046113a4565b610eaa565b60805161ffffb0b116a1526040016101b2565b34a015610478575fa0fd5b506101fb6104873660046113a4565b610eb6565b34a015610497575fa0fd5b506104ab6104a63660046120b8565b610ef4565b6080516101b2b3b2b1b06120ee565b34a0156104c5575fa0fd5b506101fb6104d43660046121b3565b610f28565b6104e1610f52565b60c06104eb610f52565b60c0a9a9a9a9a9a9a56002a0604002608051b0a101608052a0b2b1b0a260025fb25ba1a4101561052b57a235a152604001b1604001b1b2600101b261050d565bb2505050505050b550a4a4a0a0604002604001608051b0a101608052a0b3b2b1b0a1600160016101001b0316a152604001a3a35fb25ba1a4101561057f57a235a152604001b1604001b1b2600101b2610561565bb250505050505050b350b0b1b2b350a26002a0604002608051b0a101608052a0b2b1b0a260025fb25ba1a410156105c657a235a152604001b1604001b1b2600101b26105a8565bb2505050505050b250a1a1a0a0604002604001608051b0a101608052a0b3b2b1b0a1600160016101001b0316a152604001a3a35fb25ba1a4101561061a57a235a152604001b1604001b1b2600101b26105fc565bb250505050505050b050b050b350b350b350b350b650b650b650b6b2505050565b7f29d8416f597bcc46fa3c441ff72963f4a2852e9c6d77447615f782a1ca0da3576101011ba2a2608051610670b2b1b0612264565b608051a0b103b0c15050565ba061ffff16337f1398d89bb96c43f8c16ef74dee904b456a4fa8a5857191293b848ced1997a3d96101001b346080516106b5b1b06110cc565b608051a0b103b0c350565b5fa05fa05f6106cd610f70565b6106d5610f8e565b6106dd610f8e565bafafafafafafafafa26003a0604002608051b0a101608052a0b2b1b0a260035fb25ba1a4101561072157a23561ffff16a152604001b1604001b1b2600101b26106ff565bb2505050505050b250a1610734b0612394565bb15061073fa161248b565bb050b750b750b750b750b750b750b750b750b850b850b850b850b850b850b850b8b050565ba85fa1b05550a7a9a77f0971a927eb69632cd5aced366c9dd3ee5626b6c0a27cb781139eeffab9e5372f6101001baaa9a9a9a9a96080516107aab6b5b4b3b2b1b06124df565b608051a0b103b0c4a2a26080516107c2b2b1b0612524565b608051a0b103b0206101001ba5a56080516107deb2b1b0612524565b608051b0a1b003a120aba2526101001bb07f4ef7447df163d4aaeab9c66fa93651de5eebb002dcf9b60da1ebaa28ae95e8256101001bb0604001608051a0b103b0c3505050505050505050565b5fa05fa060c0a05fafafafafafafafafafa4a4a0a0603f016040a0b10402604001608051b0a101608052a0b3b2b1b0a1600160016101001b0316a152604001a3a3a0a2a4375fb201b1b0b15250506080a0516040603fa801a1b004a102a201a101b0b252600160016101001b03a716a152b3b850b5b650b3b4b2b350b0b1a5b150a4b0a1b0a401a3a2a0a2a4375fa1a40152603f19603fa20116b050a0a301b250505050505050b150b0b150b650b650b650b650b650b650b650b950b950b950b950b950b950b9b2505050565b6080a0516103c0a101a2525fa0a2526040a201a1b052b1a101a2b05260c0a101a2b052610100a101a2b052610140a101a2b052610180a101a2b0526101c0a101a2b052610200a101a2b052610240a101a2b052610280a101a2b0526102c0a101a2b052610300a101a2b052610340a101a2b052610380a101b1b0b15261098336a3b003a301a3612602565bb2b15050565b610991610f8e565b60c0a0a7a7a7a7a76109a2a56127ae565bb4506109aea3a5612802565bb350b0b1506109bda1a3612871565bb3bcb2bb50b2b950b0b75050505050505050565b5f610983a260016128e9565b7f52ebad060d7f3dc17d6ea0e956b35cfd849a6a551b539872ed459157021a97076101011ba4a4a4a4608051610a16b4b3b2b1b0612adb565b608051a0b103b0c150505050565b60c0a2a2a0a0603f016040a0b10402604001608051b0a101608052a0b3b2b1b0a1600160016101001b0316a152604001a3a3a0a2a4375fb201b1b0b15250b2b6b5505050505050565b5fa05fa0a6a6a6608051a263ffffffff166101e01ba152600401610a93b1a152604001b0565b6040608051a0a303a1a65afa15a015610aae573d5fa03e3d5ffd5b505050506080513d603f01603f19163da1a11015610acf57a0a20336a2a501375b50a1016080a1b052610ae0b1612bc2565b608051a8a152600160016101e01b03196101e0a9b01b166040a20152b0b150604401608051a0b103b0206101001b7f3e85e019f156fb371415540e280c0864415370980dac796b574f42a76aa4d08f6101021ba8a8a4608051610b45b3b2b1b0611b72565b608051a0b103b0c2b5b6b4b5b4b350505050565b610b61610fb5565b60c0610b6b610fec565b610b73611023565ba8a8a8a8a8a46002a0604002608051b0a101608052a0b2b1b05fb05ba2a21015610bee576080a051a0a201a252b0a302a50160025fa2a2a55ba1a41015610bce57a23561ffff16a152604001b1604001b1b2600101b2610bac565bb2505050505050600160016101001b0316a152604001b0600101b0610b8f565b50505050b450a3a3a0a0604002604001608051b0a101608052a0b3b2b1b0a1600160016101001b0316a1526040015fb05ba2a21015610c7e576080a051a0a201a252b0a302a60160025fa2a2a55ba1a41015610c5e57a23561ffff16a152604001b1604001b1b2600101b2610c3c565bb2505050505050600160016101001b0316a152604001b0600101b0610c1f565b5050505050b250b0b1b250a1610c93b0612d08565bb150610c9ea1612d5c565bb3bdb2bc50b0ba50b1b850b650505050505050565b60805162461bcd6101e51ba15260406004a2015260196044a20152782b269039ba30b73230b932103932bb32b93a103932b0b9b7b76101391b6084a2015260c4015b608051a0b103b0fd5ba060010ba2600160016101d81b031916a415157f19c59af463d0b89e6afb02db53c6ea998a04ce7bf1aa5c2c0d4c3ac9efc9e6596101001b608051608051a0b103b0c4505050565b610d4e612dfe565b565b5fa060c0610d5c611077565b6080a05160c0a1a101a3525f6040a301a1b052a252b1a101a2b052adadadadadadadada5a5a0a0603f016040a0b10402604001608051b0a101608052a0b3b2b1b0a1600160016101001b0316a152604001a3a3a0a2a4375fb201a2b052506080a051610100a101b0b152b4ba50b7b850b5b6b4b550b2b3b1b250a6b16002b150a2a2a55ba1a41015610e0f576040a3a1013563ffffffff16b0a20152a235a1526001b3b0b301b26080b2a301b201610de0565bb2505050505050b350a2a2a0a0608002604001608051b0a101608052a0b3b2b1b0a1600160016101001b0316a152604001a3a35fb25ba1a41015610e74576040a3a1013563ffffffff16b0a20152a235a1526001b3b0b301b26080b2a301b201610e45565bb250505050505050b150b0b150a0610e8bb0612e13565bb050b550b550b550b550b550b550b850b850b850b850b850b8b2505050565b5f610983a26001612e6f565b60805161ffffa216a1527f38e10c946553786c68c20c6706a95e949fdaa40be4cdb80e325a15f92c2e08d56101021bb0604001608051a0b103b0c150565b610efc6110a5565b60c0a0a7a7a7a7a7610f0da5612e91565bb450610f19a3a5612e9c565bb350b0b1506109bda1a3612eff565ba7a7a7a7a7a7a7a760805163a6b3cee16101e01ba152600401610cf5b8b7b6b5b4b3b2b1b0612f62565b608051a0608001608052a06002b06040a202a036a33750b1b2b15050565b608051a060c001608052a06003b06040a202a036a33750b1b2b15050565b608051a0608001608052a06002b05b60c0a1525f19b0b101b0604001a1610f9d57b05050b0565b608051a0608001608052a06002b05b610fcc610f52565b600160016101001b0316a152604001b06001b003b0a1610fc457b05050b0565b608051a0608001608052a06002b05b6110036110a5565b600160016101001b0316a152604001b06001b003b0a1610ffb57b05050b0565b6080a051610180a101b0b1525f60c0a201a1a152610100a301a2b052610140a301b1b0b152600160016101001b0316a1526040a1016110606110a5565b600160016101001b0316a15260c06040b0b10152b0565b608051a061010001608052a06002b05b5f6040a201a1b052a1525f19b0b101b0608001a161108757b05050b0565b6080a051610100a101a2525fa15260c06040a201a1b052b1a101a2b052a1a101b1b0b152b0565b600160016101001b03b1b0b116a152604001b0565ba1a352a1a16040a50137505fa2a2016040b0a101b1b0b152603fb0b101603f1916b0b10101b0565b6080a1525f61111c6080a301a5a76110e1565bb0506001a06101001b03a3166040a30152b4b350505050565ba06080a101a31015610983575fa0fd5b5fa0a3603fa40112611155575fa0fd5b50a1356001600160401b03a1111561116b575fa0fd5b6040a301b150a36040a260061ba501011115611185575fa0fd5bb250b2b050565b5fa05fa05fa0610180a7a90312156111a2575fa0fd5b6111aca8a8611135565bb5506080a701356001600160401b03a0a211156111c7575fa0fd5b6111d3aaa3ab01611145565bb0b750b550a5b1506111e8aa60c0ab01611135565bb450610140a90135b150a0a211156111fe575fa0fd5b5061120ba9a2aa01611145565bb7bab6b950b4b750b2b5b3b4b2505050565ba05f5b6002a1101561123f57a151a4526040b3a401b3b0b101b0600101611220565b50505050565b5fa151a0a4526040a0a501b4506040a4015f5ba3a1101561127457a151a752b5a201b5b0a201b0600101611258565b50b4b5b45050505050565b5f610180a2a101a3a8a45b6002a110156112a957a151a3526040b2a301b2b0b101b060010161128a565b5050506080a401b1b0b152a551b0a1b0526101c0a301b06040b0a1a8015f5ba2a110156112e457a151a552b3a301b3b0a301b06001016112c8565b505050506112f560c0a401a661121d565ba2a103610140a40152611308a1a5611245565bb7b650505050505050565b5fa0a3603fa40112611323575fa0fd5b50a1356001600160401b03a11115611339575fa0fd5b6040a301b150a36040a2a501011115611185575fa0fd5b5fa06040a3a5031215611361575fa0fd5ba2356001600160401b03a11115611376575fa0fd5b611382a5a2a601611313565bb0b6b0b550b350505050565ba03561ffffa116a11461139f575fa0fd5bb1b050565b5f6040a2a40312156113b4575fa0fd5b6113bda261138e565bb3b2505050565ba0355fa1b00ba11461139f575fa0fd5ba035600160016101001b03a116a11461139f575fa0fd5ba035601fa1b00ba11461139f575fa0fd5ba035600160016101d81b0319a116a11461139f575fa0fd5ba060c0a101a31015610983575fa0fd5b5fa05fa05fa05fa0610280a9ab03121561143c575fa0fd5ba83560ffa116a11461144c575fa0fd5bb75061145a6040aa016113c4565bb6506114686080aa016113d4565bb55061147660c0aa016113eb565bb450611485610100aa016113fc565bb350611495aa610140ab01611414565bb250610200a901356001600160401b03a0a211156114b1575fa0fd5b6114bdaca3ad01611135565bb350610240ab0135b150a0a211156114d3575fa0fd5b506114e0aba2ac01611135565bb15050b2b5b850b2b5b8b0b3b650565b5fa151a0a4525f5ba1a11015611514576040a1a501a10151a6a301a20152016114f8565ba15ba1a1101561152e575f6040a2a8010153600101611516565b5050603f01603f1916b2b0b201604001b2b15050565b5fa26080a101a35f5b6002a1101561157c57a3a303a752611566a3a3516114f0565b6040b7a801b7b0b350b1b0b101b060010161154d565b50b0b5b45050505050565b5fa26080a101a35f5b6002a1101561157c57a3a303a752a151a051a0a5526040b1a201b1a0a601b1b05f5ba2a110156115d257a45161ffff16a452b3a101b3b2a101b26001016115b2565b50b9aa01b9b1b55050b2b0b201b150600101611590565b60ffa916a1525fa8a10b6040a0a401b1b0b152600160016101001b03a9166080a40152601fa8b00b60c0a40152600160016101d81b0319a716610100a40152610280b0610140a401a7a45b6003a1101561165557a15161ffff16a352b1a301b1b0a301b0600101611634565b50505050a0610200a4015261166ca1a401a6611544565bb050a2a103610240a40152611681a1a5611587565bbbba5050505050505050505050565ba035a01515a11461139f575fa0fd5b5fa05fa05fa05fa05f6101c0aaac0312156116b8575fa0fd5ba935b8506040aa0135b7506080aa0135b65060c0aa0135b550610100aa01356001600160401b03a0a211156116eb575fa0fd5b6116f7ada3ae01611313565bb0b750b550610140ac0135b150a0a21115611710575fa0fd5b5061171daca2ad01611313565bb0b450b250611731b050610180ab01611690565bb050b2b5b850b2b5b850b2b5b8565b5f6101c0a9a352a86040a40152a76080a40152a660c0a40152a0610100a4015261176ca1a401a76114f0565bb050a2a103610140a40152611781a1a66114f0565bb15050a21515610180a30152b8b75050505050505050565b5f6103c0a2a40312156117aa575fa0fd5b50b1b050565ba1516001600160f81b0316a1526103c0a1016040a301516117d66040a401a2601e0bb052565b506080a301516117f26080a401a2600160016101001b0316b052565b5060c0a3015161180760c0a401a2601f0bb052565b50610100a3a10151600160016101081b0316b0a30152610140a0a4015160200bb0a30152610180a0a40151600160016101f81b0316b0a301526101c0a0a40151603e0bb0a30152610200a0a40151b0a30152610240a0a40151b0a30152610280a0a40151600160016101081b031916b0a301526102c0a0a40151600160016101001b031916b0a30152610300a0a401516001600160f81b031916b0a30152610340a0a4015160ff1916b0a30152610380b2a30151b2b0b101b1b0b152b0565b5fa05fa05f60c0a6a80312156118da575fa0fd5ba5356001600160401b03a0a211156118f0575fa0fd5b6118fca9a3aa01611135565bb6506040a80135b150a0a21115611911575fa0fd5b61191da9a3aa01611145565bb0b650b4506080a80135b150a0a21115611935575fa0fd5b50611942a8a2a901611145565bb6b9b5b850b3b650b2b4b3b2505050565b5fa2a251a0a5526040a0a601b5506040a260061ba401016040a6015f5ba4a1101561199e57603f19a6a40301a95261198ca3a3516114f0565bb8a401b8b250b0a301b0600101611970565b50b0b7b650505050505050565b60c0a0a2525fb0610140a301b0a301a6a35b6002a110156119ef5760bf19a6a50301a3526119daa4a3516114f0565bb3506040b2a301b2b1b0b101b06001016119bd565b5050506040a3a203a1a50152a1a651a0a452a2a401b150a2a160061ba50101a3a9015f5ba3a11015611a4157603f19a7a40301a552611a2fa3a3516114f0565bb4a601b4b250b0a501b0600101611a13565b5050a6a1036080a80152611a55a1a9611953565bbab950505050505050505050565b5f6040a2a4031215611a73575fa0fd5b5035b1b050565b5f610100a2a40312156117aa575fa0fd5b5fa05fa0610180a5a7031215611a9f575fa0fd5ba4356001600160401b03a0a21115611ab5575fa0fd5b611ac1a8a3a901611a7a565bb550611ad0a86040a901611414565bb450610100a70135b150a0a21115611ae6575fa0fd5b611af2a8a3a901611135565bb350610140a70135b150a0a21115611b08575fa0fd5b50611b15a7a2a801611135565bb15050b2b5b1b450b250565b6040a1525f6113bd6040a301a46114f0565b5fa05f60c0a4a6031215611b45575fa0fd5b5050a135b36040a3013563ffffffff16b3506080b0b20135b1b050565ba25263ffffffff166040b0b10152565b60c0a101611b81a2a5a7611b62565ba26080a30152b4b350505050565ba0610100a101a31015610983575fa0fd5b5fa0a3603fa40112611bb0575fa0fd5b50a1356001600160401b03a11115611bc6575fa0fd5b6040a301b150a36040a260071ba501011115611185575fa0fd5b5fa05fa05f6101c0a6a8031215611bf5575fa0fd5b611bffa7a7611b8f565bb450610100a601356001600160401b03a0a21115611c1b575fa0fd5b611c27a9a3aa01611ba0565bb0b650b450610140b150a7a20135a1a11115611c41575fa0fd5b611c4daaa2ab01611135565bb45050610180a80135a1a11115611c62575fa0fd5ba801b050a0a903a21315611c74575fa0fd5ba0b2505050b2b550b2b5b0b350565b5fa2a2a25b6002a11015611cab57a15161ffff16a3526040b2a301b2b0b101b0600101611c88565b5050506080a301b050b2b15050565b5fa2a251a0a5526040a0a601b550a0a260061ba40101a1a6015f5ba4a1101561199e57a5a303603f1901a952a151a051a0a552b0a501b0a5a501b05f5ba1a11015611d1757a35161ffff16a352b2a701b2b1a701b1600101611cf7565b5050b9a501b9b35050b0a301b0600101611cd5565b5f610100a251a4526040a30151a16040a60152611d4ba2a601a26114f0565bb150506080a30151a4a2036080a60152611d65a2a26114f0565bb1505060c0a30151a4a20360c0a60152611d7fa2a2611cba565bb5b45050505050565b5fa26080a101a35f5b6002a1101561157c57a3a303a752611daaa3a351611d2c565b6040b7a801b7b0b350b1b0b101b0600101611d91565b5f610140a251a051a5526040a101516040a601526080a101516080a60152506040a30151a160c0a60152611df6a2a601a2611d2c565bb150506080a30151a4a203610100a60152611d7fa2a26114f0565b5f6101c0a2a101a3a8a45b6002a11015611e4157611e30a3a351611c83565bb2506040b1b0b101b0600101611e1c565b505050610100a401b1b0b152a551b0a1b052610200a301b06040b0a1a8015f5ba2a11015611e8257611e74a5a351611c83565bb450b0a301b0600101611e61565b50505050a2a103610140a40152611e99a1a6611d88565bb050a2a103610180a40152611308a1a5611dc0565b5fa05f60c0a4a6031215611ec0575fa0fd5b611ec9a4611690565bb250611ed76040a5016113fc565bb1506080a40135a060010ba114611eec575fa0fd5ba0b15050b250b250b2565b5f60c0a2a40312156117aa575fa0fd5b5fa05fa05fa05fa0610240a9ab031215611f1f575fa0fd5ba835b7506040a9013563ffffffff16b6506080a901356001600160401b03a0a21115611f49575fa0fd5b611f55aca3ad01611313565bb0b850b650a6b150611f6aac60c0ad01611b8f565bb5506101c0ab0135b150a0a21115611f80575fa0fd5b611f8caca3ad01611ba0565bb0b550b350610200ab0135b150a0a21115611fa5575fa0fd5b506114e0aba2ac01611ef7565ba0516040b0b10151b0b163ffffffffb0b116b0565b5f611fd1a2611fb2565b611fdca5a2a4611b62565b50506080a2015160c06080a50152611ff760c0a501a26114f0565bb4b350505050565b5f61024061200ea3a9ab611b62565b6080a16080a50152612022a2a501a96114f0565bb15060c0a401a75f5b6002a1101561205a5761203da2611fb2565b612048a5a2a4611b62565b5050b1a301b1b0a301b060010161202b565b505050a3a2036101c0a50152a551a0a3526040a0a801b301b05f5ba1a110156120a357612086a5611fb2565b612091a5a2a4611b62565b5050b3a301b3b1a301b1600101612075565b5050a4a103610200a60152611681a1a7611fc7565b5fa05fa05f60c0a6a80312156120cc575fa0fd5ba5356001600160401b03a0a211156120e2575fa0fd5b6118fca9a3aa01611a7a565b60c0a1525f61210060c0a301a6611d2c565b6040a3a203a1a50152a1a651a0a452a2a401b150a2a160061ba50101a3a9015f5ba3a1101561214f57603f19a7a40301a55261213da3a351611d2c565bb4a601b4b250b0a501b0600101612121565b5050a6a1036080a80152a751a0a252a4a201b550b2506006a3b01ba101a401b150a3a8015f5ba4a110156121a357603f19a3a50301a752612191a4a351611cba565bb6a601b6b350b0a501b0600101612175565b50b1bab950505050505050505050565b5fa05fa05fa05fa06101c0a9ab0312156121cb575fa0fd5ba835b7506040a901356001600160401b03a0a211156121e8575fa0fd5b6121f4aca3ad01611313565bb0b950b7506080ab0135b150a0a2111561220c575fa0fd5b612218aca3ad01611313565bb0b750b550a5b15061222dac60c0ad01611ef7565bb450610180ab0135b150a0a21115612243575fa0fd5b50612250aba2ac01611145565bb9bcb8bb50b6b950b4b7b3b6b2b5b4505050565b6040a1525f611ff76040a301a4a66110e1565b634e487b716101e01b5f52604160045260445ffd5b6080516103c0a1016001600160401b03a111a2a21017156122af576122af612277565b608052b0565b60805160c0a1016001600160401b03a111a2a21017156122af576122af612277565b6080a051b0a1016001600160401b03a111a2a21017156122af576122af612277565b608051603fa201603f1916a1016001600160401b03a111a2a210171561232157612321612277565b608052b1b050565b5fa2603fa30112612338575fa0fd5ba1356001600160401b03a1111561235157612351612277565b612364603fa201603f19166040016122f9565ba1a152a46040a3a601011115612378575fa0fd5ba16040a5016040a301375fb1a101604001b1b0b152b3b2505050565b5f61239d6122d7565ba06080a40136a111156123ae575fa0fd5ba45ba1a110156123e857a0356001600160401b03a111156123cd575fa0fd5b6123d936a2a901612329565ba552506040b3a401b3016123b0565b50b0b4b350505050565b5f6001600160401b03a2111561240a5761240a612277565b5060061b604001b0565b5fa2603fa30112612423575fa0fd5ba1356040612438612433a36123f2565b6122f9565ba0a3a2526040a201b1506040a460061ba70101b350a6a41115612459575fa0fd5b6040a6015ba4a110156124805761ffff612472a261138e565b16a352b1a301b1a30161245e565b50b6b5505050505050565b5f6124946122d7565ba06080a40136a111156124a5575fa0fd5ba45ba1a110156123e857a0356001600160401b03a111156124c4575fa0fd5b6124d036a2a901612414565ba552506040b3a401b3016124a7565b5f610100a8a352a06040a401526124f9a1a401a8aa6110e1565bb050a2a1036080a4015261250ea1a6a86110e1565bb15050a2151560c0a30152b7b650505050505050565ba1a3a2375fb101b0a152b1b050565ba0356001600160f81b03a116a11461139f575fa0fd5ba035601ea1b00ba11461139f575fa0fd5ba035600160016101081b03a116a11461139f575fa0fd5ba0356020a1b00ba11461139f575fa0fd5ba035600160016101f81b03a116a11461139f575fa0fd5ba035603ea1b00ba11461139f575fa0fd5ba035600160016101081b0319a116a11461139f575fa0fd5ba035600160016101001b0319a116a11461139f575fa0fd5ba0356001600160f81b0319a116a11461139f575fa0fd5ba03560ff19a116a11461139f575fa0fd5b5f6103c0a2a4031215612613575fa0fd5b61261b61228c565b612634612627a4612533565b6001600160f81b0316a252565b61264d6126436040a501612549565b601e0b6040a30152565b61266d61265c6080a5016113d4565b600160016101001b03166080a30152565b61268661267c60c0a5016113eb565b601f0b60c0a30152565b6101006126a7612697a2a60161255a565b600160016101081b0316a3a30152565b506101406126c26126b9a2a601612571565b60200ba3a30152565b506101806126e46126d4a2a601612582565b600160016101f81b0316a3a30152565b506101c06126ff6126f6a2a601612599565b603e0ba3a30152565b50610200a3a10135b0a20152610240a0a40135b0a20152610280612738612727a2a6016125aa565b600160016101081b031916a3a30152565b506102c061275b61274aa2a6016125c2565b600160016101001b031916a3a30152565b5061030061277d61276da2a6016125da565b6001600160f81b031916a3a30152565b5061034061279961278fa2a6016125f1565b60ff1916a3a30152565b50610380b2a30135b2a101b2b0b25250b1b050565b5f6127b76122d7565ba06080a40136a111156127c8575fa0fd5ba45ba1a110156123e857a0356001600160401b03a111156127e7575fa0fd5b6127f336a2a901612329565ba552506040b3a401b3016127ca565b5f61280f612433a46123f2565ba0a4a2526040a0a301b250a560061ba50136a1111561282c575fa0fd5ba55ba1a1101561286557a0356001600160401b03a1111561284b575fa0fd5b61285736a2aa01612329565ba65250b3a201b3a20161282e565b50b1b6b5505050505050565b5f61287e612433a46123f2565ba0a4a2526040a0a301b250a560061ba50136a1111561289b575fa0fd5ba55ba1a1101561286557a0356001600160401b03a111156128ba575fa0fd5b6128c636a2aa01612329565ba65250b3a201b3a20161289d565b634e487b716101e01b5f52601160045260445ffd5ba0a201a0a21115610983576109836128d4565b5fa0a3356001600160401b03a0a21115612914575fa0fd5ba4a201b150603f193601a2113660401117a5a3101715612932575fa0fd5ba135b2506040a201b350a0a31115612948575fa0fd5b50a101604001a2a11036a211171561295e575fa0fd5b50b250b2b050565b5fa0a3356001600160401b03a0a2111561297e575fa0fd5ba4a201b150603f193601a2113660401117a5a310171561299c575fa0fd5ba135b2506040a201b350a0a311156129b2575fa0fd5b506006a2b01b01604001a2a11036a211171561295e575fa0fd5ba1a3525f6040a0a501b450a25f5ba5a110156112745761ffff6129eea361138e565b16a752b5a201b5b0a201b06001016129da565b5fa3a3a5526040a0a601b5506040a560061ba30101a45f5ba7a1101561199e57a4a303603f1901a952612a34a2a8612966565b612a3fa5a2a46129cc565bbaa601bab4505050b0a301b0600101612a19565b5fa26080a101a35f5b6002a1101561157c57a3a303a752612a74a2a76128fc565b612a7fa5a2a46110e1565b6040b9aa01b9b0b550b3b0b301b25050600101612a5c565b5fa26080a101a35f5b6002a1101561157c57a3a303a752612ab8a2a7612966565b612ac3a5a2a46129cc565b6040b9aa01b9b0b550b3b0b301b25050600101612aa0565b610180a0a252a535b0a201525f6040612af6a1a801a86128fc565b610100a06101c0a70152612b0f610280a701a3a56110e1565bb250612b1e6080ab01ab6128fc565bb25061017f19a0a8a60301610200a90152612b3aa5a5a46110e1565bb450612b4960c0ad01ad612966565bb450b150a0a8a60301610240a9015250612b64a4a4a3612a01565bb350506040a601b150a85f5b6003a11015612b985761ffff612b85a361138e565b16a452b2a501b2b0a501b0600101612b70565b5050a5a303b0a6015250612baca1a7612a53565bb15050a2a103610140a40152611308a1a5612a97565b5f6040a2a4031215612bd2575fa0fd5b5051b1b050565b5fa2603fa30112612be8575fa0fd5ba1356040612bf8612433a36123f2565ba2a1526006b2b0b21ba401a101b1a1a101b0a6a41115612c16575fa0fd5ba2a6015ba4a1101561248057a0356001600160401b03a11115612c37575fa0fd5b612c45a9a6a3ab0101612414565ba45250b1a301b1a301612c1a565b5f610100a0a3a5031215612c65575fa0fd5b608051b0a101b06001600160401b03a0a311a2a4101715612c8857612c88612277565ba2608052a1b350a435a2526040a50135b250a0a31115612ca6575fa0fd5b612cb2a6a4a701612329565b6040a301526080a50135b250a0a31115612cca575fa0fd5b612cd6a6a4a701612329565b6080a3015260c0a50135b250a0a31115612cee575fa0fd5b50612cfba5a3a601612bd9565b60c0a201525050b2b15050565b5f612d116122d7565ba06080a40136a11115612d22575fa0fd5ba45ba1a110156123e857a0356001600160401b03a11115612d41575fa0fd5b612d4d36a2a901612c53565ba552506040b3a401b301612d24565b5fa13603610140a11215612d6e575fa0fd5b612d766122b5565b60c0a21215612d83575fa0fd5b612d8b6122b5565ba435a1526040a0a60135b0a201526080a0a60135b0a20152a15260c0a40135b1506001600160401b03a0a31115612dc0575fa0fd5b612dcc36a4a701612c53565b6040a30152610100a50135b250a0a31115612de5575fa0fd5b50612df236a3a601612329565b6080a20152b3b2505050565b634e487b716101e01b5f52600160045260445ffd5b5f60c0a236031215612e23575fa0fd5b612e2b6122b5565ba235a1526040a0a4013563ffffffff16b0a201526080a301356001600160401b03a11115612e57575fa0fd5b612e6336a2a601612329565b6080a3015250b2b15050565b61ffffa1a116a3a21601b0a0a21115612e8a57612e8a6128d4565b50b2b15050565b5f61098336a3612c53565b5f612ea9612433a46123f2565ba0a4a2526040a0a301b250a560061ba50136a11115612ec6575fa0fd5ba55ba1a1101561286557a0356001600160401b03a11115612ee5575fa0fd5b612ef136a2aa01612c53565ba65250b3a201b3a201612ec8565b5f612f0c612433a46123f2565ba0a4a2526040a0a301b250a560061ba50136a11115612f29575fa0fd5ba55ba1a1101561286557a0356001600160401b03a11115612f48575fa0fd5b612f5436a2aa01612bd9565ba65250b3a201b3a201612f2b565b5f6101c0aaa352a06040a40152612f7ca1a401aaac6110e1565bb050a2a1036080a40152612f91a1a8aa6110e1565bb050a53560c0a401526040a60135610100a401526080a60135610140a40152a2a103610180a40152611681a1a5a7612a0156",
}

// EventEmitterABI is the input ABI used to generate the binding from.
// Deprecated: Use EventEmitterMetaData.ABI instead.
var EventEmitterABI = EventEmitterMetaData.ABI

// EventEmitterBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use EventEmitterMetaData.Bin instead.
var EventEmitterBin = EventEmitterMetaData.Bin

// DeployEventEmitter deploys a new QRL contract, binding an instance of EventEmitter to it.
func DeployEventEmitter(auth *bind.TransactOpts, backend bind.ContractBackend, initial *big.Int, note string, payload []byte, record EventEmitterRecord, numbers []uint16) (common.Address, *types.Transaction, *EventEmitter, error) {
	parsed, err := EventEmitterMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(EventEmitterBin), backend, initial, note, payload, record, numbers)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &EventEmitter{EventEmitterCaller: EventEmitterCaller{contract: contract}, EventEmitterTransactor: EventEmitterTransactor{contract: contract}, EventEmitterFilterer: EventEmitterFilterer{contract: contract}}, nil
}

// EventEmitter is an auto generated Go binding around a QRL contract.
type EventEmitter struct {
	EventEmitterCaller     // Read-only binding to the contract
	EventEmitterTransactor // Write-only binding to the contract
	EventEmitterFilterer   // Log filterer for contract events
}

// EventEmitterCaller is an auto generated read-only Go binding around a QRL contract.
type EventEmitterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EventEmitterTransactor is an auto generated write-only Go binding around a QRL contract.
type EventEmitterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EventEmitterFilterer is an auto generated log filtering Go binding around a QRL contract events.
type EventEmitterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EventEmitterSession is an auto generated Go binding around a QRL contract,
// with pre-set call and transact options.
type EventEmitterSession struct {
	Contract     *EventEmitter     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EventEmitterCallerSession is an auto generated read-only Go binding around a QRL contract,
// with pre-set call options.
type EventEmitterCallerSession struct {
	Contract *EventEmitterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// EventEmitterTransactorSession is an auto generated write-only Go binding around a QRL contract,
// with pre-set transact options.
type EventEmitterTransactorSession struct {
	Contract     *EventEmitterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// EventEmitterRaw is an auto generated low-level Go binding around a QRL contract.
type EventEmitterRaw struct {
	Contract *EventEmitter // Generic contract binding to access the raw methods on
}

// EventEmitterCallerRaw is an auto generated low-level read-only Go binding around a QRL contract.
type EventEmitterCallerRaw struct {
	Contract *EventEmitterCaller // Generic read-only contract binding to access the raw methods on
}

// EventEmitterTransactorRaw is an auto generated low-level write-only Go binding around a QRL contract.
type EventEmitterTransactorRaw struct {
	Contract *EventEmitterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEventEmitter creates a new instance of EventEmitter, bound to a specific deployed contract.
func NewEventEmitter(address common.Address, backend bind.ContractBackend) (*EventEmitter, error) {
	contract, err := bindEventEmitter(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &EventEmitter{EventEmitterCaller: EventEmitterCaller{contract: contract}, EventEmitterTransactor: EventEmitterTransactor{contract: contract}, EventEmitterFilterer: EventEmitterFilterer{contract: contract}}, nil
}

// NewEventEmitterCaller creates a new read-only instance of EventEmitter, bound to a specific deployed contract.
func NewEventEmitterCaller(address common.Address, caller bind.ContractCaller) (*EventEmitterCaller, error) {
	contract, err := bindEventEmitter(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EventEmitterCaller{contract: contract}, nil
}

// NewEventEmitterTransactor creates a new write-only instance of EventEmitter, bound to a specific deployed contract.
func NewEventEmitterTransactor(address common.Address, transactor bind.ContractTransactor) (*EventEmitterTransactor, error) {
	contract, err := bindEventEmitter(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EventEmitterTransactor{contract: contract}, nil
}

// NewEventEmitterFilterer creates a new log filterer instance of EventEmitter, bound to a specific deployed contract.
func NewEventEmitterFilterer(address common.Address, filterer bind.ContractFilterer) (*EventEmitterFilterer, error) {
	contract, err := bindEventEmitter(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EventEmitterFilterer{contract: contract}, nil
}

// bindEventEmitter binds a generic wrapper to an already deployed contract.
func bindEventEmitter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := EventEmitterMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EventEmitter *EventEmitterRaw) Call(opts *bind.CallOpts, result *[]any, method string, params ...any) error {
	return _EventEmitter.Contract.EventEmitterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EventEmitter *EventEmitterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EventEmitter.Contract.EventEmitterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EventEmitter *EventEmitterRaw) Transact(opts *bind.TransactOpts, method string, params ...any) (*types.Transaction, error) {
	return _EventEmitter.Contract.EventEmitterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EventEmitter *EventEmitterCallerRaw) Call(opts *bind.CallOpts, result *[]any, method string, params ...any) error {
	return _EventEmitter.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EventEmitter *EventEmitterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EventEmitter.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EventEmitter *EventEmitterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...any) (*types.Transaction, error) {
	return _EventEmitter.Contract.contract.Transact(opts, method, params...)
}

// Echo is a free data retrieval call binding the contract method 0x4b79d0e3.
//
// Hyperion: function echo(uint512 amount, int512 delta, bytes64 tag, address recipient, bytes payload, string note, bool enabled) pure returns(uint512, int512, bytes64, address, bytes, string, bool)
func (_EventEmitter *EventEmitterCaller) Echo(opts *bind.CallOpts, amount *big.Int, delta *big.Int, tag [64]byte, recipient common.Address, payload []byte, note string, enabled bool) (*big.Int, *big.Int, [64]byte, common.Address, []byte, string, bool, error) {
	var out []any
	err := _EventEmitter.contract.Call(opts, &out, "echo", amount, delta, tag, recipient, payload, note, enabled)

	if err != nil {
		return *new(*big.Int), *new(*big.Int), *new([64]byte), *new(common.Address), *new([]byte), *new(string), *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	out2 := *abi.ConvertType(out[2], new([64]byte)).(*[64]byte)
	out3 := *abi.ConvertType(out[3], new(common.Address)).(*common.Address)
	out4 := *abi.ConvertType(out[4], new([]byte)).(*[]byte)
	out5 := *abi.ConvertType(out[5], new(string)).(*string)
	out6 := *abi.ConvertType(out[6], new(bool)).(*bool)

	return out0, out1, out2, out3, out4, out5, out6, err

}

// Echo is a free data retrieval call binding the contract method 0x4b79d0e3.
//
// Hyperion: function echo(uint512 amount, int512 delta, bytes64 tag, address recipient, bytes payload, string note, bool enabled) pure returns(uint512, int512, bytes64, address, bytes, string, bool)
func (_EventEmitter *EventEmitterSession) Echo(amount *big.Int, delta *big.Int, tag [64]byte, recipient common.Address, payload []byte, note string, enabled bool) (*big.Int, *big.Int, [64]byte, common.Address, []byte, string, bool, error) {
	return _EventEmitter.Contract.Echo(&_EventEmitter.CallOpts, amount, delta, tag, recipient, payload, note, enabled)
}

// Echo is a free data retrieval call binding the contract method 0x4b79d0e3.
//
// Hyperion: function echo(uint512 amount, int512 delta, bytes64 tag, address recipient, bytes payload, string note, bool enabled) pure returns(uint512, int512, bytes64, address, bytes, string, bool)
func (_EventEmitter *EventEmitterCallerSession) Echo(amount *big.Int, delta *big.Int, tag [64]byte, recipient common.Address, payload []byte, note string, enabled bool) (*big.Int, *big.Int, [64]byte, common.Address, []byte, string, bool, error) {
	return _EventEmitter.Contract.Echo(&_EventEmitter.CallOpts, amount, delta, tag, recipient, payload, note, enabled)
}

// EchoBoundaries is a free data retrieval call binding the contract method 0x3b0e4d67.
//
// Hyperion: function echoBoundaries(uint8 smallUnsigned, int8 smallSigned, uint256 wideUnsigned, int256 wideSigned, bytes5 shortBytes, uint16[3] fixedNumbers, string[2] fixedStrings, uint16[][2] mixed) pure returns(uint8, int8, uint256, int256, bytes5, uint16[3], string[2], uint16[][2])
func (_EventEmitter *EventEmitterCaller) EchoBoundaries(opts *bind.CallOpts, smallUnsigned uint8, smallSigned int8, wideUnsigned *big.Int, wideSigned *big.Int, shortBytes [5]byte, fixedNumbers [3]uint16, fixedStrings [2]string, mixed [2][]uint16) (uint8, int8, *big.Int, *big.Int, [5]byte, [3]uint16, [2]string, [2][]uint16, error) {
	var out []any
	err := _EventEmitter.contract.Call(opts, &out, "echoBoundaries", smallUnsigned, smallSigned, wideUnsigned, wideSigned, shortBytes, fixedNumbers, fixedStrings, mixed)

	if err != nil {
		return *new(uint8), *new(int8), *new(*big.Int), *new(*big.Int), *new([5]byte), *new([3]uint16), *new([2]string), *new([2][]uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)
	out1 := *abi.ConvertType(out[1], new(int8)).(*int8)
	out2 := *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	out3 := *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	out4 := *abi.ConvertType(out[4], new([5]byte)).(*[5]byte)
	out5 := *abi.ConvertType(out[5], new([3]uint16)).(*[3]uint16)
	out6 := *abi.ConvertType(out[6], new([2]string)).(*[2]string)
	out7 := *abi.ConvertType(out[7], new([2][]uint16)).(*[2][]uint16)

	return out0, out1, out2, out3, out4, out5, out6, out7, err

}

// EchoBoundaries is a free data retrieval call binding the contract method 0x3b0e4d67.
//
// Hyperion: function echoBoundaries(uint8 smallUnsigned, int8 smallSigned, uint256 wideUnsigned, int256 wideSigned, bytes5 shortBytes, uint16[3] fixedNumbers, string[2] fixedStrings, uint16[][2] mixed) pure returns(uint8, int8, uint256, int256, bytes5, uint16[3], string[2], uint16[][2])
func (_EventEmitter *EventEmitterSession) EchoBoundaries(smallUnsigned uint8, smallSigned int8, wideUnsigned *big.Int, wideSigned *big.Int, shortBytes [5]byte, fixedNumbers [3]uint16, fixedStrings [2]string, mixed [2][]uint16) (uint8, int8, *big.Int, *big.Int, [5]byte, [3]uint16, [2]string, [2][]uint16, error) {
	return _EventEmitter.Contract.EchoBoundaries(&_EventEmitter.CallOpts, smallUnsigned, smallSigned, wideUnsigned, wideSigned, shortBytes, fixedNumbers, fixedStrings, mixed)
}

// EchoBoundaries is a free data retrieval call binding the contract method 0x3b0e4d67.
//
// Hyperion: function echoBoundaries(uint8 smallUnsigned, int8 smallSigned, uint256 wideUnsigned, int256 wideSigned, bytes5 shortBytes, uint16[3] fixedNumbers, string[2] fixedStrings, uint16[][2] mixed) pure returns(uint8, int8, uint256, int256, bytes5, uint16[3], string[2], uint16[][2])
func (_EventEmitter *EventEmitterCallerSession) EchoBoundaries(smallUnsigned uint8, smallSigned int8, wideUnsigned *big.Int, wideSigned *big.Int, shortBytes [5]byte, fixedNumbers [3]uint16, fixedStrings [2]string, mixed [2][]uint16) (uint8, int8, *big.Int, *big.Int, [5]byte, [3]uint16, [2]string, [2][]uint16, error) {
	return _EventEmitter.Contract.EchoBoundaries(&_EventEmitter.CallOpts, smallUnsigned, smallSigned, wideUnsigned, wideSigned, shortBytes, fixedNumbers, fixedStrings, mixed)
}

// EchoBoundaryEdges is a free data retrieval call binding the contract method 0x4dc96ec0.
//
// Hyperion: function echoBoundaryEdges((uint248,int248,uint256,int256,uint264,int264,uint504,int504,uint512,int512,bytes31,bytes32,bytes33,bytes63,bytes64) edges) pure returns((uint248,int248,uint256,int256,uint264,int264,uint504,int504,uint512,int512,bytes31,bytes32,bytes33,bytes63,bytes64))
func (_EventEmitter *EventEmitterCaller) EchoBoundaryEdges(opts *bind.CallOpts, edges EventEmitterBoundaryEdges) (EventEmitterBoundaryEdges, error) {
	var out []any
	err := _EventEmitter.contract.Call(opts, &out, "echoBoundaryEdges", edges)

	if err != nil {
		return *new(EventEmitterBoundaryEdges), err
	}

	out0 := *abi.ConvertType(out[0], new(EventEmitterBoundaryEdges)).(*EventEmitterBoundaryEdges)

	return out0, err

}

// EchoBoundaryEdges is a free data retrieval call binding the contract method 0x4dc96ec0.
//
// Hyperion: function echoBoundaryEdges((uint248,int248,uint256,int256,uint264,int264,uint504,int504,uint512,int512,bytes31,bytes32,bytes33,bytes63,bytes64) edges) pure returns((uint248,int248,uint256,int256,uint264,int264,uint504,int504,uint512,int512,bytes31,bytes32,bytes33,bytes63,bytes64))
func (_EventEmitter *EventEmitterSession) EchoBoundaryEdges(edges EventEmitterBoundaryEdges) (EventEmitterBoundaryEdges, error) {
	return _EventEmitter.Contract.EchoBoundaryEdges(&_EventEmitter.CallOpts, edges)
}

// EchoBoundaryEdges is a free data retrieval call binding the contract method 0x4dc96ec0.
//
// Hyperion: function echoBoundaryEdges((uint248,int248,uint256,int256,uint264,int264,uint504,int504,uint512,int512,bytes31,bytes32,bytes33,bytes63,bytes64) edges) pure returns((uint248,int248,uint256,int256,uint264,int264,uint504,int504,uint512,int512,bytes31,bytes32,bytes33,bytes63,bytes64))
func (_EventEmitter *EventEmitterCallerSession) EchoBoundaryEdges(edges EventEmitterBoundaryEdges) (EventEmitterBoundaryEdges, error) {
	return _EventEmitter.Contract.EchoBoundaryEdges(&_EventEmitter.CallOpts, edges)
}

// EchoCompositeContainers is a free data retrieval call binding the contract method 0xaf73dcd1.
//
// Hyperion: function echoCompositeContainers(uint16[2][2] fixedMatrix, uint16[2][] rows, (uint512,string,bytes,uint16[][])[2] records, ((uint512,address,bytes64),(uint512,string,bytes,uint16[][]),bytes) nested) pure returns(uint16[2][2], uint16[2][], (uint512,string,bytes,uint16[][])[2], ((uint512,address,bytes64),(uint512,string,bytes,uint16[][]),bytes))
func (_EventEmitter *EventEmitterCaller) EchoCompositeContainers(opts *bind.CallOpts, fixedMatrix [2][2]uint16, rows [][2]uint16, records [2]EventEmitterDynamicRecord, nested EventEmitterNestedRecord) ([2][2]uint16, [][2]uint16, [2]EventEmitterDynamicRecord, EventEmitterNestedRecord, error) {
	var out []any
	err := _EventEmitter.contract.Call(opts, &out, "echoCompositeContainers", fixedMatrix, rows, records, nested)

	if err != nil {
		return *new([2][2]uint16), *new([][2]uint16), *new([2]EventEmitterDynamicRecord), *new(EventEmitterNestedRecord), err
	}

	out0 := *abi.ConvertType(out[0], new([2][2]uint16)).(*[2][2]uint16)
	out1 := *abi.ConvertType(out[1], new([][2]uint16)).(*[][2]uint16)
	out2 := *abi.ConvertType(out[2], new([2]EventEmitterDynamicRecord)).(*[2]EventEmitterDynamicRecord)
	out3 := *abi.ConvertType(out[3], new(EventEmitterNestedRecord)).(*EventEmitterNestedRecord)

	return out0, out1, out2, out3, err

}

// EchoCompositeContainers is a free data retrieval call binding the contract method 0xaf73dcd1.
//
// Hyperion: function echoCompositeContainers(uint16[2][2] fixedMatrix, uint16[2][] rows, (uint512,string,bytes,uint16[][])[2] records, ((uint512,address,bytes64),(uint512,string,bytes,uint16[][]),bytes) nested) pure returns(uint16[2][2], uint16[2][], (uint512,string,bytes,uint16[][])[2], ((uint512,address,bytes64),(uint512,string,bytes,uint16[][]),bytes))
func (_EventEmitter *EventEmitterSession) EchoCompositeContainers(fixedMatrix [2][2]uint16, rows [][2]uint16, records [2]EventEmitterDynamicRecord, nested EventEmitterNestedRecord) ([2][2]uint16, [][2]uint16, [2]EventEmitterDynamicRecord, EventEmitterNestedRecord, error) {
	return _EventEmitter.Contract.EchoCompositeContainers(&_EventEmitter.CallOpts, fixedMatrix, rows, records, nested)
}

// EchoCompositeContainers is a free data retrieval call binding the contract method 0xaf73dcd1.
//
// Hyperion: function echoCompositeContainers(uint16[2][2] fixedMatrix, uint16[2][] rows, (uint512,string,bytes,uint16[][])[2] records, ((uint512,address,bytes64),(uint512,string,bytes,uint16[][]),bytes) nested) pure returns(uint16[2][2], uint16[2][], (uint512,string,bytes,uint16[][])[2], ((uint512,address,bytes64),(uint512,string,bytes,uint16[][]),bytes))
func (_EventEmitter *EventEmitterCallerSession) EchoCompositeContainers(fixedMatrix [2][2]uint16, rows [][2]uint16, records [2]EventEmitterDynamicRecord, nested EventEmitterNestedRecord) ([2][2]uint16, [][2]uint16, [2]EventEmitterDynamicRecord, EventEmitterNestedRecord, error) {
	return _EventEmitter.Contract.EchoCompositeContainers(&_EventEmitter.CallOpts, fixedMatrix, rows, records, nested)
}

// EchoDynamicContainers is a free data retrieval call binding the contract method 0x50aa10c9.
//
// Hyperion: function echoDynamicContainers(bytes[2] fixedBytes, bytes[] byteSlices, string[] strings) pure returns(bytes[2], bytes[], string[])
func (_EventEmitter *EventEmitterCaller) EchoDynamicContainers(opts *bind.CallOpts, fixedBytes [2][]byte, byteSlices [][]byte, strings []string) ([2][]byte, [][]byte, []string, error) {
	var out []any
	err := _EventEmitter.contract.Call(opts, &out, "echoDynamicContainers", fixedBytes, byteSlices, strings)

	if err != nil {
		return *new([2][]byte), *new([][]byte), *new([]string), err
	}

	out0 := *abi.ConvertType(out[0], new([2][]byte)).(*[2][]byte)
	out1 := *abi.ConvertType(out[1], new([][]byte)).(*[][]byte)
	out2 := *abi.ConvertType(out[2], new([]string)).(*[]string)

	return out0, out1, out2, err

}

// EchoDynamicContainers is a free data retrieval call binding the contract method 0x50aa10c9.
//
// Hyperion: function echoDynamicContainers(bytes[2] fixedBytes, bytes[] byteSlices, string[] strings) pure returns(bytes[2], bytes[], string[])
func (_EventEmitter *EventEmitterSession) EchoDynamicContainers(fixedBytes [2][]byte, byteSlices [][]byte, strings []string) ([2][]byte, [][]byte, []string, error) {
	return _EventEmitter.Contract.EchoDynamicContainers(&_EventEmitter.CallOpts, fixedBytes, byteSlices, strings)
}

// EchoDynamicContainers is a free data retrieval call binding the contract method 0x50aa10c9.
//
// Hyperion: function echoDynamicContainers(bytes[2] fixedBytes, bytes[] byteSlices, string[] strings) pure returns(bytes[2], bytes[], string[])
func (_EventEmitter *EventEmitterCallerSession) EchoDynamicContainers(fixedBytes [2][]byte, byteSlices [][]byte, strings []string) ([2][]byte, [][]byte, []string, error) {
	return _EventEmitter.Contract.EchoDynamicContainers(&_EventEmitter.CallOpts, fixedBytes, byteSlices, strings)
}

// EchoFunctions is a free data retrieval call binding the contract method 0xe558a3a7.
//
// Hyperion: function echoFunctions(function callback, string note, function[2] fixedCallbacks, function[] callbacks, (function,string) record) pure returns(function, string, function[2], function[], (function,string))
func (_EventEmitter *EventEmitterCaller) EchoFunctions(opts *bind.CallOpts, callback [68]byte, note string, fixedCallbacks [2][68]byte, callbacks [][68]byte, record EventEmitterFunctionRecord) ([68]byte, string, [2][68]byte, [][68]byte, EventEmitterFunctionRecord, error) {
	var out []any
	err := _EventEmitter.contract.Call(opts, &out, "echoFunctions", callback, note, fixedCallbacks, callbacks, record)

	if err != nil {
		return *new([68]byte), *new(string), *new([2][68]byte), *new([][68]byte), *new(EventEmitterFunctionRecord), err
	}

	out0 := *abi.ConvertType(out[0], new([68]byte)).(*[68]byte)
	out1 := *abi.ConvertType(out[1], new(string)).(*string)
	out2 := *abi.ConvertType(out[2], new([2][68]byte)).(*[2][68]byte)
	out3 := *abi.ConvertType(out[3], new([][68]byte)).(*[][68]byte)
	out4 := *abi.ConvertType(out[4], new(EventEmitterFunctionRecord)).(*EventEmitterFunctionRecord)

	return out0, out1, out2, out3, out4, err

}

// EchoFunctions is a free data retrieval call binding the contract method 0xe558a3a7.
//
// Hyperion: function echoFunctions(function callback, string note, function[2] fixedCallbacks, function[] callbacks, (function,string) record) pure returns(function, string, function[2], function[], (function,string))
func (_EventEmitter *EventEmitterSession) EchoFunctions(callback [68]byte, note string, fixedCallbacks [2][68]byte, callbacks [][68]byte, record EventEmitterFunctionRecord) ([68]byte, string, [2][68]byte, [][68]byte, EventEmitterFunctionRecord, error) {
	return _EventEmitter.Contract.EchoFunctions(&_EventEmitter.CallOpts, callback, note, fixedCallbacks, callbacks, record)
}

// EchoFunctions is a free data retrieval call binding the contract method 0xe558a3a7.
//
// Hyperion: function echoFunctions(function callback, string note, function[2] fixedCallbacks, function[] callbacks, (function,string) record) pure returns(function, string, function[2], function[], (function,string))
func (_EventEmitter *EventEmitterCallerSession) EchoFunctions(callback [68]byte, note string, fixedCallbacks [2][68]byte, callbacks [][68]byte, record EventEmitterFunctionRecord) ([68]byte, string, [2][68]byte, [][68]byte, EventEmitterFunctionRecord, error) {
	return _EventEmitter.Contract.EchoFunctions(&_EventEmitter.CallOpts, callback, note, fixedCallbacks, callbacks, record)
}

// EchoLeafContainers is a free data retrieval call binding the contract method 0x0753c06a.
//
// Hyperion: function echoLeafContainers(address[2] fixedAddresses, address[] addresses, bytes64[2] fixedTags, bytes64[] tags) pure returns(address[2], address[], bytes64[2], bytes64[])
func (_EventEmitter *EventEmitterCaller) EchoLeafContainers(opts *bind.CallOpts, fixedAddresses [2]common.Address, addresses []common.Address, fixedTags [2][64]byte, tags [][64]byte) ([2]common.Address, []common.Address, [2][64]byte, [][64]byte, error) {
	var out []any
	err := _EventEmitter.contract.Call(opts, &out, "echoLeafContainers", fixedAddresses, addresses, fixedTags, tags)

	if err != nil {
		return *new([2]common.Address), *new([]common.Address), *new([2][64]byte), *new([][64]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([2]common.Address)).(*[2]common.Address)
	out1 := *abi.ConvertType(out[1], new([]common.Address)).(*[]common.Address)
	out2 := *abi.ConvertType(out[2], new([2][64]byte)).(*[2][64]byte)
	out3 := *abi.ConvertType(out[3], new([][64]byte)).(*[][64]byte)

	return out0, out1, out2, out3, err

}

// EchoLeafContainers is a free data retrieval call binding the contract method 0x0753c06a.
//
// Hyperion: function echoLeafContainers(address[2] fixedAddresses, address[] addresses, bytes64[2] fixedTags, bytes64[] tags) pure returns(address[2], address[], bytes64[2], bytes64[])
func (_EventEmitter *EventEmitterSession) EchoLeafContainers(fixedAddresses [2]common.Address, addresses []common.Address, fixedTags [2][64]byte, tags [][64]byte) ([2]common.Address, []common.Address, [2][64]byte, [][64]byte, error) {
	return _EventEmitter.Contract.EchoLeafContainers(&_EventEmitter.CallOpts, fixedAddresses, addresses, fixedTags, tags)
}

// EchoLeafContainers is a free data retrieval call binding the contract method 0x0753c06a.
//
// Hyperion: function echoLeafContainers(address[2] fixedAddresses, address[] addresses, bytes64[2] fixedTags, bytes64[] tags) pure returns(address[2], address[], bytes64[2], bytes64[])
func (_EventEmitter *EventEmitterCallerSession) EchoLeafContainers(fixedAddresses [2]common.Address, addresses []common.Address, fixedTags [2][64]byte, tags [][64]byte) ([2]common.Address, []common.Address, [2][64]byte, [][64]byte, error) {
	return _EventEmitter.Contract.EchoLeafContainers(&_EventEmitter.CallOpts, fixedAddresses, addresses, fixedTags, tags)
}

// EchoNested is a free data retrieval call binding the contract method 0xf8041229.
//
// Hyperion: function echoNested((uint512,string,bytes,uint16[][]) record, (uint512,string,bytes,uint16[][])[] records, uint16[][][] cube) pure returns((uint512,string,bytes,uint16[][]), (uint512,string,bytes,uint16[][])[], uint16[][][])
func (_EventEmitter *EventEmitterCaller) EchoNested(opts *bind.CallOpts, record EventEmitterDynamicRecord, records []EventEmitterDynamicRecord, cube [][][]uint16) (EventEmitterDynamicRecord, []EventEmitterDynamicRecord, [][][]uint16, error) {
	var out []any
	err := _EventEmitter.contract.Call(opts, &out, "echoNested", record, records, cube)

	if err != nil {
		return *new(EventEmitterDynamicRecord), *new([]EventEmitterDynamicRecord), *new([][][]uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(EventEmitterDynamicRecord)).(*EventEmitterDynamicRecord)
	out1 := *abi.ConvertType(out[1], new([]EventEmitterDynamicRecord)).(*[]EventEmitterDynamicRecord)
	out2 := *abi.ConvertType(out[2], new([][][]uint16)).(*[][][]uint16)

	return out0, out1, out2, err

}

// EchoNested is a free data retrieval call binding the contract method 0xf8041229.
//
// Hyperion: function echoNested((uint512,string,bytes,uint16[][]) record, (uint512,string,bytes,uint16[][])[] records, uint16[][][] cube) pure returns((uint512,string,bytes,uint16[][]), (uint512,string,bytes,uint16[][])[], uint16[][][])
func (_EventEmitter *EventEmitterSession) EchoNested(record EventEmitterDynamicRecord, records []EventEmitterDynamicRecord, cube [][][]uint16) (EventEmitterDynamicRecord, []EventEmitterDynamicRecord, [][][]uint16, error) {
	return _EventEmitter.Contract.EchoNested(&_EventEmitter.CallOpts, record, records, cube)
}

// EchoNested is a free data retrieval call binding the contract method 0xf8041229.
//
// Hyperion: function echoNested((uint512,string,bytes,uint16[][]) record, (uint512,string,bytes,uint16[][])[] records, uint16[][][] cube) pure returns((uint512,string,bytes,uint16[][]), (uint512,string,bytes,uint16[][])[], uint16[][][])
func (_EventEmitter *EventEmitterCallerSession) EchoNested(record EventEmitterDynamicRecord, records []EventEmitterDynamicRecord, cube [][][]uint16) (EventEmitterDynamicRecord, []EventEmitterDynamicRecord, [][][]uint16, error) {
	return _EventEmitter.Contract.EchoNested(&_EventEmitter.CallOpts, record, records, cube)
}

// FailComplex is a free data retrieval call binding the contract method 0xfb144722.
//
// Hyperion: function failComplex(uint512 code, string reason, bytes payload, (uint512,address,bytes64) record, uint16[][] nested) pure returns()
func (_EventEmitter *EventEmitterCaller) FailComplex(opts *bind.CallOpts, code *big.Int, reason string, payload []byte, record EventEmitterRecord, nested [][]uint16) error {
	var out []any
	err := _EventEmitter.contract.Call(opts, &out, "failComplex", code, reason, payload, record, nested)

	if err != nil {
		return err
	}

	return err

}

// FailComplex is a free data retrieval call binding the contract method 0xfb144722.
//
// Hyperion: function failComplex(uint512 code, string reason, bytes payload, (uint512,address,bytes64) record, uint16[][] nested) pure returns()
func (_EventEmitter *EventEmitterSession) FailComplex(code *big.Int, reason string, payload []byte, record EventEmitterRecord, nested [][]uint16) error {
	return _EventEmitter.Contract.FailComplex(&_EventEmitter.CallOpts, code, reason, payload, record, nested)
}

// FailComplex is a free data retrieval call binding the contract method 0xfb144722.
//
// Hyperion: function failComplex(uint512 code, string reason, bytes payload, (uint512,address,bytes64) record, uint16[][] nested) pure returns()
func (_EventEmitter *EventEmitterCallerSession) FailComplex(code *big.Int, reason string, payload []byte, record EventEmitterRecord, nested [][]uint16) error {
	return _EventEmitter.Contract.FailComplex(&_EventEmitter.CallOpts, code, reason, payload, record, nested)
}

// FailPanic is a free data retrieval call binding the contract method 0xc66c9028.
//
// Hyperion: function failPanic() pure returns()
func (_EventEmitter *EventEmitterCaller) FailPanic(opts *bind.CallOpts) error {
	var out []any
	err := _EventEmitter.contract.Call(opts, &out, "failPanic")

	if err != nil {
		return err
	}

	return err

}

// FailPanic is a free data retrieval call binding the contract method 0xc66c9028.
//
// Hyperion: function failPanic() pure returns()
func (_EventEmitter *EventEmitterSession) FailPanic() error {
	return _EventEmitter.Contract.FailPanic(&_EventEmitter.CallOpts)
}

// FailPanic is a free data retrieval call binding the contract method 0xc66c9028.
//
// Hyperion: function failPanic() pure returns()
func (_EventEmitter *EventEmitterCallerSession) FailPanic() error {
	return _EventEmitter.Contract.FailPanic(&_EventEmitter.CallOpts)
}

// FailReason is a free data retrieval call binding the contract method 0xb0b75436.
//
// Hyperion: function failReason() pure returns()
func (_EventEmitter *EventEmitterCaller) FailReason(opts *bind.CallOpts) error {
	var out []any
	err := _EventEmitter.contract.Call(opts, &out, "failReason")

	if err != nil {
		return err
	}

	return err

}

// FailReason is a free data retrieval call binding the contract method 0xb0b75436.
//
// Hyperion: function failReason() pure returns()
func (_EventEmitter *EventEmitterSession) FailReason() error {
	return _EventEmitter.Contract.FailReason(&_EventEmitter.CallOpts)
}

// FailReason is a free data retrieval call binding the contract method 0xb0b75436.
//
// Hyperion: function failReason() pure returns()
func (_EventEmitter *EventEmitterCallerSession) FailReason() error {
	return _EventEmitter.Contract.FailReason(&_EventEmitter.CallOpts)
}

// Observe is a free data retrieval call binding the contract method 0x14fc78fc.
//
// Hyperion: function observe() view returns(uint512 value, address caller)
func (_EventEmitter *EventEmitterCaller) Observe(opts *bind.CallOpts) (struct {
	Value  *big.Int
	Caller common.Address
}, error) {
	var out []any
	err := _EventEmitter.contract.Call(opts, &out, "observe")

	outstruct := new(struct {
		Value  *big.Int
		Caller common.Address
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Value = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Caller = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)

	return *outstruct, err

}

// Observe is a free data retrieval call binding the contract method 0x14fc78fc.
//
// Hyperion: function observe() view returns(uint512 value, address caller)
func (_EventEmitter *EventEmitterSession) Observe() (struct {
	Value  *big.Int
	Caller common.Address
}, error) {
	return _EventEmitter.Contract.Observe(&_EventEmitter.CallOpts)
}

// Observe is a free data retrieval call binding the contract method 0x14fc78fc.
//
// Hyperion: function observe() view returns(uint512 value, address caller)
func (_EventEmitter *EventEmitterCallerSession) Observe() (struct {
	Value  *big.Int
	Caller common.Address
}, error) {
	return _EventEmitter.Contract.Observe(&_EventEmitter.CallOpts)
}

// PlusOne is a free data retrieval call binding the contract method 0x79531c40.
//
// Hyperion: function plusOne(uint512 value) pure returns(uint512)
func (_EventEmitter *EventEmitterCaller) PlusOne(opts *bind.CallOpts, value *big.Int) (*big.Int, error) {
	var out []any
	err := _EventEmitter.contract.Call(opts, &out, "plusOne", value)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PlusOne is a free data retrieval call binding the contract method 0x79531c40.
//
// Hyperion: function plusOne(uint512 value) pure returns(uint512)
func (_EventEmitter *EventEmitterSession) PlusOne(value *big.Int) (*big.Int, error) {
	return _EventEmitter.Contract.PlusOne(&_EventEmitter.CallOpts, value)
}

// PlusOne is a free data retrieval call binding the contract method 0x79531c40.
//
// Hyperion: function plusOne(uint512 value) pure returns(uint512)
func (_EventEmitter *EventEmitterCallerSession) PlusOne(value *big.Int) (*big.Int, error) {
	return _EventEmitter.Contract.PlusOne(&_EventEmitter.CallOpts, value)
}

// Transform is a free data retrieval call binding the contract method 0x9e420a8f.
//
// Hyperion: function transform(string value) pure returns(string)
func (_EventEmitter *EventEmitterCaller) Transform(opts *bind.CallOpts, value string) (string, error) {
	var out []any
	err := _EventEmitter.contract.Call(opts, &out, "transform", value)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Transform is a free data retrieval call binding the contract method 0x9e420a8f.
//
// Hyperion: function transform(string value) pure returns(string)
func (_EventEmitter *EventEmitterSession) Transform(value string) (string, error) {
	return _EventEmitter.Contract.Transform(&_EventEmitter.CallOpts, value)
}

// Transform is a free data retrieval call binding the contract method 0x9e420a8f.
//
// Hyperion: function transform(string value) pure returns(string)
func (_EventEmitter *EventEmitterCallerSession) Transform(value string) (string, error) {
	return _EventEmitter.Contract.Transform(&_EventEmitter.CallOpts, value)
}

// Transform0 is a free data retrieval call binding the contract method 0xed928c96.
//
// Hyperion: function transform(uint16 value) pure returns(uint16)
func (_EventEmitter *EventEmitterCaller) Transform0(opts *bind.CallOpts, value uint16) (uint16, error) {
	var out []any
	err := _EventEmitter.contract.Call(opts, &out, "transform0", value)

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// Transform0 is a free data retrieval call binding the contract method 0xed928c96.
//
// Hyperion: function transform(uint16 value) pure returns(uint16)
func (_EventEmitter *EventEmitterSession) Transform0(value uint16) (uint16, error) {
	return _EventEmitter.Contract.Transform0(&_EventEmitter.CallOpts, value)
}

// Transform0 is a free data retrieval call binding the contract method 0xed928c96.
//
// Hyperion: function transform(uint16 value) pure returns(uint16)
func (_EventEmitter *EventEmitterCallerSession) Transform0(value uint16) (uint16, error) {
	return _EventEmitter.Contract.Transform0(&_EventEmitter.CallOpts, value)
}

// EmitComposite is a paid mutator transaction binding the contract method 0x99cf235f.
//
// Hyperion: function emitComposite((uint512,string,bytes,uint16[][]) record, uint16[3] fixedNumbers, string[2] fixedStrings, uint16[][2] mixed) returns()
func (_EventEmitter *EventEmitterTransactor) EmitComposite(opts *bind.TransactOpts, record EventEmitterDynamicRecord, fixedNumbers [3]uint16, fixedStrings [2]string, mixed [2][]uint16) (*types.Transaction, error) {
	return _EventEmitter.contract.Transact(opts, "emitComposite", record, fixedNumbers, fixedStrings, mixed)
}

// EmitComposite is a paid mutator transaction binding the contract method 0x99cf235f.
//
// Hyperion: function emitComposite((uint512,string,bytes,uint16[][]) record, uint16[3] fixedNumbers, string[2] fixedStrings, uint16[][2] mixed) returns()
func (_EventEmitter *EventEmitterSession) EmitComposite(record EventEmitterDynamicRecord, fixedNumbers [3]uint16, fixedStrings [2]string, mixed [2][]uint16) (*types.Transaction, error) {
	return _EventEmitter.Contract.EmitComposite(&_EventEmitter.TransactOpts, record, fixedNumbers, fixedStrings, mixed)
}

// EmitComposite is a paid mutator transaction binding the contract method 0x99cf235f.
//
// Hyperion: function emitComposite((uint512,string,bytes,uint16[][]) record, uint16[3] fixedNumbers, string[2] fixedStrings, uint16[][2] mixed) returns()
func (_EventEmitter *EventEmitterTransactorSession) EmitComposite(record EventEmitterDynamicRecord, fixedNumbers [3]uint16, fixedStrings [2]string, mixed [2][]uint16) (*types.Transaction, error) {
	return _EventEmitter.Contract.EmitComposite(&_EventEmitter.TransactOpts, record, fixedNumbers, fixedStrings, mixed)
}

// EmitIndexedScalars is a paid mutator transaction binding the contract method 0xb94d6fa6.
//
// Hyperion: function emitIndexedScalars(bool flag, bytes5 code, int16 delta) returns()
func (_EventEmitter *EventEmitterTransactor) EmitIndexedScalars(opts *bind.TransactOpts, flag bool, code [5]byte, delta int16) (*types.Transaction, error) {
	return _EventEmitter.contract.Transact(opts, "emitIndexedScalars", flag, code, delta)
}

// EmitIndexedScalars is a paid mutator transaction binding the contract method 0xb94d6fa6.
//
// Hyperion: function emitIndexedScalars(bool flag, bytes5 code, int16 delta) returns()
func (_EventEmitter *EventEmitterSession) EmitIndexedScalars(flag bool, code [5]byte, delta int16) (*types.Transaction, error) {
	return _EventEmitter.Contract.EmitIndexedScalars(&_EventEmitter.TransactOpts, flag, code, delta)
}

// EmitIndexedScalars is a paid mutator transaction binding the contract method 0xb94d6fa6.
//
// Hyperion: function emitIndexedScalars(bool flag, bytes5 code, int16 delta) returns()
func (_EventEmitter *EventEmitterTransactorSession) EmitIndexedScalars(flag bool, code [5]byte, delta int16) (*types.Transaction, error) {
	return _EventEmitter.Contract.EmitIndexedScalars(&_EventEmitter.TransactOpts, flag, code, delta)
}

// EmitTransformed is a paid mutator transaction binding the contract method 0x1e3ed7e4.
//
// Hyperion: function emitTransformed(string value) returns()
func (_EventEmitter *EventEmitterTransactor) EmitTransformed(opts *bind.TransactOpts, value string) (*types.Transaction, error) {
	return _EventEmitter.contract.Transact(opts, "emitTransformed", value)
}

// EmitTransformed is a paid mutator transaction binding the contract method 0x1e3ed7e4.
//
// Hyperion: function emitTransformed(string value) returns()
func (_EventEmitter *EventEmitterSession) EmitTransformed(value string) (*types.Transaction, error) {
	return _EventEmitter.Contract.EmitTransformed(&_EventEmitter.TransactOpts, value)
}

// EmitTransformed is a paid mutator transaction binding the contract method 0x1e3ed7e4.
//
// Hyperion: function emitTransformed(string value) returns()
func (_EventEmitter *EventEmitterTransactorSession) EmitTransformed(value string) (*types.Transaction, error) {
	return _EventEmitter.Contract.EmitTransformed(&_EventEmitter.TransactOpts, value)
}

// EmitTransformed0 is a paid mutator transaction binding the contract method 0xf404ae99.
//
// Hyperion: function emitTransformed(uint16 value) returns()
func (_EventEmitter *EventEmitterTransactor) EmitTransformed0(opts *bind.TransactOpts, value uint16) (*types.Transaction, error) {
	return _EventEmitter.contract.Transact(opts, "emitTransformed0", value)
}

// EmitTransformed0 is a paid mutator transaction binding the contract method 0xf404ae99.
//
// Hyperion: function emitTransformed(uint16 value) returns()
func (_EventEmitter *EventEmitterSession) EmitTransformed0(value uint16) (*types.Transaction, error) {
	return _EventEmitter.Contract.EmitTransformed0(&_EventEmitter.TransactOpts, value)
}

// EmitTransformed0 is a paid mutator transaction binding the contract method 0xf404ae99.
//
// Hyperion: function emitTransformed(uint16 value) returns()
func (_EventEmitter *EventEmitterTransactorSession) EmitTransformed0(value uint16) (*types.Transaction, error) {
	return _EventEmitter.Contract.EmitTransformed0(&_EventEmitter.TransactOpts, value)
}

// ExerciseFunction is a paid mutator transaction binding the contract method 0xa43e73c9.
//
// Hyperion: function exerciseFunction(function callback, uint512 value) returns(function, uint512)
func (_EventEmitter *EventEmitterTransactor) ExerciseFunction(opts *bind.TransactOpts, callback [68]byte, value *big.Int) (*types.Transaction, error) {
	return _EventEmitter.contract.Transact(opts, "exerciseFunction", callback, value)
}

// ExerciseFunction is a paid mutator transaction binding the contract method 0xa43e73c9.
//
// Hyperion: function exerciseFunction(function callback, uint512 value) returns(function, uint512)
func (_EventEmitter *EventEmitterSession) ExerciseFunction(callback [68]byte, value *big.Int) (*types.Transaction, error) {
	return _EventEmitter.Contract.ExerciseFunction(&_EventEmitter.TransactOpts, callback, value)
}

// ExerciseFunction is a paid mutator transaction binding the contract method 0xa43e73c9.
//
// Hyperion: function exerciseFunction(function callback, uint512 value) returns(function, uint512)
func (_EventEmitter *EventEmitterTransactorSession) ExerciseFunction(callback [68]byte, value *big.Int) (*types.Transaction, error) {
	return _EventEmitter.Contract.ExerciseFunction(&_EventEmitter.TransactOpts, callback, value)
}

// Pay is a paid mutator transaction binding the contract method 0x2fb0dbcd.
//
// Hyperion: function pay(uint16 marker) payable returns()
func (_EventEmitter *EventEmitterTransactor) Pay(opts *bind.TransactOpts, marker uint16) (*types.Transaction, error) {
	return _EventEmitter.contract.Transact(opts, "pay", marker)
}

// Pay is a paid mutator transaction binding the contract method 0x2fb0dbcd.
//
// Hyperion: function pay(uint16 marker) payable returns()
func (_EventEmitter *EventEmitterSession) Pay(marker uint16) (*types.Transaction, error) {
	return _EventEmitter.Contract.Pay(&_EventEmitter.TransactOpts, marker)
}

// Pay is a paid mutator transaction binding the contract method 0x2fb0dbcd.
//
// Hyperion: function pay(uint16 marker) payable returns()
func (_EventEmitter *EventEmitterTransactorSession) Pay(marker uint16) (*types.Transaction, error) {
	return _EventEmitter.Contract.Pay(&_EventEmitter.TransactOpts, marker)
}

// Store is a paid mutator transaction binding the contract method 0x3d0e1089.
//
// Hyperion: function store(uint512 amount, int512 delta, bytes64 tag, address recipient, bytes payload, string note, bool enabled) returns()
func (_EventEmitter *EventEmitterTransactor) Store(opts *bind.TransactOpts, amount *big.Int, delta *big.Int, tag [64]byte, recipient common.Address, payload []byte, note string, enabled bool) (*types.Transaction, error) {
	return _EventEmitter.contract.Transact(opts, "store", amount, delta, tag, recipient, payload, note, enabled)
}

// Store is a paid mutator transaction binding the contract method 0x3d0e1089.
//
// Hyperion: function store(uint512 amount, int512 delta, bytes64 tag, address recipient, bytes payload, string note, bool enabled) returns()
func (_EventEmitter *EventEmitterSession) Store(amount *big.Int, delta *big.Int, tag [64]byte, recipient common.Address, payload []byte, note string, enabled bool) (*types.Transaction, error) {
	return _EventEmitter.Contract.Store(&_EventEmitter.TransactOpts, amount, delta, tag, recipient, payload, note, enabled)
}

// Store is a paid mutator transaction binding the contract method 0x3d0e1089.
//
// Hyperion: function store(uint512 amount, int512 delta, bytes64 tag, address recipient, bytes payload, string note, bool enabled) returns()
func (_EventEmitter *EventEmitterTransactorSession) Store(amount *big.Int, delta *big.Int, tag [64]byte, recipient common.Address, payload []byte, note string, enabled bool) (*types.Transaction, error) {
	return _EventEmitter.Contract.Store(&_EventEmitter.TransactOpts, amount, delta, tag, recipient, payload, note, enabled)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Hyperion: fallback() payable returns()
func (_EventEmitter *EventEmitterTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _EventEmitter.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Hyperion: fallback() payable returns()
func (_EventEmitter *EventEmitterSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _EventEmitter.Contract.Fallback(&_EventEmitter.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Hyperion: fallback() payable returns()
func (_EventEmitter *EventEmitterTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _EventEmitter.Contract.Fallback(&_EventEmitter.TransactOpts, calldata)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Hyperion: receive() payable returns()
func (_EventEmitter *EventEmitterTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EventEmitter.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Hyperion: receive() payable returns()
func (_EventEmitter *EventEmitterSession) Receive() (*types.Transaction, error) {
	return _EventEmitter.Contract.Receive(&_EventEmitter.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Hyperion: receive() payable returns()
func (_EventEmitter *EventEmitterTransactorSession) Receive() (*types.Transaction, error) {
	return _EventEmitter.Contract.Receive(&_EventEmitter.TransactOpts)
}

// EventEmitterCompositeIterator is returned from FilterComposite and is used to iterate over the raw logs and unpacked data for Composite events raised by the EventEmitter contract.
type EventEmitterCompositeIterator struct {
	Event *EventEmitterComposite // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log   // Log channel receiving the found contract events
	sub  qrl.Subscription // Subscription for errors, completion and termination
	done bool             // Whether the subscription completed delivering logs
	fail error            // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EventEmitterCompositeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EventEmitterComposite)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EventEmitterComposite)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EventEmitterCompositeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EventEmitterCompositeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EventEmitterComposite represents a Composite event raised by the EventEmitter contract.
type EventEmitterComposite struct {
	Record       EventEmitterDynamicRecord
	FixedNumbers [3]uint16
	FixedStrings [2]string
	Mixed        [2][]uint16
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterComposite is a free log retrieval operation binding the contract event 0xa5d75a0c1afe7b82fadd41d2ad66b9fb0934d4aa36a730e5da8b22ae04352e0e.
//
// Hyperion: event Composite((uint512,string,bytes,uint16[][]) record, uint16[3] fixedNumbers, string[2] fixedStrings, uint16[][2] mixed)
func (_EventEmitter *EventEmitterFilterer) FilterComposite(opts *bind.FilterOpts) (*EventEmitterCompositeIterator, error) {

	logs, sub, err := _EventEmitter.contract.FilterLogs(opts, "Composite")
	if err != nil {
		return nil, err
	}
	return &EventEmitterCompositeIterator{contract: _EventEmitter.contract, event: "Composite", logs: logs, sub: sub}, nil
}

// WatchComposite is a free log subscription operation binding the contract event 0xa5d75a0c1afe7b82fadd41d2ad66b9fb0934d4aa36a730e5da8b22ae04352e0e.
//
// Hyperion: event Composite((uint512,string,bytes,uint16[][]) record, uint16[3] fixedNumbers, string[2] fixedStrings, uint16[][2] mixed)
func (_EventEmitter *EventEmitterFilterer) WatchComposite(opts *bind.WatchOpts, sink chan<- *EventEmitterComposite) (event.Subscription, error) {

	logs, sub, err := _EventEmitter.contract.WatchLogs(opts, "Composite")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EventEmitterComposite)
				if err := _EventEmitter.contract.UnpackLog(event, "Composite", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseComposite is a log parse operation binding the contract event 0xa5d75a0c1afe7b82fadd41d2ad66b9fb0934d4aa36a730e5da8b22ae04352e0e.
//
// Hyperion: event Composite((uint512,string,bytes,uint16[][]) record, uint16[3] fixedNumbers, string[2] fixedStrings, uint16[][2] mixed)
func (_EventEmitter *EventEmitterFilterer) ParseComposite(log types.Log) (*EventEmitterComposite, error) {
	event := new(EventEmitterComposite)
	if err := _EventEmitter.contract.UnpackLog(event, "Composite", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EventEmitterDeployedIterator is returned from FilterDeployed and is used to iterate over the raw logs and unpacked data for Deployed events raised by the EventEmitter contract.
type EventEmitterDeployedIterator struct {
	Event *EventEmitterDeployed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log   // Log channel receiving the found contract events
	sub  qrl.Subscription // Subscription for errors, completion and termination
	done bool             // Whether the subscription completed delivering logs
	fail error            // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EventEmitterDeployedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EventEmitterDeployed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EventEmitterDeployed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EventEmitterDeployedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EventEmitterDeployedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EventEmitterDeployed represents a Deployed event raised by the EventEmitter contract.
type EventEmitterDeployed struct {
	Value   *big.Int
	Note    string
	Payload []byte
	Record  EventEmitterRecord
	Numbers []uint16
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterDeployed is a free log retrieval operation binding the contract event 0x100ec3f67dbb9bac62c3c46aa9e2acda3788c474549e77c16e730ee2d023be00.
//
// Hyperion: event Deployed(uint512 value, string note, bytes payload, (uint512,address,bytes64) record, uint16[] numbers)
func (_EventEmitter *EventEmitterFilterer) FilterDeployed(opts *bind.FilterOpts) (*EventEmitterDeployedIterator, error) {

	logs, sub, err := _EventEmitter.contract.FilterLogs(opts, "Deployed")
	if err != nil {
		return nil, err
	}
	return &EventEmitterDeployedIterator{contract: _EventEmitter.contract, event: "Deployed", logs: logs, sub: sub}, nil
}

// WatchDeployed is a free log subscription operation binding the contract event 0x100ec3f67dbb9bac62c3c46aa9e2acda3788c474549e77c16e730ee2d023be00.
//
// Hyperion: event Deployed(uint512 value, string note, bytes payload, (uint512,address,bytes64) record, uint16[] numbers)
func (_EventEmitter *EventEmitterFilterer) WatchDeployed(opts *bind.WatchOpts, sink chan<- *EventEmitterDeployed) (event.Subscription, error) {

	logs, sub, err := _EventEmitter.contract.WatchLogs(opts, "Deployed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EventEmitterDeployed)
				if err := _EventEmitter.contract.UnpackLog(event, "Deployed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDeployed is a log parse operation binding the contract event 0x100ec3f67dbb9bac62c3c46aa9e2acda3788c474549e77c16e730ee2d023be00.
//
// Hyperion: event Deployed(uint512 value, string note, bytes payload, (uint512,address,bytes64) record, uint16[] numbers)
func (_EventEmitter *EventEmitterFilterer) ParseDeployed(log types.Log) (*EventEmitterDeployed, error) {
	event := new(EventEmitterDeployed)
	if err := _EventEmitter.contract.UnpackLog(event, "Deployed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EventEmitterDynamicIterator is returned from FilterDynamic and is used to iterate over the raw logs and unpacked data for Dynamic events raised by the EventEmitter contract.
type EventEmitterDynamicIterator struct {
	Event *EventEmitterDynamic // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log   // Log channel receiving the found contract events
	sub  qrl.Subscription // Subscription for errors, completion and termination
	done bool             // Whether the subscription completed delivering logs
	fail error            // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EventEmitterDynamicIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EventEmitterDynamic)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EventEmitterDynamic)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EventEmitterDynamicIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EventEmitterDynamicIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EventEmitterDynamic represents a Dynamic event raised by the EventEmitter contract.
type EventEmitterDynamic struct {
	Payload common.Hash
	Note    common.Hash
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterDynamic is a free log retrieval operation binding the contract event 0x4ef7447df163d4aaeab9c66fa93651de5eebb002dcf9b60da1ebaa28ae95e825.
//
// Hyperion: event Dynamic(bytes indexed payload, string indexed note, uint512 amount)
func (_EventEmitter *EventEmitterFilterer) FilterDynamic(opts *bind.FilterOpts, payload [][]byte, note []string) (*EventEmitterDynamicIterator, error) {

	var payloadRule []any
	for _, payloadItem := range payload {
		payloadRule = append(payloadRule, payloadItem)
	}
	var noteRule []any
	for _, noteItem := range note {
		noteRule = append(noteRule, noteItem)
	}

	logs, sub, err := _EventEmitter.contract.FilterLogs(opts, "Dynamic", payloadRule, noteRule)
	if err != nil {
		return nil, err
	}
	return &EventEmitterDynamicIterator{contract: _EventEmitter.contract, event: "Dynamic", logs: logs, sub: sub}, nil
}

// WatchDynamic is a free log subscription operation binding the contract event 0x4ef7447df163d4aaeab9c66fa93651de5eebb002dcf9b60da1ebaa28ae95e825.
//
// Hyperion: event Dynamic(bytes indexed payload, string indexed note, uint512 amount)
func (_EventEmitter *EventEmitterFilterer) WatchDynamic(opts *bind.WatchOpts, sink chan<- *EventEmitterDynamic, payload [][]byte, note []string) (event.Subscription, error) {

	var payloadRule []any
	for _, payloadItem := range payload {
		payloadRule = append(payloadRule, payloadItem)
	}
	var noteRule []any
	for _, noteItem := range note {
		noteRule = append(noteRule, noteItem)
	}

	logs, sub, err := _EventEmitter.contract.WatchLogs(opts, "Dynamic", payloadRule, noteRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EventEmitterDynamic)
				if err := _EventEmitter.contract.UnpackLog(event, "Dynamic", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDynamic is a log parse operation binding the contract event 0x4ef7447df163d4aaeab9c66fa93651de5eebb002dcf9b60da1ebaa28ae95e825.
//
// Hyperion: event Dynamic(bytes indexed payload, string indexed note, uint512 amount)
func (_EventEmitter *EventEmitterFilterer) ParseDynamic(log types.Log) (*EventEmitterDynamic, error) {
	event := new(EventEmitterDynamic)
	if err := _EventEmitter.contract.UnpackLog(event, "Dynamic", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EventEmitterFallbackCalledIterator is returned from FilterFallbackCalled and is used to iterate over the raw logs and unpacked data for FallbackCalled events raised by the EventEmitter contract.
type EventEmitterFallbackCalledIterator struct {
	Event *EventEmitterFallbackCalled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log   // Log channel receiving the found contract events
	sub  qrl.Subscription // Subscription for errors, completion and termination
	done bool             // Whether the subscription completed delivering logs
	fail error            // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EventEmitterFallbackCalledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EventEmitterFallbackCalled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EventEmitterFallbackCalled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EventEmitterFallbackCalledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EventEmitterFallbackCalledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EventEmitterFallbackCalled represents a FallbackCalled event raised by the EventEmitter contract.
type EventEmitterFallbackCalled struct {
	Payload []byte
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterFallbackCalled is a free log retrieval operation binding the contract event 0xe5b92b8ba08394dd9b027fafca0dc888f149e8f420b55893ecee14ea148aa08b.
//
// Hyperion: event FallbackCalled(bytes payload, uint256 amount)
func (_EventEmitter *EventEmitterFilterer) FilterFallbackCalled(opts *bind.FilterOpts) (*EventEmitterFallbackCalledIterator, error) {

	logs, sub, err := _EventEmitter.contract.FilterLogs(opts, "FallbackCalled")
	if err != nil {
		return nil, err
	}
	return &EventEmitterFallbackCalledIterator{contract: _EventEmitter.contract, event: "FallbackCalled", logs: logs, sub: sub}, nil
}

// WatchFallbackCalled is a free log subscription operation binding the contract event 0xe5b92b8ba08394dd9b027fafca0dc888f149e8f420b55893ecee14ea148aa08b.
//
// Hyperion: event FallbackCalled(bytes payload, uint256 amount)
func (_EventEmitter *EventEmitterFilterer) WatchFallbackCalled(opts *bind.WatchOpts, sink chan<- *EventEmitterFallbackCalled) (event.Subscription, error) {

	logs, sub, err := _EventEmitter.contract.WatchLogs(opts, "FallbackCalled")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EventEmitterFallbackCalled)
				if err := _EventEmitter.contract.UnpackLog(event, "FallbackCalled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFallbackCalled is a log parse operation binding the contract event 0xe5b92b8ba08394dd9b027fafca0dc888f149e8f420b55893ecee14ea148aa08b.
//
// Hyperion: event FallbackCalled(bytes payload, uint256 amount)
func (_EventEmitter *EventEmitterFilterer) ParseFallbackCalled(log types.Log) (*EventEmitterFallbackCalled, error) {
	event := new(EventEmitterFallbackCalled)
	if err := _EventEmitter.contract.UnpackLog(event, "FallbackCalled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EventEmitterFunctionObservedIterator is returned from FilterFunctionObserved and is used to iterate over the raw logs and unpacked data for FunctionObserved events raised by the EventEmitter contract.
type EventEmitterFunctionObservedIterator struct {
	Event *EventEmitterFunctionObserved // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log   // Log channel receiving the found contract events
	sub  qrl.Subscription // Subscription for errors, completion and termination
	done bool             // Whether the subscription completed delivering logs
	fail error            // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EventEmitterFunctionObservedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EventEmitterFunctionObserved)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EventEmitterFunctionObserved)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EventEmitterFunctionObservedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EventEmitterFunctionObservedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EventEmitterFunctionObserved represents a FunctionObserved event raised by the EventEmitter contract.
type EventEmitterFunctionObserved struct {
	IndexedCallback common.Hash
	Callback        [68]byte
	Result          *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterFunctionObserved is a free log retrieval operation binding the contract event 0xfa178067c55becdc50555038a0302191054dc26036b1e5ad5d3d0a9daa93423c.
//
// Hyperion: event FunctionObserved(function indexed indexedCallback, function callback, uint512 result)
func (_EventEmitter *EventEmitterFilterer) FilterFunctionObserved(opts *bind.FilterOpts, indexedCallback [][68]byte) (*EventEmitterFunctionObservedIterator, error) {

	var indexedCallbackRule []any
	for _, indexedCallbackItem := range indexedCallback {
		indexedCallbackRule = append(indexedCallbackRule, indexedCallbackItem)
	}

	logs, sub, err := _EventEmitter.contract.FilterLogs(opts, "FunctionObserved", indexedCallbackRule)
	if err != nil {
		return nil, err
	}
	return &EventEmitterFunctionObservedIterator{contract: _EventEmitter.contract, event: "FunctionObserved", logs: logs, sub: sub}, nil
}

// WatchFunctionObserved is a free log subscription operation binding the contract event 0xfa178067c55becdc50555038a0302191054dc26036b1e5ad5d3d0a9daa93423c.
//
// Hyperion: event FunctionObserved(function indexed indexedCallback, function callback, uint512 result)
func (_EventEmitter *EventEmitterFilterer) WatchFunctionObserved(opts *bind.WatchOpts, sink chan<- *EventEmitterFunctionObserved, indexedCallback [][68]byte) (event.Subscription, error) {

	var indexedCallbackRule []any
	for _, indexedCallbackItem := range indexedCallback {
		indexedCallbackRule = append(indexedCallbackRule, indexedCallbackItem)
	}

	logs, sub, err := _EventEmitter.contract.WatchLogs(opts, "FunctionObserved", indexedCallbackRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EventEmitterFunctionObserved)
				if err := _EventEmitter.contract.UnpackLog(event, "FunctionObserved", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFunctionObserved is a log parse operation binding the contract event 0xfa178067c55becdc50555038a0302191054dc26036b1e5ad5d3d0a9daa93423c.
//
// Hyperion: event FunctionObserved(function indexed indexedCallback, function callback, uint512 result)
func (_EventEmitter *EventEmitterFilterer) ParseFunctionObserved(log types.Log) (*EventEmitterFunctionObserved, error) {
	event := new(EventEmitterFunctionObserved)
	if err := _EventEmitter.contract.UnpackLog(event, "FunctionObserved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EventEmitterIndexedScalarsIterator is returned from FilterIndexedScalars and is used to iterate over the raw logs and unpacked data for IndexedScalars events raised by the EventEmitter contract.
type EventEmitterIndexedScalarsIterator struct {
	Event *EventEmitterIndexedScalars // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log   // Log channel receiving the found contract events
	sub  qrl.Subscription // Subscription for errors, completion and termination
	done bool             // Whether the subscription completed delivering logs
	fail error            // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EventEmitterIndexedScalarsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EventEmitterIndexedScalars)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EventEmitterIndexedScalars)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EventEmitterIndexedScalarsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EventEmitterIndexedScalarsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EventEmitterIndexedScalars represents a IndexedScalars event raised by the EventEmitter contract.
type EventEmitterIndexedScalars struct {
	Flag  bool
	Code  [5]byte
	Delta int16
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterIndexedScalars is a free log retrieval operation binding the contract event 0x19c59af463d0b89e6afb02db53c6ea998a04ce7bf1aa5c2c0d4c3ac9efc9e659.
//
// Hyperion: event IndexedScalars(bool indexed flag, bytes5 indexed code, int16 indexed delta)
func (_EventEmitter *EventEmitterFilterer) FilterIndexedScalars(opts *bind.FilterOpts, flag []bool, code [][5]byte, delta []int16) (*EventEmitterIndexedScalarsIterator, error) {

	var flagRule []any
	for _, flagItem := range flag {
		flagRule = append(flagRule, flagItem)
	}
	var codeRule []any
	for _, codeItem := range code {
		codeRule = append(codeRule, codeItem)
	}
	var deltaRule []any
	for _, deltaItem := range delta {
		deltaRule = append(deltaRule, deltaItem)
	}

	logs, sub, err := _EventEmitter.contract.FilterLogs(opts, "IndexedScalars", flagRule, codeRule, deltaRule)
	if err != nil {
		return nil, err
	}
	return &EventEmitterIndexedScalarsIterator{contract: _EventEmitter.contract, event: "IndexedScalars", logs: logs, sub: sub}, nil
}

// WatchIndexedScalars is a free log subscription operation binding the contract event 0x19c59af463d0b89e6afb02db53c6ea998a04ce7bf1aa5c2c0d4c3ac9efc9e659.
//
// Hyperion: event IndexedScalars(bool indexed flag, bytes5 indexed code, int16 indexed delta)
func (_EventEmitter *EventEmitterFilterer) WatchIndexedScalars(opts *bind.WatchOpts, sink chan<- *EventEmitterIndexedScalars, flag []bool, code [][5]byte, delta []int16) (event.Subscription, error) {

	var flagRule []any
	for _, flagItem := range flag {
		flagRule = append(flagRule, flagItem)
	}
	var codeRule []any
	for _, codeItem := range code {
		codeRule = append(codeRule, codeItem)
	}
	var deltaRule []any
	for _, deltaItem := range delta {
		deltaRule = append(deltaRule, deltaItem)
	}

	logs, sub, err := _EventEmitter.contract.WatchLogs(opts, "IndexedScalars", flagRule, codeRule, deltaRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EventEmitterIndexedScalars)
				if err := _EventEmitter.contract.UnpackLog(event, "IndexedScalars", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseIndexedScalars is a log parse operation binding the contract event 0x19c59af463d0b89e6afb02db53c6ea998a04ce7bf1aa5c2c0d4c3ac9efc9e659.
//
// Hyperion: event IndexedScalars(bool indexed flag, bytes5 indexed code, int16 indexed delta)
func (_EventEmitter *EventEmitterFilterer) ParseIndexedScalars(log types.Log) (*EventEmitterIndexedScalars, error) {
	event := new(EventEmitterIndexedScalars)
	if err := _EventEmitter.contract.UnpackLog(event, "IndexedScalars", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EventEmitterPaidIterator is returned from FilterPaid and is used to iterate over the raw logs and unpacked data for Paid events raised by the EventEmitter contract.
type EventEmitterPaidIterator struct {
	Event *EventEmitterPaid // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log   // Log channel receiving the found contract events
	sub  qrl.Subscription // Subscription for errors, completion and termination
	done bool             // Whether the subscription completed delivering logs
	fail error            // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EventEmitterPaidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EventEmitterPaid)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EventEmitterPaid)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EventEmitterPaidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EventEmitterPaidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EventEmitterPaid represents a Paid event raised by the EventEmitter contract.
type EventEmitterPaid struct {
	Sender common.Address
	Marker uint16
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterPaid is a free log retrieval operation binding the contract event 0x1398d89bb96c43f8c16ef74dee904b456a4fa8a5857191293b848ced1997a3d9.
//
// Hyperion: event Paid(address indexed sender, uint16 indexed marker, uint256 amount)
func (_EventEmitter *EventEmitterFilterer) FilterPaid(opts *bind.FilterOpts, sender []common.Address, marker []uint16) (*EventEmitterPaidIterator, error) {

	var senderRule []any
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var markerRule []any
	for _, markerItem := range marker {
		markerRule = append(markerRule, markerItem)
	}

	logs, sub, err := _EventEmitter.contract.FilterLogs(opts, "Paid", senderRule, markerRule)
	if err != nil {
		return nil, err
	}
	return &EventEmitterPaidIterator{contract: _EventEmitter.contract, event: "Paid", logs: logs, sub: sub}, nil
}

// WatchPaid is a free log subscription operation binding the contract event 0x1398d89bb96c43f8c16ef74dee904b456a4fa8a5857191293b848ced1997a3d9.
//
// Hyperion: event Paid(address indexed sender, uint16 indexed marker, uint256 amount)
func (_EventEmitter *EventEmitterFilterer) WatchPaid(opts *bind.WatchOpts, sink chan<- *EventEmitterPaid, sender []common.Address, marker []uint16) (event.Subscription, error) {

	var senderRule []any
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var markerRule []any
	for _, markerItem := range marker {
		markerRule = append(markerRule, markerItem)
	}

	logs, sub, err := _EventEmitter.contract.WatchLogs(opts, "Paid", senderRule, markerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EventEmitterPaid)
				if err := _EventEmitter.contract.UnpackLog(event, "Paid", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePaid is a log parse operation binding the contract event 0x1398d89bb96c43f8c16ef74dee904b456a4fa8a5857191293b848ced1997a3d9.
//
// Hyperion: event Paid(address indexed sender, uint16 indexed marker, uint256 amount)
func (_EventEmitter *EventEmitterFilterer) ParsePaid(log types.Log) (*EventEmitterPaid, error) {
	event := new(EventEmitterPaid)
	if err := _EventEmitter.contract.UnpackLog(event, "Paid", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EventEmitterReceivedIterator is returned from FilterReceived and is used to iterate over the raw logs and unpacked data for Received events raised by the EventEmitter contract.
type EventEmitterReceivedIterator struct {
	Event *EventEmitterReceived // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log   // Log channel receiving the found contract events
	sub  qrl.Subscription // Subscription for errors, completion and termination
	done bool             // Whether the subscription completed delivering logs
	fail error            // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EventEmitterReceivedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EventEmitterReceived)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EventEmitterReceived)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EventEmitterReceivedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EventEmitterReceivedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EventEmitterReceived represents a Received event raised by the EventEmitter contract.
type EventEmitterReceived struct {
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterReceived is a free log retrieval operation binding the contract event 0xa8142743f8f70a4c26f3691cf4ed59718381fb2f18070ec52be1f1022d855557.
//
// Hyperion: event Received(uint256 amount)
func (_EventEmitter *EventEmitterFilterer) FilterReceived(opts *bind.FilterOpts) (*EventEmitterReceivedIterator, error) {

	logs, sub, err := _EventEmitter.contract.FilterLogs(opts, "Received")
	if err != nil {
		return nil, err
	}
	return &EventEmitterReceivedIterator{contract: _EventEmitter.contract, event: "Received", logs: logs, sub: sub}, nil
}

// WatchReceived is a free log subscription operation binding the contract event 0xa8142743f8f70a4c26f3691cf4ed59718381fb2f18070ec52be1f1022d855557.
//
// Hyperion: event Received(uint256 amount)
func (_EventEmitter *EventEmitterFilterer) WatchReceived(opts *bind.WatchOpts, sink chan<- *EventEmitterReceived) (event.Subscription, error) {

	logs, sub, err := _EventEmitter.contract.WatchLogs(opts, "Received")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EventEmitterReceived)
				if err := _EventEmitter.contract.UnpackLog(event, "Received", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseReceived is a log parse operation binding the contract event 0xa8142743f8f70a4c26f3691cf4ed59718381fb2f18070ec52be1f1022d855557.
//
// Hyperion: event Received(uint256 amount)
func (_EventEmitter *EventEmitterFilterer) ParseReceived(log types.Log) (*EventEmitterReceived, error) {
	event := new(EventEmitterReceived)
	if err := _EventEmitter.contract.UnpackLog(event, "Received", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EventEmitterStoredIterator is returned from FilterStored and is used to iterate over the raw logs and unpacked data for Stored events raised by the EventEmitter contract.
type EventEmitterStoredIterator struct {
	Event *EventEmitterStored // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log   // Log channel receiving the found contract events
	sub  qrl.Subscription // Subscription for errors, completion and termination
	done bool             // Whether the subscription completed delivering logs
	fail error            // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EventEmitterStoredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EventEmitterStored)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EventEmitterStored)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EventEmitterStoredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EventEmitterStoredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EventEmitterStored represents a Stored event raised by the EventEmitter contract.
type EventEmitterStored struct {
	Recipient common.Address
	Amount    *big.Int
	Delta     *big.Int
	Tag       [64]byte
	Payload   []byte
	Note      string
	Enabled   bool
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterStored is a free log retrieval operation binding the contract event 0x0971a927eb69632cd5aced366c9dd3ee5626b6c0a27cb781139eeffab9e5372f.
//
// Hyperion: event Stored(address indexed recipient, uint512 indexed amount, int512 indexed delta, bytes64 tag, bytes payload, string note, bool enabled)
func (_EventEmitter *EventEmitterFilterer) FilterStored(opts *bind.FilterOpts, recipient []common.Address, amount []*big.Int, delta []*big.Int) (*EventEmitterStoredIterator, error) {

	var recipientRule []any
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}
	var amountRule []any
	for _, amountItem := range amount {
		amountRule = append(amountRule, amountItem)
	}
	var deltaRule []any
	for _, deltaItem := range delta {
		deltaRule = append(deltaRule, deltaItem)
	}

	logs, sub, err := _EventEmitter.contract.FilterLogs(opts, "Stored", recipientRule, amountRule, deltaRule)
	if err != nil {
		return nil, err
	}
	return &EventEmitterStoredIterator{contract: _EventEmitter.contract, event: "Stored", logs: logs, sub: sub}, nil
}

// WatchStored is a free log subscription operation binding the contract event 0x0971a927eb69632cd5aced366c9dd3ee5626b6c0a27cb781139eeffab9e5372f.
//
// Hyperion: event Stored(address indexed recipient, uint512 indexed amount, int512 indexed delta, bytes64 tag, bytes payload, string note, bool enabled)
func (_EventEmitter *EventEmitterFilterer) WatchStored(opts *bind.WatchOpts, sink chan<- *EventEmitterStored, recipient []common.Address, amount []*big.Int, delta []*big.Int) (event.Subscription, error) {

	var recipientRule []any
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}
	var amountRule []any
	for _, amountItem := range amount {
		amountRule = append(amountRule, amountItem)
	}
	var deltaRule []any
	for _, deltaItem := range delta {
		deltaRule = append(deltaRule, deltaItem)
	}

	logs, sub, err := _EventEmitter.contract.WatchLogs(opts, "Stored", recipientRule, amountRule, deltaRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EventEmitterStored)
				if err := _EventEmitter.contract.UnpackLog(event, "Stored", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStored is a log parse operation binding the contract event 0x0971a927eb69632cd5aced366c9dd3ee5626b6c0a27cb781139eeffab9e5372f.
//
// Hyperion: event Stored(address indexed recipient, uint512 indexed amount, int512 indexed delta, bytes64 tag, bytes payload, string note, bool enabled)
func (_EventEmitter *EventEmitterFilterer) ParseStored(log types.Log) (*EventEmitterStored, error) {
	event := new(EventEmitterStored)
	if err := _EventEmitter.contract.UnpackLog(event, "Stored", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EventEmitterTransformedIterator is returned from FilterTransformed and is used to iterate over the raw logs and unpacked data for Transformed events raised by the EventEmitter contract.
type EventEmitterTransformedIterator struct {
	Event *EventEmitterTransformed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log   // Log channel receiving the found contract events
	sub  qrl.Subscription // Subscription for errors, completion and termination
	done bool             // Whether the subscription completed delivering logs
	fail error            // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EventEmitterTransformedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EventEmitterTransformed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EventEmitterTransformed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EventEmitterTransformedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EventEmitterTransformedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EventEmitterTransformed represents a Transformed event raised by the EventEmitter contract.
type EventEmitterTransformed struct {
	Value uint16
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransformed is a free log retrieval operation binding the contract event 0xe3843251954de1b1a308319c1aa57a527f6a902f9336e038c96857e4b0b82354.
//
// Hyperion: event Transformed(uint16 value)
func (_EventEmitter *EventEmitterFilterer) FilterTransformed(opts *bind.FilterOpts) (*EventEmitterTransformedIterator, error) {

	logs, sub, err := _EventEmitter.contract.FilterLogs(opts, "Transformed")
	if err != nil {
		return nil, err
	}
	return &EventEmitterTransformedIterator{contract: _EventEmitter.contract, event: "Transformed", logs: logs, sub: sub}, nil
}

// WatchTransformed is a free log subscription operation binding the contract event 0xe3843251954de1b1a308319c1aa57a527f6a902f9336e038c96857e4b0b82354.
//
// Hyperion: event Transformed(uint16 value)
func (_EventEmitter *EventEmitterFilterer) WatchTransformed(opts *bind.WatchOpts, sink chan<- *EventEmitterTransformed) (event.Subscription, error) {

	logs, sub, err := _EventEmitter.contract.WatchLogs(opts, "Transformed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EventEmitterTransformed)
				if err := _EventEmitter.contract.UnpackLog(event, "Transformed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransformed is a log parse operation binding the contract event 0xe3843251954de1b1a308319c1aa57a527f6a902f9336e038c96857e4b0b82354.
//
// Hyperion: event Transformed(uint16 value)
func (_EventEmitter *EventEmitterFilterer) ParseTransformed(log types.Log) (*EventEmitterTransformed, error) {
	event := new(EventEmitterTransformed)
	if err := _EventEmitter.contract.UnpackLog(event, "Transformed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EventEmitterTransformed0Iterator is returned from FilterTransformed0 and is used to iterate over the raw logs and unpacked data for Transformed0 events raised by the EventEmitter contract.
type EventEmitterTransformed0Iterator struct {
	Event *EventEmitterTransformed0 // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log   // Log channel receiving the found contract events
	sub  qrl.Subscription // Subscription for errors, completion and termination
	done bool             // Whether the subscription completed delivering logs
	fail error            // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EventEmitterTransformed0Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EventEmitterTransformed0)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(EventEmitterTransformed0)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EventEmitterTransformed0Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EventEmitterTransformed0Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EventEmitterTransformed0 represents a Transformed0 event raised by the EventEmitter contract.
type EventEmitterTransformed0 struct {
	Value string
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransformed0 is a free log retrieval operation binding the contract event 0x53b082deb2f7988df478883fee52c7e9450a5d38daee88ec2bef0543941b46ae.
//
// Hyperion: event Transformed(string value)
func (_EventEmitter *EventEmitterFilterer) FilterTransformed0(opts *bind.FilterOpts) (*EventEmitterTransformed0Iterator, error) {

	logs, sub, err := _EventEmitter.contract.FilterLogs(opts, "Transformed0")
	if err != nil {
		return nil, err
	}
	return &EventEmitterTransformed0Iterator{contract: _EventEmitter.contract, event: "Transformed0", logs: logs, sub: sub}, nil
}

// WatchTransformed0 is a free log subscription operation binding the contract event 0x53b082deb2f7988df478883fee52c7e9450a5d38daee88ec2bef0543941b46ae.
//
// Hyperion: event Transformed(string value)
func (_EventEmitter *EventEmitterFilterer) WatchTransformed0(opts *bind.WatchOpts, sink chan<- *EventEmitterTransformed0) (event.Subscription, error) {

	logs, sub, err := _EventEmitter.contract.WatchLogs(opts, "Transformed0")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EventEmitterTransformed0)
				if err := _EventEmitter.contract.UnpackLog(event, "Transformed0", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransformed0 is a log parse operation binding the contract event 0x53b082deb2f7988df478883fee52c7e9450a5d38daee88ec2bef0543941b46ae.
//
// Hyperion: event Transformed(string value)
func (_EventEmitter *EventEmitterFilterer) ParseTransformed0(log types.Log) (*EventEmitterTransformed0, error) {
	event := new(EventEmitterTransformed0)
	if err := _EventEmitter.contract.UnpackLog(event, "Transformed0", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
