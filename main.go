package main

import (
  "context"
  "bufio"
  "strings"
  "strconv"
  "os"
  "fmt"

  "github.com/jolatechno/ipfs-mpi/core"
)

func main(){
  ctx := context.Background()

  config, err := ParseFlag()
  if err != nil {
    panic(err)
  }

  store, err := core.NewMpi(ctx, config)
  if err != nil {
    panic(err)
  }

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

    if splitted[0] == "Start" {
      if len(splitted) < 3 {
        panic("No size given")
      }

      n, err := strconv.Atoi(splitted[2])
      if err != nil {
        panic(err)
      }

      store.Start(splitted[1], n, splitted[3:]...)

    } else if splitted[0] == "Add" {

      if len(splitted) < 2 {
        panic("No file given")
      }

      for _, f := range splitted[1:] {
        store.Add(f)
      }

    } else if splitted[0] == "Del" {
      if len(splitted) < 2 {
        panic("No file given")
      }

      for _, f := range splitted[1:] {
        store.Del(f)
      }

    } else if splitted[0] == "exit" {
      panic(store.Close())
    } else {
      panic("Command not understood")
    }
  }
}
