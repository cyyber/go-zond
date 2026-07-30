# ABI suite

```bash
# Compile the complete suite without running live tests.
go test -tags=e2e -run '^$' ./testing/endtoend/suites/abi

# Regeneration requires `hypc --version` to report commit.2b9a0f1d.
go generate ./testing/endtoend/suites/abi

# Run against an already-running development network.
make e2e-test E2E_PACKAGES=./testing/endtoend/suites/abi
```

## Coverage contract

The target is 100% of documented supported feature classes: every class has
core ABI coverage, every Go representation is generated and compiled, and each
compiler/VM/RPC boundary has a representative live assertion. It is not a
promise to exercise every type-width and nesting cross-product or to reach 100%
statement coverage.

Covered:

- Calls: VM integers, booleans, 64-byte addresses, fixed and dynamic bytes,
  strings, fixed and dynamic arrays, tuples, views, and overloaded methods.
- Errors: a complex custom error, `Error(string)`, `Panic(uint256)`, RPC revert
  data, and failed transaction receipts.
- Events and filters: successful transactions, scalar and composite events,
  supported indexed topics, overloaded events, and positive, negative,
  wildcard, and OR filters.
- Function values: two-word 68-byte values, dynamic offsets, callback
  execution, generated calls and events, and indexed hashing and filtering.
- Payable entrypoints: a named payable method plus distinct receive and
  fallback transactions.
- WebSocket subscriptions: a generated watcher with matching and non-matching
  indexed predicates.

Not currently supported and excluded from this target:

- Anonymous-event binding decoding and filtering.
- Composite indexed-topic construction and tuple reconstruction.
- Overloaded custom errors and generic `ABI.Unpack` error-name decoding.
- Anonymous tuple fields and unused fixed-point and hash ABI types.
- The upstream v2 binding API.
