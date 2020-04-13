# libp2p-mpi

####  *__WARNING : Development in progress, might contain bug, Please download releases to avoid bug__*

[![](https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square)](https://ipfs.io/)
[![](https://img.shields.io/badge/freenode-%23ipfs-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23ipfs)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

Message Passing Interface computing using libp2p as a back-bone to build computational pool.

## How to build an interpreter directory ?

This example needs to be added to ipfs :

```
ipfs add -r example_interpretor
```

Wich will give an output ressembling :

```
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD interpreter/dummy/init.py
added Qmb5jKmyFQFDceBXLCkjdfQbJrJ5fQ6KLGJHESTJ8E5mZo interpreter/dummy/run.py
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD interpreter/echo/init.py
added Qmccr1oeWtn2kumHvQScsVhCRqvTECsv9oBZu9i2QscUvY interpreter/echo/run.py
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD interpreter/multi_echo/init.py
added QmTcxaF2K1CrcZ4H6GT1GXU814tRwGbUEu6rNFXohR98R5 interpreter/multi_echo/run.py
added QmYjH4sibLpCSNCVbHXhfCZY78Ckj2aDDMWFwup7NGJ3JD interpreter/round_echo/init.py
added QmQVR7biYJP8578BADqmR2GKJrDMJAU8hmPzcd4ZJDjumt interpreter/round_echo/run.py
added QmUQT4c4btHFZGgcrSzzxXTstFWSe3eS32YhVCJYGqtut7 interpreter/dummy
added QmT29r7Pgba1q2crW8MW4dKwaGw3Wg72E6uLRTp17rFFyC interpreter/echo
added QmQyvtdRGSS6dq3dyDLRJNGsMoadKjC59wykyJHFTLRBiW interpreter/multi_echo
added Qme4447ijhvXDEqBNaRqPBpMgCCFNSzKRr2DshWrjD39vL interpreter/round_echo
added QmQvghovY96ujyrFGgPPRu6UKXuEBLiZrNuQePLCigCWKS interpreter
 1.90 KiB / 1.90 KiB [=========================================================================] 100.00%
 ```

Where the line corresponding to the whole directory is `added QmQvghovY96ujyrFGgPPRu6UKXuEBLiZrNuQePLCigCWKS interpreter`, so the ipfs-store that you should use will be `QmPL9GyMFUZmgVmgNrSGDQaQQzSmivu2K8sNVYPNp4LAw3/`

## How to build ?

```
go build
```

## Getting started

You should first launch the ipfs daemon with `ipfs daemon` wich will output :

```
API server listening on /ip4/127.0.0.1/tcp/5001
```

Where `/ip4/127.0.0.1/tcp/5001` is the url of the ipfs daemon.

You can then run the libp2p-mpi dameon using :

```
./libp2p-mpi -ipfs-api  /ip4/127.0.0.1/tcp/5001 -ipfs-store SomeIpfsDirectory/
```

For example :

```
./libp2p-mpi -ipfs-api  /ip4/127.0.0.1/tcp/5001 -ipfs-store Qmb5pxxiBDKiX9zZT3uPHeXYdAQ2keyNyk7QjzvbuAPkSe/
```

### Commands

#### List

You can list all installed interpreters using :

```
List
```

#### Start

You can start a interpreter using :

```
Start file n args...
```

with `file` being the name of the interpreter, `n` the size the number of peer to connect to and `args` a list of argument to pass to the interpreter

#### Add

You can add interpreters using :

```
Add interpreters...
```

with `interpreters` being a list of interpreter names.

#### Del

You can delete interpreters using :

```
Del interpreters...
```

with `interpreters` being a list of interpreter names.

#### exit

You can close the interface using `exit`.

## Usage

See [example](./example) for info on how to design an interpretor for ipfs-mpi.

### WARNING : Development in progress, might contain bug

# ToDo

__main changes:__

- [x] implement basic message passing using libp2p
- [x] handle fault (like unexpected peer hangup)
- [ ] automatically delete unused file

__'cosmetic' changes:__

- [x] write an example

__long-term goals:__

- [ ] implement a _filecoin-like_ cryptocurrency rewarding computation
