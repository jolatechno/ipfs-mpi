package core

import (
  "context"
  
  "github.com/libp2p/go-libp2p-core/host"

  "fmt"
)

func StartDiscovery(ctx context.Context, host host.Host, rendezvous string) {
  peerChan := initMDNS(ctx, host, rendezvous)

  go func() {
    for {
      peer := <- peerChan
      go func(){
        err := host.Connect(ctx, peer)

        if err != nil {
          fmt.Println("found peer ", peer.ID)
        }

      }()
    }
  }()
}
