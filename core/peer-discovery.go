package core

import (
  "context"

  "github.com/libp2p/go-libp2p-core/host"
)

func StartDiscovery(ctx context.Context, host host.Host, rendezvous string) {
  peerChan := initMDNS(ctx, host, rendezvous)

  go func() {
    for {
      peer := <- peerChan
      go func(){
        host.Connect(ctx, peer)
      }()
    }
  }()
}
