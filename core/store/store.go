package store

import (
  "context"

  "github.com/jolatechno/mpi-peerstore"
  "github.com/jolatechno/mpi-peerstore/utils"
  "github.com/jolatechno/ipfs-mpi/core/ipfs-interface"

  "github.com/libp2p/go-libp2p"
  "github.com/libp2p/go-libp2p-discovery"
  "github.com/libp2p/go-libp2p-core/host"
  maddr "github.com/multiformats/go-multiaddr"
)

type Store struct{
  store map[file.File]Entry
  host *host.Host
  routingDiscovery *discovery.RoutingDiscovery
  shell *file.IpfsShell
}

func NewStore(ctx context.Context, url string, host host.Host, BootstrapPeers []maddr.Multiaddr) (*Store, err){
  shell, err := file.NewShell(url)
  if err != nil {
    return nil, err
  }

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

  return &Store{ store:store, host:&host, routingDiscovery:routingDiscovery, shell:shell }
}

func (s *Store)Add(f file.File){
  e := NewEntry(s.host, s.routingDiscovery, f, s.shell)
  s.store[f] = e
}

func (s *Store)Del(f file.File) error{
  return s.shell.Del(f)
}

func (s *Store)Start() error{
  files, err := s.shell.List()
  if err != nil {
    return err
  }

  for _, f := range files {
    s.Add(f)
    s.store[f].LoadEntry()
  }

  return nil
}

func (s *Store)Get(){
  f := file.Get()
  s.Add(f)
  s.store[f].InitEntry()
  s.store[f].LoadEntry()
}
