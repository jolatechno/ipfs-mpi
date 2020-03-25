# ipfs-mpi-plugin

[![](https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square)](https://ipfs.io/)
[![](https://img.shields.io/badge/freenode-%23ipfs-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23ipfs)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

> mpi plugin for go-ipfs

## How to build ?

```
go build -o main
```

## Getting started

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

## Usage

This service will launch an api on a local port like `/127.0.0.1:8000`, push a message to other peers you need to formulate a request to this port with :

- a header named `Expected` which will be a string-formatted integer which will tel the api how many messages it should expect as a response.
- a header named `File``File` formatted as `interpreter_name/version`
- the request body with the message formatted as follow :

```json
{
  "Pid": 11,
  "messages" : [
      {
        "Pid":11,
        "To":"ToAdress1",
        "From":"YourAdress",
        "Data": [12, 32, 40]
      },
      {
        "Pid":11,
        "To":"ToAdress2",
        "From":"YourAdress",
        "Data": [20, 50, 51, 54]
      },
      {
        "Pid":11,
        "To":"ToAdress3",
        "From":"YourAdress",
        "Data": [12, 20]
      }
  ]
}
```

You will then receive the same json formatted message in response.

If you add a header named `List` containing any non-empty string you will get a response as follow :

```json
{
  "host":"HostAdress",
  "peers": ["Peer1Adress", "Peer2Adress", "Peer3Adress"...]
}
```

## Architecture

This plugin dynamically matches peer using __libp2b__'s _rendezvous_ to handle message passing.

### WARNING : Development in progress, might contain bug

## License

MIT
