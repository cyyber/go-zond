# Live API E2E suite

This suite validates the APIs exposed by the disposable qrl-package network.
It covers:

- HTTP JSON-RPC namespaces and read-only node metadata
- blocks, transactions, receipts, account state, storage, proofs, and calls
- transaction-pool inspection
- log and block filters
- WebSocket head, log, pending-transaction, and sync subscriptions
- read-only debug and tracing methods
- GraphQL queries, nested fields, and raw-transaction mutation

The fixture deploys a small contract that stores and returns a non-zero
64-byte value and emits a non-zero 64-byte topic. This keeps VM64 API
serialization part of the live assertions.

Run it against the separately started development network:

```bash
make e2e-test E2E_PACKAGES=./testing/endtoend/suites/api
```

The suite does not call APIs that change node configuration, rewrite chain
state, write files inside the execution container, or require the authenticated
Engine endpoint. It covers the node-managed signing methods; the standalone
`account_*` Clef API remains in the separate Clef live suite.
