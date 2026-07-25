#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd -- "$SCRIPT_DIR/../.." && pwd)

E2E_EXECUTION_IMAGE=${E2E_EXECUTION_IMAGE:?E2E_EXECUTION_IMAGE is required}
DOCKER_BIN=${E2E_DOCKER_BIN:-docker}

SOURCE_COMMIT=$(git -C "$REPO_ROOT" rev-parse HEAD)
SOURCE_STATUS=$(git -C "$REPO_ROOT" status --porcelain=v1 --untracked-files=all)
if [ -n "$SOURCE_STATUS" ]; then
    echo "Refusing to build an unattestable network image from a dirty checkout." >&2
    echo "$SOURCE_STATUS" >&2
    exit 1
fi

export E2E_EXECUTION_IMAGE
export GO_QRL_COMMIT="$SOURCE_COMMIT"

cd "$REPO_ROOT"
"$DOCKER_BIN" buildx bake \
    --file "$SCRIPT_DIR/docker-bake.hcl" \
    network
