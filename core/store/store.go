package store

import (
  "context"

  "github.com/jolatechno/mpi-peerstore/utils"
  "github.com/jolatechno/ipfs-mpi/core/ipfs-interface"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-discovery"
  "github.com/libp2p/go-libp2p-core/host"
  maddr "github.com/multiformats/go-multiaddr"

  "fmt"
)

type Store struct {
  store map[file.File] *Entry
  host host.Host
  routingDiscovery *discovery.RoutingDiscovery
  shell *file.IpfsShell
  protocol protocol.ID
  maxsize uint64
}

func NewStore(ctx context.Context, url string, host host.Host, BootstrapPeers []maddr.Multiaddr, base protocol.ID, path string, ipfs_store string, maxsize uint64) (*Store, error) {
  shell, err := file.NewShell(url, path, ipfs_store)
  if err != nil {
    return nil, err
  }
  routingDiscovery, err := utils.NewKadmeliaDHT(ctx, host, BootstrapPeers)

  if err != nil {
    return nil, err
  }

  store := make(map[file.File] *Entry)

  return &Store{ store:store, host:host, routingDiscovery:routingDiscovery, shell:shell, protocol:base, maxsize:maxsize }, nil
}


func (s *Store)Add(f file.File, ctx context.Context) error {
  e := NewEntry(&s.host, s.routingDiscovery, f, s.shell)

  err := e.InitEntry()
  if err != nil {
    return err
  }

  err = e.LoadEntry(ctx, s.protocol)
  fmt.Println("store/store.go/Add ~ LoadEntry v")
  if err != nil {
    return err
  }

  s.store[f] = e
  return nil
}

func (s *Store)Del(f file.File) error {
  return s.shell.Del(f)
}

func (s *Store)Start(ctx context.Context) error {
  files := (*s.shell).List()

  fmt.Println("files : ", files)

  for _, f := range files {
    err := s.Add(f, ctx)
    if err != nil {
      return err
    }
  }

  go func(){
    for{
      err := s.Get(ctx)
      if err != nil { //No new file to add
        return
      }
    }
  }()

  return nil
}

func (s *Store)Get(ctx context.Context) error {
  used, err := s.shell.Occupied()
  if err != nil {
    return err
  }

  f, err := s.shell.Get(s.maxsize - used)
  if err != nil {
    return err
  }

  return s.Add(*f, ctx)
}
