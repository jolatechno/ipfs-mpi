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

type Store struct {
  store map[file.File] *Entry
  host *host.Host
  routingDiscovery *discovery.RoutingDiscovery
  shell *file.IpfsShell
  protocol protocol.ID
}

func NewStore(ctx context.Context, url string, host host.Host, BootstrapPeers []maddr.Multiaddr, base protocol.ID) (*Store, error){
  shell, err := file.NewShell(url)
  if err != nil {
    return nil, err
  }
  routingDiscovery, err := utils.NewKadmeliaDHT(ctx, host, BootstrapPeers)

  if err != nil {
    return nil, err
  }

  store := make(map[file.File] *Entry)

  return &Store{ store:store, host:nil, routingDiscovery:routingDiscovery, shell:shell, protocol:base }, nil
}


func (s *Store)Add(f file.File, ctx context.Context){
  e := NewEntry(s.host, s.routingDiscovery, f, s.shell)
  e.InitEntry()
  e.LoadEntry(ctx, s.protocol)
  s.store[f] = e
}

func (s *Store)Del(f file.File) error{
  return s.shell.Del(f)
}

func (s *Store)Start(ctx context.Context) error{
  files := (*s.shell).List()

  for _, f := range files {
    s.Add(f, ctx)
    s.store[f].LoadEntry(ctx, s.protocol)
  }

  return nil
}

func (s *Store)Get(ctx context.Context){
  f := s.shell.Get()
  s.Add(f, ctx)
  s.store[f].InitEntry()
  s.store[f].LoadEntry(ctx, s.protocol)
}
