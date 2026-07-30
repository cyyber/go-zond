# ABI suite

```bash
# Compile the complete suite without running live tests.
go test -tags=e2e -run '^$' ./testing/devnet/suites/abi

# Regeneration requires `hypc --version` to report commit.2b9a0f1d.
go generate ./testing/devnet/suites/abi

# Run against an already-running development network.
make e2e-test E2E_PACKAGES=./testing/devnet/suites/abi
```

## Coverage

Covered:

- Calls: VM integers, booleans, 64-byte addresses, fixed and dynamic bytes,
  strings, fixed and dynamic arrays, nested tuples, views, and overloaded
  methods through generic ABI, generated bindings, and raw RPC.
- Errors: a complex custom error, `Error(string)`, `Panic(uint256)`, RPC revert
  data, and failed transaction receipts.
- Events and filters: exact `Stored`, indexed dynamic, and composite event data
  and topics; generated decoding; and positive, negative, wildcard, and OR
  filters.

Function values, additional payable entrypoints, overloaded events, indexed
scalar events, and WebSocket subscriptions remain disabled pending their
underlying support.
