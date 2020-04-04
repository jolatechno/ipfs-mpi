package core

import (
  "errors"
  "context"
  "fmt"

  "github.com/libp2p/go-libp2p-core/protocol"
)

func NewMpi(ctx context.Context, url string, path string, ipfs_store string, maxsize uint64, base protocol.ID) (Mpi, error) {
  host, err := NewHost(ctx)
  if err != nil {
    return nil, err
  }

  store, err := NewStore(url, path, ipfs_store)
  if err != nil {
    return nil, err
  }

  mpi := BasicMpi{
    Ctx:ctx,
    Maxsize: maxsize,
    Path: path,
    EndChan: make(chan bool),
    Ipfs_store: ipfs_store,
    MpiHost: host,
    MpiStore: store,
    MasterComms: make(map[int]MasterComm),
    SlaveComms: make(map[string]SlaveComm),
    Id: 0,
  }

  return &mpi, nil
}

type BasicMpi struct {
  Ctx context.Context
  Maxsize uint64
  Path string
  EndChan chan bool
  Ipfs_store string
  MpiHost ExtHost
  MpiStore Store
  MasterComms map[int]MasterComm
  SlaveComms map[string]SlaveComm
  Id int
}

func (m *BasicMpi)Close() error {
  m.EndChan <- true
  err := m.Store().Close()
  if err != nil {
    return err
  }

  err = m.Host().Close()
  if err != nil {
    return err
  }

  for _, comm := range m.SlaveComms {
    err = comm.Close()
    if err != nil {
      return err
    }
  }

  for _, comm := range m.MasterComms {
    err = comm.Close()
    if err != nil {
      return err
    }
  }

  return nil
}

func (m *BasicMpi)CloseChan() chan bool {
  return m.EndChan
}

func (m *BasicMpi)Add(f string) error {
  return errors.New("not yet implemented")
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

  m.MasterComms[m.Id] = comm
  go func(id int){
    <- comm.CloseChan()
    delete(m.MasterComms, id)
  }(m.Id)

  m.Id++


  return nil
}
