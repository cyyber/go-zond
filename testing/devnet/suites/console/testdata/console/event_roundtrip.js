// End-to-end event/topic round-trip suite, run through `gqrl attach --exec`.
// Deploys precompiled Hyperion bytecode whose constructor emits
// Deployed(uint256), using a transaction pre-signed by the Go test from the
// prefunded dev account. It then asserts that the emitted topic0
// is keccak256(signature) LEFT-aligned in the 64-byte topic (hash||zeros)
// and that both the contract event filter and a raw qrl.getLogs filter using
// the full VM64 topic match it.
//
// The Go test generates console/.params.js with the ABI and signed deployment
// transaction before running this suite.

var _pass = 0;
var _fail = 0;

function check(desc, fn) {
    try {
        if (fn() === false) {
            throw new Error("assertion returned false");
        }
        _pass++;
        console.log("PASS: " + desc);
    } catch (e) {
        _fail++;
        console.log("FAIL: " + desc + " -- " + e);
    }
}

function zeros(n) {
    return new Array(n + 1).join("0");
}

function patternedHex(length, multiplier, addend) {
    var out = "0x";
    for (var i = 0; i < length; i++) {
        var value = (i * multiplier + addend) & 0xff;
        out += (value < 16 ? "0" : "") + value.toString(16);
    }
    return out;
}

// loadScript throws when the file is missing or does not compile.
var _abort = false;
try {
    loadScript("console/.params.js");
} catch (e) {
    console.log("FAIL: could not load test scripts -- " + e);
    _fail++;
    _abort = true;
}

var receipt = null;

if (!_abort) {
    check("prefunded account matches the signed transaction", function () {
        if (!/^Q[0-9a-fA-F]{128}$/.test(PARAMS.address)) {
            throw new Error("unexpected deployer address: " + PARAMS.address);
        }
        var balance = qrl.getBalance(PARAMS.address);
        if (!(balance > 0)) {
            throw new Error("deployer " + PARAMS.address + " has zero balance; check network_params.yaml prefunded_accounts");
        }
        return true;
    });

    check("deployment transaction is accepted and mined", function () {
        var txHash = PARAMS.txHash;
        if (typeof txHash !== "string" || !/^0x[0-9a-f]{64}$/.test(txHash)) {
            throw new Error("invalid locally prepared transaction hash: " + txHash);
        }
        var responseHash = qrl.sendRawTransaction(PARAMS.rawTransaction);
        if (responseHash !== txHash) {
            throw new Error("tx hash mismatch: have " + responseHash + " want " + txHash);
        }

        // Allow transient missed slots without hiding a stalled chain.
        for (var i = 0; i < 60; i++) {
            receipt = qrl.getTransactionReceipt(txHash);
            if (receipt !== null && receipt.blockNumber !== null) {
                break;
            }
            admin.sleep(5);
        }
        if (receipt === null || receipt.blockNumber === null) {
            throw new Error("transaction not mined within timeout");
        }
        if (Number(receipt.status) !== 1) {
            throw new Error("deployment reverted, status " + receipt.status);
        }
        if (!receipt.contractAddress) {
            throw new Error("receipt has no contractAddress");
        }
        return true;
    });
}

