# Development network and end-to-end tests

This nested Go module contains the standalone QRL development-network command
and the Ginkgo suites that use it. Keeping both in one module isolates Kurtosis
and Ginkgo from go-qrl's root dependency graph without adding another module or
a public framework API.

Network lifecycle and suite execution are deliberately separate:

```bash
make network-start
make network-status
make live-test E2E_PACKAGES=./suites/<suite>
make network-stop
```

Starting a network never runs tests. Running tests never creates or destroys a
network. The root `make test` and `make lint` commands include this module.

## Development network

`network-start` builds the current go-qrl tree with the root
[`Dockerfile`](../Dockerfile), then runs the commit-pinned qrl-package through
Kurtosis. Docker, the Kurtosis CLI, and a running local Kurtosis 1.20 engine are
required.

The built-in profile creates one go-qrl execution client, one Qrysm beacon node,
one Qrysm validator, one QRL genesis generator, and one funded ML-DSA wallet.
The support images and qrl-package commit are defined in
[`internal/network/config.go`](internal/network/config.go).

Use `DEVNET_DIR` to select an independent deterministic network slot:

```bash
DEVNET_DIR=/tmp/my-go-qrl-devnet make network-start
DEVNET_DIR=/tmp/my-go-qrl-devnet make network-status
DEVNET_DIR=/tmp/my-go-qrl-devnet \
  make live-test E2E_PACKAGES=./suites/<suite>
DEVNET_DIR=/tmp/my-go-qrl-devnet make network-stop
```

Available settings are:

- `DEVNET_DIR` (default `/tmp/go-qrl-devnet`)
- `DEVNET_EXECUTION_IMAGE` (default `local/go-qrl:devnet`)
- `DEVNET_START_TIMEOUT` (default `30m`)
- `DEVNET_PARAMS_FILE` (optional complete JSON qrl-package parameters)
- `E2E_PACKAGES` (required by `live-test`)
- `E2E_SUITE_TIMEOUT` (default `25m`)

Relative runtime and parameter-file paths supplied through Make are normalized
to absolute paths.

### Custom parameters

Set `DEVNET_PARAMS_FILE` to replace the built-in profile with a complete JSON
qrl-package argument object. The file is passed through without schema merging
or numeric decoding, except for two exact JSON string tokens:

```text
__DEVNET_EXECUTION_IMAGE__
__DEVNET_WALLET_ADDRESS__
```

The first participant's `el_image` must be the execution-image token.
`network_params.prefunded_accounts` must contain the wallet token as a key. The
wallet token may also be used as a value, for example as
`withdrawal_address`. Exact tokens are replaced with the image built by Make
and the generated development-wallet address; token text embedded inside a
larger string is left unchanged.

Custom parameters may use any qrl-package fields, but the current controller
expects the primary execution service `el-1-gqrl-qrysm` with public `rpc` and
`ws` ports. Readiness requires a valid positive chain ID, a post-genesis block,
continued block production, and a positive generated-wallet balance. No
parameter copy, manifest, or checkpoint is written to the runtime directory.

## Lifecycle and private state

The runtime directory contains only `wallet.seed`. It must be an absolute
canonical `0700` directory; the seed is an exclusive, non-symlink `0600` file.
Do not upload it, qrl-package output, or raw enclave dumps.

The enclave name contains a deterministic 192-bit digest of the canonical
runtime directory. Each command resolves its current UUID from Kurtosis.
`network-stop` confirms independently that neither that UUID nor its exact name
remains and leaves the shared Kurtosis engine running. A failed create or
provisioning operation is never adopted as success; run `network-stop` before
retrying because its deterministic enclave slot may remain occupied.

Lifecycle commands and suites using the same `DEVNET_DIR` must run serially.
Use separate directories for concurrent networks. If accidental concurrency
creates duplicate exact-name enclaves, status and stop refuse the ambiguous
slot; remove the duplicates by UUID with the Kurtosis CLI.

## Suites

Create a suite in `devnet/suites/<suite>`. Put its live bootstrap in a file with
the `e2e` build tag and call `network.Inspect(ctx)` to obtain the RPC, GraphQL,
WebSocket, and funded-seed locations. Suites own their clients, contracts,
consoles, and cleanup; they must not start or stop the network.

`E2E_PACKAGES` accepts one or more module-relative package paths:

```bash
make live-test E2E_PACKAGES='./suites/goabi ./suites/commands'
```

Ginkgo owns discovery, progress, timeouts, and the pass/fail exit status. Keep
portable helpers and unit tests untagged. Compile live-only files without
starting a network with:

```bash
go -C devnet test -tags=e2e -run '^$' ./suites/...
```

Prefer `SpecContext`, `By`, `DeferCleanup`, and explicit timeouts over another
custom runner. State-changing scenarios should remain rerunnable through fresh
contracts and current nonces.
