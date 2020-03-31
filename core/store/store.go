package store

import (
  "context"
  "os"
  "errors"

  "github.com/jolatechno/ipfs-mpi/core/store/interfaces"
  "github.com/jolatechno/ipfs-mpi/core/store/store/ipfs"
  "github.com/jolatechno/ipfs-mpi/core/store/peerstore"
)

type Store struct {
  shell *file.IpfsShell
  host host.Host
  routingDiscovery *discovery.RoutingDiscovery
  protocol protocol.ID
  maxsize uint64
  path string
  Ipfs_store string
}

func NewStore(ctx context.Context, host host.Host, config Config) (*Store, error) {
  return nil, errors.New("Not yet implemented")
}

func (s *Store)Init(ctx context.Context) error {
  files := (*s.shell).List()

  for _, f := range files {
    err := errors.New("Not yet implemented")
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
  return errors.New("Not yet implemented")
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
