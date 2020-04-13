package core

import (
  "bufio"

  "github.com/libp2p/go-libp2p-core/peerstore"
  "github.com/libp2p/go-libp2p-core/peer"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/network"
)

//-------

type standardFunctionsCloser interface {
  standardFunctions

  Close() error
}

type standardFunctions interface {
  Check() bool
  CloseChan() chan bool
  ErrorChan() chan error
}

type Message struct {
  To int
  Content string
}

//-------

type Mpi interface {
  standardFunctionsCloser

  Add(string) error
  Del(string) error
  Get(uint64) error

  Host() ExtHost
  Store() Store
  Start(string, int, ...string) error
}

type ExtHost interface {
  host.Host
  standardFunctions

  PeerstoreProtocol(protocol.ID) (peerstore.Peerstore, error)
  NewPeer(protocol.ID) (peer.ID, error)
  Listen(protocol.ID, string)
  SelfStream(...protocol.ID) (SelfStream, error)
}

type Store interface {
  standardFunctionsCloser

  Add(string)
  List() []string
  Has(string) bool
  Del(string) error
  Dowload(string) error
  Occupied() (uint64, error)
  Get(uint64) (string, error)
}

type MasterComm interface {
  standardFunctionsCloser

  SlaveComm() SlaveComm
  Connect(int, peer.ID, bool)
  CheckPeer(int) bool
  Reset(int)
}

type SlaveComm interface {
  standardFunctionsCloser

  Host() ExtHost
  Interface() Interface
  Remote(int) Remote
  Connect(int, peer.ID) error
}

type Remote interface {
  standardFunctionsCloser

  Stream() *bufio.ReadWriter
  Reset(*bufio.ReadWriter)
  Get() string
  GetHandshake() string
  Send(string)
  StreamHandler() (network.StreamHandler, error)
}

type Interface interface {
  standardFunctionsCloser

  Message() chan Message
  Request() chan int
  Push(string) error
}

type SelfStream interface {
  Reverse() (SelfStream, error)

  network.Stream
}
