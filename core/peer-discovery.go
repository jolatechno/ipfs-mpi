package core

import (
  "context"
  "time"

  "github.com/libp2p/go-libp2p-core/protocol"
)

func StartDiscovery(ctx context.Context, host ExtHost, file string, base protocol.ID) {
  string := file + string(base)
  //proto := protocol.ID(string)
  go func() {
    for host.Check() {
      peerChan := initMDNS(ctx, host, string)
      for {
        select {
        case peer := <- peerChan:
          go func(){
            host.Connect(ctx, peer)
          }()
        case <- time.After(ScanDuration):
          continue
        }
      }
    }
  }()
}
