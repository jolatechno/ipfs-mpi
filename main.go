package main

import (
  "errors"
  "context"
  "bufio"
  "strings"
  "strconv"
  "os"
  "fmt"
  "log"

  "github.com/jolatechno/ipfs-mpi/core"
)

const (
  errorFormat = "\033[31m%s\033[0m\n"
  prompt = "libp2p-mpi>"
)

func main(){
  ctx := context.Background()

  config, quiet, err := ParseFlag()
  if err != nil {
    panic(err)
  }

  store, err := core.NewMpi(ctx, config)
  if err != nil {
    panic(err)
  }

  store.SetErrorHandler(func(err error) {
    if !quiet {
      log.Printf(errorFormat, err.Error())
    }
  })

  fmt.Println("Our adress is ", store.Host().ID())

  for _, addr := range store.Host().Addrs() {
    fmt.Println("swarm listening on ", addr)
  }

  reader := bufio.NewReader(os.Stdin)
  for store.Check() {
    cmd, err := reader.ReadString('\n')
    if err != nil {
      panic(err)
    }

    splitted := strings.Split(cmd, " ")
    if len(splitted) == 0 {
      continue
    }

    end_idx := len(splitted) - 1
    last_size := len(splitted[end_idx]) - 1
    splitted[end_idx] = splitted[end_idx][:last_size]

    switch splitted[0] {
    default:
      store.Raise(errors.New("Command not understood"))

    case "List":
      list := store.Store().List()
      for _, f := range list {
        fmt.Println(" ", f)
      }

    case "Start":
      if len(splitted) < 3 {
        store.Raise(errors.New("No size given"))
        continue
      }

      n, err := strconv.Atoi(splitted[2])
      if n <= 0 && err == nil {
        err = errors.New("Size not understood")
      }
      if err != nil {
        store.Raise(err)
        continue
      }

      go func() {
        err := store.Start(splitted[1], n, splitted[3:]...)
        if err != nil {
          store.Raise(err)
        }
      }()

    case "Add":
      if len(splitted) < 2 {
        store.Raise(errors.New("No file given"))
        continue
      }

      for _, f := range splitted[1:] {
        go func() {
          store.Add(f)
        }()
      }

    case "Del":
      if len(splitted) < 2 {
        store.Raise(errors.New("No file given"))
        continue
      }

      for _, f := range splitted[1:] {
        go func() {
          store.Del(f)
        }()
      }

    case "exit":
      store.Close()
      return
    }
  }
}
