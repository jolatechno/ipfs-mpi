# ipfs-mpi-shell

[![](https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square)](https://ipfs.io/)
[![](https://img.shields.io/badge/freenode-%23ipfs-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23ipfs)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

> mpi plugin using go-ipfs and go-libp2p

## Usage

To create a new shell use `Shell, c, err := shell.NewShell(port, pid)` where `port` is the port of the ipfs-mpi api.

`c` is a channel on which you will receive incoming messages.

`Shell.List(file)` will return `host, peers` where `host` is the host address and `peers` is a list of the addresses of all peers listening fore the `file` interpreter.

`Shell.Send(msg)` will send `msg` to `msg.To` (see the Message type of [ipfs-mpi/core/mpi-interface](../core/mpi-interface) for more information).

### WARNING : Development in progress, might contain bug

## License

MIT
