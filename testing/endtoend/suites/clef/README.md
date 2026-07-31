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
  password prompts, and the account remains listed after Clef restarts.
- `account_list` returns the imported full Q-address.
- `account_signData` returns a valid ML-DSA-87 signature for plain text.
- The ruleset rejects marked signing data and the external API returns the
  rejection to the caller.
- `account_signData` returns a valid signature for `data/validator`.
- `account_signTypedData` returns a valid signature for QRL typed data.
- Typed data for a chain ID other than Clef's configured chain is rejected.
- The ML-DSA-87 precompile verifies the signature produced by
  `account_signTypedData` with the wallet's descriptor-bound context.
- `account_signTransaction` returns matching raw and decoded VM64 transaction
  representations with the expected public key, descriptor, signature, and
  sender.
- The ruleset rejects the marked transaction and propagates the denial.
- A newly created password-protected account can still sign after Clef restarts.

The suite uses the live network's chain ID for Clef startup, typed-data signing,
and transaction signing. The transaction scenario also derives its nonce and
fees from the network, submits the Clef-signed transaction, and requires a
successful receipt.

Representative boundaries include full Q-addresses, ML-DSA signer metadata,
typed and validator-bound digests, raw transaction equivalence, and an
independently verified sender.

Excluded:

- Manual interactive rejection and use outside the controlled test input.
- UI-only methods and node-managed `qrl_*` signing APIs.
