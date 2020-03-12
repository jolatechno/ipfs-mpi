# ipfs-compute

[![](https://img.shields.io/badge/project-IPFS-blue.svg?style=flat-square)](https://ipfs.io/)
[![](https://img.shields.io/badge/freenode-%23ipfs-blue.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23ipfs)
[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

Distributed computing using ipfs as a back-bone

# ToDo

__main changes:__

- [ ] import __vanillaBlock.go__'s function directly from [ipfs/core/coreapi/block](https://github.com/ipfs/go-ipfs/blob/master/core/coreapi/block.go)
- [x] implement interpreter installation/uninstallation
- [x] implement interpreter execution
- [ ] call executer and installer from a custom BlockApi
- [ ] implement argument-passing through context

__'cosmetic' changes:__

- [x] write a boilerplate interpreter

__security:__

- [ ] limit the risk of malicious use (like _DDOS_ for example)

__long-term goals:__

- [ ] implement a _filecoin-like_ cryptocurrency rewarding computation
