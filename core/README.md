# libp2p-mpi/core

[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

## How does it work?

### Interfaces

All of the interfaces use by the main [mpi](./mpi.go) interface are defined in the [type_definition.go](./type_definition.go) file.

Using [mpi](./mpi.go)`.SetInitFunctions()` you can set the init functions for all other interfaces.

#### ExtHost

`ExtHost` is an extended go-libp2p host interface that implements functions to manage peerstores for each interpreter.

#### Store

The `Store` interface is an ipfs interface to store interpreters.

#### Remote

The `Remote` interface implements the connection between two peers and peer reseting.

#### Interface

The `Interface` interface implements the interactions between a `SlaveComm` interface and a local interpreter.

#### SlaveComm

The `SlaveComm` interface handles the interactions between the `Remotes` and a local `Interface`.

#### MasterComm

The `MasterComm` interface is a wrap-around of the `SlaveComm` interface.

#### standardFunctionsCloser

A `standardFunctionsCloser` interface is defined in the [standardInterface.go](./standardInterface.go) file, and is used in all other classes to handle the functions in the `standardFunctionsCloser` interface.

## Peer reset

The peer reset algorithm of libp2p-mpi is also defined in the [type_definition.go](./type_definition.go) file in the `ResetReader` function:

```go
func ResetReader(received int, sent []string, sendToRemote func(string), pushToComm func(string)) (readFromRemote func(string)) {
  offset := received

  for _, msg := range sent {
    sendToRemote(msg)
  }

  return func(msg string) {
    if offset > 0 {
      offset--
      return
    }

    pushToComm(msg)
  }
}
```

This function takes as argument:
 - the number of messages already received (`received`) to know how many messages to discard from the remote,
 - the list of all sent messages (`sent`) to re-send them,
 - the function that handles sending messages to the remote (`sendToRemote`),
 - and the function that pushes messages back to the comm (`readFromRemote`).

It returns a function that handles new messages from the remote (`readFromRemote`).
