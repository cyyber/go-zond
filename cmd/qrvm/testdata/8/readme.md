## EIP-2930 testing

This test contains testcases for EIP-2930, which uses transactions with access lists.

### Prestate

The alloc portion contains one contract (`Q0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000aaaa`), containing the following code: `0x5854505854`: `PC; SLOAD; POP; PC; SLOAD`.

Essentially, this contract does `SLOAD(0)` and `SLOAD(3)`.

The alloc also contains some funds on `Q4d8A013e3E0F6Ae973e9ef0C3247932950741486fB1a1ead60214cB5e35B1406869D47ec7A16B2280e2c6fe931465C7BC0B1844F34d535800F81EaA6590a2895`.

## Transactions

There are three transactions, each invokes the contract above.

1. ACL-transaction, which contains some non-used slots
2. Regular transaction
3. ACL-transaction, which contains the slots `0` and `3` in the contract above

## Execution

Running it yields:
```console
$ go run . t8n --state.fork=Zond --input.alloc=testdata/8/alloc.json --input.txs=testdata/8/txs.json --input.env=testdata/8/env.json --trace 2>/dev/null && grep SLOAD trace-*
{"pc":1,"op":84,"gas":"0x484be","gasCost":"0x834","memSize":0,"stack":[[0,0,0,0,0,0,0,0]],"depth":1,"refund":0,"opName":"SLOAD"}
{"pc":4,"op":84,"gas":"0x47c86","gasCost":"0x834","memSize":0,"stack":[[3,0,0,0,0,0,0,0]],"depth":1,"refund":0,"opName":"SLOAD"}
{"pc":1,"op":84,"gas":"0x49cf6","gasCost":"0x834","memSize":0,"stack":[[0,0,0,0,0,0,0,0]],"depth":1,"refund":0,"opName":"SLOAD"}
{"pc":4,"op":84,"gas":"0x494be","gasCost":"0x834","memSize":0,"stack":[[3,0,0,0,0,0,0,0]],"depth":1,"refund":0,"opName":"SLOAD"}
{"pc":1,"op":84,"gas":"0x484be","gasCost":"0x64","memSize":0,"stack":[[0,0,0,0,0,0,0,0]],"depth":1,"refund":0,"opName":"SLOAD"}
{"pc":4,"op":84,"gas":"0x48456","gasCost":"0x64","memSize":0,"stack":[[3,0,0,0,0,0,0,0]],"depth":1,"refund":0,"opName":"SLOAD"}
```
