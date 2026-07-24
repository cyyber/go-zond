#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd -- "$SCRIPT_DIR/../.." && pwd)

if (( $# != 0 )); then
    echo "This wrapper does not accept flags." >&2
    echo "Set E2E_NETWORK_DIR or E2E_DOCKER_BIN and run make network-start." >&2
    exit 2
fi

exec make -C "$REPO_ROOT" network-start \
    E2E_NETWORK_DIR="${E2E_NETWORK_DIR:-/tmp/go-qrl-e2e-network}"
