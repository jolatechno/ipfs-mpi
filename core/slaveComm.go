package core

import (
  "io"
  "fmt"
  "errors"
  "context"
  "strings"
  "strconv"
  "sync"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/peer"
  "github.com/libp2p/go-libp2p-core/network"
  "github.com/libp2p/go-libp2p-core/helpers"
  "github.com/ipfs/go-log"

  "github.com/jolatechno/go-timeout"
)

var (
  SlaveCommHeader = "SlaveComm"
  SlaveLogger = log.Logger(SlaveCommHeader)
)

func ParamFromString(msg string) (_ Param, err error) {
  param := Param{}

  splitted := strings.Split(msg, ",")
  if len(splitted) != 6 {
    return param, errors.New("Param dosen't have the right number fields")
  }

  if splitted[0] == "0" {
    param.Init = false
  } else if splitted[0] == "1" {
    param.Init = true
  } else {
    return param, errors.New("bool header not understood")
  }

  param.Idx, err = strconv.Atoi(splitted[1])
  if err != nil {
    return param, err
  }

  param.N, err = strconv.Atoi(splitted[2])
  if err != nil {
    return param, err
  }

  param.Id = splitted[3]

  slaveIds := strings.Split(splitted[4], ";")
  param.SlaveIds = make([]int, param.N)

  if len(slaveIds) != param.N {
    return param, errors.New("list length and comm size don't match")
  }

  for i, id := range slaveIds {
    param.SlaveIds[i], err = strconv.Atoi(id)
    if err != nil {
      return param, err
    }
  }

  addrs := strings.Split(splitted[5], ";")
  param.Addrs = make([]peer.ID, param.N)

  if len(addrs) != param.N {
    return param, errors.New("list length and comm size don't match")
  }

  for i, addr := range addrs {
    if addr != "" {
      param.Addrs[i], err = peer.IDB58Decode(addr)
      if err != nil {
        return param, err
      }
    }
  }

  return param, nil
}

type Param struct {
  Init bool
  Idx int
  N int
  Id string
  SlaveIds []int
  Addrs []peer.ID

}

func (p *Param)String() string {
  addrs := make([]string, p.N)
  slaveIds := make([]string, p.N)

  for i := 0; i < p.N; i++ {
    slaveIds[i] = fmt.Sprint(p.SlaveIds[i])

    if i == 0 || i == p.Idx || (p.Init && i <= p.Idx){
      continue
    }
    addrs[i] = peer.IDB58Encode(p.Addrs[i])
  }

  initInt := 0
  if p.Init {
    initInt = 1
  }

  joinedAddress := strings.Join(addrs, ";")
  joinedSlaveIds := strings.Join(slaveIds, ";")
  return fmt.Sprintf("%d,%d,%d,%s,%s,%s", initInt, p.Idx, p.N, p.Id, joinedSlaveIds, joinedAddress)
}

func NewSlaveComm(ctx context.Context, host ExtHost, zeroRw io.ReadWriteCloser, base protocol.ID, param Param, inter Interface, remotes []Remote) (_ SlaveComm, err error) { //fmt.Println("[SlaveComm] New", param) //--------------------------
  SlaveLogger.Debug("New slaveComm with param ", param) //--------------------------

  comm := BasicSlaveComm {
    Ctx: ctx,
    Inter: inter,
    Param: param,
    CommHost: host,
    Base: base,
    Remotes: remotes,
  }

  proto := protocol.ID(fmt.Sprintf("%d/%d/%s/%s", comm.Param.Idx, param.SlaveIds[comm.Param.Idx], comm.Param.Id, string(comm.Base)))

  close := func() error {
    SlaveLogger.Debugf("Closing the %dth reset of %d", comm.Param.SlaveIds[comm.Param.Idx], comm.Param.Idx) //--------------------------

    go comm.Interface().Close()

    comm.CommHost.RemoveStreamHandler(proto)

    for j := 0; j < comm.Param.N; j++ {
      i := j

      if i == comm.Param.Idx {
        continue
      }

      comm.Remote(i).SetErrorHandler(nilErrorHandler)
      comm.Remote(i).SetCloseHandler(nilEndHandler)

      go comm.Remote(i).Close()
    }

    return nil
  }

  comm.Standard = NewStandardInterface(close)

  defer func() {
    if err := recover(); err != nil {
      comm.Raise(err.(error))
    }
  }()

  comm.Remote(0).SetErrorHandler(func(err error) {
    RemoteLogger.Error(err) //--------------------------
    comm.Close()
  })

  comm.Remote(0).SetCloseHandler(func() {
    comm.Close()
  })

  comm.Remote(0).Reset(zeroRw, 0)

  for j := 1; j < comm.Param.N; j++ {
    i := j

    if i == comm.Param.Idx {
      continue
    }

    comm.Remote(i).SetErrorHandler(func(err error) {
      RemoteLogger.Warn(err) //--------------------------
      SlaveLogger.Warnf("%d hanged-up on %d", i, comm.Param.Idx) //--------------------------

      comm.Remote(i).Reset(io.ReadWriteCloser(nil), 0)
      comm.RequestReset(i)
    })

    comm.Remote(i).SetCloseHandler(func() {
      comm.Close()
    })
  }

  matcher, err := helpers.MultistreamSemverMatcher(proto)
  if err != nil {
    return nil, err
  }

  match := func(p string) bool {
    splitted := strings.Split(p, "/")
    if len(splitted) < 3 {
      return false
    }

    joined := strings.Join(splitted[2:], "/")
    return matcher(joined)
  }

  handler := func(stream network.Stream) {
    pid := string(stream.Protocol())
    splitted := strings.Split(pid, "/")

    i, err := strconv.Atoi(splitted[0])
    if err != nil {
      stream.Close()
      return
    }

    slaveId, err := strconv.Atoi(splitted[1])
    if err != nil {
      stream.Close()
      return
    }

    comm.Remote(i).Reset(stream, slaveId)
  }

  host.SetStreamHandlerMatch(proto, match, handler)

  if param.Init {
    go comm.Remote(0).SendHandshake()
    comm.Remote(0).WaitHandshake()
  }

  var wg sync.WaitGroup

  j := 1
  if param.Init {
    j = comm.Param.Idx + 1
    wg.Add(param.N - param.Idx - 1)

  } else {
    wg.Add(param.N - 2)
  }

  for ;j < comm.Param.N; j++ {
    i := j

    if i == comm.Param.Idx {
      continue
    }

    go func(wp *sync.WaitGroup) {
      err := comm.Connect(i, param.Addrs[i])
      if err != nil {
          go comm.Remote(i).Raise(err)
      }

      wp.Done()
    }(&wg)
  }

  wg.Wait()

  if param.Init {
    go comm.Remote(0).SendHandshake()
    comm.Remote(0).WaitHandshake()
  }

  comm.Interface().SetResetHandler(func(i int) {
    comm.RequestReset(i)
    comm.Remote(i).Reset(io.ReadWriteCloser(nil), 0)
  })

  comm.Start()

  return &comm, nil
}

