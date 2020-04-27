  package core

import (
  "sync"
  "errors"
  "context"
  "fmt"
  "bufio"
  "io"
  "strings"
  "time"

  "github.com/libp2p/go-libp2p-core/peer"
  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/network"
  "github.com/ipfs/go-log"

  maddr "github.com/multiformats/go-multiaddr"
)

var (
  MpiHeader = "Mpi"
  MpiLogger = log.Logger(MpiHeader)

  ThrottleDuration = 50 * time.Millisecond
)

type safeInt struct {
  Value int
  Mutex sync.Mutex
}

func (i *safeInt)ReadIncrement() int {
  i.Mutex.Lock()
  defer i.Mutex.Unlock()
  return i.Value
}

type addrList []maddr.Multiaddr

func (al *addrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

func (al *addrList) Set(value string) error {
	addr, err := maddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}

type Config struct {
  Url string
  Path string
  Ipfs_store string
  Maxsize uint64
  Base string
  BootstrapPeers addrList
}

func NewMpi(ctx context.Context, config Config, host ExtHost, store Store) (Mpi, error) {
  proto := protocol.ID(config.Ipfs_store + config.Base)
  mpi := BasicMpi {
    Ctx:ctx,
    Pid: proto,
    Maxsize: config.Maxsize,
    Path: config.Path,
    Ipfs_store: config.Ipfs_store,
    MpiHost: host,
    MpiStore: store,
  }

  close := func() error {
    mpi.Ctx.Done()

    err :=mpi.Store().Close()
    if err != nil {
      return err
    }

    err = mpi.Host().Close()
    if err != nil {
      return err
    }

    mpi.ToClose.Range(func(key interface{}, value interface{}) bool {
      closer, ok := value.(io.Closer)
      if ok {
        closer.Close()
      }

      return true
    })

    return nil
  }

  mpi.Standard = NewStandardInterface(MpiHeader, close)

  defer func() {
    if err := recover(); err != nil {
      mpi.Raise(err.(error))
    }
  }()

  for _, f := range store.List() {
    mpi.Add(f)
  }

  go func() {
    for mpi.Check() {
      time.Sleep(ThrottleDuration)

      occupied, err := store.Occupied()
      if err != nil {
        return
      }

      left := config.Maxsize - occupied
      if left <= 0 {
        return
      }

      f, err := store.Get(left)
      if err != nil {
        store.Raise(SetNonPanic(err))
        return
      }

      err = mpi.Add(f)
      if err != nil {
        store.Raise(SetNonPanic(err))
      }
    }
  }()

  store.SetErrorHandler(func(err error) {
    mpi.Raise(err)
    if IsPanic(err) {
      go mpi.Close()
    }
  })

  store.SetCloseHandler(func() {
    go mpi.Close()
  })

  host.SetErrorHandler(func(err error) {
    mpi.Raise(err)
    if IsPanic(err) {
      go mpi.Close()
    }
  })

  host.SetCloseHandler(func() {
    go mpi.Close()
  })

  return &mpi, nil
}

type BasicMpi struct {
  ToClose sync.Map
  Ctx context.Context
  Pid protocol.ID
  Maxsize uint64
  Path string
  Ipfs_store string
  MpiHost ExtHost
  MpiStore Store
  Id safeInt
  Standard standardFunctionsCloser

  NewSlaveComm func(context.Context, ExtHost, io.ReadWriteCloser, protocol.ID, Param, Interface, []Remote) (SlaveComm, error)
  NewMasterSlaveComm func(context.Context, ExtHost, protocol.ID, Param, Interface, []Remote) (SlaveComm, error)
  NewMasterComm func(context.Context, SlaveComm, Param) (MasterComm, error)
  NewInterface func(ctx context.Context, file string, n int, i int, args ...string) (Interface, error)
  NewRemote func(context.Context, int) (Remote, error)
  NewLogger func(string, int, int) (func(string), error)
}

func (m *BasicMpi) SetInitFunctions(
  newSlaveComm func(context.Context, ExtHost, io.ReadWriteCloser, protocol.ID, Param, Interface, []Remote) (SlaveComm, error),
  newMasterSlaveComm func(context.Context, ExtHost, protocol.ID, Param, Interface, []Remote) (SlaveComm, error),
  newMasterComm func(context.Context, SlaveComm, Param) (MasterComm, error),
  newInterface func(context.Context, string, int, int, ...string) (Interface, error),
  newRemote func(context.Context, int) (Remote, error),
  newLogger func(string, int, int) (func(string), error)) {

  m.NewSlaveComm = newSlaveComm
  m.NewMasterSlaveComm = newMasterSlaveComm
  m.NewMasterComm = newMasterComm
  m.NewInterface = newInterface
  m.NewRemote = newRemote
  m.NewLogger = newLogger
}

func (m *BasicMpi)SetCloseHandler(handler func()) {
  m.Standard.SetCloseHandler(handler)
}

func (m *BasicMpi)SetErrorHandler(handler func(error)) {
  m.Standard.SetErrorHandler(handler)
}

func (m *BasicMpi)Raise(err error) {
  m.Standard.Raise(err)
}

func (m *BasicMpi)Check() bool {
  return m.Standard.Check()
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

func (m *BasicMpi)Del(f string) error {
  err := m.Store().Del(f, false)
  if err != nil {
    return err
  }

  proto := protocol.ID(f + string(m.Pid))
  m.Host().RemoveStreamHandler(proto)
  return nil
}

func (m *BasicMpi)Close() error {
  return m.Standard.Close()
}

func (m *BasicMpi)Add(file string) error {
  defer func() {
    if err := recover(); err != nil {
      m.Raise(err.(error))
    }
  }()

  if !m.Store().Has(file) {
    err := m.Store().Dowload(file)
    if err != nil {
      return err
    }
  }

  proto := protocol.ID(fmt.Sprintf("/%s/%s", file, m.Pid))
  m.Host().Listen(proto, fmt.Sprintf("/%s/%s", file, m.Ipfs_store))
  m.Host().SetStreamHandler(proto, func(stream network.Stream) {
    defer func() {
      if err := recover(); err != nil {
        m.Raise(err.(error))
      }
    }()

    reader := bufio.NewReader(stream)

    str, err := reader.ReadString('\n')
    if err != nil {
      return
    }

    MpiLogger.Debugf("Requested a new slaveComm, first message : %q", str) //--------------------------

    param, err := ParamFromString(str[:len(str) - 1])
    if err != nil {
      return
    }

    inter, err := m.NewInterface(m.Ctx, m.Path + InstalledHeader + file, param.N, param.Idx)
    if err != nil {
      return
    }

    logger, err := m.NewLogger(file, param.N, param.Idx)
    if err != nil {
      return
    }

    inter.SetLogger(logger)

    remotes := make([]Remote, param.N)
    for i := 0; i < param.N; i++ {
      if i == param.Idx {
        continue
      }

      remotes[i], err = m.NewRemote(m.Ctx, 0)
      if err != nil {
        return
      }
    }

    comm, err := m.NewSlaveComm(m.Ctx, m.Host(), stream.(io.ReadWriteCloser), proto, param, inter, remotes)
    if err != nil {
      return
    }

    stringId := string(param.Idx) + "/" + param.Id
    m.ToClose.Store(stringId, comm)

    comm.SetErrorHandler(func(err error) {
      if IsPanic(err) {
        comm.Close()
      }

      m.Raise(err)
    })

    comm.SetCloseHandler(func() {
      go m.ToClose.Delete(stringId)
    })
  })
  return nil
}

func (m *BasicMpi)Start(file string, n int, args ...string) (err error) {
  defer func() {
    if err := recover(); err != nil {
      m.Raise(err.(error))
    }
  }()

  if !m.MpiStore.Has(file) {
    return errors.New("no such file")
  }

  id := m.Id.ReadIncrement()

  proto := protocol.ID(fmt.Sprintf("/%s/%s", file, m.Pid))
  stringId := fmt.Sprintf("%d.%s", id, m.Host().ID())

  addrs := make([]peer.ID, n)
  for i := 0; i < n; i++ {
    if i == 0 {
      addrs[i] = m.Host().ID()
    } else {
      addrs[i], err = m.Host().NewPeer(proto)
      if err != nil {
        return err
      }
    }
  }

  inter, err := m.NewInterface(m.Ctx, m.Path + InstalledHeader + file, n, 0, args...)
  if err != nil {
    return err
  }

  logger, err := m.NewLogger(file, n, 0)
  if err != nil {
    return
  }

  inter.SetLogger(logger)

  param := Param {
    Addrs: addrs,
    Id: stringId,
    SlaveIds: make([]int, n),
    N: n,
  }

  remotes := make([]Remote, n)
  for i := 1; i < n; i++ {
    remotes[i], err = m.NewRemote(m.Ctx, 0)
    if err != nil {
      return err
    }
  }

  slaveComm, err := m.NewMasterSlaveComm(m.Ctx, m.Host(), proto, param, inter, remotes)
  if err != nil {
    return err
  }

  comm, err := m.NewMasterComm(m.Ctx, slaveComm, param)

  if err != nil {
    return err
  }

  m.ToClose.Store(stringId, comm)

  comm.SetErrorHandler(func(err error) {
    if IsPanic(err) {
      comm.Close()
    }

    m.Raise(err)
  })

  comm.SetCloseHandler(func() {
    go m.ToClose.Delete(stringId)
  })

  return nil
}
