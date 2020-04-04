package core

import (
  "errors"
  "context"
  "fmt"

  "github.com/libp2p/go-libp2p-core/protocol"
)

func NewMpi(ctx context.Context, url string, path string, ipfs_store string, maxsize uint64, base protocol.ID) (Mpi, error) {
  var nilMpi *BasicMpi

  host, err := NewHost(ctx)
  if err != nil {
    return nilMpi, err
  }

  store, err := NewStore(url, path, ipfs_store)
  if err != nil {
    return nilMpi, err
  }

  mpi := BasicMpi{
    Ctx:ctx,
    Maxsize: maxsize,
    Path: path,
    Ipfs_store: ipfs_store,
    MpiHost: host,
    MpiStore: store,
    MasterComms: []MasterComm{},
    SlaveComms: []SlaveComm{},
    Id: 0,
  }

  return &mpi, nil
}

type BasicMpi struct {
  Ctx context.Context
  Maxsize uint64
  Path string
  Ipfs_store string
  MpiHost ExtHost
  MpiStore Store
  MasterComms []MasterComm
  SlaveComms []SlaveComm
  Id int
}

func (m *BasicMpi)Close() {
  m.Store().Close()
  m.Host().Close()

  for _, comm := range m.SlaveComms {
    comm.Close()
  }

  for _, comm := range m.MasterComms {
    comm.Close()
  }
}

func (m *BasicMpi)Host() ExtHost {
  return m.MpiHost
}

func (m *BasicMpi)Store() Store {
  return m.MpiStore
}

func (m *BasicMpi)Get(maxsize uint64) error {
  f, err := m.MpiStore.Get(maxsize)
  if err != nil {
    return err
  }

  return m.MpiStore.Dowload(f)
}

func (m *BasicMpi)Start(file string, n int) error {
  if !m.MpiStore.Has(file) {
    return errors.New("no such file")
  }

  inter, err := NewInterface(m.Path + file, n)
  if err != nil {
    return err
  }

  comm, err := NewMasterComm(m.Ctx, m.Host(), n, protocol.ID(file), inter, fmt.Sprint(m.Host().ID(), m.Id))
  if err != nil {
    return err
  }

  m.Id++
  m.MasterComms = append(m.MasterComms, comm)

  return nil
}
