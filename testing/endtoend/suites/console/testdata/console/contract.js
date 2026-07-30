var suite = createConsoleSuite("contract");
var check = suite.check;
var deployedValue = 1337;
var eventSignature = "Deployed(uint256)";

function patternedHex(length, multiplier, addend) {
    var out = "0x";
    for (var i = 0; i < length; i++) {
        var value = (i * multiplier + addend) & 0xff;
        out += (value < 16 ? "0" : "") + value.toString(16);
    }
    return out;
}

function expectTopicError(topic, blockNumber, address) {
    try {
        qrl.getLogs({
            fromBlock: web3.toHex(blockNumber),
            toBlock: web3.toHex(blockNumber),
            address: address,
            topics: [topic]
        });
    } catch (e) {
        return true;
    }
    throw new Error("RPC unexpectedly accepted topic " + topic);
}

loadScript(".params.js");

var receipt = null;
check("prefunded account matches the signed transaction", function () {
    if (!/^Q[0-9a-fA-F]{128}$/.test(PARAMS.address)) {
        throw new Error("unexpected deployer address: " + PARAMS.address);
    }
    if (!(qrl.getBalance(PARAMS.address) > 0)) {
        throw new Error("deployer has zero balance: " + PARAMS.address);
    }
    return true;
});

check("deployment transaction is accepted and mined", function () {
    if (typeof PARAMS.txHash !== "string" || !/^0x[0-9a-f]{64}$/.test(PARAMS.txHash)) {
        throw new Error("invalid prepared transaction hash: " + PARAMS.txHash);
    }
    var responseHash = qrl.sendRawTransaction(PARAMS.rawTransaction);
    if (responseHash !== PARAMS.txHash) {
        throw new Error("tx hash mismatch: have " + responseHash + " want " + PARAMS.txHash);
    }
    for (var i = 0; i < 60; i++) {
        receipt = qrl.getTransactionReceipt(PARAMS.txHash);
        if (receipt !== null && receipt.blockNumber !== null) {
            break;
        }
        admin.sleep(5);
    }
    if (receipt === null || receipt.blockNumber === null) {
        throw new Error("transaction not mined within timeout");
    }
    if (Number(receipt.status) !== 1 || !receipt.contractAddress) {
        throw new Error("deployment failed: " + JSON.stringify(receipt));
    }
    return true;
});

var signatureHash = web3.sha3(eventSignature);
var expectedTopic = signatureHash + zeros(64);
var otherTopic = "0x" + new Array(65).join("ab");
var contract = qrl.contract(PARAMS.abi).at(receipt.contractAddress);

check("transaction and block APIs expose the deployment", function () {
    var tx = qrl.getTransaction(PARAMS.txHash);
    if (tx === null || tx.hash !== PARAMS.txHash || tx.from !== PARAMS.address || tx.to !== null) {
        throw new Error("unexpected transaction: " + JSON.stringify(tx));
    }
    var blockWithHashes = qrl.getBlock(receipt.blockNumber, false);
    var blockWithTransactions = qrl.getBlock(receipt.blockNumber, true);
    if (blockWithHashes.transactions.indexOf(PARAMS.txHash) < 0) {
        throw new Error("block does not include deployment hash");
    }
    for (var i = 0; i < blockWithTransactions.transactions.length; i++) {
        if (blockWithTransactions.transactions[i].hash === PARAMS.txHash) {
            return true;
        }
    }
    throw new Error("block does not include deployment transaction");
});

check("receipt APIs expose one VM64 event", function () {
    if (receipt.logs.length !== 1) {
        throw new Error("expected one receipt log, got " + receipt.logs.length);
    }
    if (receipt.logs[0].topics[0] !== expectedTopic) {
        throw new Error("unexpected receipt topic: " + receipt.logs[0].topics[0]);
    }
    var expectedData = "0x" + zeros(125) + "539";
    if (receipt.logs[0].data !== expectedData) {
        throw new Error("unexpected event data: " + receipt.logs[0].data);
    }
    var receipts = qrl.getBlockReceipts(web3.toHex(receipt.blockNumber));
    for (var i = 0; i < receipts.length; i++) {
        if (receipts[i].transactionHash === PARAMS.txHash) {
            if (receipts[i].logs.length !== 1 ||
                receipts[i].logs[0].topics[0] !== expectedTopic) {
                throw new Error("unexpected block receipt");
            }
            return true;
        }
    }
    throw new Error("block receipts omit deployment transaction");
});

