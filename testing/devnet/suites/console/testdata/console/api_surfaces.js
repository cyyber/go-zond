// JSON-RPC API surface suite, run through `gqrl attach --exec`. It checks both
// direct provider calls and the console wrappers
// that should be available on a local testnet node.

var _pass = 0;
var _fail = 0;
var _nextID = 1;

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

function rpc(method, params) {
    var response = web3.currentProvider.send({
        jsonrpc: "2.0",
        id: _nextID++,
        method: method,
        params: params || []
    });
    if (response.error) {
        throw new Error(method + ": " + JSON.stringify(response.error));
    }
    return response.result;
}

function requireInvalidArgumentRPCError(method, params, expectedMessage) {
    var response;
    try {
        response = web3.currentProvider.send({
            jsonrpc: "2.0",
            id: _nextID++,
            method: method,
            params: params || []
        });
    } catch (e) {
        throw new Error(method + " provider threw instead of returning a JSON-RPC error: " + e);
    }
    if (!response || !response.error) {
        throw new Error(method + " unexpectedly accepted " + JSON.stringify(params) +
            ": " + JSON.stringify(response));
    }
    if (response.error.code !== -32602) {
        throw new Error(method + " returned error code " + response.error.code +
            ", want -32602: " + JSON.stringify(response.error));
    }
    if (response.error.message !== expectedMessage) {
        throw new Error(method + " returned error message " + JSON.stringify(response.error.message) +
            ", want " + JSON.stringify(expectedMessage));
    }
    return true;
}

function requireHexQuantity(name, value) {
    if (typeof value !== "string" || !/^0x[0-9a-f]+$/i.test(value)) {
        throw new Error(name + " is not a hex quantity: " + value);
    }
}

function requireHash(name, value) {
    if (typeof value !== "string" || !/^0x[0-9a-f]{64}$/i.test(value)) {
        throw new Error(name + " is not a 32-byte hash: " + value);
    }
}

function requireAddress(name, value) {
    if (typeof value !== "string" || !/^Q[0-9a-fA-F]{128}$/.test(value)) {
        throw new Error(name + " is not a QRL address: " + value);
    }
}

check("rpc.modules exposes expected namespaces", function () {
    var modules = rpc("rpc_modules");
    ["admin", "net", "qrl", "txpool", "web3"].forEach(function (name) {
        if (typeof modules[name] !== "string") {
            throw new Error("missing rpc module " + name + ": " + JSON.stringify(modules));
        }
    });
    return true;
});

check("web3_clientVersion responds", function () {
    var version = rpc("web3_clientVersion");
    return typeof version === "string" && version.length > 0;
});

check("qrl_chainId matches console qrl.chainId()", function () {
    var raw = rpc("qrl_chainId");
    requireHexQuantity("qrl_chainId", raw);
    var viaConsole = qrl.chainId();
    if (web3.toDecimal(raw) !== web3.toDecimal(viaConsole)) {
        throw new Error("chain id mismatch: raw " + raw + " console " + viaConsole);
    }
    return true;
});

check("qrl_getBlockByNumber and qrl_getBlockByHash agree", function () {
    var byNumber = rpc("qrl_getBlockByNumber", ["latest", false]);
    requireHash("latest block hash", byNumber.hash);
    var byHash = rpc("qrl_getBlockByHash", [byNumber.hash, false]);
    if (byHash.hash !== byNumber.hash || byHash.number !== byNumber.number) {
        throw new Error("block mismatch: " + JSON.stringify({byNumber: byNumber, byHash: byHash}));
    }
    requireAddress("block fee recipient", byNumber.miner);
    return true;
});

check("qrl_getHeaderByNumber returns the latest header", function () {
    var header = qrl.getHeaderByNumber("latest");
    requireHash("header hash", header.hash);
    requireHash("header parentHash", header.parentHash);
    requireAddress("header fee recipient", header.miner);
    return true;
});

check("qrl balance, nonce, gas price and priority fee APIs respond", function () {
    var miner = qrl.getBlock("latest").miner;
    var balance = qrl.getBalance(miner, "latest");
    var nonce = qrl.getTransactionCount(miner, "latest");
    var gasPrice = qrl.gasPrice;
    var priorityFee = qrl.maxPriorityFeePerGas;

    if (!(balance >= 0)) {
        throw new Error("unexpected balance: " + balance);
    }
    if (typeof nonce !== "number" || nonce < 0) {
        throw new Error("unexpected nonce: " + nonce);
    }
    if (!(gasPrice > 0)) {
        throw new Error("unexpected gasPrice: " + gasPrice);
    }
    if (!(priorityFee >= 0)) {
        throw new Error("unexpected maxPriorityFeePerGas: " + priorityFee);
    }
    return true;
});

check("RPC rejects legacy, truncated, over-wide, and wrong-prefix addresses", function () {
    var malformed = [
        {
            address: "Q" + new Array(21).join("11"),
            message: "invalid argument 0: hex string has length 40, want 128 for common.Address"
        },
        {
            address: "Q" + new Array(64).join("11"),
            message: "invalid argument 0: hex string has length 126, want 128 for common.Address"
        },
        {
            address: "Q" + new Array(66).join("11"),
            message: "invalid argument 0: hex string has length 130, want 128 for common.Address"
        },
        {
            address: "0x" + new Array(65).join("11"),
            message: "invalid argument 0: json: cannot unmarshal hex string without Q prefix into Go value of type common.Address"
        }
    ];
    malformed.forEach(function (test) {
        requireInvalidArgumentRPCError("qrl_getBalance", [test.address, "latest"], test.message);
    });
    return true;
});

check("qrl_feeHistory returns coherent history", function () {
    var history = qrl.feeHistory(1, "latest", []);
    requireHexQuantity("oldestBlock", history.oldestBlock);
    if (!(history.baseFeePerGas instanceof Array) || history.baseFeePerGas.length < 1) {
        throw new Error("missing baseFeePerGas: " + JSON.stringify(history));
    }
    if (!(history.gasUsedRatio instanceof Array) || history.gasUsedRatio.length !== 1) {
        throw new Error("unexpected gasUsedRatio: " + JSON.stringify(history));
    }
    return true;
});

check("qrl_getBlockReceipts returns an array", function () {
    var receipts = qrl.getBlockReceipts("latest");
    if (!(receipts instanceof Array)) {
        throw new Error("receipts is not an array: " + JSON.stringify(receipts));
    }
    return true;
});

check("txpool content and inspect APIs respond", function () {
    if (typeof txpool.content !== "object") {
        throw new Error("txpool.content is not an object");
    }
    if (typeof txpool.inspect !== "object") {
        throw new Error("txpool.inspect is not an object");
    }
    return true;
});

console.log("VM64_E2E_RESULT " + JSON.stringify({schema: 1, suite: "api_surfaces", status: _fail === 0 ? "passed" : "failed", passed: _pass, failed: _fail, total: _pass + _fail}));
console.log("SUITE api_surfaces: " + (_fail === 0 ? "PASSED" : "FAILED") +
    " (" + _pass + "/" + (_pass + _fail) + " checks)");
