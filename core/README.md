# ipfs-mpi-plugin

[![](https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square)](https://ipfs.io/)
[![](https://img.shields.io/badge/freenode-%23ipfs-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23ipfs)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

> mpi plugin for go-ipfs

## How to build ?

```
go build -o main
```

For more detail consult : `./core -h`

## Usage

Use different terminal windows to run

```
./main
```

## Architecture

This plugin dynamically matches peer using __libp2b__'s _rendezvous_ to handle message passing.

### WARNING : Development in progress, might contain bug

## License

MIT
