# Devnet and end-to-end tests

This module provides a Kurtosis-backed QRL development network and the Ginkgo
suites that run against it. It requires Docker, the Kurtosis CLI, and a running
local Kurtosis 1.20 engine.

## Run

```bash
make network-start
make e2e-test E2E_PACKAGES=./suites/<suite>
make network-stop
```

`network-start` builds the current go-qrl tree and waits for readiness. Tests
never create or destroy the network.

## Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `DEVNET_ENCLAVE_NAME` | `go-qrl-devnet` (CLI default) | Kurtosis enclave |
| `DEVNET_EXECUTION_IMAGE` | `local/go-qrl:devnet` | Tag for the locally built execution image |
| `DEVNET_START_TIMEOUT` | `30m` (CLI default) | Network startup budget |
| `DEVNET_PARAMS_FILE` | unset | Complete qrl-package JSON parameters |
| `E2E_PACKAGES` | required | Suite packages passed to Ginkgo |
| `E2E_SUITE_TIMEOUT` | `25m` | Suite execution budget |

Use the same enclave name for every command:

```bash
DEVNET_ENCLAVE_NAME=my-devnet make network-start
DEVNET_ENCLAVE_NAME=my-devnet \
  make e2e-test E2E_PACKAGES=./suites/<suite>
DEVNET_ENCLAVE_NAME=my-devnet make network-stop
```

Kurtosis restricts enclave names to letters, digits, and dashes. Operations using the same name
must run serially. Concurrent networks need different names; concurrent builds
from different source trees also need different `DEVNET_EXECUTION_IMAGE` tags.

## Custom parameters

`DEVNET_PARAMS_FILE` replaces the built-in single-node profile with a complete
qrl-package JSON argument object. Two exact JSON string tokens are substituted:

```text
__DEVNET_EXECUTION_IMAGE__
__DEVNET_WALLET_ADDRESS__
```

The first participant's `el_image` must use the image token.
`network_params.prefunded_accounts` must contain the wallet token as a key; the
wallet token may also be used as a value, such as `withdrawal_address`.

The controller expects service `el-1-gqrl-qrysm` with public `rpc` and `ws`
ports; the reported GraphQL URL is live only if the profile enables GraphQL on
the rpc port (the built-in profile passes `--graphql`). Readiness requires
advancing blocks and a funded development wallet.

## Adding a suite

Add suites under `suites/<suite>`. Live bootstrap files use the `e2e` build tag
and call `network.Inspect(ctx)` for endpoints. Use
`devnet.UnsafeDevelopmentWallet()` for the funded signer. Suites must not manage
the network lifecycle.

## Safety

[`testdata/unsafe-development-wallet.seed`](testdata/unsafe-development-wallet.seed)
is public and embedded in the devnet binary. Never fund or use it outside
disposable local development networks.

After a failed start, run `make network-stop` with the same enclave name before
retrying.
