package main

import (
  //"fmt"
  "context"

  //"github.com/jolatechno/go-timeout"
  "github.com/jolatechno/ipfs-mpi/core"
)

func main(){

  host, err := core.NewHost(context.Background())
  if err != nil {
    panic(err)
  }

  fmt.Println("Our adress is: ", host.ID())
  for _, addr := range host.Addrs() {
    fmt.Println("Swarm listenning on: ", addr)
  }
}
