# Local test network

This directory contains the image builder for the independently managed
qrl-package network used by E2E suites. The built-in network runs one execution
client, one Qrysm beacon node, one Qrysm validator, and the QRL genesis
generator in Docker through Kurtosis.

Network lifecycle remains separate from test execution:

```bash
make network-start
make live-test E2E_SUITES=<suite>
make network-stop
```

The first and third commands can be used without running any tests.

## Requirements

- Docker with a responsive daemon
- Kurtosis CLI 1.20.0
- Git
- Go 1.26
- Network access for the pinned source revisions, container bases, and
  qrl-package

The controller has one built-in qrl-package revision and topology. Its compact
configuration pins every source revision and builder/base image digest. At
startup it derives one-participant parameters with one fresh ML-DSA-87 prefund
and withdrawal address.

## Image preparation

`network-start` invokes the generic image builder as a provisioning stage. The
builder intentionally requires controller-provided output refs, so invoke it
through the lifecycle target:

```bash
E2E_NETWORK_DIR=/tmp/my-go-qrl-network make network-start
```

It builds:

- go-qrl from the clean current checkout with the exact commit embedded in the
  binary;
- beacon and validator binaries from the pinned Qrysm source revision; and
- the genesis generator and deposit runtime from pinned generator/Qrysm
  revisions.

Published Qrysm and generator images are used only as digest-pinned runtime
filesystems. Their older binaries and generator sources are replaced during
the build. The builder rejects tracked or untracked checkout changes so the
execution image cannot claim a commit that differs from the repository source.

## Start and inspect

Use a private runtime directory:

```bash
E2E_NETWORK_DIR=/tmp/my-go-qrl-network make network-start
```

Starting writes one `private/ownership.json`. A name-only creation intent blocks
replay if Kurtosis loses the create response; the full enclave UUID replaces it
as soon as creation succeeds. If provisioning then fails, run `network-stop`
before starting again. A ready network also blocks another start; use `status`
to inspect it or `network-stop` before rebuilding. Provisioning is never resumed
or replayed automatically.

Inspect the exact recorded enclave and endpoints:

```bash
go -C scripts/testing/e2e run ./cmd/e2e status \
  --network-dir /tmp/my-go-qrl-network
```

`network.json` contains only `{"ready":true}`. Status verifies the exact
recorded enclave, required services and endpoints, advancing chain ID, and
funded wallet. Enclave ownership, the mutation lock, and secret wallet material
live below `private/` and must not be shared.

## Stop

```bash
E2E_NETWORK_DIR=/tmp/my-go-qrl-network make network-stop
```

Stop validates the persisted enclave name and full UUID before destroying
only that enclave. It never stops the shared Kurtosis engine. Raw enclave dumps
are deliberately not uploadable because they can contain JWTs, keystores, and
funded-account material.

`start_local_testnet.sh` and `stop_local_testnet.sh` remain as thin wrappers
around the same generic Make targets; they contain no second lifecycle
implementation. Configure them with `E2E_NETWORK_DIR`; the start wrapper also
accepts `E2E_DOCKER_BIN` through the Make target.

See `scripts/testing/e2e/README.md` for suite selection and live coverage.
