# VM and precompile E2E suite

This suite executes hand-written QRVM bytecode against the live development
network. It does not depend on Hyperion output.

## Coverage contract

- `PUSH33` through `PUSH64`, shifted `DUP1..DUP16` and `SWAP1..SWAP16`, and
  aligned and unaligned 64-byte `MSTORE`/`MLOAD` behavior.
- 512-bit arithmetic and modular operations, upper-half comparisons and bitwise
  operations, signed comparisons, shifts, `BYTE`, and `SIGNEXTEND` across the
  former 32-byte boundary.
- Full-width `SSTORE`/`SLOAD` values and jump analysis when `JUMPDEST` bytes
  occur inside `PUSH33..PUSH64` data.
- `CALLDATALOAD`, `CALLDATACOPY`, `CODECOPY`, `EXTCODECOPY`,
  `RETURNDATACOPY`, and `KECCAK256` at 63, 64, and 65 bytes.
- `CALL`, `STATICCALL`, and `DELEGATECALL` success plus address, caller, and
  value context, including a successful non-zero value transfer and failed-call
  state and value rollback.
- `CREATE` and `CREATE2` address derivation, deployed code size, and child-code
  execution with zero and non-zero value, plus failed-creation state and value
  rollback.
- Exact full-width address, account, block, fee, randomness, and transaction
  context opcode results, including historical `BLOCKHASH`.
- `LOG0` through `LOG4` with full-width log data and ordered topic values.
- Every registered precompile with independently specified output and gas
  vectors, defined empty-input behavior, and an out-of-gas call. SHA-256 and
  identity gas are checked at 63, 64, and 65 bytes. Deposit-root uses a static
  non-zero vector, and ML-DSA-87 covers both malformed and full-length invalid
  input; the other precompiles define empty input as valid. A hand-written
  `STATICCALL` also verifies the SHA-256 success flag, output, and exact gas
  boundary from inside QRVM bytecode.