var vm64Amount = "6703903964971298549787012499102923063739682910296196688861780721860882015036773488400937149083451713845015929093243025426876941405973284973216824503046708";
var vm64Delta = "-3351951982485649274893506249551461531869841455148098344430890360930441007518386744200468574541725856922507964546621512713438470702986642486608412251520982";
var vm64Tag = patternedHex(64, 1, 0x80);
var vm64Payload = patternedHex(129, 29, 7);
var vm64Note = "VM64 string crosses the 64-byte ABI word boundary: 0123456789abcdef0123456789abcdef";

check("contract echoes VM64 scalar and dynamic values", function () {
    var echoed = contract.echo(
        vm64Amount,
        vm64Delta,
        vm64Tag,
        PARAMS.address,
        vm64Payload,
        vm64Note,
        true
    );
    if (!(echoed instanceof Array) || echoed.length !== 7) {
        throw new Error("unexpected echo result: " + JSON.stringify(echoed));
    }
    if (echoed[0].toString(10) !== vm64Amount || echoed[1].toString(10) !== vm64Delta) {
        throw new Error("integer mismatch");
    }
    if (echoed[2].toLowerCase() !== vm64Tag) {
        throw new Error("fixed-width mismatch");
    }
    var expectedAddress = web3.toChecksumAddress(PARAMS.address);
    if (echoed[3] !== expectedAddress || !web3.isChecksumAddress(echoed[3])) {
        throw new Error("decoded address is not canonical: " + echoed[3]);
    }
    if (echoed[4].toLowerCase() !== vm64Payload || echoed[5] !== vm64Note || echoed[6] !== true) {
        throw new Error("dynamic-value mismatch");
    }
    return true;
});

check("contract echoes fixed-byte boundaries", function () {
    var values = [
        "0xa5",
        patternedHex(32, 1, 1),
        patternedHex(33, 1, 0x40),
        vm64Tag
    ];
    var echoed = contract.echoFixed(values[0], values[1], values[2], values[3]);
    if (!(echoed instanceof Array) || echoed.length !== values.length) {
        throw new Error("unexpected fixed-bytes result: " + JSON.stringify(echoed));
    }
    for (var i = 0; i < values.length; i++) {
        if (echoed[i].toLowerCase() !== values[i]) {
            throw new Error("fixed bytes " + i + " mismatch");
        }
    }
    return true;
});

check("contract echoes fixed and dynamic arrays", function () {
    var secondTag = vm64Payload.substr(0, 130);
    var echoed = contract.echoArrays([0, 1, vm64Amount], [vm64Tag, secondTag]);
    if (!(echoed instanceof Array) || echoed.length !== 2 ||
        echoed[0].length !== 3 || echoed[1].length !== 2) {
        throw new Error("unexpected array result: " + JSON.stringify(echoed));
    }
    if (echoed[0][0].toString(10) !== "0" ||
        echoed[0][1].toString(10) !== "1" ||
        echoed[0][2].toString(10) !== vm64Amount) {
        throw new Error("integer array mismatch");
    }
    if (echoed[1][0].toLowerCase() !== vm64Tag ||
        echoed[1][1].toLowerCase() !== secondTag) {
        throw new Error("bytes64 array mismatch");
    }
    return true;
});

check("contract event filter decodes the emitted log", function () {
    var filter = contract.Deployed({}, {
        fromBlock: web3.toHex(receipt.blockNumber),
        toBlock: web3.toHex(receipt.blockNumber)
    });
    var events = filter.get();
    if (events.length !== 1 ||
        events[0].transactionHash !== receipt.transactionHash ||
        Number(events[0].args.value) !== deployedValue) {
        throw new Error("unexpected contract event: " + JSON.stringify(events));
    }
    return true;
});

check("raw log filters support exact, wildcard, and OR topics", function () {
    var options = {
        fromBlock: web3.toHex(receipt.blockNumber),
        toBlock: web3.toHex(receipt.blockNumber),
        address: receipt.contractAddress
    };
    options.topics = [expectedTopic];
    var exact = qrl.getLogs(options);
    options.topics = [null];
    var wildcard = qrl.getLogs(options);
    options.topics = [[otherTopic, expectedTopic]];
    var alternatives = qrl.getLogs(options);
    if (exact.length !== 1 || wildcard.length !== 1 || alternatives.length !== 1) {
        throw new Error("unexpected filtered logs");
    }
    if (exact[0].topics[0] !== expectedTopic) {
        throw new Error("unexpected exact topic: " + exact[0].topics[0]);
    }
    return true;
});

check("raw log filters reject 32-byte topics", function () {
    return expectTopicError(signatureHash, receipt.blockNumber, receipt.contractAddress);
});

suite.finish();
