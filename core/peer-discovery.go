package core

import (
  "context"
  "time"
)

func StartDiscovery(ctx context.Context, host ExtHost, rendezvous string, scanDuration time.Duration) {
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
