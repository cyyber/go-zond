#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd -- "$SCRIPT_DIR/../.." && pwd)

E2E_LOCAL_EL_IMAGE=${E2E_LOCAL_EL_IMAGE:?E2E_LOCAL_EL_IMAGE is required}
if [[ ! "$E2E_LOCAL_EL_IMAGE" =~ ^[A-Za-z0-9./:_-]+$ ]] ||
    [[ "$E2E_LOCAL_EL_IMAGE" == *"@"* ]]; then
    echo "Output image reference is invalid: $E2E_LOCAL_EL_IMAGE" >&2
    exit 2
fi

DOCKER_BIN=${E2E_DOCKER_BIN:-docker}

for command in "$DOCKER_BIN" git; do
    if ! command -v "$command" >/dev/null 2>&1; then
        echo "$command is not installed or not on PATH." >&2
        exit 1
    fi
done
if ! "$DOCKER_BIN" info >/dev/null 2>&1; then
    echo "Docker is installed but its daemon is unavailable." >&2
    exit 1
fi
if ! "$DOCKER_BIN" buildx version >/dev/null 2>&1; then
    echo "Docker Buildx is required." >&2
    exit 1
fi

if ! git -C "$REPO_ROOT" rev-parse --show-toplevel >/dev/null 2>&1; then
    echo "$REPO_ROOT is not a Git checkout." >&2
    exit 1
fi

SOURCE_COMMIT=$(git -C "$REPO_ROOT" rev-parse HEAD)
SOURCE_STATUS=$(git -C "$REPO_ROOT" status --porcelain=v1 --untracked-files=all)
if [ -n "$SOURCE_STATUS" ]; then
    echo "Refusing to build an unattestable network image from a dirty checkout." >&2
    echo "$SOURCE_STATUS" >&2
    exit 1
fi

export E2E_LOCAL_EL_IMAGE
export GO_QRL_COMMIT=$SOURCE_COMMIT

cd "$REPO_ROOT"
"$DOCKER_BIN" buildx bake \
    --file "$SCRIPT_DIR/docker-bake.hcl" \
    network
