package main

import (
  "context"
  "fmt"
  "time"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/network"

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

  string := "/test"
  proto := protocol.ID(string + "/0.0.0")
  host.Listen(proto, string)

  host.SetStreamHandler(proto, func(_ network.Stream) {
    fmt.Println("got a new stream")
  })

  go func() {
    l := 0
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

        _, err = host.NewStream(ctx, peer, proto)
        if err != nil {
          panic(err)
        }

        l++
      }
    }
  }()

  fmt.Println("Closed ", <- host.CloseChan())
}
