package main

import (
  "context"
  "fmt"
  "time"

  "github.com/libp2p/go-libp2p-core/protocol"

  "github.com/jolatechno/ipfs-mpi/core"

  dht "github.com/libp2p/go-libp2p-kad-dht"
)

func main() {
  ctx := context.Background()
  host, err := core.NewHost(ctx, dht.DefaultBootstrapPeers...)
  if err != nil {
    panic(err)
  }

  for _, addr := range host.Addrs() {
    fmt.Println("swarm listening on ", addr)
  }

  string := "test"
  proto := protocol.ID(string + "/0.0.0")
  host.Listen(proto, string)

  go func() {
    l := -1
    for {
      time.Sleep(time.Second)

      pstore, err := host.PeerstoreProtocol(proto)
      if err != nil {
        panic(err)
      }

      if len(pstore.Peers()) > l {
        fmt.Print(pstore.Peers())

        peer, err := host.NewPeer(proto)
        if err != nil {
          panic(err)
        }
        fmt.Println(", random peer : ", peer)
        l++
      }
    }
  }()

  fmt.Println("Closed ", <- host.CloseChan())
}
