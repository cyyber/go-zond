// VM64 ABI suite, run through `gqrl attach --exec` by the devnet console suite. It
// exercises the embedded web3 ABI coder without relying on account management:
// getData builds calldata locally, and a provider shim returns synthetic data
// through the public contract call path used for real qrl.call results.

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

function word(hex) {
    hex = hex.replace(/^0x/, "").toLowerCase();
    return zeros(128 - hex.length) + hex;
}

function fixedBytes(hex) {
    hex = hex.replace(/^0x/, "").toLowerCase();
    return hex + zeros(128 - hex.length);
}

function selector(signature) {
    return web3.sha3(signature).slice(2, 10);
}

function assertEqual(name, have, want) {
    if (have !== want) {
        throw new Error(name + ": have " + have + " want " + want);
    }
}

var abi = [{
    name: "store",
    type: "function",
    constant: false,
    inputs: [
        {name: "amount", type: "uint512"},
        {name: "tag", type: "bytes4"},
        {name: "recipient", type: "address"},
        {name: "payload", type: "bytes"}
    ],
    outputs: []
}, {
    name: "read",
    type: "function",
    constant: true,
    inputs: [],
    outputs: [
        {name: "amount", type: "uint512"},
        {name: "tag", type: "bytes4"},
        {name: "recipient", type: "address"},
        {name: "payload", type: "bytes"}
    ]
}, {
    name: "acceptBytes64",
    type: "function",
    constant: false,
    inputs: [{name: "value", type: "bytes64"}],
    outputs: []
}];

var miner = qrl.getBlock("latest").miner;
var minerHex = miner.slice(1).toLowerCase();
var contract = qrl.contract(abi).at(miner);

function callReadWithOutput(output) {
    var previousProvider = web3.currentProvider;
    var callPayload = null;
    var provider = {
        send: function (payload) {
            if (payload.method === "qrl_call") {
                callPayload = payload;
                return {jsonrpc: "2.0", id: payload.id, result: output};
            }
            return previousProvider.send(payload);
        },
        sendAsync: function (payload, callback) {
            if (payload.method === "qrl_call") {
                callPayload = payload;
                callback(null, {jsonrpc: "2.0", id: payload.id, result: output});
                return;
            }
            return previousProvider.sendAsync(payload, callback);
        },
        isConnected: function () {
            return true;
        }
    };

    try {
        web3.setProvider(provider);
        return {decoded: contract.read.call(), payload: callPayload};
    } finally {
        web3.setProvider(previousProvider);
    }
}

check("ABI calldata uses 64-byte VM words", function () {
    var data = contract.store.getData(web3.toBigNumber(1337), "0x01020304", miner, "0xabcd");
    var expected = "0x" +
        selector("store(uint512,bytes4,address,bytes)") +
        word("539") +
        fixedBytes("01020304") +
        minerHex +
        word("100") +
        word("2") +
        fixedBytes("abcd");

    assertEqual("calldata", data.toLowerCase(), expected);
    return true;
});

check("ABI bytesN values are left-aligned", function () {
    var value = "0x" + new Array(65).join("ab");
    var data = contract.acceptBytes64.getData(value);
    var expected = "0x" + selector("acceptBytes64(bytes64)") + value.slice(2);
    assertEqual("bytes64 calldata", data.toLowerCase(), expected);
    return true;
});

check("ABI output decoding consumes 64-byte dynamic offsets", function () {
    var encoded = "0x" +
        word("539") +
        fixedBytes("01020304") +
        minerHex +
        word("100") +
        word("2") +
        fixedBytes("abcd");
    var result = callReadWithOutput(encoded);
    var decoded = result.decoded;

    assertEqual("RPC method", result.payload.method, "qrl_call");
    assertEqual("call recipient", result.payload.params[0].to.toLowerCase(), miner.toLowerCase());
    assertEqual("read selector", result.payload.params[0].data.toLowerCase(), "0x" + selector("read()"));

    assertEqual("decoded amount", decoded[0].toString(10), "1337");
    assertEqual("decoded bytes4", decoded[1].toLowerCase(), "0x01020304");
    assertEqual("decoded address bytes", "Q" + decoded[2].slice(1).toLowerCase(), "Q" + minerHex);
    assertEqual("decoded payload", decoded[3].toLowerCase(), "0xabcd");
    return true;
});

check("ABI rejects malformed trailing partial words", function () {
    var encoded = "0x" + word("1") + "00";
    try {
        callReadWithOutput(encoded);
    } catch (e) {
        if (String(e).indexOf("complete 64-byte words") === -1) {
            throw new Error("unexpected decoder error: " + e);
        }
        return true;
    }
    throw new Error("expected public contract call to reject trailing partial word");
});

console.log("VM64_E2E_RESULT " + JSON.stringify({schema: 1, suite: "abi_vm64", status: _fail === 0 ? "passed" : "failed", passed: _pass, failed: _fail, total: _pass + _fail}));
console.log("SUITE abi_vm64: " + (_fail === 0 ? "PASSED" : "FAILED") +
    " (" + _pass + "/" + (_pass + _fail) + " checks)");
