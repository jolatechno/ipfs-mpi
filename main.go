package main

import (
  "fmt"
  "context"
  "time"

  //"github.com/jolatechno/go-timeout"
  "github.com/jolatechno/ipfs-mpi/core"
)

func main(){
  ctx := context.Background()

  host, err := core.NewHost(ctx)
  if err != nil {
    panic(err)
  }

  core.StartDiscovery(ctx, host, "meet me here")

  fmt.Println("Our adress is: ", host.ID())
  for _, addr := range host.Addrs() {
    fmt.Println("Swarm listenning on: ", addr)
  }


  for {
    time.Sleep(time.Second)
    fmt.Println(host.Peerstore().Peers())
  }
}
