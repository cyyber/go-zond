# Live API E2E suite

This suite validates the APIs exposed by the disposable qrl-package network.

## Coverage contract

The suite maps every exposed HTTP/WebSocket RPC method to a named live scenario
or an explicit exclusion. Manifest entries distinguish behavioral assertions
from response-shape checks and expected error dispatch.

Each method is validated at the strongest stable level:

| API class | Assertion |
| --- | --- |
| Deterministic fixture data | Exact value |
| Runtime or network-dependent data | Stable invariant |
| Passive or variable response | Typed response shape |
| Unavailable prerequisite | Registered method and expected error |
| Unsafe or disabled method | Explicit exclusion |

Covered:

- HTTP JSON-RPC namespaces and read-only node metadata
- blocks, transactions, receipts, account state, storage, cryptographically
  verified account/storage proofs, exact access-list generation, and calls
- non-empty pending and queued transaction-pool inspection
- log and block filters
- WebSocket head, log, pending hash, and full pending-transaction subscriptions
- read-only debug and tracing methods
- GraphQL schema, historical and pending calls and gas estimates, exact block
  fields, transactions, access lists, nested account fields, and raw-transaction
  mutation
- structured transaction traces with VM64 opcode ordering and scalar stack values

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
- `admin_peerEvents` delivery. The fixed devnet topology covers subscription
  registration without adding or removing peers.
- Active-sync GraphQL state and blocks containing withdrawals. The standard
  profile is already synced and does not create consensus withdrawals; those
  require dedicated lifecycle fixtures.
- Successful debug paths that require bad blocks, preimages, external trace
  files, or ancient data. Where applicable, the suite covers registration and
  expected-error behavior instead.
- Standalone `account_*` methods, which are covered by the Clef suite.
