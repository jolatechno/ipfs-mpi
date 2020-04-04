package core

import (
  "github.com/libp2p/go-libp2p-core/peer"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/protocol"
)

type ExtHost interface {
  host.Host

  NewPeer(protocol.ID) peer.ID
}

type Store interface {
  Add(string)
  List() []string
  Has(string) bool
  Del(string) error
  Dowload(string) error
  Occupied() (uint64, error)
  Get(uint64) (string, error)
}

type SlaveComm interface {
  Close()
  Send(int, string)
  Get(int) string
}

type MasterComm interface {
  SlaveComm
  
  Present(int) bool
  Reset(int)
  Connect(int, peer.ID)
}
