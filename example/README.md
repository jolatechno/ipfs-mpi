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
added QmWRhweY1vALcp4U6EiBSeBBRQbiuv7bG9KVVdVrHYJV9L test_interpreters/echo/0.0.0/run.py
added QmUQT4c4btHFZGgcrSzzxXTstFWSe3eS32YhVCJYGqtut7 test_interpreters/dummy/0.0.0
added QmZ3cvbBb82beSnMZV1KP8rhzoSFJJETrtWgPm8N7xWBPM test_interpreters/dummy
added QmSeGjgV2zoMMJUT1CTWSYEyc8qudhFQthLuJZtDixAKFu test_interpreters/echo/0.0.0
added QmSxJdY2PoebhaE4fgDcwr4fTUyqyLF1aAvHXZPPCe9X84 test_interpreters/echo
added QmeqRpxbbjTWfPbr54WLTuPekUDeAaqmJbFQCEkLb6ABRT test_interpreters
 526 B / 526 B [======================================================================================================================================] 100.00%
```

Where the line corresponding to the whole directory is `added QmeqRpxbbjTWfPbr54WLTuPekUDeAaqmJbFQCEkLb6ABRT test_interpreters`, so the ipfs-store that you should use will be `QmeqRpxbbjTWfPbr54WLTuPekUDeAaqmJbFQCEkLb6ABRT/`

## How to interact with an interpreter ?

Using [ipfs-mpi/shell](../shell) as demonstrated in the [client.go](./) file (which demonstrate the [echo/0.0.0](./test_interpreters/echo/0.0.0) interpreter), you can send and receive messages and list connected peers.

The [client.go](./) file can be built using `go build -o client` and used as follow :

```
./client ipfs-mpi-port pid
```
