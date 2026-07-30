# Console suite

```bash
# Compile the live suite without running it.
go test -tags=e2e -run '^$' ./testing/devnet/suites/console

# Regeneration requires `hypc --version` to report commit.2b9a0f1d.
go -C testing/devnet generate ./suites/console

# Run against an already-running development network.
make e2e-test E2E_PACKAGES=./testing/devnet/suites/console
```

## Coverage

The `api` scenario verifies block, header, state, fee, receipt, namespace, chain
ID, and QIP-55 address behavior through the embedded console and raw RPC.

The `contract` scenario signs and submits a deployment transaction, then
verifies VM64 scalar, dynamic, fixed-byte, and array ABI values; transaction and
receipt lookup; 64-byte event data and topics; generated event decoding; and
exact, wildcard, and OR log filters.

Every JavaScript check has a descriptive name that is printed on success and
included in failures.
