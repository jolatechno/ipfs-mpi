package core

import (
  "github.com/libp2p/go-libp2p-core/peer"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/protocol"
)

type Mpi interface {
  Close()
  Host() ExtHost
  Store() Store
  Get(uint64) error
  Start(string) error
}

type ExtHost interface {
  host.Host

  NewPeer(protocol.ID) peer.ID
}

type Store interface {
  Close()
  Add(string)
  List() []string
  Has(string) bool
  Del(string) error
  Dowload(string) error
  Occupied() (uint64, error)
  Get(uint64) (string, error)
}

type MasterComm interface {
  SlaveComm

  Present(int) bool
  Reset(int)
  Connect(int, peer.ID)
}

type SlaveComm interface {
  Interface() Interface
  Close()
  Send(int, string)
  Get(int) string
}

type Interface interface {
  Message() chan Message
  Request() chan int
  Push(string) error
}

type Message struct {
  To int
  Content string
}