if (receipt !== null) {
    var sigHash = web3.sha3(PARAMS.signature); // 0x + 64 hex chars
    var expectedTopic = sigHash + zeros(64);    // left-aligned in 64-byte topic

    check("transaction and block APIs expose the mined deployment", function () {
        var tx = qrl.getTransaction(PARAMS.txHash);
        if (tx === null) {
            throw new Error("qrl.getTransaction returned null");
        }
        if (tx.hash !== PARAMS.txHash || tx.from !== PARAMS.address) {
            throw new Error("unexpected transaction: " + JSON.stringify(tx));
        }
        if (tx.to !== null) {
            throw new Error("deployment transaction has non-null to: " + tx.to);
        }

        var blockWithHashes = qrl.getBlock(receipt.blockNumber, false);
        var blockWithTxs = qrl.getBlock(receipt.blockNumber, true);
        if (blockWithHashes.transactions.indexOf(PARAMS.txHash) < 0) {
            throw new Error("block hash list does not include deployment tx");
        }
        var found = false;
        for (var i = 0; i < blockWithTxs.transactions.length; i++) {
            if (blockWithTxs.transactions[i].hash === PARAMS.txHash) {
                found = true;
                break;
            }
        }
        if (!found) {
            throw new Error("full transaction block does not include deployment tx object");
        }
        return true;
    });

    check("block receipt API exposes the deployment log", function () {
        var receipts = qrl.getBlockReceipts(web3.toHex(receipt.blockNumber));
        if (!(receipts instanceof Array)) {
            throw new Error("qrl.getBlockReceipts did not return an array");
        }
        for (var i = 0; i < receipts.length; i++) {
            if (receipts[i].transactionHash === PARAMS.txHash) {
                if (receipts[i].logs.length !== 1) {
                    throw new Error("expected one log in block receipt, got " + receipts[i].logs.length);
                }
                if (receipts[i].logs[0].topics[0] !== expectedTopic) {
                    throw new Error("receipt topic mismatch: " + receipts[i].logs[0].topics[0]);
                }
                return true;
            }
        }
        throw new Error("qrl.getBlockReceipts did not include deployment receipt");
    });

    check("emitted topic0 is keccak256(signature) left-aligned (hash||zeros)", function () {
        if (receipt.logs.length !== 1) {
            throw new Error("expected 1 log, got " + receipt.logs.length);
        }
        var topic0 = receipt.logs[0].topics[0];
        if (topic0 !== expectedTopic) {
            throw new Error("have " + topic0 + " want " + expectedTopic);
        }
        return true;
    });

    check("VM64 event data encodes the value in a 64-byte word", function () {
        var data = receipt.logs[0].data;
        var expected = "0x" + zeros(125) + "539"; // 1337
        if (data !== expected) {
            throw new Error("have " + data + " want " + expected);
        }
        return true;
    });

    var contract = qrl.contract(PARAMS.abi).at(receipt.contractAddress);

    var vm64Amount = "6703903964971298549787012499102923063739682910296196688861780721860882015036773488400937149083451713845015929093243025426876941405973284973216824503046708";
    var vm64Delta = "-3351951982485649274893506249551461531869841455148098344430890360930441007518386744200468574541725856922507964546621512713438470702986642486608412251520982";
    var vm64Tag = patternedHex(64, 1, 0x80);
    var vm64Payload = patternedHex(129, 29, 7);
    var vm64Note = "VM64 string crosses the 64-byte ABI word boundary: 0123456789abcdef0123456789abcdef";

    check("live contract echoes uint512/int512/address/bytes64/dynamic values", function () {
        var echoed = contract.echo(vm64Amount, vm64Delta, vm64Tag, PARAMS.address, vm64Payload, vm64Note, true);
        if (!(echoed instanceof Array) || echoed.length !== 7) {
            throw new Error("unexpected echo result: " + JSON.stringify(echoed));
        }
        if (echoed[0].toString(10) !== vm64Amount || echoed[1].toString(10) !== vm64Delta) {
            throw new Error("integer mismatch: " + echoed[0] + ", " + echoed[1]);
        }
        if (echoed[2].toLowerCase() !== vm64Tag || echoed[3].toLowerCase() !== PARAMS.address.toLowerCase()) {
            throw new Error("fixed-width mismatch: " + echoed[2] + ", " + echoed[3]);
        }
        if (echoed[4].toLowerCase() !== vm64Payload || echoed[5] !== vm64Note || echoed[6] !== true) {
            throw new Error("dynamic-value mismatch: " + JSON.stringify(echoed));
        }
        return true;
    });

    check("live contract echoes bytes1/bytes32/bytes33/bytes64 boundaries", function () {
        var value1 = "0xa5";
        var value32 = patternedHex(32, 1, 1);
        var value33 = patternedHex(33, 1, 0x40);
        var echoed = contract.echoFixed(value1, value32, value33, vm64Tag);
        var want = [value1, value32, value33, vm64Tag];
        if (!(echoed instanceof Array) || echoed.length !== want.length) {
            throw new Error("unexpected fixed-bytes result: " + JSON.stringify(echoed));
        }
        for (var i = 0; i < want.length; i++) {
            if (echoed[i].toLowerCase() !== want[i]) {
                throw new Error("fixed bytes " + i + " mismatch: have " + echoed[i] + " want " + want[i]);
            }
        }
        return true;
    });

    check("live contract echoes dynamic and fixed VM64 arrays", function () {
        var secondTag = vm64Payload.substr(0, 2 + 64 * 2);
        var echoed = contract.echoArrays([0, 1, vm64Amount], [vm64Tag, secondTag]);
        if (!(echoed instanceof Array) || echoed.length !== 2 || echoed[0].length !== 3 || echoed[1].length !== 2) {
            throw new Error("unexpected array result: " + JSON.stringify(echoed));
        }
        if (echoed[0][0].toString(10) !== "0" || echoed[0][1].toString(10) !== "1" || echoed[0][2].toString(10) !== vm64Amount) {
            throw new Error("integer array mismatch: " + JSON.stringify(echoed[0]));
        }
        if (echoed[1][0].toLowerCase() !== vm64Tag || echoed[1][1].toLowerCase() !== secondTag) {
            throw new Error("bytes64 array mismatch: " + JSON.stringify(echoed[1]));
        }
        return true;
    });

    check("live contract default storage decodes as zero values", function () {
        var stored = contract.read();
        if (!(stored instanceof Array) || stored.length !== 7) {
            throw new Error("unexpected read result: " + JSON.stringify(stored));
        }
        if (stored[0].toString(10) !== "0" || stored[1].toString(10) !== "0") {
            throw new Error("default integer state is nonzero");
        }
        if (stored[2] !== "0x" + zeros(128) || stored[3] !== "Q" + zeros(128)) {
            throw new Error("default fixed-width state is nonzero: " + stored[2] + ", " + stored[3]);
        }
        if (stored[4] !== "0x" || stored[5] !== "" || stored[6] !== false) {
            throw new Error("default dynamic state is nonzero: " + JSON.stringify(stored));
        }
        return true;
    });

    check("contract event filter matches the emitted log", function () {
        var filter = contract.Deployed({}, {
            fromBlock: web3.toHex(receipt.blockNumber),
            toBlock: web3.toHex(receipt.blockNumber)
        });
        var events = filter.get();
        if (events.length !== 1) {
            throw new Error("expected 1 event, got " + events.length);
        }
        if (events[0].transactionHash !== receipt.transactionHash) {
            throw new Error("event from unexpected transaction: " + events[0].transactionHash);
        }
        return true;
    });

    check("contract event filter decodes the emitted value", function () {
        var filter = contract.Deployed({}, {
            fromBlock: web3.toHex(receipt.blockNumber),
            toBlock: web3.toHex(receipt.blockNumber)
        });
        var events = filter.get();
        if (events.length === 1 && Number(events[0].args.value) === PARAMS.value) {
            return true;
        }
        throw new Error("unexpected decoded args: " + JSON.stringify(events.length === 1 ? events[0].args : events));
    });

    check("qrl.getLogs with the full VM64 signature topic matches", function () {
        var logs = qrl.getLogs({
            fromBlock: web3.toHex(receipt.blockNumber),
            toBlock: web3.toHex(receipt.blockNumber),
            topics: [expectedTopic]
        });
        if (logs.length !== 1) {
            throw new Error("expected 1 log, got " + logs.length);
        }
        if (logs[0].address !== receipt.contractAddress) {
            throw new Error("log from unexpected address: " + logs[0].address);
        }
        if (logs[0].topics[0] !== expectedTopic) {
            throw new Error("have topic " + logs[0].topics[0] + " want " + expectedTopic);
        }
        return true;
    });

    check("qrl.getLogs with only the raw 32-byte signature hash is rejected", function () {
        try {
            qrl.getLogs({
                fromBlock: web3.toHex(receipt.blockNumber),
                toBlock: web3.toHex(receipt.blockNumber),
                topics: [sigHash]
            });
        } catch (e) {
            return true;
        }
        throw new Error("expected an error");
    });
}

console.log("VM64_E2E_RESULT " + JSON.stringify({schema: 1, suite: "event_roundtrip", status: _fail === 0 ? "passed" : "failed", passed: _pass, failed: _fail, total: _pass + _fail}));
console.log("SUITE event_roundtrip: " + (_fail === 0 ? "PASSED" : "FAILED") +
    " (" + _pass + "/" + (_pass + _fail) + " checks)");
