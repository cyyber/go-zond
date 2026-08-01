var suite = createConsoleSuite("topics");
var check = suite.check;

loadScript(".params.js");

function sameTopics(got, want) {
    if (got.length !== want.length) {
        return false;
    }
    for (var i = 0; i < got.length; i++) {
        if (got[i].toLowerCase() !== want[i].toLowerCase()) {
            return false;
        }
    }
    return true;
}

var txHash = qrl.sendRawTransaction(PARAMS.indexedRawTransaction);
if (txHash !== PARAMS.indexedTxHash) {
    throw new Error("indexed event transaction hash mismatch");
}

var receipt = null;
for (var i = 0; i < 60; i++) {
    receipt = qrl.getTransactionReceipt(txHash);
    if (receipt !== null && receipt.blockNumber !== null) {
        break;
    }
    admin.sleep(5);
}
if (receipt === null || receipt.blockNumber === null || Number(receipt.status) !== 1) {
    throw new Error("indexed event transaction failed: " + JSON.stringify(receipt));
}

check("receipt preserves indexed VM64 scalar topics", function () {
    if (receipt.logs.length !== 2 ||
        !sameTopics(receipt.logs[0].topics, PARAMS.numberTopics) ||
        !sameTopics(receipt.logs[1].topics, PARAMS.bytesTopics)) {
        throw new Error("unexpected indexed topics: " + JSON.stringify(receipt.logs));
    }
    return true;
});

var contract = qrl.contract(PARAMS.indexedABI).at(receipt.contractAddress);
var block = web3.toHex(receipt.blockNumber);

check("generated filters encode and decode indexed bool and 512-bit integers", function () {
    var events = contract.IndexedNumbers({
        flag: true,
        delta: PARAMS.indexedDelta,
        amount: PARAMS.indexedAmount
    }, {fromBlock: block, toBlock: block}).get();
    if (events.length !== 1 ||
        events[0].args.flag !== true ||
        events[0].args.delta.toString(10) !== PARAMS.indexedDelta ||
        events[0].args.amount.toString(10) !== PARAMS.indexedAmount) {
        throw new Error("unexpected indexed scalar event: " + JSON.stringify(events));
    }
    return true;
});

check("generated filters preserve indexed bytes33 alignment", function () {
    var events = contract.IndexedBytes({
        code: PARAMS.indexedCode
    }, {fromBlock: block, toBlock: block}).get();
    if (events.length !== 1 ||
        events[0].args.code.toLowerCase() !== PARAMS.indexedCode.toLowerCase()) {
        throw new Error("unexpected indexed bytes33 event: " + JSON.stringify(events));
    }
    return true;
});

suite.finish();
