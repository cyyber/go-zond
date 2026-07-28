// Basic RPC/console sanity suite, run through `gqrl attach --exec`.

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

check("qrl.blockNumber is a positive number", function () {
    var bn = qrl.blockNumber;
    return typeof bn === "number" && bn > 0;
});

check("qrl.getBlock round-trips by number and hash", function () {
    var byNumber = qrl.getBlock(qrl.blockNumber);
    if (!byNumber || !byNumber.hash) {
        throw new Error("no block returned for qrl.blockNumber");
    }
    if (!/^0x[0-9a-f]{64}$/.test(byNumber.hash)) {
        throw new Error("unexpected block hash format: " + byNumber.hash);
    }
    var byHash = qrl.getBlock(byNumber.hash);
    return byHash.hash === byNumber.hash && byHash.number === byNumber.number;
});

check("parent hashes chain correctly", function () {
    var head = qrl.getBlock(qrl.blockNumber);
    var parent = qrl.getBlock(head.number - 1);
    return head.parentHash === parent.hash;
});

check("net namespace responds", function () {
    return typeof net.version === "string" && net.version.length > 0 &&
        typeof net.listening === "boolean" && typeof net.peerCount === "number";
});

check("admin namespace responds", function () {
    var info = admin.nodeInfo;
    if (!info || typeof info.id !== "string" || typeof info.name !== "string") {
        throw new Error("unexpected admin.nodeInfo: " + JSON.stringify(info));
    }
    var peers = admin.peers;
    if (!(peers instanceof Array)) {
        throw new Error("admin.peers is not an array");
    }
    return true;
});

check("txpool namespace responds", function () {
    var status = txpool.status;
    return typeof status.pending === "number" && typeof status.queued === "number";
});

// QIP-55 checksum round-trip on the head block's fee recipient.
check("QIP-55 Q-address checksum round-trips", function () {
    var miner = qrl.getBlock(qrl.blockNumber).miner;
    if (!/^Q[0-9a-fA-F]{128}$/.test(miner)) {
        throw new Error("unexpected miner address format: " + miner);
    }
    var lower = "Q" + miner.slice(1).toLowerCase();
    var checksummed = web3.toChecksumAddress(lower);
    if (!web3.isChecksumAddress(checksummed)) {
        throw new Error("toChecksumAddress output fails isChecksumAddress: " + checksummed);
    }
    if (("Q" + checksummed.slice(1).toLowerCase()) !== lower) {
        throw new Error("checksumming changed the address bytes: " + checksummed);
    }
    if (!web3.isAddress(checksummed)) {
        throw new Error("isAddress rejects checksummed address: " + checksummed);
    }
    // Flipping the case of one letter must break the checksum.
    var mangled = null;
    for (var i = 1; i < checksummed.length; i++) {
        var c = checksummed.charAt(i);
        if (/[a-fA-F]/.test(c)) {
            var flipped = (c === c.toLowerCase()) ? c.toUpperCase() : c.toLowerCase();
            mangled = checksummed.slice(0, i) + flipped + checksummed.slice(i + 1);
            break;
        }
    }
    if (mangled !== null && web3.isChecksumAddress(mangled)) {
        throw new Error("case-mangled address still passes isChecksumAddress: " + mangled);
    }
    return true;
});

console.log("VM64_E2E_RESULT " + JSON.stringify({schema: 1, suite: "web3_sanity", status: _fail === 0 ? "passed" : "failed", passed: _pass, failed: _fail, total: _pass + _fail}));
console.log("SUITE web3_sanity: " + (_fail === 0 ? "PASSED" : "FAILED") +
    " (" + _pass + "/" + (_pass + _fail) + " checks)");
