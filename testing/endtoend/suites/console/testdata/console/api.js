var suite = createConsoleSuite("api");
var check = suite.check;
var nextID = 1;

function rpc(method, params) {
    var response = web3.currentProvider.send({
        jsonrpc: "2.0",
        id: nextID++,
        method: method,
        params: params || []
    });
    if (response.error) {
        throw new Error(method + ": " + JSON.stringify(response.error));
    }
    return response.result;
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

check("block APIs agree", function () {
    var blockNumber = qrl.blockNumber;
    if (typeof blockNumber !== "number" || blockNumber <= 0) {
        throw new Error("unexpected qrl.blockNumber: " + blockNumber);
    }
    var block = qrl.getBlock(blockNumber);
    var byHash = qrl.getBlock(block.hash);
    requireHash("block hash", block.hash);
    requireAddress("block fee recipient", block.miner);
    if (byHash.hash !== block.hash || byHash.number !== block.number) {
        throw new Error("block lookup mismatch");
    }
    return true;
});

check("parent hashes chain correctly", function () {
    var head = qrl.getBlock(qrl.blockNumber);
    var parent = qrl.getBlock(head.number - 1);
    return head.parentHash === parent.hash;
});

check("provider dispatch and console namespaces respond", function () {
    var modules = rpc("rpc_modules");
    ["admin", "net", "qrl", "txpool", "web3"].forEach(function (name) {
        if (typeof modules[name] !== "string") {
            throw new Error("missing rpc module " + name);
        }
    });
    if (typeof web3.version.node !== "string") {
        throw new Error("web3.version.node did not return a string");
    }
    if (typeof net.version !== "string" || typeof net.listening !== "boolean" ||
        typeof net.peerCount !== "number") {
        throw new Error("unexpected net namespace");
    }
    if (!admin.nodeInfo || !(admin.peers instanceof Array)) {
        throw new Error("unexpected admin namespace");
    }
    if (typeof txpool.status.pending !== "number" || typeof txpool.status.queued !== "number") {
        throw new Error("unexpected txpool namespace");
    }
    return true;
});

check("qrl.chainId matches the network ID", function () {
    var chainID = qrl.chainId();
    requireHexQuantity("qrl.chainId", chainID);
    return web3.toDecimal(chainID) === web3.toDecimal(net.version);
});

check("header API returns the latest header", function () {
    var header = qrl.getHeaderByNumber("latest");
    requireHash("header hash", header.hash);
    requireHash("header parentHash", header.parentHash);
    requireAddress("header fee recipient", header.miner);
    return true;
});

check("state and fee APIs respond", function () {
    var miner = qrl.getBlock("latest").miner;
    if (!(qrl.getBalance(miner, "latest") >= 0)) {
        throw new Error("invalid balance");
    }
    var nonce = qrl.getTransactionCount(miner, "latest");
    if (typeof nonce !== "number" || nonce < 0) {
        throw new Error("invalid nonce: " + nonce);
    }
    if (!(qrl.gasPrice > 0) || !(qrl.maxPriorityFeePerGas >= 0)) {
        throw new Error("invalid fee data");
    }
    return true;
});

check("qrl.feeHistory returns coherent history", function () {
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

check("qrl.getBlockReceipts returns an array", function () {
    return qrl.getBlockReceipts("latest") instanceof Array;
});

check("QIP-55 Q-address checksum round-trips", function () {
    var miner = qrl.getBlock(qrl.blockNumber).miner;
    requireAddress("miner", miner);
    var lower = "Q" + miner.slice(1).toLowerCase();
    var checksummed = web3.toChecksumAddress(lower);
    if (!web3.isChecksumAddress(checksummed) || !web3.isAddress(checksummed)) {
        throw new Error("invalid checksummed address: " + checksummed);
    }
    if ("Q" + checksummed.slice(1).toLowerCase() !== lower) {
        throw new Error("checksumming changed the address bytes");
    }
    for (var i = 1; i < checksummed.length; i++) {
        var character = checksummed.charAt(i);
        if (/[a-fA-F]/.test(character)) {
            var flipped = character === character.toLowerCase() ?
                character.toUpperCase() : character.toLowerCase();
            var mangled = checksummed.slice(0, i) + flipped + checksummed.slice(i + 1);
            if (web3.isChecksumAddress(mangled)) {
                throw new Error("case-mangled address passes checksum validation");
            }
            break;
        }
    }
    return true;
});

suite.finish();
