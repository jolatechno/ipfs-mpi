package store

import (
  "context"

  "github.com/jolatecno/mpi-peerstore"
  "github.com/jolatecno/mpi-peerstore/utils"
  "github.com/jolatecno/ipfs-mpi/ipfs-interface"

  "github.com/libp2p/go-libp2p"
  "github.com/libp2p/go-libp2p-discovery"
  "github.com/libp2p/go-libp2p-core/host"
  maddr "github.com/multiformats/go-multiaddr"
)

type Store struct{
  store map[file.File]Entry
  host *host.Host
  routingDiscovery *discovery.RoutingDiscovery
}

func NewStore(ctx context.Context, host host.Host, BootstrapPeers []maddr.Multiaddr) (Store, err){
  host, err := libp2p.New(ctx,
		libp2p.ListenAddrs([]multiaddr.Multiaddr(config.ListenAddresses)...),
	)

  if err != nil {
    return nil, err
  }

  routingDiscovery, err := utils.NewKadmeliaDHT(ctx, host, BootstrapPeers)

  if err != nil {
    return nil, err
  }

  store := make(map[file.File]Entry)

  return Store{ store:store, host:&host, routingDiscovery:routingDiscovery}
}

func (s *Store)Add(f file.File){
  e := NewEntry(s.host, s.routingDiscovery, f)
  s.store[f] = e
}

func (s *Store)Start(){
  files := file.List()
  for _, f := files {
    s.Add(f)
  }
}
