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

The Make interface has one required package selector, `E2E_PACKAGES`, and five
optional overrides: `E2E_NETWORK_DIR`, `E2E_NETWORK_TIMEOUT`,
`E2E_SUITE_TIMEOUT`, `E2E_EXECUTION_IMAGE`, and `E2E_REQUIRE_CLEAN`.
`E2E_NETWORK_TIMEOUT` bounds provisioning; `E2E_SUITE_TIMEOUT` bounds Ginkgo
execution. The lifecycle targets normalize `E2E_NETWORK_DIR` to an absolute path
before passing it to the Go commands.

## Live network

The deliberately minimal configuration starts a real qrl-package network with
one go-qrl execution client, one Qrysm beacon node, one Qrysm validator, one
funded wallet, and one QRL genesis generator. Multi-node, Clef, explorer, and
transaction-spammer topologies are outside this minimal profile. They can be
added later as separate network profiles with suites that target them. The
required Go and Kurtosis versions are declared in [`go.mod`](go.mod); Docker and
Git are also required. A local Kurtosis 1.20 engine must already be running;
the lifecycle command connects to it but does not install or start it.

`network-start` first builds the current go-qrl working tree. Clean builds embed
the full commit; dirty builds embed `working-tree-<short-commit>` and print a
warning. Set `E2E_REQUIRE_CLEAN=1` to reject dirty builds. Build the image without
starting a network with `make network-image`. The root
[`Dockerfile`](../../Dockerfile) pins its builder and runtime bases by
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

Status is deliberately pass/fail: it prints `network ready` only after resolving
the directory's deterministic enclave slot and probing the advancing funded
chain. Unavailable or incomplete networks return a non-zero exit status.

The framework writes only the funded wallet seed.
The enclave name is a deterministic 192-bit digest of the canonical directory;
each lifecycle command resolves its current UUID from Kurtosis. `network-stop`
destroys that UUID, independently confirms that neither the UUID nor its
deterministic name remains, and leaves the shared Kurtosis engine running. A
create error is never adopted as success; run `network-stop` before retrying
because the deterministic slot may have been created despite a lost response.
If a normal runtime directory was deleted, `network-stop` recreates its private
slot directory. Existing directories must already have `0700` permissions; the
tool never changes permissions on a supplied path. Never upload the runtime
directory, qrl-package output, or raw enclave dumps.

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

Create `testing/endtoend/suites/<suite>` with a `TestE2E` bootstrap, then run:

```bash
make live-test E2E_PACKAGES=./suites/<suite>
```

Call `suitekit.InspectLiveNetwork(ctx)` to validate and inspect the shared network.
It exposes RPC, GraphQL, and WebSocket endpoints plus the funded seed path. Each
suite owns the clients, wallets, consoles, or command processes it needs.

Lifecycle commands and separate live-test invocations for the same
`E2E_NETWORK_DIR` must run serially. Use a different network directory for
independent concurrent runs.

Prefer Ginkgo specs with `SpecContext`, `By`, `DeferCleanup`, and explicit
timeouts over another custom runner. Keep state-changing scenarios rerunnable
through fresh contracts and current nonces. Do not start or stop the network
from a suite hook.
