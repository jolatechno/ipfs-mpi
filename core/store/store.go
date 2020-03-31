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
)

type Store struct {
  store map[file.File] Entry
  host host.Host
  routingDiscovery *discovery.RoutingDiscovery
  shell *file.IpfsShell
  Api *api.Api
  protocol protocol.ID
  maxsize uint64
  path string
  Ipfs_store string
}

func NewStore(ctx context.Context, host host.Host, config Config) (*Store, error) {
  store := make(map[file.File] Entry)
  proto := protocol.ID(config.Ipfs_store + config.ProtocolID)

  routingDiscovery, err := utils.NewKadmeliaDHT(ctx, host, config.BootstrapPeers)
  if err != nil {
    return nil, err
  }

if _, err := os.Stat(config.Path); os.IsNotExist(err) {
    os.MkdirAll(config.Path, file.ModePerm)
  } else if err != nil {
    return nil, err
  }

  shell, err := file.NewShell(config.Url, config.Path, config.Ipfs_store)
  if err != nil {
    return nil, err
  }

  api, err := api.NewApi(config.Api_port)
  if err != nil {
    return nil, err
  }

  return &Store{ store:store, host:host, routingDiscovery:routingDiscovery, shell:shell, Api:api, protocol:proto, maxsize:config.Maxsize, path:config.Path }, nil
}

func (s *Store)Init(ctx context.Context) error {
  files := (*s.shell).List()

  for _, f := range files {
    e := NewEntry(&s.host, s.routingDiscovery, f, s.shell, s.Api, s.path)
    err := e.LoadEntry(ctx, s.protocol)
    if err != nil {
      return err
    }

    s.store[f] = *e
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

func (s *Store)Add(f file.File, ctx context.Context) error {
  e := NewEntry(&s.host, s.routingDiscovery, f, s.shell, s.Api, s.path )

  err := e.InitEntry()
  if err != nil {
    return err
  }

  err = e.LoadEntry(ctx, s.protocol)
  if err != nil {
    return err
  }

  s.store[f] = *e
  return nil
}

func (s *Store)Del(f file.File) error {
  return s.shell.Del(f)
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
