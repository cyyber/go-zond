var suite = createConsoleSuite("constructor");
var check = suite.check;

loadScript(".params.js");

var managed = qrl.accounts;
if (!(managed instanceof Array) || managed.length !== 1) {
    throw new Error("unexpected node-managed accounts: " + JSON.stringify(managed));
}

var contractABI = PARAMS.abi.filter(function (entry) {
    return entry.type !== "constructor";
}).concat(PARAMS.constructorABI);
var callbackCount = 0;

qrl.contract(contractABI).new(
    PARAMS.storeValue,
    PARAMS.address,
    PARAMS.constructorTag,
    PARAMS.constructorPayload,
    {from: managed[0], data: PARAMS.bytecode, gas: PARAMS.constructorGas},
    function (error, contract) {
        try {
            if (error) {
                throw error;
            }
            callbackCount++;
            if (callbackCount === 1) {
                if (!contract.transactionHash || contract.address) {
                    throw new Error("unexpected initial constructor callback");
                }
                return;
            }
            if (callbackCount !== 2) {
                throw new Error("unexpected constructor callback count: " + callbackCount);
            }

            check("ContractFactory.new encodes and mines VM64 constructor data", function () {
                var receipt = qrl.getTransactionReceipt(contract.transactionHash);
                var transaction = qrl.getTransaction(contract.transactionHash);
                if (receipt === null || Number(receipt.status) !== 1 ||
                    receipt.contractAddress !== contract.address ||
                    transaction.input.toLowerCase() !== PARAMS.constructorInput.toLowerCase()) {
                    throw new Error("unexpected constructor deployment: " +
                        JSON.stringify({receipt: receipt, transaction: transaction}));
                }
                if (!web3.isChecksumAddress(contract.address) || qrl.getCode(contract.address) === "0x") {
                    throw new Error("constructor returned an invalid deployed contract");
                }
                if (contract.stored().toString(10) !== "0") {
                    throw new Error("deployed contract methods were not attached");
                }
                return true;
            });
            suite.finish();
        } catch (failure) {
            console.error("CONSOLE_E2E_FAIL constructor " + failure);
        }
    }
);
