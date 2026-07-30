# Live API E2E suite

This suite validates the APIs exposed by the disposable qrl-package network.

## Coverage contract

The suite maps every exposed HTTP/WebSocket RPC method to a named live scenario
or an explicit exclusion. Manifest entries distinguish behavioral assertions
from response-shape checks and expected error dispatch.

Covered:

- HTTP JSON-RPC namespaces and read-only node metadata
- blocks, transactions, receipts, account state, storage, proofs, and calls
- non-empty pending and queued transaction-pool inspection
- log and block filters
- WebSocket head, log, pending hash, and full pending-transaction subscriptions
- read-only debug and tracing methods
- GraphQL schema, historical and pending transactions, access lists, nested
  account fields, and raw-transaction mutation

The fixture deploys a small contract that stores and returns a non-zero
64-byte value and emits a non-zero 64-byte topic. This keeps VM64 API
serialization part of the live assertions.

Run it against the separately started development network:

```bash
make e2e-test E2E_PACKAGES=./testing/endtoend/suites/api
```

Representative boundaries include full 64-byte addresses, storage values, and
log topics; pending and mined transactions; HTTP and WebSocket transports; and
successful, shape-only, and expected-error responses.

Excluded:

- APIs that change node configuration, rewrite chain state, or write files
  inside the execution container.
- The authenticated Engine endpoint and APIs disabled by the devnet profile.
- Active-sync GraphQL state and blocks containing withdrawals. The standard
  profile is already synced and does not create consensus withdrawals; those
  require dedicated lifecycle fixtures.
- Standalone `account_*` methods, which are covered by the Clef suite.
