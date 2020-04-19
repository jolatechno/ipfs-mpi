  package core

import (
  "sync"
  "errors"
  "context"
  "fmt"
  "bufio"
  "io"
  "strings"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/network"

  maddr "github.com/multiformats/go-multiaddr"
)

var (
  MpiHeader = "Mpi"
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

func NewMpi(ctx context.Context, config Config) (Mpi, error) {
  host, err := NewHost(ctx, config.BootstrapPeers...)
  if err != nil {
    return nil, err
  }

  store, err := NewStore(config.Url, config.Path, config.Ipfs_store)
  if err != nil {
    return nil, err
  }

  proto := protocol.ID(config.Ipfs_store + config.Base)
  mpi := BasicMpi {
    Ctx:ctx,
    Pid: proto,
    Maxsize: config.Maxsize,
    Path: config.Path,
    Ipfs_store: config.Ipfs_store,
    MpiHost: host,
    MpiStore: store,
    Standard: NewStandardInterface(),
  }

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
        return
      }

      err = mpi.Add(f)
      if err != nil {
        return
      }
    }
  }()

  store.SetErrorHandler(func(err error) {
    mpi.Raise(err)
    go mpi.Close()
  })

  store.SetCloseHandler(func() {
    go mpi.Close()
  })

  host.SetErrorHandler(func(err error) {
    mpi.Raise(err)
    go mpi.Close()
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
}

func (m *BasicMpi)SetCloseHandler(handler func()) {
  m.Standard.SetCloseHandler(handler)
}

func (m *BasicMpi)SetErrorHandler(handler func(error)) {
  m.Standard.SetErrorHandler(handler)
}

func (m *BasicMpi)Raise(err error) {
  hErr := NewHeadedError(err, true, MpiHeader)
  m.Standard.Raise(hErr)
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
  err := m.Store().Del(f)
  if err != nil {
    return err
  }

  proto := protocol.ID(f + string(m.Pid))
  m.Host().RemoveStreamHandler(proto)
  return nil
}

func (m *BasicMpi)Close() error {
  defer recover()

  if m.Check() {
    m.Standard.Close()

    m.Ctx.Done()

    err := m.Store().Close()
    if err != nil {
      return err
    }

    err = m.Host().Close()
    if err != nil {
      return err
    }
  }

  return nil
}

func (m *BasicMpi)Add(f string) error {
  defer func() {
    if err := recover(); err != nil {
      m.Raise(err.(error))
    }
  }()

  if !m.Store().Has(f) {
    err := m.Store().Dowload(f)
    if err != nil {
      return err
    }
  }

  proto := protocol.ID(fmt.Sprintf("/%s/%s", f, m.Pid))
  m.Host().Listen(proto, fmt.Sprintf("/%s/%s", f, m.Ipfs_store))
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

    param, err := ParamFromString(str[:len(str) - 1])
    if err != nil {
      return
    }

    comm, err := NewSlaveComm(m.Ctx, m.Host(), stream.(io.ReadWriteCloser), proto, param, m.Path + f, param.N, param.Idx)
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

func (m *BasicMpi)Start(file string, n int, args ...string) error {
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

  comm, err := NewMasterComm(m.Ctx, m.Host(), n, proto, stringId, m.Path + file, args...)

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
