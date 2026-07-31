# VM and precompile E2E suite

This suite executes hand-written QRVM bytecode against the live development
network. It does not depend on Hyperion output.

## Coverage contract

- `PUSH33` through `PUSH64`, shifted `DUP1..DUP16` and `SWAP1..SWAP16`, and
  aligned and unaligned 64-byte `MSTORE`/`MLOAD` behavior.
- `CALLDATACOPY` and return-data behavior at 63, 64, and 65 bytes.
- `CALL`, `STATICCALL`, and `DELEGATECALL` success plus address, caller, and
  value context.
- `CREATE` and `CREATE2` address derivation, deployed code size, and child-code
  execution.
- `LOG0` through `LOG4` with full-width log data and ordered topic values.
- Every registered precompile with independently specified output and gas
  vectors, defined empty-input behavior, and an out-of-gas call. SHA-256 and
  identity gas are checked at 63, 64, and 65 bytes. ML-DSA-87 additionally
  covers malformed input rejection; the other precompiles define empty input
  as valid.
