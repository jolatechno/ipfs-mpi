# ipfs-mpi-core

[![](https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square)](https://ipfs.io/)
[![](https://img.shields.io/badge/freenode-%23ipfs-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23ipfs)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

> mpi plugin using go-ipfs and go-libp2p

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

You can then run ipfs-mpi using :

```
./core -ipfs-api  /ip4/127.0.0.1/tcp/5001 -ipfs-store SomeIpfsDirectory/ -listen /ip4/YourIp/tcp/6666
```

For example :

```
./core -ipfs-api  /ip4/127.0.0.1/tcp/5001 -ipfs-store QmRfk8DdfrPQUxxThhgRxpPYvoa9qpjwV1veqXaSYgrrWf/ -listen /ip4/192.168.1.12/tcp/6666
```

## Usage

See [ipfs-mpi/shell](../shell) for a Gide on designing software that use ipfs-mpi and [ipfs-mpi/example](../example) to design an interpretor for ipfs-mpi.

## Architecture

This plugin dynamically matches peer using __libp2b__'s _rendezvous_ to handle message passing.

### WARNING : Development in progress, might contain bug

## License

MIT
