# End-to-end tests

This isolated Go module contains independently selectable Ginkgo v2 suites and
the small QRL-specific helpers they share. Kurtosis owns the live network;
Ginkgo owns suite discovery, timeouts, progress, cleanup, and the pass/fail
exit status.

Network lifecycle and suite execution remain separate:

```bash
make e2e-unit
make network-start
make network-status
make live-test E2E_PACKAGES=./suites/<suite>
make network-stop
```

Starting a network never runs tests. Running tests never creates or destroys a
network.

The Make interface has one required package selector, `E2E_PACKAGES`, and three
optional overrides: `E2E_NETWORK_DIR`, `E2E_TIMEOUT`, and
`E2E_EXECUTION_IMAGE`. The lifecycle targets normalize `E2E_NETWORK_DIR` to an
absolute path before passing it to the Go commands.

## Live network

The built-in configuration starts a real qrl-package network with one go-qrl
execution client, one Qrysm beacon node, one Qrysm validator, and one QRL
genesis generator. The required Go and Kurtosis versions are declared in
[`go.mod`](go.mod); Docker and Git are also required.

`network-start` first builds the current clean go-qrl checkout. Build the image
without starting a network with `make network-image`. The root
[`Dockerfile`](../../../Dockerfile) pins its builder and runtime bases by
digest. The organization-published support images are consumed directly.

Use one private runtime directory for the complete lifecycle:

```bash
E2E_NETWORK_DIR=/tmp/my-go-qrl-network make network-start
E2E_NETWORK_DIR=/tmp/my-go-qrl-network \
  make live-test E2E_PACKAGES=./suites/<suite>
E2E_NETWORK_DIR=/tmp/my-go-qrl-network make network-stop
```

Inspect readiness without running a suite:

```bash
E2E_NETWORK_DIR=/tmp/my-go-qrl-network make network-status
```

Status is deliberately pass/fail: it prints `network ready` only after
authenticating the recorded enclave and probing the advancing funded chain.
Unavailable or incomplete networks return a non-zero exit status.

The runtime directory contains only private lifecycle data: the exact enclave
name and UUID, the funded wallet seed, and a mutation lock. `network-stop`
destroys only that full recorded identity and leaves the shared Kurtosis engine
running. Never upload the runtime directory, qrl-package output, or raw enclave
dumps.

The qrl-package locator is commit-pinned and the organization-published
support-image references are defined once in
[`internal/network/config.go`](internal/network/config.go).

## Suite runner

`E2E_PACKAGES` is required and accepts one or more space-separated Go package
paths relative to this module. Package selection is the single source of suite
identity.

Quote the value when selecting multiple suites:

```bash
make live-test E2E_PACKAGES='./suites/goabi ./suites/commands'
```

Create `scripts/testing/e2e/suites/<suite>` with a `TestE2E` bootstrap, then run:

```bash
make live-test E2E_PACKAGES=./suites/<suite>
```

Call `suitekit.OpenLiveNetwork(ctx)` to authenticate the shared network and
hold its mutation lease. It exposes RPC, GraphQL, and WebSocket endpoints plus
the funded seed path. Each suite owns the clients, wallets, consoles, or command
processes it needs. Close the live network handle when the suite finishes.

Prefer Ginkgo specs with `SpecContext`, `By`, `DeferCleanup`, and explicit
timeouts over another custom runner. Keep state-changing scenarios rerunnable
through fresh contracts and current nonces. Do not start or stop the network
from a suite hook.
