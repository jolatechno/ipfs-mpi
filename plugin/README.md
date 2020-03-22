# WARNING : Development in progress

# ipfs-mpi-plugin

[![](https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square)](https://ipfs.io/)
[![](https://img.shields.io/badge/freenode-%23ipfs-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23ipfs)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

> mpi plugin for go-ipfs

## Architecture

This plugin dynamically matches peer using __libp2b__'s _rendezvous_ to handle message passing.

## ToDo

__main changes:__

- [ ] create an interface with standard Message Passing Interface (or MPI)
- [ ] implement message-passing through ipfs-key
- [ ] make the plugin buildable using the __build/build.sh__ file

## License

MIT
