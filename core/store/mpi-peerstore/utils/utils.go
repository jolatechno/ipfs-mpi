package utils

import (
  "context"
  "sync"

  "github.com/libp2p/go-libp2p-discovery"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/peer"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	maddr "github.com/multiformats/go-multiaddr"
)

func NewKadmeliaDHT(ctx context.Context, host host.Host, BootstrapPeers []maddr.Multiaddr) (*discovery.RoutingDiscovery, error){
  if len(BootstrapPeers) == 0 {
    BootstrapPeers = dht.DefaultBootstrapPeers
  }

  kademliaDHT, err := dht.New(ctx, host)
  if err != nil {
    return nil, err
  }

  var wg sync.WaitGroup
  for _, peerAddr := range BootstrapPeers {
    peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
    wg.Add(1)
    go func() {
      defer wg.Done()
      host.Connect(ctx, *peerinfo)
    }()
  }
  wg.Wait()

  return discovery.NewRoutingDiscovery(kademliaDHT), nil
}
