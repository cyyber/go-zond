# VM and precompile E2E suite

This suite executes hand-written QRVM bytecode against the live development
network. It does not depend on Hyperion output.

## Coverage contract

- `PUSH33` through `PUSH64`, shifted `DUP1..DUP16` and `SWAP1..SWAP16`, and
  full 64-byte `MSTORE`/`MLOAD` behavior.
- `CALL`, `STATICCALL`, `DELEGATECALL`, `CREATE`, and `CREATE2`.
- A mined transaction with full-width log data and topic values.
- Every registered precompile with a successful vector, defined empty-input
  behavior, and an out-of-gas call. ML-DSA-87 additionally covers malformed
  input rejection; the other precompiles define empty input as valid.
