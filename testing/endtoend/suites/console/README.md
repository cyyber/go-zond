# Console suite

```bash
# Compile the live suite without running it.
go test -tags=e2e -run '^$' ./testing/endtoend/suites/console

# Regeneration requires `hypc --version` to report commit.2b9a0f1d.
go generate ./testing/endtoend/suites/console

# Run against an already-running development network.
make e2e-test E2E_PACKAGES=./testing/endtoend/suites/console
```

## Coverage contract

The console suite targets behavior implemented by the embedded web3 wrapper,
not exhaustive raw RPC coverage.

Covered:

- Provider dispatch plus block, header, state, fee, receipt, namespace, chain
  ID, and QIP-55 wrapper behavior.
- Contract deployment, pure calls, state-changing wrapper request formatting,
  signed raw submission, and canonical ABI-decoded Q-addresses.
- VM64 scalars, dynamic values, fixed-byte boundaries, fixed and dynamic
  arrays.
- Receipt logs, exact/wildcard/OR filters, generated event decoding, and
  WebSocket event watches with indexed dynamic string and bytes fields.

Representative boundaries include values with non-zero upper 256 bits,
`bytes33`/`bytes64`, data crossing a 64-byte boundary, and full 64-byte topics.

Excluded:

- Exhaustive raw JSON-RPC behavior, covered by the API suite.
- Generated Go bindings and unsupported ABI classes, covered by the ABI suite.
- Direct node-managed `qrl_sendTransaction`, covered by the external signer
  suite.
- Nested arrays containing dynamic elements, which remain outside the current
  embedded web3 ABI coverage.
- Node lifecycle and unsafe debug operations.
