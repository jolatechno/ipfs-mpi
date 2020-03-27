# ipfs-mpi example

## How to build an interpreter directory ?

This example needs to be added to ipfs :

```
ipfs add -r example_interpretor
```

Wich will give an output ressembling :

```
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD test_interpreters/dummy/0.0.0/init.py
added Qmb5jKmyFQFDceBXLCkjdfQbJrJ5fQ6KLGJHESTJ8E5mZo test_interpreters/dummy/0.0.0/run.py
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD test_interpreters/echo/0.0.0/init.py
added QmdU1T4T4qMy7yr43s69G9ssFsHcHKxRHDZ4QFJWbpkmMd test_interpreters/echo/0.0.0/run.py
added QmUQT4c4btHFZGgcrSzzxXTstFWSe3eS32YhVCJYGqtut7 test_interpreters/dummy/0.0.0
added QmZ3cvbBb82beSnMZV1KP8rhzoSFJJETrtWgPm8N7xWBPM test_interpreters/dummy
added QmSkPvqwokmVkRKBe1Cz2sLUwy7nGGaVLuEAaHBPSNkjJy test_interpreters/echo/0.0.0
added QmbbPbSuEMCdND7zqyjcttCtaPXTq6CXnWinwJS5FnVF8e test_interpreters/echo
added QmVsod5dSMxnwZfmCVE2GWXDocPRwdUjLgkggx8LV8GVPq test_interpreters
 519 B / 519 B [======================================================================================================================================] 100.00%
```

Where the line corresponding to the whole directory is `added QmZyay35BsdK49tDCJpJgtJzXmtimupEbvWgisFbnPpvyi test_interpreters`, so the ipfs-store that you should use will be `QmZyay35BsdK49tDCJpJgtJzXmtimupEbvWgisFbnPpvyi/`

## How to interact with an interpreter ?

Using [ipfs-mpi/shell](../shell) as demonstrated in the [client.go](./) file (which demonstrate the [echo/0.0.0](./test_interpreters/echo/0.0.0) interpreter), you can send and receive messages and list connected peers.

The [client.go](./) file can be built using `go build -o client` and used as follow :

```
./client ipfs-mpi-port pid
```
