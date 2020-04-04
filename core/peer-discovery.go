package core

import (
  "context"
  "time"
)

func StartDiscovery(ctx context.Context, host ExtHost, rendezvous string) {
  go func() {
    for {
      peerChan := initMDNS(ctx, host, rendezvous)
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
