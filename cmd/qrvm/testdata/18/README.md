# Invalid RLP

This folder contains a sample of invalid RLP, and it's expected that `t9n`
handles this properly:

```console
$ go run . t9n --input.txs=./testdata/18/invalid.rlp --state.fork=Zond
ERROR(11): rlp: value size exceeds available input length
```

Run `WRITE_FIXTURES=1 go test -run TestRegenerateT8nFixtures ./cmd/qrvm` to refresh.
