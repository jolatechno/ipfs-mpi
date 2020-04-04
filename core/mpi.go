package core

import (
  "errors"
  "context"
  "fmt"
  
  "github.com/libp2p/go-libp2p-core/protocol"
)

func NewMpi(ctx context.Context) (Mpi, error) {
  return &BasicMpi{Ctx:ctx}, errors.New("not yet implemented")
}

type BasicMpi struct {
  Ctx context.Context
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

  inter, err := NewInterface(file)
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
