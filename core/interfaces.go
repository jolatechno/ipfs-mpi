package core

import (
  "context"

  "github.com/libp2p/go-libp2p-core/peer"
)

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
  Stop()
  Send(int, string)
  Get(int) string
}

type MasterComm interface {
  Stop()
  Send(int, string)
  Get(int) string
  Present(context.Context, int) bool
  Reset(context.Context, int)
  Connect(context.Context, int, peer.ID)
}
