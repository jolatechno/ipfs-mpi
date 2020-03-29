# ipfs-mpi example

## How to build an interpreter directory ?

This example needs to be added to ipfs :

```
ipfs add -r example_interpretor
```

Wich will give an output ressembling :

```
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD test_interpreters/dummy/init.py
added Qmb5jKmyFQFDceBXLCkjdfQbJrJ5fQ6KLGJHESTJ8E5mZo test_interpreters/dummy/run.py
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD test_interpreters/echo/init.py
added QmZLAfUWm3y4xwnU9LQEmVAmRFKq5Yvtf6XDdJhgFS4xJq test_interpreters/echo/run.py
added QmV9f1et5dLVR8DCtgUWtgTjVbGsmQh8vcW6iBbTpBdbfd test_interpreters/pingBack/client
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD test_interpreters/pingBack/init.py
added QmfHbm9xzeRAAmrA282Sr6ruPvoQaZyAwZYM3p9Fh8Df53 test_interpreters/pingBack/run.py
added QmUQT4c4btHFZGgcrSzzxXTstFWSe3eS32YhVCJYGqtut7 test_interpreters/dummy
added QmfXzW7Xtd9zxLWM46NxX9cU6XsipT1FrPjJu92Mzyjb5s test_interpreters/echo
added QmVA3KKJ55NStY9SUw8ihqz97kaRo8nJs5TG9xejiy3uyA test_interpreters/pingBack
added QmWUSKJjVupUHSZm4kwA8R6TD55mLhwVXAyoSEjXx2r1z9 test_interpreters
 2.29 MiB / 2.29 MiB [================================================================================================================================] 100.00%
```

Where the line corresponding to the whole directory is `added QmWUSKJjVupUHSZm4kwA8R6TD55mLhwVXAyoSEjXx2r1z9 test_interpreters`, so the ipfs-store that you should use will be `QmWUSKJjVupUHSZm4kwA8R6TD55mLhwVXAyoSEjXx2r1z9/`

## How to interact with an interpreter ?

Using [ipfs-mpi/shell](../shell) as demonstrated in the [client.go](./) file (which demonstrate the [echo/0.0.0](./test_interpreters/echo/0.0.0) interpreter), you can send and receive messages and list connected peers.

The [client.go](./) file can be built using `go build -o client` and used as follow :

```
./client ipfs-mpi-port pid
```
