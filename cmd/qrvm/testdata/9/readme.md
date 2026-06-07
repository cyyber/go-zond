## EIP-1559 testing

This test contains testcases for EIP-1559, which uses a new transaction type and has a new block parameter.

### Prestate

The alloc portion contains one contract (`Q0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000aaaa`), containing the following code: `0x58585454`: `PC; PC; SLOAD; SLOAD`.

Essentially, this contract does `SLOAD(0)` and `SLOAD(1)`.

The alloc also contains some funds on `Q4d8A013e3E0F6Ae973e9ef0C3247932950741486fB1a1ead60214cB5e35B1406869D47ec7A16B2280e2c6fe931465C7BC0B1844F34d535800F81EaA6590a2895`.

## Transactions

The transaction invokes the contract above.

1. EIP-1559 ACL-transaction, which contains the `0x0` slot for the contract above

## Execution

Running it yields:
```console
$ go run . t8n --state.fork=Zond --input.alloc=testdata/9/alloc.json --input.txs=testdata/9/txs.json --input.env=testdata/9/env.json --trace 2>/dev/null && grep SLOAD trace-*
{"pc":2,"op":84,"gas":"0x48c28","gasCost":"0x834","memSize":0,"stack":[[0,0,0,0,0,0,0,0],[1,0,0,0,0,0,0,0]],"depth":1,"refund":0,"opName":"SLOAD"}
{"pc":3,"op":84,"gas":"0x483f4","gasCost":"0x64","memSize":0,"stack":[[0,0,0,0,0,0,0,0],[0,0,0,0,0,0,0,0]],"depth":1,"refund":0,"opName":"SLOAD"}
```

We can also get the post-alloc:
```console
$ go run . t8n --state.fork=Zond --input.alloc=testdata/9/alloc.json --input.txs=testdata/9/txs.json --input.env=testdata/9/env.json --output.alloc=stdout 2>/dev/null
{
  "alloc": {
    "Q0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000aAaa": {
      "code": "0x58585454",
      "balance": "0x3",
      "nonce": "0x1"
    },
    "Q22222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222222": {
      "balance": "0xd6e0"
    },
    "Q4d8A013e3E0F6Ae973e9ef0C3247932950741486fB1a1ead60214cB5e35B1406869D47ec7A16B2280e2c6fe931465C7BC0B1844F34d535800F81EaA6590a2895": {
      "balance": "0xffe6fc39d8c920",
      "nonce": "0x1"
    }
  }
}
```
