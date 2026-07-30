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

// EventEmitterRecord is an auto generated low-level Go binding around an user-defined struct.
type EventEmitterRecord struct {
	Amount    *big.Int
	Recipient common.Address
	Tag       [64]byte
}

// EventEmitterMetaData contains all meta data concerning the EventEmitter contract.
var EventEmitterMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint512\",\"name\":\"initial\",\"type\":\"uint512\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint512\",\"name\":\"code\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"bytes64\",\"name\":\"tag\",\"type\":\"bytes64\"}],\"internalType\":\"structEventEmitter.Record\",\"name\":\"record\",\"type\":\"tuple\"},{\"internalType\":\"uint16[][]\",\"name\":\"nested\",\"type\":\"uint16[][]\"}],\"name\":\"ComplexFailure\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint16[][]\",\"name\":\"values\",\"type\":\"uint16[][]\"}],\"indexed\":false,\"internalType\":\"structEventEmitter.DynamicRecord\",\"name\":\"record\",\"type\":\"tuple\"},{\"indexed\":false,\"internalType\":\"uint16[3]\",\"name\":\"fixedNumbers\",\"type\":\"uint16[3]\"},{\"indexed\":false,\"internalType\":\"string[2]\",\"name\":\"fixedStrings\",\"type\":\"string[2]\"},{\"indexed\":false,\"internalType\":\"uint16[][2]\",\"name\":\"mixed\",\"type\":\"uint16[][2]\"}],\"name\":\"Composite\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint512\",\"name\":\"value\",\"type\":\"uint512\"}],\"name\":\"Deployed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"}],\"name\":\"Dynamic\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"FallbackCalled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"function(uint512)pureexternalreturns(uint512)\",\"name\":\"indexedCallback\",\"type\":\"function\"},{\"indexed\":false,\"internalType\":\"function(uint512)pureexternalreturns(uint512)\",\"name\":\"callback\",\"type\":\"function\"},{\"indexed\":false,\"internalType\":\"uint512\",\"name\":\"result\",\"type\":\"uint512\"}],\"name\":\"FunctionObserved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bool\",\"name\":\"flag\",\"type\":\"bool\"},{\"indexed\":true,\"internalType\":\"bytes5\",\"name\":\"code\",\"type\":\"bytes5\"},{\"indexed\":true,\"internalType\":\"int16\",\"name\":\"delta\",\"type\":\"int16\"}],\"name\":\"IndexedScalars\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint16\",\"name\":\"marker\",\"type\":\"uint16\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Paid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Received\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"indexed\":true,\"internalType\":\"int512\",\"name\":\"delta\",\"type\":\"int512\"},{\"indexed\":false,\"internalType\":\"bytes64\",\"name\":\"tag\",\"type\":\"bytes64\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"enabled\",\"type\":\"bool\"}],\"name\":\"Stored\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"value\",\"type\":\"uint16\"}],\"name\":\"Transformed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"}],\"name\":\"Transformed\",\"type\":\"event\"},{\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"inputs\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"int512\",\"name\":\"delta\",\"type\":\"int512\"},{\"internalType\":\"bytes64\",\"name\":\"tag\",\"type\":\"bytes64\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bool\",\"name\":\"enabled\",\"type\":\"bool\"}],\"name\":\"echo\",\"outputs\":[{\"internalType\":\"uint512\",\"name\":\"\",\"type\":\"uint512\"},{\"internalType\":\"int512\",\"name\":\"\",\"type\":\"int512\"},{\"internalType\":\"bytes64\",\"name\":\"\",\"type\":\"bytes64\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"},{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"smallUnsigned\",\"type\":\"uint8\"},{\"internalType\":\"int8\",\"name\":\"smallSigned\",\"type\":\"int8\"},{\"internalType\":\"uint256\",\"name\":\"wideUnsigned\",\"type\":\"uint256\"},{\"internalType\":\"int256\",\"name\":\"wideSigned\",\"type\":\"int256\"},{\"internalType\":\"bytes5\",\"name\":\"shortBytes\",\"type\":\"bytes5\"},{\"internalType\":\"uint16[3]\",\"name\":\"fixedNumbers\",\"type\":\"uint16[3]\"},{\"internalType\":\"string[2]\",\"name\":\"fixedStrings\",\"type\":\"string[2]\"},{\"internalType\":\"uint16[][2]\",\"name\":\"mixed\",\"type\":\"uint16[][2]\"}],\"name\":\"echoBoundaries\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"},{\"internalType\":\"int8\",\"name\":\"\",\"type\":\"int8\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"int256\",\"name\":\"\",\"type\":\"int256\"},{\"internalType\":\"bytes5\",\"name\":\"\",\"type\":\"bytes5\"},{\"internalType\":\"uint16[3]\",\"name\":\"\",\"type\":\"uint16[3]\"},{\"internalType\":\"string[2]\",\"name\":\"\",\"type\":\"string[2]\"},{\"internalType\":\"uint16[][2]\",\"name\":\"\",\"type\":\"uint16[][2]\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"function(uint512)pureexternalreturns(uint512)\",\"name\":\"callback\",\"type\":\"function\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"function(uint512)pureexternalreturns(uint512)[2]\",\"name\":\"fixedCallbacks\",\"type\":\"function[2]\"},{\"internalType\":\"function(uint512)pureexternalreturns(uint512)[]\",\"name\":\"callbacks\",\"type\":\"function[]\"},{\"components\":[{\"internalType\":\"function(uint512)pureexternalreturns(uint512)\",\"name\":\"callback\",\"type\":\"function\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"}],\"internalType\":\"structEventEmitter.FunctionRecord\",\"name\":\"record\",\"type\":\"tuple\"}],\"name\":\"echoFunctions\",\"outputs\":[{\"internalType\":\"function(uint512)pureexternalreturns(uint512)\",\"name\":\"\",\"type\":\"function\"},{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"function(uint512)pureexternalreturns(uint512)[2]\",\"name\":\"\",\"type\":\"function[2]\"},{\"internalType\":\"function(uint512)pureexternalreturns(uint512)[]\",\"name\":\"\",\"type\":\"function[]\"},{\"components\":[{\"internalType\":\"function(uint512)pureexternalreturns(uint512)\",\"name\":\"callback\",\"type\":\"function\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"}],\"internalType\":\"structEventEmitter.FunctionRecord\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint16[][]\",\"name\":\"values\",\"type\":\"uint16[][]\"}],\"internalType\":\"structEventEmitter.DynamicRecord\",\"name\":\"record\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint16[][]\",\"name\":\"values\",\"type\":\"uint16[][]\"}],\"internalType\":\"structEventEmitter.DynamicRecord[]\",\"name\":\"records\",\"type\":\"tuple[]\"},{\"internalType\":\"uint16[][][]\",\"name\":\"cube\",\"type\":\"uint16[][][]\"}],\"name\":\"echoNested\",\"outputs\":[{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint16[][]\",\"name\":\"values\",\"type\":\"uint16[][]\"}],\"internalType\":\"structEventEmitter.DynamicRecord\",\"name\":\"\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint16[][]\",\"name\":\"values\",\"type\":\"uint16[][]\"}],\"internalType\":\"structEventEmitter.DynamicRecord[]\",\"name\":\"\",\"type\":\"tuple[]\"},{\"internalType\":\"uint16[][][]\",\"name\":\"\",\"type\":\"uint16[][][]\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint16[][]\",\"name\":\"values\",\"type\":\"uint16[][]\"}],\"internalType\":\"structEventEmitter.DynamicRecord\",\"name\":\"record\",\"type\":\"tuple\"},{\"internalType\":\"uint16[3]\",\"name\":\"fixedNumbers\",\"type\":\"uint16[3]\"},{\"internalType\":\"string[2]\",\"name\":\"fixedStrings\",\"type\":\"string[2]\"},{\"internalType\":\"uint16[][2]\",\"name\":\"mixed\",\"type\":\"uint16[][2]\"}],\"name\":\"emitComposite\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"flag\",\"type\":\"bool\"},{\"internalType\":\"bytes5\",\"name\":\"code\",\"type\":\"bytes5\"},{\"internalType\":\"int16\",\"name\":\"delta\",\"type\":\"int16\"}],\"name\":\"emitIndexedScalars\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"}],\"name\":\"emitTransformed\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"value\",\"type\":\"uint16\"}],\"name\":\"emitTransformed\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"function(uint512)pureexternalreturns(uint512)\",\"name\":\"callback\",\"type\":\"function\"},{\"internalType\":\"uint512\",\"name\":\"value\",\"type\":\"uint512\"}],\"name\":\"exerciseFunction\",\"outputs\":[{\"internalType\":\"function(uint512)pureexternalreturns(uint512)\",\"name\":\"\",\"type\":\"function\"},{\"internalType\":\"uint512\",\"name\":\"\",\"type\":\"uint512\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint512\",\"name\":\"code\",\"type\":\"uint512\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"bytes64\",\"name\":\"tag\",\"type\":\"bytes64\"}],\"internalType\":\"structEventEmitter.Record\",\"name\":\"record\",\"type\":\"tuple\"},{\"internalType\":\"uint16[][]\",\"name\":\"nested\",\"type\":\"uint16[][]\"}],\"name\":\"failComplex\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"failPanic\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"failReason\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"observe\",\"outputs\":[{\"internalType\":\"uint512\",\"name\":\"value\",\"type\":\"uint512\"},{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"marker\",\"type\":\"uint16\"}],\"name\":\"pay\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint512\",\"name\":\"value\",\"type\":\"uint512\"}],\"name\":\"plusOne\",\"outputs\":[{\"internalType\":\"uint512\",\"name\":\"\",\"type\":\"uint512\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint512\",\"name\":\"amount\",\"type\":\"uint512\"},{\"internalType\":\"int512\",\"name\":\"delta\",\"type\":\"int512\"},{\"internalType\":\"bytes64\",\"name\":\"tag\",\"type\":\"bytes64\"},{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"string\",\"name\":\"note\",\"type\":\"string\"},{\"internalType\":\"bool\",\"name\":\"enabled\",\"type\":\"bool\"}],\"name\":\"store\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"}],\"name\":\"transform\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"value\",\"type\":\"uint16\"}],\"name\":\"transform\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x61010060805234a015610010575fa0fd5b50608051613d9c3803a0613d9ca339a1a101608052a101b0610032b1b06100cc565ba05fa1b055509f3c76630ff64425a2fc18c61f7cff5d13ea1914ced4ed353a5a1d2f6292146da10000000000000000000000000000000000000000000000000000000000000000a1608051610087b1b0610106565b608051a0b103b0c15061011f565b5fa0fd5b5fa1b050b1b050565b6100aba1610099565ba1146100b5575fa0fd5b50565b5fa151b0506100c6a16100a2565bb2b15050565b5f6040a2a40312156100e1576100e0610095565b5b5f6100eea4a2a5016100b8565bb15050b2b15050565b610100a1610099565ba2525050565b5f6040a201b0506101195fa301a46100f7565bb2b15050565b613c70a061012c5f395ff3fe6101006080526004361061010e575f356101e01ca063a43e73c91161009657a063e558a3a71161006557a063e558a3a71461041757a063ed928c961461045857a063f404ae991461049457a063f8041229146104bc57a063fb144722146104fa5761016c565ba063a43e73c91461038557a063b0b75436146103c357a063b94d6fa6146103d957a063c66c9028146104015761016c565ba0633d0e1089116100dd57a0633d0e10891461027b57a0634b79d0e3146102a357a06379531c40146102e557a06399cf235f1461032157a0639e420a8f146103495761016c565ba06314fc78fc146101c957a0631e3ed7e4146101f457a0632fb0dbcd1461021c57a0633b0e4d67146102385761016c565b3661016c579fa8142743f8f70a4c26f3691cf4ed59718381fb2f18070ec52be1f1022d855557000000000000000000000000000000000000000000000000000000000000000034608051610162b1b0611046565b608051a0b103b0c1005b9fe5b92b8ba08394dd9b027fafca0dc888f149e8f420b55893ecee14ea148aa08b00000000000000000000000000000000000000000000000000000000000000005f36346080516101bfb3b2b1b06110b9565b608051a0b103b0c1005b34a0156101d4575fa0fd5b506101dd610522565b6080516101ebb2b1b0611121565b608051a0b103b0f35b34a0156101ff575fa0fd5b5061021a6004a03603a101b0610215b1b06111ba565b61052f565b005b6102366004a03603a101b0610231b1b061123c565b61058c565b005b34a015610243575fa0fd5b5061025e6004a03603a101b0610259b1b061140a565b6105ec565b608051610272b8b7b6b5b4b3b2b1b0611857565b608051a0b103b0f35b34a015610286575fa0fd5b506102a16004a03603a101b061029cb1b0611a29565b610691565b005b34a0156102ae575fa0fd5b506102c96004a03603a101b06102c4b1b0611a29565b610795565b6080516102dcb7b6b5b4b3b2b1b0611bc0565b608051a0b103b0f35b34a0156102f0575fa0fd5b5061030b6004a03603a101b0610306b1b0611c3f565b61089f565b608051610318b1b0611c6a565b608051a0b103b0f35b34a01561032c575fa0fd5b506103476004a03603a101b0610342b1b0611ca6565b6108b4565b005b34a015610354575fa0fd5b5061036f6004a03603a101b061036ab1b06111ba565b610917565b60805161037cb1b0611d61565b608051a0b103b0f35b34a015610390575fa0fd5b506103ab6004a03603a101b06103a6b1b0611db2565b610987565b6080516103bab3b2b1b0611e31565b608051a0b103b0f35b34a0156103ce575fa0fd5b506103d7610a89565b005b34a0156103e4575fa0fd5b506103ff6004a03603a101b06103fab1b0611e90565b610ae4565b005b34a01561040c575fa0fd5b50610415610b7b565b005b34a015610422575fa0fd5b5061043d6004a03603a101b0610438b1b0611f74565b610b8b565b60805161044fb6b5b4b3b2b1b0612206565b608051a0b103b0f35b34a015610463575fa0fd5b5061047e6004a03603a101b0610479b1b061123c565b610d25565b60805161048bb1b0612280565b608051a0b103b0f35b34a01561049f575fa0fd5b506104ba6004a03603a101b06104b5b1b061123c565b610d3a565b005b34a0156104c7575fa0fd5b506104e26004a03603a101b06104ddb1b0612343565b610d94565b6080516104f1b3b2b1b0612743565b608051a0b103b0f35b34a015610505575fa0fd5b506105206004a03603a101b061051bb1b0612800565b610de5565b005b5fa05f5433b150b150b0b1565b9f53b082deb2f7988df478883fee52c7e9450a5d38daee88ec2bef0543941b46ae0000000000000000000000000000000000000000000000000000000000000000a2a2608051610580b2b1b0612905565b608051a0b103b0c15050565ba061ffff16339f1398d89bb96c43f8c16ef74dee904b456a4fa8a5857191293b848ced1997a3d90000000000000000000000000000000000000000000000000000000000000000346080516105e1b1b0611046565b608051a0b103b0c350565b5fa05fa05f6105f9610e50565b610601610e72565b610609610ebb565bafafafafafafafafa26003a0604002608051b0a101608052a0b2b1b0a260035fb25ba1a4101561064d57a23561ffff16a152604001b1604001b1b2600101b261062b565bb2505050505050b250a1610660b0612b08565bb150a061066cb0612ca6565bb050b750b750b750b750b750b750b750b750b850b850b850b850b850b850b850b8b050565ba85fa1b05550a7a9a79f0971a927eb69632cd5aced366c9dd3ee5626b6c0a27cb781139eeffab9e5372f0000000000000000000000000000000000000000000000000000000000000000aaa9a9a9a9a96080516106f3b6b5b4b3b2b1b0612cba565b608051a0b103b0c4a2a260805161070bb2b1b0612d3e565b608051a0b103b0206101001ba5a5608051610727b2b1b0612d84565b608051a0b103b0206101001b9f4ef7447df163d4aaeab9c66fa93651de5eebb002dcf9b60da1ebaa28ae95e8250000000000000000000000000000000000000000000000000000000000000000ab608051610782b1b0611c6a565b608051a0b103b0c3505050505050505050565b5fa05fa060c0a05fafafafafafafafafafa4a4a0a0603f016040a0b10402604001608051b0a101608052a0b3b2b1b0a17fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff16a152604001a3a3a0a2a4375fa1a40152603f19603fa20116b050a0a301b250505050505050b350b0b1b2b350a2a2a0a0603f016040a0b10402604001608051b0a101608052a0b3b2b1b0a17fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff16a152604001a3a3a0a2a4375fa1a40152603f19603fa20116b050a0a301b250505050505050b150b0b150b650b650b650b650b650b650b650b950b950b950b950b950b950b9b2505050565b5f6001a26108adb1b0612dd1565bb050b1b050565b9fa5d75a0c1afe7b82fadd41d2ad66b9fb0934d4aa36a730e5da8b22ae04352e0e0000000000000000000000000000000000000000000000000000000000000000a4a4a4a4608051610909b4b3b2b1b0613419565b608051a0b103b0c150505050565b60c0a2a2a0a0603f016040a0b10402604001608051b0a101608052a0b3b2b1b0a17fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff16a152604001a3a3a0a2a4375fa1a40152603f19603fa20116b050a0a301b250505050505050b050b2b15050565b5fa05fa0a6a6a6608051a263ffffffff166101e01ba1526004016109abb1b0611c6a565b6040608051a0a303a1a65afa15a0156109c6573d5fa03e3d5ffd5b505050506080513d603f01603f19163da1a110156109e757a0a20336a2a501375b50a0a201a060805250a101b06109fdb1b0613488565bb050a6a6608051610a0fb2b1b06134dd565b608051a0b103b0206101001b9ffa178067c55becdc50555038a0302191054dc26036b1e5ad5d3d0a9daa93423c0000000000000000000000000000000000000000000000000000000000000000a8a8a4608051610a6eb3b2b1b0611e31565b608051a0b103b0c2a6a6a2b350b350b35050b350b350b3b050565b6080519f08c379a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a152600401610adbb0613563565b608051a0b103b0fd5ba060010ba29affffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916a415159f19c59af463d0b89e6afb02db53c6ea998a04ce7bf1aa5c2c0d4c3ac9efc9e6590000000000000000000000000000000000000000000000000000000000000000608051608051a0b103b0c4505050565b5f610b8957610b88613581565b5b565b5fa060c0610b97610f04565b60c0610ba1610f37565badadadadadadadada5a5a0a0603f016040a0b10402604001608051b0a101608052a0b3b2b1b0a17fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff16a152604001a3a3a0a2a4375fa1a40152603f19603fa20116b050a0a301b250505050505050b450b0b1b2b3b450a36002a0608002608051b0a101608052a0b2b1b0a260025fb25ba1a41015610c6957a2a035b0a06040013563ffffffff16b05063ffffffff16a260400152a152608001b1608001b1b2600101b2610c31565bb2505050505050b350a2a2a0a0608002604001608051b0a101608052a0b3b2b1b0a17fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff16a152604001a3a35fb25ba1a41015610cef57a2a035b0a06040013563ffffffff16b05063ffffffff16a260400152a152608001b1608001b1b2600101b2610cb7565bb250505050505050b150b0b150a0610d06b061364f565bb050b550b550b550b550b550b550b850b850b850b850b850b8b2505050565b5f6001a2610d33b1b0613661565bb050b1b050565b9fe3843251954de1b1a308319c1aa57a527f6a902f9336e038c96857e4b0b823540000000000000000000000000000000000000000000000000000000000000000a1608051610d89b1b0612280565b608051a0b103b0c150565b610d9c610f7e565b60c0a0a7a7a7a7a7a4610daeb0613927565bb450a3a3b0610dbdb1b06139fa565bb250b0b1b250a1a1b0610dd0b1b0613ac8565bb050b050b250b250b250b550b550b5b2505050565ba7a7a7a7a7a7a7a76080519fa6b3cee1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a152600401610e47b8b7b6b5b4b3b2b1b0613c02565b608051a0b103b0fd5b608051a060c001608052a06003b06040a202a036a337a0a201b15050b05050b0565b608051a0608001608052a06002b05b60c07fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff16a152604001b06001b003b0a1610e8157b05050b0565b608051a0608001608052a06002b05b60c07fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff16a152604001b06001b003b0a1610eca57b05050b0565b608051a061010001608052a06002b05b5fa063ffffffff16a260400152a152608001b06001b003b0a1610f1457b05050b0565b608051a060c001608052a05fa063ffffffff16a260400152a15260800160c07fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff16a15250b0565b608051a061010001608052a05fa15260400160c07fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff16a15260400160c07fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff16a15260400160c07fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff16a15250b0565b5f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffa216b050b1b050565b611040a161100c565ba2525050565b5f6040a201b0506110595fa301a4611037565bb2b15050565b5fa2a2526040a201b050b2b15050565ba2a1a3375fa3a30152505050565b5f603f19603fa30116b050b1b050565b5f611098a3a561105f565bb3506110a5a3a5a461106f565b6110aea361107d565ba401b050b3b2505050565b5f6080a201b050a1a1035fa301526110d2a1a5a761108d565bb0506110e16040a301a4611037565bb4b350505050565b5fa1b050b1b050565b6110fba16110e9565ba2525050565b5f61110ba26110e9565bb050b1b050565b61111ba1611101565ba2525050565b5f6080a201b0506111345fa301a56110f2565b6111416040a301a4611112565bb3b2505050565b5f608051b050b0565b5fa0fd5b5fa0fd5b5fa0fd5b5fa0fd5b5fa0fd5b5fa0a3603fa4011261117a57611179611159565b5ba235b05067ffffffffffffffffa111156111975761119661115d565b5b6040a301b150a36001a202a30111156111b3576111b2611161565b5bb250b2b050565b5fa06040a3a50312156111d0576111cf611151565b5b5fa3013567ffffffffffffffffa111156111ed576111ec611155565b5b6111f9a5a2a601611165565bb250b25050b250b2b050565b5f61ffffa216b050b1b050565b61121ba1611205565ba114611225575fa0fd5b50565b5fa135b050611236a1611212565bb2b15050565b5f6040a2a403121561125157611250611151565b5b5f61125ea4a2a501611228565bb15050b2b15050565b5f60ffa216b050b1b050565b61127ca1611267565ba114611286575fa0fd5b50565b5fa135b050611297a1611273565bb2b15050565b5fa15f0bb050b1b050565b6112b1a161129d565ba1146112bb575fa0fd5b50565b5fa135b0506112cca16112a8565bb2b15050565b6112dba161100c565ba1146112e5575fa0fd5b50565b5fa135b0506112f6a16112d2565bb2b15050565b5fa1601f0bb050b1b050565b611311a16112fc565ba11461131b575fa0fd5b50565b5fa135b05061132ca1611308565bb2b15050565b5f9fffffffffff0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a216b050b1b050565b611386a1611332565ba114611390575fa0fd5b50565b5fa135b0506113a1a161137d565bb2b15050565b5fa1b050a26040600302a20111156113c2576113c1611161565b5bb2b15050565b5fa1b050a26040600202a20111156113e3576113e2611161565b5bb2b15050565b5fa1b050a26040600202a201111561140457611403611161565b5bb2b15050565b5fa05fa05fa05fa0610280a9ab03121561142757611426611151565b5b5f611434aba2ac01611289565bb850506040611445aba2ac016112be565bb750506080611456aba2ac016112e8565bb6505060c0611467aba2ac0161131e565bb55050610100611479aba2ac01611393565bb4505061014061148baba2ac016113a7565bb35050610200a9013567ffffffffffffffffa111156114ad576114ac611155565b5b6114b9aba2ac016113c8565bb25050610240a9013567ffffffffffffffffa111156114db576114da611155565b5b6114e7aba2ac016113e9565bb15050b2b5b850b2b5b8b0b3b650565b611500a1611267565ba2525050565b61150fa161129d565ba2525050565b61151ea16112fc565ba2525050565b61152da1611332565ba2525050565b5f6003b050b1b050565b5fa1b050b2b15050565b5fa1b050b1b050565b611559a1611205565ba2525050565b5f61156aa3a3611550565b6040a301b050b2b15050565b5f6040a201b050b1b050565b61158ba1611533565b611595a1a461153d565bb2506115a0a2611547565ba05f5ba3a110156115d057a1516115b7a7a261155f565bb6506115c2a3611576565bb250506001a101b0506115a3565b505050505050565b5f6002b050b1b050565b5fa1b050b2b15050565b5fa1b050b1b050565b5fa151b050b1b050565b5fa2a2526040a201b050b2b15050565b5f5ba3a1101561162c57a0a20151a1a401526040a101b050611611565ba35ba1a11015611646575fa1a501536001a101b05061162e565b5050505050565b5f611657a26115f5565b611661a1a56115ff565bb350611671a1a56040a60161160f565b61167aa161107d565ba401b15050b2b15050565b5f611690a3a361164d565bb050b2b15050565b5f6040a201b050b1b050565b5f6116aea26115d8565b6116b8a1a56115e2565bb350a36040a202a5016116caa56115ec565ba05f5ba5a1101561170557a4a403a952a1516116e6a5a2611685565bb4506116f1a3611698565bb2506040aa01b950506001a101b0506116cd565b50a2b750a7b5505050505050b2b15050565b5f6002b050b1b050565b5fa1b050b2b15050565b5fa1b050b1b050565b5fa151b050b1b050565b5fa2a2526040a201b050b2b15050565b5fa1b0506040a201b050b1b050565b5f6040a201b050b1b050565b5f611773a2611734565b61177da1a561173e565bb350611788a361174e565ba05f5ba3a110156117b857a15161179fa8a261155f565bb7506117aaa361175d565bb250506001a101b05061178b565b50a5b350505050b2b15050565b5f6117d0a3a3611769565bb050b2b15050565b5f6040a201b050b1b050565b5f6117eea2611717565b6117f8a1a5611721565bb350a36040a202a50161180aa561172b565ba05f5ba5a1101561184557a4a403a952a151611826a5a26117c5565bb450611831a36117d8565bb2506040aa01b950506001a101b05061180d565b50a2b750a7b5505050505050b2b15050565b5f610280a201b05061186b5fa301ab6114f7565b6118786040a301aa611506565b6118856080a301a9611037565b61189260c0a301a8611515565b6118a0610100a301a7611524565b6118ae610140a301a6611582565ba1a103610200a301526118c1a1a56116a4565bb050a1a103610240a301526118d6a1a46117e4565bb050b9b8505050505050505050565b6118eea16110e9565ba1146118f8575fa0fd5b50565b5fa135b050611909a16118e5565bb2b15050565b5fa1b050b1b050565b611921a161190f565ba11461192b575fa0fd5b50565b5fa135b05061193ca1611918565bb2b15050565b5fa1b050b1b050565b611954a1611942565ba11461195e575fa0fd5b50565b5fa135b05061196fa161194b565bb2b15050565b61197ea1611101565ba114611988575fa0fd5b50565b5fa135b050611999a1611975565bb2b15050565b5fa0a3603fa401126119b4576119b3611159565b5ba235b05067ffffffffffffffffa111156119d1576119d061115d565b5b6040a301b150a36001a202a30111156119ed576119ec611161565b5bb250b2b050565b5fa11515b050b1b050565b611a08a16119f4565ba114611a12575fa0fd5b50565b5fa135b050611a23a16119ff565bb2b15050565b5fa05fa05fa05fa05f6101c0aaac031215611a4757611a46611151565b5b5f611a54aca2ad016118fb565bb950506040611a65aca2ad0161192e565bb850506080611a76aca2ad01611961565bb7505060c0611a87aca2ad0161198b565bb65050610100aa013567ffffffffffffffffa11115611aa957611aa8611155565b5b611ab5aca2ad0161199f565bb550b55050610140aa013567ffffffffffffffffa11115611ad957611ad8611155565b5b611ae5aca2ad01611165565bb350b35050610180611af9aca2ad01611a15565bb15050b2b5b850b2b5b850b2b5b8565b611b12a161190f565ba2525050565b611b21a1611942565ba2525050565b5fa151b050b1b050565b5f611b3ba2611b27565b611b45a1a561105f565bb350611b55a1a56040a60161160f565b611b5ea161107d565ba401b15050b2b15050565b5fa2a2526040a201b050b2b15050565b5f611b83a26115f5565b611b8da1a5611b69565bb350611b9da1a56040a60161160f565b611ba6a161107d565ba401b15050b2b15050565b611bbaa16119f4565ba2525050565b5f6101c0a201b050611bd45fa301aa6110f2565b611be16040a301a9611b09565b611bee6080a301a8611b18565b611bfb60c0a301a7611112565ba1a103610100a30152611c0ea1a6611b31565bb050a1a103610140a30152611c23a1a5611b79565bb050611c33610180a301a4611bb1565bb8b75050505050505050565b5f6040a2a4031215611c5457611c53611151565b5b5f611c61a4a2a5016118fb565bb15050b2b15050565b5f6040a201b050611c7d5fa301a46110f2565bb2b15050565b5fa0fd5b5f610100a2a4031215611c9d57611c9c611c83565b5ba1b050b2b15050565b5fa05fa0610180a5a7031215611cbf57611cbe611151565b5b5fa5013567ffffffffffffffffa11115611cdc57611cdb611155565b5b611ce8a7a2a801611c87565bb450506040611cf9a7a2a8016113a7565bb35050610100a5013567ffffffffffffffffa11115611d1b57611d1a611155565b5b611d27a7a2a8016113c8565bb25050610140a5013567ffffffffffffffffa11115611d4957611d48611155565b5b611d55a7a2a8016113e9565bb15050b2b5b1b450b250565b5f6040a201b050a1a1035fa30152611d79a1a4611b79565bb050b2b15050565b5fa0a235b15063ffffffff6040a4013516b050b250b2b050565b5fa0611da7a4a4611d81565bb150b150b250b2b050565b5fa05f60c0a4a6031215611dc957611dc8611151565b5b5f611dd6a6a2a701611d9b565bb350b350506080611de9a6a2a7016118fb565bb15050b250b250b2565b5fa0a2b150a3b050b250b2b050565b5fa16101e01bb050b1b050565b611e19a2a2611df3565bb250b050a0a35263ffffffffa2166040a40152505050565b5f60c0a201b050611e455fa301a5a7611e0f565b611e526080a301a46110f2565bb4b350505050565b5fa160010bb050b1b050565b611e6fa1611e5a565ba114611e79575fa0fd5b50565b5fa135b050611e8aa1611e66565bb2b15050565b5fa05f60c0a4a6031215611ea757611ea6611151565b5b5f611eb4a6a2a701611a15565bb350506040611ec5a6a2a701611393565bb250506080611ed6a6a2a701611e7c565bb15050b250b250b2565b5fa1b050a26080600202a2011115611efb57611efa611161565b5bb2b15050565b5fa0a3603fa40112611f1657611f15611159565b5ba235b05067ffffffffffffffffa11115611f3357611f3261115d565b5b6040a301b150a36080a202a3011115611f4f57611f4e611161565b5bb250b2b050565b5f60c0a2a4031215611f6b57611f6a611c83565b5ba1b050b2b15050565b5fa05fa05fa05fa0610240a9ab031215611f9157611f90611151565b5b5f611f9eaba2ac01611d9b565bb850b850506080a9013567ffffffffffffffffa11115611fc157611fc0611155565b5b611fcdaba2ac01611165565bb650b6505060c0611fe0aba2ac01611ee0565bb450506101c0a9013567ffffffffffffffffa1111561200257612001611155565b5b61200eaba2ac01611f01565bb350b35050610200a9013567ffffffffffffffffa1111561203257612031611155565b5b61203eaba2ac01611f56565bb15050b2b5b850b2b5b8b0b3b650565b5f6002b050b1b050565b5fa1b050b2b15050565b5fa1b050b1b050565b612075a2a2611df3565bb250b050a0a35263ffffffffa2166040a40152505050565b5f612099a4a4a461206b565b6080a401b050b3b2505050565b5fa0a251b15063ffffffff6040a4015116b050b150b1565b5f6080a201b050b1b050565b6120d3a161204e565b6120dda1a4612058565bb2506120e8a2612062565ba05f5ba3a11015612121576120fca26120a6565b612107a8a2a461208d565bb750612112a46120be565bb35050506001a101b0506120eb565b505050505050565b5fa151b050b1b050565b5fa2a2526040a201b050b2b15050565b5fa1b0506040a201b050b1b050565b5f6080a201b050b1b050565b5f612168a2612129565b612172a1a5612133565bb35061217da3612143565ba05f5ba3a110156121b657612191a26120a6565b61219ca9a2a461208d565bb8506121a7a4612152565bb35050506001a101b050612180565b50a5b350505050b2b15050565b5f60c0a3016121d35fa4016120a6565b6121e05fa701a2a461206b565b50506080a30151a4a2036080a601526121f9a2a261164d565bb15050a0b15050b2b15050565b5f610240a201b05061221b5fa301a8aa611e0f565ba1a1036080a3015261222da1a7611b79565bb05061223c60c0a301a66120ca565ba1a1036101c0a3015261224fa1a561215e565bb050a1a103610200a30152612264a1a46121c3565bb050b7b650505050505050565b61227aa1611205565ba2525050565b5f6040a201b0506122935fa301a4612271565bb2b15050565b5fa0a3603fa401126122ae576122ad611159565b5ba235b05067ffffffffffffffffa111156122cb576122ca61115d565b5b6040a301b150a36040a202a30111156122e7576122e6611161565b5bb250b2b050565b5fa0a3603fa4011261230357612302611159565b5ba235b05067ffffffffffffffffa111156123205761231f61115d565b5b6040a301b150a36040a202a301111561233c5761233b611161565b5bb250b2b050565b5fa05fa05f60c0a6a803121561235c5761235b611151565b5b5fa6013567ffffffffffffffffa1111561237957612378611155565b5b612385a8a2a901611c87565bb550506040a6013567ffffffffffffffffa111156123a6576123a5611155565b5b6123b2a8a2a901612299565bb450b450506080a6013567ffffffffffffffffa111156123d5576123d4611155565b5b6123e1a8a2a9016122ee565bb250b25050b2b550b2b5b0b350565b6123f9a16110e9565ba2525050565b5fa2a2526040a201b050b2b15050565b5f612419a2611b27565b612423a1a56123ff565bb350612433a1a56040a60161160f565b61243ca161107d565ba401b15050b2b15050565b5fa151b050b1b050565b5fa2a2526040a201b050b2b15050565b5fa1b0506040a201b050b1b050565b5f6040a201b050b1b050565b5f612486a2612447565b612490a1a5612451565bb350a36040a202a5016124a2a5612461565ba05f5ba5a110156124dd57a4a403a952a1516124bea5a26117c5565bb4506124c9a3612470565bb2506040aa01b950506001a101b0506124a5565b50a2b750a7b5505050505050b2b15050565b5f610100a3015fa301516125055fa601a26123f0565b506040a30151a4a2036040a6015261251da2a261164d565bb150506080a30151a4a2036080a60152612537a2a261240f565bb1505060c0a30151a4a20360c0a60152612551a2a261247c565bb15050a0b15050b2b15050565b5fa151b050b1b050565b5fa2a2526040a201b050b2b15050565b5fa1b0506040a201b050b1b050565b5f610100a3015fa3015161259d5fa601a26123f0565b506040a30151a4a2036040a601526125b5a2a261164d565bb150506080a30151a4a2036080a601526125cfa2a261240f565bb1505060c0a30151a4a20360c0a601526125e9a2a261247c565bb15050a0b15050b2b15050565b5f612601a3a3612587565bb050b2b15050565b5f6040a201b050b1b050565b5f61261fa261255e565b612629a1a5612568565bb350a36040a202a50161263ba5612578565ba05f5ba5a1101561267657a4a403a952a151612657a5a26125f6565bb450612662a3612609565bb2506040aa01b950506001a101b05061263e565b50a2b750a7b5505050505050b2b15050565b5fa151b050b1b050565b5fa2a2526040a201b050b2b15050565b5fa1b0506040a201b050b1b050565b5f6126bca3a361247c565bb050b2b15050565b5f6040a201b050b1b050565b5f6126daa2612688565b6126e4a1a5612692565bb350a36040a202a5016126f6a56126a2565ba05f5ba5a1101561273157a4a403a952a151612712a5a26126b1565bb45061271da36126c4565bb2506040aa01b950506001a101b0506126f9565b50a2b750a7b5505050505050b2b15050565b5f60c0a201b050a1a1035fa3015261275ba1a66124ef565bb050a1a1036040a3015261276fa1a5612615565bb050a1a1036080a30152612783a1a46126d0565bb050b4b350505050565b5f60c0a2a40312156127a2576127a1611c83565b5ba1b050b2b15050565b5fa0a3603fa401126127c0576127bf611159565b5ba235b05067ffffffffffffffffa111156127dd576127dc61115d565b5b6040a301b150a36040a202a30111156127f9576127f8611161565b5bb250b2b050565b5fa05fa05fa05fa06101c0a9ab03121561281d5761281c611151565b5b5f61282aaba2ac016118fb565bb850506040a9013567ffffffffffffffffa1111561284b5761284a611155565b5b612857aba2ac01611165565bb750b750506080a9013567ffffffffffffffffa1111561287a57612879611155565b5b612886aba2ac0161199f565bb550b5505060c0612899aba2ac0161278d565bb35050610180a9013567ffffffffffffffffa111156128bb576128ba611155565b5b6128c7aba2ac016127ab565bb250b25050b2b5b850b2b5b8b0b3b650565b5f6128e4a3a5611b69565bb3506128f1a3a5a461106f565b6128faa361107d565ba401b050b3b2505050565b5f6040a201b050a1a1035fa3015261291ea1a4a66128d9565bb050b3b2505050565b5fa16101001bb050b1b050565b61295d7f4e487b7100000000000000000000000000000000000000000000000000000000612927565b5f52604160045260445ffd5b612972a261107d565ba101a1a11067ffffffffffffffffa211171561299157612990612934565b5ba0608052505050565b5f6129a3611148565bb0506129afa2a2612969565bb1b050565b5f67ffffffffffffffffa211156129ce576129cd612934565b5b6040a202b050b1b050565ba1a1525050565b5fa0fd5b5f67ffffffffffffffffa211156129fe576129fd612934565b5b612a07a261107d565bb0506040a101b050b1b050565b5f612a26612a21a46129e4565b61299a565bb050a2a1526040a101a4a4a4011115612a4257612a416129e0565b5b612a4da4a2a561106f565b50b3b2505050565b5fa2603fa30112612a6957612a68611159565b5ba135612a79a4a26040a601612a14565bb15050b2b15050565b5f612a94612a8fa46129b4565b61299a565bb050a06040a402a301a5a11115612aae57612aad611161565b5ba35ba1a11015612afe57a03567ffffffffffffffffa11115612ad357612ad2611159565b5ba0a601612ae0a9a2612a55565b612aeaa1a76129d9565b6040a601b5505050506040a101b050612ab0565b505050b3b2505050565b5f612b15366002a4612a82565bb050b1b050565b5f67ffffffffffffffffa21115612b3657612b35612934565b5b6040a202b050b1b050565ba1a1525050565b5f67ffffffffffffffffa21115612b6257612b61612934565b5b6040a202b0506040a101b050b1b050565b612b7ca2611205565ba1525050565b5f612b94612b8fa4612b48565b61299a565bb050a0a3a2526040a201b0506040a402a301a5a11115612bb757612bb6611161565b5ba35ba1a11015612be957a0612bcca8a2611228565b612bd6a1a6612b73565b6040a501b45050506040a101b050612bb9565b505050b3b2505050565b5fa2603fa30112612c0757612c06611159565b5ba135612c17a4a26040a601612b82565bb15050b2b15050565b5f612c32612c2da4612b1c565b61299a565bb050a06040a402a301a5a11115612c4c57612c4b611161565b5ba35ba1a11015612c9c57a03567ffffffffffffffffa11115612c7157612c70611159565b5ba0a601612c7ea9a2612bf3565b612c88a1a7612b41565b6040a601b5505050506040a101b050612c4e565b505050b3b2505050565b5f612cb3366002a4612c20565bb050b1b050565b5f610100a201b050612cce5fa301a9611b18565ba1a1036040a30152612ce1a1a7a961108d565bb050a1a1036080a30152612cf6a1a5a76128d9565bb050612d0560c0a301a4611bb1565bb7b650505050505050565b5fa1b050b2b15050565b5f612d25a3a5612d10565bb350612d32a3a5a461106f565ba2a401b050b3b2505050565b5f612d4aa2a4a6612d1a565bb150a1b050b3b2505050565b5fa1b050b2b15050565b5f612d6ba3a5612d56565bb350612d78a3a5a461106f565ba2a401b050b3b2505050565b5f612d90a2a4a6612d60565bb150a1b050b3b2505050565b612dc57f4e487b7100000000000000000000000000000000000000000000000000000000612927565b5f52601160045260445ffd5b5f612ddba26110e9565bb150612de6a36110e9565bb250a2a201b050a0a21115612dfe57612dfd612d9c565b5bb2b15050565b5f612e126040a401a46118fb565bb050b2b15050565b5fa0fd5b5fa0fd5b5fa0fd5b5fa0a33567ffffffffffffffffa11115612e4357612e42612e22565b5ba3a101b25060403603a3113660401117a4a4101715612e6557612e64612e22565b5ba235b1506040a301b25067ffffffffffffffffa21115612e8857612e87612e1a565b5b6001a202a30136a111a4a2101715612ea357612ea2612e1e565b5b5050b250b2b050565b5f612eb7a3a56115ff565bb350612ec4a3a5a461106f565b612ecda361107d565ba401b050b3b2505050565b5fa0a33567ffffffffffffffffa11115612ef557612ef4612e22565b5ba3a101b25060403603a3113660401117a4a4101715612f1757612f16612e22565b5ba235b1506040a301b25067ffffffffffffffffa21115612f3a57612f39612e1a565b5b6001a202a30136a111a4a2101715612f5557612f54612e1e565b5b5050b250b2b050565b5f612f69a3a56123ff565bb350612f76a3a5a461106f565b612f7fa361107d565ba401b050b3b2505050565b5fa0a33567ffffffffffffffffa11115612fa757612fa6612e22565b5ba3a101b25060403603a3113660401117a4a4101715612fc957612fc8612e22565b5ba235b1506040a301b25067ffffffffffffffffa21115612fec57612feb612e1a565b5b6040a202a30136a111a4a210171561300757613006612e1e565b5b5050b250b2b050565b5fa1b050b1b050565b5fa1b050b1b050565b5f6130306040a401a4611228565bb050b2b15050565b5f6040a201b050b1b050565b5f61304fa3a561173e565bb35061305aa2613019565ba05f5ba5a110156130925761306fa2a4613022565b613079a8a261155f565bb750613084a3613038565bb250506001a101b05061305d565b50a5b2505050b3b2505050565b5f6130aba4a4a4613044565bb050b3b2505050565b5fa0a33567ffffffffffffffffa111156130d1576130d0612e22565b5ba3a101b25060403603a3113660401117a4a41017156130f3576130f2612e22565b5ba235b1506040a301b25067ffffffffffffffffa2111561311657613115612e1a565b5b6040a202a30136a111a4a210171561313157613130612e1e565b5b5050b250b2b050565b5f6040a201b050b1b050565b5f613151a3a5612451565bb350a36040a402a501613163a4613010565ba05f5ba7a110156131a857a4a403a95261317da2a46130b4565b613188a6a2a461309f565bb550613193a461313a565bb3506040ab01ba5050506001a101b050613166565b50a2b750a7b45050505050b3b2505050565b5f610100a3016131cc5fa401a4612e04565b6131d85fa601a26123f0565b506131e66040a401a4612e26565ba5a3036040a701526131f9a3a2a4612eac565bb250505061320a6080a401a4612ed8565ba5a3036080a7015261321da3a2a4612f5e565bb250505061322e60c0a401a4612f8a565ba5a30360c0a70152613241a3a2a4613146565bb2505050a0b15050b2b15050565b5f6003b050b1b050565b5fa1b050b1b050565b5f6040a201b050b1b050565b613277a161324f565b613281a1a461153d565bb25061328ca2613259565ba05f5ba3a110156132c4576132a1a2a4613022565b6132aba7a261155f565bb6506132b6a3613262565bb250506001a101b05061328f565b505050505050565b5f6002b050b1b050565b5fa1b050b1b050565b5f6132eba4a4a4612eac565bb050b3b2505050565b5f6040a201b050b1b050565b5f61330aa26132cc565b613314a1a56115e2565bb350a36040a202a501613326a56132d6565ba05f5ba5a1101561336b57a4a403a952613340a2a4612e26565b61334ba6a2a46132df565bb550613356a46132f4565bb3506040ab01ba5050506001a101b050613329565b50a2b750a7b5505050505050b2b15050565b5f6002b050b1b050565b5fa1b050b1b050565b5f6040a201b050b1b050565b5f6133a6a261337d565b6133b0a1a5611721565bb350a36040a202a5016133c2a5613387565ba05f5ba5a1101561340757a4a403a9526133dca2a46130b4565b6133e7a6a2a461309f565bb5506133f2a4613390565bb3506040ab01ba5050506001a101b0506133c5565b50a2b750a7b5505050505050b2b15050565b5f610180a201b050a1a1035fa30152613432a1a76131ba565bb0506134416040a301a661326e565ba1a103610100a30152613454a1a5613300565bb050a1a103610140a30152613469a1a461339c565bb050b5b45050505050565b5fa151b050613482a16118e5565bb2b15050565b5f6040a2a403121561349d5761349c611151565b5b5f6134aaa4a2a501613474565bb15050b2b15050565b6134bda2a2611df3565bb250b050a0a3526134d363ffffffffa316611e02565b6040a40152505050565b5f6134e9a2a4a66134b3565b6044a201b150a1b050b3b2505050565b9f564d207374616e646172642072657665727420726561736f6e0000000000000000000000000000000000000000000000000000000000000000000000000000005fa2015250565b5f61354d6019a3611b69565bb150613558a26134f9565b6040a201b050b1b050565b5f6040a201b050a1a1035fa3015261357aa1613541565bb050b1b050565b6135aa7f4e487b7100000000000000000000000000000000000000000000000000000000612927565b5f52600160045260445ffd5b5fa0fd5b5fa0fd5ba1a15263ffffffffa3166040a20152505050565b5f60c0a2a40312156135e7576135e66135b6565b5b6135f160c061299a565bb0505f613600a4a2a501611d81565b61360da1a35fa7016135be565b5050506080a2013567ffffffffffffffffa1111561362e5761362d6135ba565b5b61363aa4a2a501612a55565b613647a16080a5016129d9565b5050b2b15050565b5f61365a36a36135d2565bb050b1b050565b5f61366ba2611205565bb150613676a3611205565bb250a2a201b05061ffffa111156136905761368f612d9c565b5bb2b15050565b61369fa26110e9565ba1525050565b5f67ffffffffffffffffa211156136bf576136be612934565b5b6136c8a261107d565bb0506040a101b050b1b050565b5f6136e76136e2a46136a5565b61299a565bb050a2a1526040a101a4a4a4011115613703576137026129e0565b5b61370ea4a2a561106f565b50b3b2505050565b5fa2603fa3011261372a57613729611159565b5ba13561373aa4a26040a6016136d5565bb15050b2b15050565ba1a1525050565b5f67ffffffffffffffffa2111561376457613763612934565b5b6040a202b0506040a101b050b1b050565b5f613787613782a461374a565b61299a565bb050a0a3a2526040a201b0506040a402a301a5a111156137aa576137a9611161565b5ba35ba1a110156137fa57a03567ffffffffffffffffa111156137cf576137ce611159565b5ba0a6016137dca9a2612bf3565b6137e6a1a7612b41565b6040a601b5505050506040a101b0506137ac565b505050b3b2505050565b5fa2603fa3011261381857613817611159565b5ba135613828a4a26040a601613775565bb15050b2b15050565ba1a1525050565b5f610100a2a403121561384e5761384d6135b6565b5b61385961010061299a565bb0505f613868a4a2a5016118fb565b613874a15fa501613696565b50506040a2013567ffffffffffffffffa11115613894576138936135ba565b5b6138a0a4a2a501612a55565b6138ada16040a5016129d9565b50506080a2013567ffffffffffffffffa111156138cd576138cc6135ba565b5b6138d9a4a2a501613716565b6138e6a16080a501613743565b505060c0a2013567ffffffffffffffffa11115613906576139056135ba565b5b613912a4a2a501613804565b61391fa160c0a501613831565b5050b2b15050565b5f61393236a3613838565bb050b1b050565b5f67ffffffffffffffffa2111561395357613952612934565b5b6040a202b0506040a101b050b1b050565ba1a1525050565b5f61397d613978a4613939565b61299a565bb050a0a3a2526040a201b0506040a402a301a5a111156139a05761399f611161565b5ba35ba1a110156139f057a03567ffffffffffffffffa111156139c5576139c4611159565b5ba0a6016139d2a9a2613838565b6139dca1a7613964565b6040a601b5505050506040a101b0506139a2565b505050b3b2505050565b5f613a0636a4a461396b565bb050b2b15050565b5f67ffffffffffffffffa21115613a2857613a27612934565b5b6040a202b0506040a101b050b1b050565b5f613a4b613a46a4613a0e565b61299a565bb050a0a3a2526040a201b0506040a402a301a5a11115613a6e57613a6d611161565b5ba35ba1a11015613abe57a03567ffffffffffffffffa11115613a9357613a92611159565b5ba0a601613aa0a9a2613804565b613aaaa1a7613831565b6040a601b5505050506040a101b050613a70565b505050b3b2505050565b5f613ad436a4a4613a39565bb050b2b15050565b5f613aea6040a401a461198b565bb050b2b15050565b613afba1611101565ba2525050565b5f613b0f6040a401a4611961565bb050b2b15050565b613b20a1611942565ba2525050565b60c0a201613b365fa301a3612e04565b613b425fa501a26123f0565b50613b506040a301a3613adc565b613b5d6040a501a2613af2565b50613b6b6080a301a3613b01565b613b786080a501a2613b17565b50505050565b5fa2a2526040a201b050b2b15050565b5f613b99a3a5613b7e565bb350a36040a402a501613baba4613010565ba05f5ba7a11015613bf057a4a403a952613bc5a2a46130b4565b613bd0a6a2a461309f565bb550613bdba461313a565bb3506040ab01ba5050506001a101b050613bae565b50a2b750a7b45050505050b3b2505050565b5f6101c0a201b050613c165fa301ab6110f2565ba1a1036040a30152613c29a1a9ab6128d9565bb050a1a1036080a30152613c3ea1a7a961108d565bb050613c4d60c0a301a6613b26565ba1a103610180a30152613c61a1a4a6613b8e565bb050b9b850505050505050505056",
}

