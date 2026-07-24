#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd -- "$SCRIPT_DIR/../.." && pwd)

if (( $# != 0 )); then
    echo "This compatibility wrapper no longer accepts an enclave name." >&2
    echo "Set E2E_NETWORK_DIR and run make network-stop." >&2
    exit 2
fi

# The Go controller validates and destroys only the full UUID recorded in the
# selected network directory. It deliberately leaves the shared engine alive.
exec make -C "$REPO_ROOT" network-stop \
    E2E_NETWORK_DIR="${E2E_NETWORK_DIR:-/tmp/go-qrl-e2e-network}"
