# Local test-network images

This directory contains the reproducible image build used by the independently
managed E2E network. Network lifecycle and suite execution are documented in
[`scripts/testing/e2e/README.md`](../testing/e2e/README.md).

Build all four images without starting a network:

```bash
make network-images
```

`docker-bake.hcl` declares four BuildKit targets:

- `execution` builds the current clean go-qrl commit;
- `beacon` and `validator` build the pinned Qrysm revision; and
- `genesis` builds the canonical VM64 genesis-generator revision.

The three support targets share `Dockerfile.support` and its Qrysm source
builder. Bake can therefore schedule them together and reuse the expensive
source and Go build layers.

## Requirements

- Docker with Buildx and a responsive daemon;
- Git; and
- network access for the digest-pinned bases and exact source revisions.

The build rejects a dirty go-qrl checkout because the execution binary and
image metadata embed the current Git commit. Support images retain the pinned
upstream provenance declared in `Dockerfile.support`. Builder and runtime bases
are pinned by digest in `docker-bake.hcl`, and source repositories are fetched
at exact 40-character commits.

The genesis source is
`cyyber/qrl-genesis-generator@360410c72353c3a337f078018b36877dbbe40799`,
which includes the canonical 64-byte-address, VM64 deposit-runtime, packed
deposit-storage, and base-fee fixes. The image build executes that generator
and validates those properties before producing the image.

The Qrysm source temporarily remains pinned to
`rgeraldes24/qrysm@8b80fa0c3f5a98f2edc3fc8b7b9c67808373cafb`. Its two
post-`cyyber/main` commits fix validator aggregation/REST error handling and a
stale Zond protobuf descriptor. Replace the fork locator once those focused
fixes land in `cyyber/qrysm`.

Support-image tags include their pinned source revisions so concurrent
worktrees cannot silently replace each other's Qrysm or genesis inputs. Only
the execution-image name is configurable:

```bash
make network-images E2E_EXECUTION_IMAGE=local/go-qrl:mine
```
