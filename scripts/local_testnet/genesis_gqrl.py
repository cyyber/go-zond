#!/usr/bin/env python3
"""Wrap the pinned genesis generator with the required deposit allocation.

The pinned generator widens storage values but retains the legacy
one-bytes32-per-slot layout. The current runtime packs two bytes32 array
elements into each 64-byte slot, so that allocation leaves the deposit
contract's zero-hash tree uninitialized.
"""

from __future__ import annotations

import hashlib
import json
import subprocess
import sys
from pathlib import Path


UPSTREAM_GENERATOR = Path("/apps/el-gen/genesis_gqrl_upstream.py")
RUNTIME_CODE_FILE = Path("/apps/el-gen/deposit-runtime.hex")
EMPTY_DEPOSIT_ROOT = "d70a234731285c6804c2a4f56711ddb8c82c99740f207854891028af34e27e5e"
BROKEN_EMPTY_DEPOSIT_ROOT = "691a2cb303bfa42437412cd455155952c395370b31a4be3adfb0a373e7ee7c5c"


def zero_hashes() -> list[bytes]:
    hashes = [bytes(32)]
    for _ in range(31):
        hashes.append(hashlib.sha256(hashes[-1] + hashes[-1]).digest())
    return hashes


def packed_storage() -> dict[str, str]:
    """Return constructor-equivalent storage for zero_hashes[32]."""
    hashes = zero_hashes()
    storage: dict[str, str] = {}
    for pair in range(16):
        # SLOAD returns the low/right half for the even element and the
        # high/left half for the odd element. Serialize the word as odd|even.
        value = hashes[2 * pair + 1] + hashes[2 * pair]
        storage[f"0x{0x11 + pair:064x}"] = "0x" + value.hex()
    return storage


def legacy_storage() -> dict[str, str]:
    """Return the stale layout embedded by the pinned upstream revision."""
    hashes = zero_hashes()
    return {
        f"0x{0x21 + index:064x}": "0x" + (hashes[index] + bytes(32)).hex()
        for index in range(1, 32)
    }


def deposit_root(siblings: list[bytes]) -> str:
    node = bytes(32)
    for sibling in siblings:
        node = hashlib.sha256(node + sibling).digest()
    return hashlib.sha256(node + bytes(32)).hexdigest()


def load_runtime_code() -> str:
    code = RUNTIME_CODE_FILE.read_text(encoding="ascii").strip()
    if not code.startswith("0x") or len(code) <= 2:
        raise RuntimeError(f"invalid pinned Qrysm runtime in {RUNTIME_CODE_FILE}")
    bytes.fromhex(code[2:])
    return code


def repair_genesis(genesis: dict[str, object]) -> None:
    alloc = genesis.get("alloc")
    if not isinstance(alloc, dict):
        raise RuntimeError("generated genesis has no alloc object")
    stale = legacy_storage()
    candidates = [
        account
        for account in alloc.values()
        if isinstance(account, dict) and account.get("storage") == stale
    ]
    if len(candidates) != 1:
        raise RuntimeError(
            "expected exactly one deposit allocation with the pinned legacy "
            f"storage layout, found {len(candidates)}"
        )
    candidates[0]["storage"] = packed_storage()
    candidates[0]["code"] = load_runtime_code()


def manifest() -> dict[str, object]:
    runtime = bytes.fromhex(load_runtime_code()[2:])
    storage = packed_storage()
    canonical_storage = json.dumps(
        storage, sort_keys=True, separators=(",", ":")
    ).encode()
    return {
        "schema": 1,
        "runtime_code_bytes": len(runtime),
        "runtime_code_sha256": hashlib.sha256(runtime).hexdigest(),
        "empty_deposit_root": "0x" + EMPTY_DEPOSIT_ROOT,
        "storage_sha256": hashlib.sha256(canonical_storage).hexdigest(),
        "storage_layout": "packed-bytes32-pairs-v1",
    }


def self_test() -> None:
    hashes = zero_hashes()
    if deposit_root(hashes) != EMPTY_DEPOSIT_ROOT:
        raise RuntimeError("standard empty deposit root self-test failed")
    if deposit_root([bytes(32)] * 32) != BROKEN_EMPTY_DEPOSIT_ROOT:
        raise RuntimeError("legacy-allocation regression vector self-test failed")
    storage = packed_storage()
    if (
        len(storage) != 16
        or min(storage) != f"0x{0x11:064x}"
        or max(storage) != f"0x{0x20:064x}"
    ):
        raise RuntimeError("deposit storage slot range self-test failed")
    for pair, value in enumerate(storage.values()):
        expected = "0x" + (hashes[2 * pair + 1] + hashes[2 * pair]).hex()
        if value != expected:
            raise RuntimeError(f"deposit storage pair {pair} self-test failed")


def main() -> None:
    if sys.argv[1:] == ["--self-test"]:
        self_test()
        print("deposit genesis self-test: PASS")
        return
    if sys.argv[1:] == ["--manifest"]:
        self_test()
        print(json.dumps(manifest(), sort_keys=True, separators=(",", ":")))
        return

    completed = subprocess.run(
        [sys.executable, str(UPSTREAM_GENERATOR), *sys.argv[1:]],
        check=False,
        stdout=subprocess.PIPE,
    )
    if completed.returncode != 0:
        raise SystemExit(completed.returncode)
    genesis = json.loads(completed.stdout)
    repair_genesis(genesis)
    json.dump(genesis, sys.stdout, indent="  ")
    sys.stdout.write("\n")


if __name__ == "__main__":
    main()
