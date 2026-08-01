# External signer E2E suite

This suite verifies the live node-to-Clef integration configured by the built-in
development-network profile.

## Coverage contract

- Discover the Clef account through `qrl_accounts`.
- Sign and verify text through `qrl_sign`.
- Propagate Clef rejection through `qrl_sign`.
- Sign a transaction with input and an access list through
  `qrl_signTransaction`, then verify all signed fields and its sender.
- Propagate Clef transaction rejection through `qrl_signTransaction` and
  `qrl_sendTransaction` without changing the account nonce or transaction pool.
- Sign and submit through `qrl_sendTransaction`, then verify the mined sender
  and receipt.
- Restart Clef and verify the running node reconnects for account discovery,
  data signing, and transaction signing.
- Reject requests while Clef is unavailable and recover after it restarts.
- Cancel a pending approval and verify the transaction is never submitted.

Clef's direct `account_*` API and account-management behavior remain covered by
the standalone Clef suite.
