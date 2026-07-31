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

check("state-changing wrapper executes through the node-managed signer", function () {
    var managed = qrl.accounts;
    if (!(managed instanceof Array) || managed.length !== 1) {
        throw new Error("unexpected node-managed accounts: " + JSON.stringify(managed));
    }
    var txHash = contract.store(
        PARAMS.storeValue,
        "Clef-backed console transaction",
        "0x010203",
        {from: managed[0], gas: 500000}
    );
    var receipt = null;
    for (var i = 0; i < 60; i++) {
        receipt = qrl.getTransactionReceipt(txHash);
        if (receipt !== null && receipt.blockNumber !== null) {
            break;
        }
        admin.sleep(5);
    }
    if (receipt === null || receipt.blockNumber === null || Number(receipt.status) !== 1) {
        throw new Error("Clef-backed wrapper transaction failed: " + JSON.stringify(receipt));
    }
    var transaction = qrl.getTransaction(txHash);
    if (transaction.from !== managed[0] || contract.stored().toString(10) !== PARAMS.storeValue) {
        throw new Error("unexpected Clef-backed wrapper result");
    }
    return true;
});

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

var watcher = contract.Stored({
    sender: PARAMS.address,
    label: PARAMS.storeLabel,
    payload: PARAMS.storePayload
}, {fromBlock: "latest"});
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

        check("state wrappers return the full VM64 storage value", function () {
            var expected = "0x" + web3.toBigNumber(PARAMS.storeValue).toString(16);
            if (expected.length !== 130) {
                throw new Error("fixture is not a full-width VM64 value: " + expected);
            }
            var stored = qrl.getStorageAt(deployment.contractAddress, "0x0", "latest");
            if (stored.toLowerCase() !== expected) {
                throw new Error("unexpected storage value: " + stored);
            }
            var proof = qrl.getProof(deployment.contractAddress, ["0x0"], "latest");
            if (!proof.storageProof || proof.storageProof.length !== 1 ||
                proof.storageProof[0].value.toLowerCase() !== expected) {
                throw new Error("unexpected storage proof: " + JSON.stringify(proof));
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

        check("indexed event filters reject non-matching dynamic values", function () {
            var events = contract.Stored({
                sender: PARAMS.address,
                label: PARAMS.storeLabel + "-missing",
                payload: PARAMS.storePayload
            }, {
                fromBlock: web3.toHex(receipt.blockNumber),
                toBlock: web3.toHex(receipt.blockNumber)
            }).get();
            if (events.length !== 0) {
                throw new Error("non-matching indexed filter returned events: " + JSON.stringify(events));
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
