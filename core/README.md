# ipfs-mpi-plugin

[![](https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square)](https://ipfs.io/)
[![](https://img.shields.io/badge/freenode-%23ipfs-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23ipfs)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

> mpi plugin for go-ipfs

## How to build ?

```
go build -o main
```

## Usage

You should first launch the ipfs daemon with `ipfs daemon` wich will output :

```
API server listening on /ip4/127.0.0.1/tcp/5001
```

Where `/ip4/127.0.0.1/tcp/5001` is the url of the ipfs daemon.

You can then run ipfs-mpi using :

```
./main -ipfs-api  /ip4/127.0.0.1/tcp/5001 -ipfs-store SomeIpfsDirectory/ -listen /ip4/YourIp/tcp/6666
```

For example :

```
./main -ipfs-api  /ip4/127.0.0.1/tcp/5001 -ipfs-store QmRfk8DdfrPQUxxThhgRxpPYvoa9qpjwV1veqXaSYgrrWf/ -listen /ip4/192.168.1.12/tcp/6666
```

## Architecture

This plugin dynamically matches peer using __libp2b__'s _rendezvous_ to handle message passing.

### WARNING : Development in progress, might contain bug

## License

MIT
