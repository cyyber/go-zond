#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd -- "$SCRIPT_DIR/../.." && pwd)

# The Go network controller derives these refs and immutable inputs from its
# pinned configuration. The inherited E2E namespace is scrubbed before these
# controller-owned values are passed to this script.
OUTPUT_EL_IMAGE=${E2E_LOCAL_EL_IMAGE:?E2E_LOCAL_EL_IMAGE is required}
OUTPUT_CL_IMAGE=${E2E_LOCAL_CL_IMAGE:?E2E_LOCAL_CL_IMAGE is required}
OUTPUT_VC_IMAGE=${E2E_LOCAL_VC_IMAGE:?E2E_LOCAL_VC_IMAGE is required}
OUTPUT_GENESIS_IMAGE=${E2E_LOCAL_GENESIS_IMAGE:?E2E_LOCAL_GENESIS_IMAGE is required}
for output_ref in "$OUTPUT_EL_IMAGE" "$OUTPUT_CL_IMAGE" "$OUTPUT_VC_IMAGE" "$OUTPUT_GENESIS_IMAGE"; do
    if [[ ! "$output_ref" =~ ^[A-Za-z0-9./:_-]+$ ]] || [[ "$output_ref" == *"@"* ]]; then
        echo "Controller-provided output image reference is invalid." >&2
        exit 2
    fi
done

LOCAL_EL_IMAGE=$OUTPUT_EL_IMAGE
LOCAL_CL_IMAGE=$OUTPUT_CL_IMAGE
LOCAL_VC_IMAGE=$OUTPUT_VC_IMAGE
LOCAL_GENESIS_IMAGE=$OUTPUT_GENESIS_IMAGE
DOCKER_BIN=${E2E_DOCKER_BIN:-docker}
PINNED_GO_BUILDER_IMAGE=${E2E_PINNED_GO_BUILDER_IMAGE:?E2E_PINNED_GO_BUILDER_IMAGE is required}
PINNED_ALPINE_RUNTIME_IMAGE=${E2E_PINNED_ALPINE_RUNTIME_IMAGE:?E2E_PINNED_ALPINE_RUNTIME_IMAGE is required}
PINNED_QRYSM_GO_BUILDER_IMAGE=${E2E_PINNED_QRYSM_GO_BUILDER_IMAGE:?E2E_PINNED_QRYSM_GO_BUILDER_IMAGE is required}
PINNED_CL_BASE_IMAGE=${E2E_PINNED_CL_BASE_IMAGE:?E2E_PINNED_CL_BASE_IMAGE is required}
PINNED_VC_BASE_IMAGE=${E2E_PINNED_VC_BASE_IMAGE:?E2E_PINNED_VC_BASE_IMAGE is required}
PINNED_QRYSM_GIT_REPO=${E2E_PINNED_QRYSM_GIT_REPO:?E2E_PINNED_QRYSM_GIT_REPO is required}
PINNED_QRYSM_GIT_COMMIT=${E2E_PINNED_QRYSM_GIT_COMMIT:?E2E_PINNED_QRYSM_GIT_COMMIT is required}
PINNED_GENESIS_GO_BUILDER_IMAGE=${E2E_PINNED_GENESIS_GO_BUILDER_IMAGE:?E2E_PINNED_GENESIS_GO_BUILDER_IMAGE is required}
PINNED_GENESIS_BASE_IMAGE=${E2E_PINNED_GENESIS_BASE_IMAGE:?E2E_PINNED_GENESIS_BASE_IMAGE is required}
PINNED_GENERATOR_GIT_REPO=${E2E_PINNED_GENERATOR_GIT_REPO:?E2E_PINNED_GENERATOR_GIT_REPO is required}
PINNED_GENERATOR_GIT_COMMIT=${E2E_PINNED_GENERATOR_GIT_COMMIT:?E2E_PINNED_GENERATOR_GIT_COMMIT is required}

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

echo "Building execution image from go-qrl $SOURCE_COMMIT."
"$DOCKER_BIN" build \
    --build-arg "COMMIT=$SOURCE_COMMIT" \
    --build-arg "GO_BUILDER_IMAGE=$PINNED_GO_BUILDER_IMAGE" \
    --build-arg "ALPINE_RUNTIME_IMAGE=$PINNED_ALPINE_RUNTIME_IMAGE" \
    --tag "$LOCAL_EL_IMAGE" \
    "$REPO_ROOT"

echo "Building consensus images from Qrysm $PINNED_QRYSM_GIT_COMMIT."
for target in beacon validator; do
    if [ "$target" = beacon ]; then
        output_image=$LOCAL_CL_IMAGE
    else
        output_image=$LOCAL_VC_IMAGE
    fi
    "$DOCKER_BIN" build \
        --file "$SCRIPT_DIR/Dockerfile.qrysm-consensus" \
        --target "$target" \
        --build-arg "QRYSM_GO_BUILDER_IMAGE=$PINNED_QRYSM_GO_BUILDER_IMAGE" \
        --build-arg "QRYSM_CL_BASE_IMAGE=$PINNED_CL_BASE_IMAGE" \
        --build-arg "QRYSM_VC_BASE_IMAGE=$PINNED_VC_BASE_IMAGE" \
        --build-arg "QRYSM_GIT_REPO=$PINNED_QRYSM_GIT_REPO" \
        --build-arg "QRYSM_GIT_COMMIT=$PINNED_QRYSM_GIT_COMMIT" \
        --tag "$output_image" \
        "$SCRIPT_DIR"
done

echo "Building genesis image from generator $PINNED_GENERATOR_GIT_COMMIT."
"$DOCKER_BIN" build \
    --file "$SCRIPT_DIR/Dockerfile.genesis-generator" \
    --build-arg "GENESIS_GO_BUILDER_IMAGE=$PINNED_GENESIS_GO_BUILDER_IMAGE" \
    --build-arg "GENESIS_BASE_IMAGE=$PINNED_GENESIS_BASE_IMAGE" \
    --build-arg "QRYSM_GIT_REPO=$PINNED_QRYSM_GIT_REPO" \
    --build-arg "QRYSM_GIT_COMMIT=$PINNED_QRYSM_GIT_COMMIT" \
    --build-arg "GENERATOR_GIT_REPO=$PINNED_GENERATOR_GIT_REPO" \
    --build-arg "GENERATOR_GIT_COMMIT=$PINNED_GENERATOR_GIT_COMMIT" \
    --tag "$LOCAL_GENESIS_IMAGE" \
    "$SCRIPT_DIR"

printf 'Built network images:\n  %s\n  %s\n  %s\n  %s\n' \
    "$LOCAL_EL_IMAGE" "$LOCAL_CL_IMAGE" "$LOCAL_VC_IMAGE" "$LOCAL_GENESIS_IMAGE"
