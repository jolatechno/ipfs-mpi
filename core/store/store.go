package store

import (
  "context"
  "os"

  "github.com/jolatechno/mpi-peerstore/utils"
  "github.com/jolatechno/ipfs-mpi/core/ipfs-interface"
  "github.com/jolatechno/ipfs-mpi/core/api"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-discovery"
  "github.com/libp2p/go-libp2p-core/host"
  maddr "github.com/multiformats/go-multiaddr"
)

type Store struct {
  store map[file.File] *Entry
  host host.Host
  routingDiscovery *discovery.RoutingDiscovery
  shell *file.IpfsShell
  api *api.Api
  protocol protocol.ID
  maxsize uint64
  path string
}

func NewStore(ctx context.Context, host host.Host, base protocol.ID) (*Store, error) {
  store := make(map[file.File] *Entry)
  return &Store{ store:store, host:host, routingDiscovery:nil, shell:nil, api:nil, protocol:base, maxsize:1, path:"" }, nil
}

func (s *Store)StartDiscovery(ctx context.Context, BootstrapPeers []maddr.Multiaddr) error{
  routingDiscovery, err := utils.NewKadmeliaDHT(ctx, s.host, BootstrapPeers)
  if err != nil {
    return err
  }

  s.routingDiscovery = routingDiscovery
  return nil
}

func (s *Store)StartShell(url string, path string, ipfs_store string, maxsize uint64) error {
  if _, err := os.Stat(path); os.IsNotExist(err) {
    os.MkdirAll(path, file.ModePerm)
  } else if err != nil {
    return err
  }

  shell, err := file.NewShell(url, path, ipfs_store)
  if err != nil {
    return err
  }

  s.shell = shell
  s.path = path
  s.maxsize = maxsize
  s.shell = shell
  return nil
}

func (s *Store)StartApi(port int, ReadTimeout int, WriteTimeout int) error {
  api, err := api.NewApi(port, ReadTimeout, WriteTimeout)
  if err != nil {
    return err
  }

  s.api = api
  return nil
}

func (s *Store)Add(f file.File, ctx context.Context) error {
  e := NewEntry(&s.host, s.routingDiscovery, f, s.shell, s.api, s.path )

  err := e.InitEntry()
  if err != nil {
    return err
  }

  err = e.LoadEntry(ctx, s.protocol)
  if err != nil {
    return err
  }

  s.store[f] = e
  return nil
}

func (s *Store)Del(f file.File) error {
  return s.shell.Del(f)
}

func (s *Store)Init(ctx context.Context) error {
  files := (*s.shell).List()

  for _, f := range files {
    err := s.Add(f, ctx)
    if err != nil {
      return err
    }
  }

  /*go func(){
    for{
      err := s.Get(ctx)
      if err != nil { //No new file to add
        return
      }
    }
  }()*/

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
