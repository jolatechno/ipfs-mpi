package store

import (
  "context"

  "github.com/jolatechno/mpi-peerstore/utils"
  "github.com/jolatechno/ipfs-mpi/core/ipfs-interface"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-discovery"
  "github.com/libp2p/go-libp2p-core/host"
  maddr "github.com/multiformats/go-multiaddr"
)

type Store struct{
  store map[file.File] *Entry
  host *host.Host
  routingDiscovery *discovery.RoutingDiscovery
  shell *file.IpfsShell
}

func NewStore(ctx context.Context, url string, host host.Host, BootstrapPeers []maddr.Multiaddr) (*Store, error){
  shell, err := file.NewShell(url)
  if err != nil {
    return nil, err
  }
  routingDiscovery, err := utils.NewKadmeliaDHT(ctx, host, BootstrapPeers)

  if err != nil {
    return nil, err
  }

  store := make(map[file.File] *Entry)

  return &Store{ store:store, host:nil, routingDiscovery:routingDiscovery, shell:shell }, nil
}


func (s *Store)Add(f file.File){
  e := NewEntry(s.host, s.routingDiscovery, f, s.shell)
  s.store[f] = e
}

func (s *Store)Del(f file.File) error{
  return s.shell.Del(f)
}

func (s *Store)Start(ctx context.Context, base protocol.ID) error{
  files := (*s.shell).List()

  for _, f := range files {
    s.Add(f)
    s.store[f].LoadEntry(ctx, base)
  }

  return nil
}

func (s *Store)Get(ctx context.Context, base protocol.ID){
  f := s.shell.Get()
  s.Add(f)
  s.store[f].InitEntry()
  s.store[f].LoadEntry(ctx, base)
}
