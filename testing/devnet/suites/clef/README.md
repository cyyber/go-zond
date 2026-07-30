# Clef suite

```bash
# Compile the live suite without running it.
go test -tags=e2e -run '^$' ./testing/devnet/suites/clef

# Run against an already-running development network.
make e2e-test E2E_PACKAGES=./testing/devnet/suites/clef
```

## Coverage

The suite initializes a standalone Clef instance with the funded development
wallet and verifies these account RPC workflows:

- `account_list` returns the imported full Q-address.
- `account_signData` returns a valid ML-DSA-87 signature for plain text.
- `account_signTypedData` returns a valid signature for QRL typed data.
- `account_signTransaction` returns matching raw and decoded VM64 transaction
  representations with the expected public key, descriptor, signature, and
  sender.

The suite uses the live network's chain ID for Clef startup, typed-data signing,
and transaction signing. The transaction scenario also derives its nonce and
fees from the network, submits the Clef-signed transaction, and requires a
successful receipt.
