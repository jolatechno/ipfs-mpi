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
added QmYauZgRAAvxarnenQ9u8kQDRwAHRq4utRfSeqbbTpqws1 test_interpreters/echo/0.0.0/run.py
added QmUQT4c4btHFZGgcrSzzxXTstFWSe3eS32YhVCJYGqtut7 test_interpreters/dummy/0.0.0
added QmZ3cvbBb82beSnMZV1KP8rhzoSFJJETrtWgPm8N7xWBPM test_interpreters/dummy
added QmS9cQp6qcgQ5GfrhQdEte8DHurgVzQNzUVWToatFfNEnv test_interpreters/echo/0.0.0
added QmSsobFQjpYNXqQb6FLFzZHrLnZNENjh93brQYEaCRxfEf test_interpreters/echo
added QmasBzS9wap9gJjyrd8idTa1Ray4g1URNbJa8k8sKNMNxu test_interpreters
 519 B / 519 B [======================================================================================================================================] 100.00%
```

Where the line corresponding to the whole directory is `added QmasBzS9wap9gJjyrd8idTa1Ray4g1URNbJa8k8sKNMNxu test_interpreters`, so the ipfs-store that you should use will be `QmasBzS9wap9gJjyrd8idTa1Ray4g1URNbJa8k8sKNMNxu/`

## How to interact with an interpreter ?

Using [ipfs-mpi/shell](../shell) as demonstrated in the [client.go](./) file (which demonstrate the [echo/0.0.0](./test_interpreters/echo/0.0.0) interpreter), you can send and receive messages and list connected peers.

The [client.go](./) file can be built using `go build -o client` and used as follow :

```
./client ipfs-mpi-port pid
```
