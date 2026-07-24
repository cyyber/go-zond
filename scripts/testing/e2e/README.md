# End-to-end tests

This module contains independently selectable Ginkgo v2 suites and the small
QRL-specific libraries they share. Kurtosis owns the real test network; Ginkgo
owns suite discovery, ordering, timeouts, progress, cleanup, and the pass/fail
exit status.

Network lifecycle and test execution stay separate:

```bash
make network-start
make live-test E2E_SUITES=<suite>
make network-stop
```

Starting a network never runs tests. Running tests never creates or destroys a
network.

## Live network

The built-in configuration runs a real qrl-package network in Kurtosis with one
current-source go-qrl execution client, one pinned Qrysm beacon node, one pinned
Qrysm validator client, and one pinned genesis generator.

Requirements are Docker, Kurtosis CLI 1.20.0, Git, and Go 1.26. Set
`E2E_DOCKER_BIN` if the Docker CLI is not named `docker`.

Use one private directory for the complete lifecycle:

```bash
E2E_NETWORK_DIR=/tmp/my-go-qrl-network make network-start
E2E_NETWORK_DIR=/tmp/my-go-qrl-network \
  make live-test E2E_SUITES=<suite>
E2E_NETWORK_DIR=/tmp/my-go-qrl-network make network-stop
```

`network-start` requires a clean checkout, builds the current go-qrl image and
the pinned network dependencies, creates one uniquely named enclave, and runs a
pinned qrl-package revision. It refuses to reuse an existing network, so source
changes cannot silently run against a stale execution image.

Inspect the network without running tests:

```bash
go -C scripts/testing/e2e run ./cmd/e2e status \
  --network-dir /tmp/my-go-qrl-network
```

If provisioning is interrupted after enclave creation, exact ownership is
retained for `network-stop`. A lost create response leaves a name-only intent
that blocks replay but cannot authorize destruction without manual inspection.
Provisioning is never replayed automatically.

`network-stop` validates and destroys only the recorded enclave. It does not
stop the shared Kurtosis engine.

The network directory is private runtime state. Never upload `private/`, raw
qrl-package output, or raw enclave dumps. The root `network.json` is only the
sanitized `{"ready":true}` marker; exact ownership and the funded wallet remain
below `private/`.

## Suite runner

`E2E_SUITES` is required and accepts comma-separated suite directory names. The
Make target maps each name to `./suites/<name>` and the matching Ginkgo label,
then runs selected suites serially against the existing network.

Create `scripts/testing/e2e/suites/<suite>` with a `TestE2E` bootstrap and label
every live spec `e2e`, `live`, and `<suite>`. Run it with:

```bash
make live-test E2E_SUITES=<suite>
```

Call `suitekit.OpenLiveSession(ctx)` to authenticate the shared network and open
its RPC client and funded signer. The session exposes the RPC, GraphQL, and
WebSocket URLs, and `Close` releases both the client and the network mutation
lease.

Prefer one Ginkgo `Serial` spec with `SpecContext`, `By`, `DeferCleanup`, and a
timeout over another custom runner. Keep state-changing scenarios rerunnable
through fresh contracts and current nonces. Do not start or stop the network
from a suite hook.
