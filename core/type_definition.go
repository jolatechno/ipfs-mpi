package core

import (
  "io"
  "time"
  "context"

  "github.com/libp2p/go-libp2p-core/peerstore"
  "github.com/libp2p/go-libp2p-core/peer"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/network"
)

//-------

type standardFunctionsCloser interface {
  standardFunctions
  io.Closer
}

type standardFunctions interface {
  Check() bool
  Raise(error)
  SetCloseHandler(func())
  SetErrorHandler(func(error))
}

//-------

func ResetReader(received int, sent []string, sendToRemote func(string), pushToComm func(string)) (readFromRemote func(string)) {
  offset := received

  for _, msg := range sent {
    sendToRemote(msg)
  }

  return func(msg string) {
    if offset > 0 {
      offset--
      return
    }

    pushToComm(msg)
  }
}

//-------

type Mpi interface {
  standardFunctionsCloser

  SetInitFunctions(
    newSlaveComm func(context.Context, ExtHost, io.ReadWriteCloser, protocol.ID, Param, Interface, []Remote) (SlaveComm, error),
    newMasterSlaveComm func(context.Context, ExtHost, protocol.ID, Param, Interface, []Remote) (SlaveComm, error),
    newMasterComm func(context.Context, SlaveComm, Param) (MasterComm, error),
    newInterface func(context.Context, string, int, int, ...string) (Interface, error),
    newRemote func(context.Context, int) (Remote, error),
    newLogger func(string, int, int) (func(string), error),
  )

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
  Del(name string, failed bool) error
  Dowload(string) error
  Occupied() (uint64, error)
  Get(uint64) (string, error)
}

type MasterComm interface {
  standardFunctionsCloser

  SlaveComm() SlaveComm
  Reset(idx int, ith_time int)
}

type SlaveComm interface {
  standardFunctionsCloser

  Protocol() protocol.ID
  RequestReset(int)
  Start()
  Host() ExtHost
  Interface() Interface
  Remote(int) Remote
  Connect(int, peer.ID, ...interface{}) error
}

type Remote interface {
  standardFunctionsCloser

  SlaveId() int
  SetResetHandler(func(int, int))
  RequestReset(int, int)
  CloseRemote()
  SetPingInterval(time.Duration)
  SetPingTimeout(time.Duration)
  Stream() io.ReadWriteCloser
  Reset(stream io.ReadWriteCloser, slaveId int, msgs...interface{})
  Get() string
  WaitHandshake()
  Send(string)
  SendHandshake()
}

type Interface interface {
  standardFunctionsCloser

  Start()
  SetLogger(func(string))
  SetResetHandler(func(int))
  SetMessageHandler(func(int, string))
  SetRequestHandler(func(int))
  Push(string) error
}

type SelfStream interface {
  Reverse() (SelfStream, error)

  network.Stream
}