// EventEmitterABI is the input ABI used to generate the binding from.
// Deprecated: Use EventEmitterMetaData.ABI instead.
var EventEmitterABI = EventEmitterMetaData.ABI

// EventEmitterBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use EventEmitterMetaData.Bin instead.
var EventEmitterBin = EventEmitterMetaData.Bin

// DeployEventEmitter deploys a new QRL contract, binding an instance of EventEmitter to it.
func DeployEventEmitter(auth *bind.TransactOpts, backend bind.ContractBackend, initial *big.Int) (common.Address, *types.Transaction, *EventEmitter, error) {
	parsed, err := EventEmitterMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(EventEmitterBin), backend, initial)
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
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDeployed is a free log retrieval operation binding the contract event 0x3c76630ff64425a2fc18c61f7cff5d13ea1914ced4ed353a5a1d2f6292146da1.
//
// Hyperion: event Deployed(uint512 value)
func (_EventEmitter *EventEmitterFilterer) FilterDeployed(opts *bind.FilterOpts) (*EventEmitterDeployedIterator, error) {

	logs, sub, err := _EventEmitter.contract.FilterLogs(opts, "Deployed")
	if err != nil {
		return nil, err
	}
	return &EventEmitterDeployedIterator{contract: _EventEmitter.contract, event: "Deployed", logs: logs, sub: sub}, nil
}

// WatchDeployed is a free log subscription operation binding the contract event 0x3c76630ff64425a2fc18c61f7cff5d13ea1914ced4ed353a5a1d2f6292146da1.
//
// Hyperion: event Deployed(uint512 value)
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

// ParseDeployed is a log parse operation binding the contract event 0x3c76630ff64425a2fc18c61f7cff5d13ea1914ced4ed353a5a1d2f6292146da1.
//
// Hyperion: event Deployed(uint512 value)
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
