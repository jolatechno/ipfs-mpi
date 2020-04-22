# libp2p-mpi/core

[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

## How does it work?

All of the interfaces use by libp2p-mpi are defined in the [type_definition.go](./type_definition.go) file.

A `standardFunctionsCloser` interface is defined in the [standardInterface.go](./standardInterface.go) file, and is used in all other classes to handle the functions in the `standardFunctionsCloser` interface.

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
