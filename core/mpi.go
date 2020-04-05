package core

import (
  "errors"
  "context"
  "fmt"
  "bufio"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/network"
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
    Pid: base,
    Ended: false,
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

  go func(){
    <- store.CloseChan()
    if mpi.Check() {
      mpi.Close()
    }
  }()

  go func(){
    <- host.CloseChan()
    if mpi.Check() {
      mpi.Close()
    }
  }()

  return &mpi, nil
}

type BasicMpi struct {
  Ctx context.Context
  Pid protocol.ID
  Ended bool
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
  m.Ended = true
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

func (m *BasicMpi)Check() bool {
  return !m.Ended
}

func (m *BasicMpi)Add(f string) error {
  if !m.Store().Has(f) {
    err := m.Store().Dowload(f)
    if err != nil {
      return err
    }
  }

  proto := protocol.ID(f + string(m.Pid))
  m.Host().SetStreamHandler(proto, func(stream network.Stream) {
    rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
    str, err := rw.ReadString('\n')
    if err != nil {
      return
    }

    param, err := ParamFromString(str[:len(str) - 1])
    if err != nil {
      return
    }

    inter, err := NewInterface(f, param.N, param.Idx)
    if err != nil {
      return
    }

    comm, err := NewSlaveComm(m.Ctx, m.Host(), rw, proto, inter, param)
    if err != nil {
      return
    }

    m.SlaveComms[param.Id] = comm
    go func(id string){
      <- comm.CloseChan()
      delete(m.SlaveComms, id)
    }(param.Id)
  })
  return nil
}

func (m *BasicMpi)Del(f string) error {
  err := m.Store().Del(f)
  if err != nil {
    return err
  }

  proto := protocol.ID(f + string(m.Pid))
  m.Host().RemoveStreamHandler(proto)
  return nil
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

  return m.Add(f)
}

func (m *BasicMpi)Start(file string, n int) error {
  if !m.MpiStore.Has(file) {
    return errors.New("no such file")
  }

  inter, err := NewInterface(m.Path + file, n, 0)
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