type BasicSlaveComm struct {
  SlaveIdMutex sync.Mutex
  RemotesMutex sync.Mutex

  Param Param

  Ctx context.Context
  Inter Interface
  CommHost ExtHost
  Base protocol.ID
  Remotes []Remote
  Standard standardFunctionsCloser
}

func (c *BasicSlaveComm)Start() {
  if c.Param.Idx != 0  { //--------------------------
    SlaveLogger.Debugf("Starting the %dth reset of %d", c.Param.SlaveIds[c.Param.Idx], c.Param.Idx) //--------------------------
  } else { //--------------------------
    MasterLogger.Debug("Starting") //--------------------------
  } //--------------------------

  defer func() {
    if err := recover(); err != nil {
      c.Raise(err.(error))
    }
  }()

  c.Interface().SetErrorHandler(func(err error) {

    fmt.Println(err) //--------------------------

    InterfaceLogger.Error(err) //--------------------------
    c.Raise(err)
  })

  c.Interface().SetCloseHandler(func() {
    if c.Param.Idx == 0 {
      c.Close()
    }
  })

  c.Interface().SetMessageHandler(func(to int, content string) {
    c.Remote(to).Send(content)
  })

  c.Interface().SetRequestHandler(func(i int) {
    c.Interface().Push(c.Remote(i).Get())
  })

  c.Interface().SetErrorHandler(func(err error) {
    c.Raise(err)
  })

  c.Interface().Start()
}

func (c *BasicSlaveComm)Protocol() protocol.ID {
  return c.Base
}

func (c *BasicSlaveComm)Interface() Interface {
  return c.Inter
}

func (c *BasicSlaveComm)RequestReset(i int) {
  c.Remote(0).RequestReset(i, c.Remote(i).SlaveId())
}

func (c *BasicSlaveComm)SetErrorHandler(handler func(error)) {
  c.Standard.SetErrorHandler(handler)
}

func (c *BasicSlaveComm)SetCloseHandler(handler func()) {
  c.Standard.SetCloseHandler(handler)
}

func (c *BasicSlaveComm)Raise(err error) {
  c.Standard.Raise(err)
}

func (c *BasicSlaveComm)Check() bool {
  return c.Standard.Check()
}

func (c *BasicSlaveComm)Remote(idx int) Remote {
  if c.Param.Idx == idx {
    c.Raise(errors.New("Requesting self remote"))
    return Remote(nil)
  }

  c.RemotesMutex.Lock()
  defer c.RemotesMutex.Unlock()
  return c.Remotes[idx]
}

func (c *BasicSlaveComm)Host() ExtHost {
  return c.CommHost
}

func (c *BasicSlaveComm)Close() error {
  return c.Standard.Close()
}

func (c *BasicSlaveComm)Connect(i int, addr peer.ID, msgs ...interface{}) error {
  if c.Param.Idx != 0  { //--------------------------
    SlaveLogger.Debugf("%dth connecting to the %dth reset of %d", c.Param.Idx, c.Param.SlaveIds[i], i) //--------------------------
  } else { //--------------------------
    MasterLogger.Debugf("Connecting to the %dth reset of %d", c.Param.SlaveIds[i], i) //--------------------------
  } //--------------------------

  c.RemotesMutex.Lock()
  defer func() {
    c.RemotesMutex.Unlock()
    if err := recover(); err != nil {
      c.Raise(err.(error))
    }
  }()

  remote := c.Remotes[i]
  slaveId := remote.SlaveId()

  pid := c.Base
  if c.Param.Idx != 0 {
    pid = protocol.ID(fmt.Sprintf("%d/%d/%d/%d/%s/%s", c.Param.Idx, c.Param.SlaveIds[c.Param.Idx], i, slaveId, c.Param.Id, string(c.Base)))
  }

  rwi, err := timeout.MakeTimeout(func() (interface{}, error) {
    stream, err := c.CommHost.NewStream(c.Ctx, addr, pid)
    if err != nil {
      return nil, err
    }

    return stream, nil
  }, StandardTimeout)

  if err != nil {
    return err
  }

  rwc, ok := rwi.(io.ReadWriteCloser)
  if !ok {
    return errors.New("couldn't convert interface")
  }

  remote.Reset(rwc, slaveId, msgs...)
  return nil
}
