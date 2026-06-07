## Test 1559 balance + gasCap

This test contains an EIP-1559 consensus issue which happened on Ropsten, where `gqrl` did not properly account for the value transfer while doing the check on `max_fee_per_gas * gas_limit`.

Before the issue was fixed, this invocation allowed the transaction to pass into a block:
```console
$ go run . t8n --state.fork=Zond --input.alloc=testdata/12/alloc.json --input.txs=testdata/12/txs.json --input.env=testdata/12/env.json --output.alloc=stdout --output.result=stdout
```

With the fix applied, the result is:
```json
{
  "alloc": {
    "Q4d8A013e3E0F6Ae973e9ef0C3247932950741486fB1a1ead60214cB5e35B1406869D47ec7A16B2280e2c6fe931465C7BC0B1844F34d535800F81EaA6590a2895": {
      "balance": "0x501bd00"
    }
  },
  "result": {
    "stateRoot": "0x6ccffbf3ab9d461db1ed359cd9caa068c7a678dfcc738f133271a2d0ee46503b",
    "txRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
    "receiptsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
    "logsHash": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
    "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
    "receipts": [],
    "rejected": [
      {
        "index": 0,
        "error": "insufficient funds for gas * price + value: address Q4d8A013e3E0F6Ae973e9ef0C3247932950741486fB1a1ead60214cB5e35B1406869D47ec7A16B2280e2c6fe931465C7BC0B1844F34d535800F81EaA6590a2895 have 84000000 want 84000032"
      }
    ],
    "gasUsed": "0x0",
    "currentBaseFee": "0x20",
    "withdrawalsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
  }
}
```

The transaction is rejected.
