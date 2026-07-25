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
make live-test E2E_SUITES=<suite>
make network-stop
```

Starting a network never runs tests. Running tests never creates or destroys a
network.

## Live network

The built-in configuration starts a real qrl-package network with one go-qrl
execution client, one Qrysm beacon node, one Qrysm validator, and one QRL
genesis generator. Requirements are Docker with Buildx, Kurtosis 1.20.0, Git,
and Go 1.26.

Use one private runtime directory for the complete lifecycle:

```bash
E2E_NETWORK_DIR=/tmp/my-go-qrl-network make network-start
E2E_NETWORK_DIR=/tmp/my-go-qrl-network \
  make live-test E2E_SUITES=<suite>
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

The network temporarily runs
`rgeraldes24/qrl-package@3892c3d2596403c080424d9e8fc99ff172483fe0`,
whose fork metadata supports remote execution. The pin omits only the later
Clef automation commit because this topology has remote signing disabled. Move
it to `cyyber/qrl-package` once that metadata lands there.

## Suite runner

`E2E_SUITES` is required and accepts comma-separated suite directory names.
Each name maps to `./suites/<name>`; package selection is the single source of
suite identity. Ginkgo labels describe capabilities, not package names. Put
`e2e` and `live` once on the suite's top-level Ginkgo container so its specs
inherit them; additional useful labels include `slow`, `serial`, and
`mutates-chain`.

Create `scripts/testing/e2e/suites/<suite>` with a `TestE2E` bootstrap, then run:

```bash
make live-test E2E_SUITES=<suite>
```

Call `suitekit.OpenLiveNetwork(ctx)` to authenticate the shared network and
hold its mutation lease. It exposes RPC, GraphQL, and WebSocket endpoints plus
the funded seed path. Each suite owns the clients, wallets, consoles, or command
processes it needs. Close the live network handle when the suite finishes.

Prefer Ginkgo `Serial` specs with `SpecContext`, `By`, `DeferCleanup`, and
explicit timeouts over another custom runner. Keep state-changing scenarios
rerunnable through fresh contracts and current nonces. Do not start or stop the
network from a suite hook.
