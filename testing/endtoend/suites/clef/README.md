# Clef suite

```bash
# Compile the live suite without running it.
go test -tags=e2e -run '^$' ./testing/endtoend/suites/clef

# Run against an already-running development network.
make e2e-test E2E_PACKAGES=./testing/endtoend/suites/clef
```

## Coverage contract

Covered:

The suite initializes a standalone Clef instance with the funded development
wallet and verifies:

- `account_version` matches the advertised external API version.
- `account_new` creates a distinct full Q-address through the approval and
  password prompts.
- `account_list` returns the imported full Q-address.
- `account_signData` returns a valid ML-DSA-87 signature for plain text.
- `account_signData` returns a valid signature for `data/validator`.
- `account_signTypedData` returns a valid signature for QRL typed data.
- `account_signTransaction` returns matching raw and decoded VM64 transaction
  representations with the expected public key, descriptor, signature, and
  sender.

The suite uses the live network's chain ID for Clef startup, typed-data signing,
and transaction signing. The transaction scenario also derives its nonce and
fees from the network, submits the Clef-signed transaction, and requires a
successful receipt.

Representative boundaries include full Q-addresses, ML-DSA signer metadata,
typed and validator-bound digests, raw transaction equivalence, and an
independently verified sender.

Excluded:

- Manual rejection paths and interactive use outside the controlled test input.
- UI-only methods and node-managed `qrl_*` signing APIs.
