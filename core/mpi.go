package core

import (
  "errors"
)

func NewMpi() (Mpi, error) {
  return &BasicMpi{}, errors.New("not yet implemented")
}

type BasicMpi struct {
  MpiHost ExtHost
  MpiStore Store
  MasterComms []MasterComm
  SlaveComms []SlaveComm
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

func (m *BasicMpi)Get(uint64) error {
  return errors.New("not yet implemented")
}

func (m *BasicMpi)Start(file string) error {
  return errors.New("not yet implemented")
}
