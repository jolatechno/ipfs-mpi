package core

import (
  "context"
  "time"

  "github.com/libp2p/go-libp2p-core/host"
)

func StartDiscovery(ctx context.Context, host host.Host, rendezvous string, scanDuration time.Duration) {
  go func() {
    for {
      peerChan := initMDNS(ctx, host, rendezvous, scanDuration)
      timeoutChan := time.After(scanDuration)
      for {
        select {
        case peer := <- peerChan:
          go func(){
            host.Connect(ctx, peer)
          }()
        case <- timeoutChan:
          break;
        }
      }
    }
  }()
}
