# End-to-end suites

These Ginkgo suites validate go-qrl against an already-running
[development network](../devnet/README.md). Suites inspect the network but never
create or destroy it.

## Run

```bash
make network-start
make e2e-test E2E_PACKAGES=./testing/endtoend/suites/...
make network-stop
```

Use the same `DEVNET_ENCLAVE_NAME` for all commands when overriding the default.
`E2E_SUITE_TIMEOUT` controls the Ginkgo execution budget and defaults to `25m`.
Ginkgo continues through all selected suite packages and writes unified JUnit
and JSON reports to `build/cache/e2e`. Override `E2E_REPORT_DIR` to change that
location.

## Adding a suite

Add suites under `suites/<suite>`. Live bootstrap files use the `e2e` build tag
and open `internal/live.Session` for endpoints, clients, the development wallet,
and chain ID. Keep network lifecycle management outside the suites.

The [ABI](suites/abi/README.md), [API](suites/api/README.md),
[console](suites/console/README.md), and [Clef](suites/clef/README.md) suites
document their focused commands and coverage.
