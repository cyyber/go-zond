# External signer E2E suite

This suite verifies the live node-to-Clef integration configured by the built-in
development-network profile.

## Coverage contract

- Discover the Clef account through `qrl_accounts`.
- Sign and verify text through `qrl_sign`.
- Sign a transaction through `qrl_signTransaction` and verify its sender.
- Sign and submit through `qrl_sendTransaction`, then verify the mined sender
  and receipt.

Clef's direct `account_*` API and account-management behavior remain covered by
the standalone Clef suite.
