var suite = createConsoleSuite("events");
var check = suite.check;

loadScript(".params.js");

var deployment = qrl.getTransactionReceipt(PARAMS.txHash);
if (deployment === null || !deployment.contractAddress) {
    throw new Error("deployment receipt is unavailable");
}

var contract = qrl.contract(PARAMS.abi).at(deployment.contractAddress);
var expectedLabelTopic = web3.sha3(PARAMS.storeLabel) + zeros(64);
var expectedPayloadTopic = web3.sha3(PARAMS.storePayload, {encoding: "hex"}) + zeros(64);

var request = contract.store.request(
    PARAMS.storeValue,
    PARAMS.storeLabel,
    PARAMS.storePayload,
    {from: PARAMS.address, gas: 500000}
);
if (request.method !== "qrl_sendTransaction" ||
    request.params.length !== 1 ||
    request.params[0].data !== PARAMS.storeData) {
    throw new Error("unexpected state-changing wrapper request");
}

var watcher = contract.Stored({}, {fromBlock: "latest"});
watcher.watch(function (error, event) {
    try {
        if (error) {
            throw error;
        }
        var receipt = qrl.getTransactionReceipt(PARAMS.storeTxHash);
        check("state-changing contract wrapper call is mined", function () {
            if (receipt === null || receipt.blockNumber === null || Number(receipt.status) !== 1) {
                throw new Error("store transaction failed: " + JSON.stringify(receipt));
            }
            if (contract.stored().toString(10) !== PARAMS.storeValue) {
                throw new Error("stored value mismatch");
            }
            return true;
        });

        check("WebSocket event watch decodes indexed dynamic fields", function () {
            if (event.transactionHash !== PARAMS.storeTxHash) {
                throw new Error("event watch returned the wrong transaction");
            }
            var expectedSender = web3.toChecksumAddress(PARAMS.address);
            if (event.args.sender !== expectedSender ||
                !web3.isChecksumAddress(event.args.sender)) {
                throw new Error("event sender is not canonical: " + event.args.sender);
            }
            if (event.args.label !== expectedLabelTopic ||
                event.args.payload !== expectedPayloadTopic) {
                throw new Error("indexed dynamic topic mismatch: " + JSON.stringify(event.args));
            }
            if (event.args.value.toString(10) !== PARAMS.storeValue) {
                throw new Error("event value mismatch");
            }
            return true;
        });
        watcher.stopWatching();
        suite.finish();
    } catch (failure) {
        watcher.stopWatching();
        console.error("CONSOLE_E2E_FAIL events " + failure);
    }
});

var txHash = qrl.sendRawTransaction(PARAMS.storeRawTransaction);
if (txHash !== PARAMS.storeTxHash) {
    watcher.stopWatching();
    console.error("CONSOLE_E2E_FAIL events store transaction hash mismatch");
}
