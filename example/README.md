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
added QmQXxoEsJSnN4BYuAXscbRLcwUP9U566YWkTtERF8MqVbq test_interpreters/echo/0.0.0/run.py
added QmUQT4c4btHFZGgcrSzzxXTstFWSe3eS32YhVCJYGqtut7 test_interpreters/dummy/0.0.0
added QmZ3cvbBb82beSnMZV1KP8rhzoSFJJETrtWgPm8N7xWBPM test_interpreters/dummy
added QmPCFULBU5tXMiJ3BnqFHWHA6w5UCA1u5WUbDD5gRP4r5s test_interpreters/echo/0.0.0
added Qmf5WSF2TedYanhtyMFM9t8zwCSkb9ASR9PJ21TuJ1ZRXT test_interpreters/echo
added QmPY2JaF8ZWqYY2zMm1c2RpfUKuZm4mk59ojMe62C6LYjc test_interpreters
 518 B / 518 B [======================================================================================================================================] 100.00%
```

Where the line corresponding to the whole directory is `added QmPY2JaF8ZWqYY2zMm1c2RpfUKuZm4mk59ojMe62C6LYjc test_interpreters`, so the ipfs-store that you should use will be `QmPY2JaF8ZWqYY2zMm1c2RpfUKuZm4mk59ojMe62C6LYjc/`

## How to interact with an interpreter ?

Using [ipfs-mpi/shell](../shell) as demonstrated in the [client.go](./) file (which demonstrate the [echo/0.0.0](./test_interpreters/echo/0.0.0) interpreter), you can send and receive messages and list connected peers.

The [client.go](./) file can be built using `go build -o client` and used as follow :

```
./client ipfs-mpi-port pid
```
