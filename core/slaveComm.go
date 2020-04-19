package core

import (
  "io"
  "fmt"
  "errors"
  "context"
  "strings"
  "strconv"
  "sync"
  "bufio"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/peer"
  "github.com/libp2p/go-libp2p-core/network"

  "github.com/jolatechno/go-timeout"
)

var (
  SlaveCommHeader = "SlaveComm"
)

func ParamFromString(msg string) (Param, error) {
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

  idx, err := strconv.Atoi(splitted[1])
  if err != nil {
    return param, err
  }

  slaveId, err := strconv.Atoi(splitted[2])
  if err != nil {
    return param, err
  }

  n, err := strconv.Atoi(splitted[3])
  if err != nil {
    return param, err
  }

  len_addrs := len(splitted[5]) - 1
  if splitted[5][len_addrs] == '\n' {
    splitted[5] = splitted[5][:len_addrs]
  }

  addrs := strings.Split(splitted[5], ";")
  list := make([]peer.ID, n)

  if len(addrs) != n {
    return param, errors.New("lsit length and comm size don't match")
  }

  for i, addr := range addrs {
    if addr != "" {
      list[i], err = peer.IDB58Decode(addr)
      if err != nil {
        return param, err
      }
    }
  }

  param.Idx = idx
  param.SlaveId = slaveId
  param.N = n
  param.Id = splitted[4]
  param.Addrs = &list

  return param, nil
}

type Param struct {
  Init bool
  Idx int
  SlaveId int
  N int
  Id string
  Addrs *[]peer.ID
}

func (p *Param)String() string {
  addrs := make([]string, p.N)

  i := 1
  if p.Init {
    i = p.Idx + 1
  }

  for ;i < p.N; i++ {
    if i == p.Idx {
      continue
    }

    addrs[i] = peer.IDB58Encode((*p.Addrs)[i])
  }

  initInt := 0
  if p.Init {
    initInt = 1
  }

  joinedAddress := strings.Join(addrs, ";")
  return fmt.Sprintf("%d,%d,%d,%d,%s,%s", initInt, p.Idx, p.SlaveId, p.N, p.Id, joinedAddress)
}

func NewSlaveComm(ctx context.Context, host ExtHost, zeroRw io.ReadWriteCloser, base protocol.ID, param Param, file string, n int, i int) (_ SlaveComm, err error) {

  fmt.Println("[SlaveComm] New", param) //--------------------------

  inter, err := NewInterface(ctx, file, n, i)
  if err != nil {
    return nil, err
  }

  remotes := make([]Remote, param.N)
  comm := BasicSlaveComm {
    SlaveIds: make([]int, n),
    SlaveId: param.SlaveId,
    Ctx: ctx,
    Inter: inter,
    Id: param.Id,
    N: param.N,
    Idx: param.Idx,
    CommHost: host,
    Base: base,
    Remotes: &remotes,
    Standard: NewStandardInterface(),
  }

  defer func() {
    if err := recover(); err != nil {
      comm.Raise(err.(error))
    }
  }()

  (*comm.Remotes)[0], err = NewRemote()
  if err != nil {
    return nil, err
  }

  comm.Remote(0).Reset(zeroRw)

  comm.Remote(0).SetErrorHandler(func(err error) {
    comm.Raise(err)
    comm.Close()
  })

  comm.Remote(0).SetCloseHandler(func() {
    comm.Close()
  })

  for j := 1; j < comm.N; j++ {
    i := j

    if i == comm.Idx {
      continue
    }

    (*comm.Remotes)[i], err = NewRemote()
    if err != nil {
      return nil, err
    }

    comm.Remote(i).SetErrorHandler(func(err error) {

      fmt.Printf("[SlaveComm] %d disconnected from %d\n", comm.Idx, i) //--------------------------

      go comm.Raise(SetNonPanic(err))
      comm.Remote(i).Reset(io.ReadWriteCloser(nil))

      comm.RequestReset(i)
    })

    comm.Remote(i).SetCloseHandler(func() {
      comm.Close()
    })

    proto := protocol.ID(fmt.Sprintf("%d/%s/%s", i, param.Id, string(base)))
    host.SetStreamHandler(proto, func(stream network.Stream) {
      comm.Mutex.Lock()
      defer comm.Mutex.Unlock()

      fmt.Println("[Remote] [StreamHandler]") //--------------------------

      str, err := bufio.NewReader(stream).ReadString('\n')
      if err != nil {
        return
      }

      slaveId, err := strconv.Atoi(str[:len(str) - 1])
      if err != nil {
        return
      }

      comm.SlaveIds[i] = slaveId
      comm.Remote(i).Reset(stream)
    })
  }

  if param.Init {
    comm.Remote(0).SendHandshake()
    <- comm.Remote(0).GetHandshake()
  }

  var wg sync.WaitGroup

  j := 1
  if param.Init {
    j = comm.Idx + 1
    wg.Add(param.N - param.Idx - 1)

  } else {
    wg.Add(param.N - 2)
  }

  for ;j < comm.N; j++ {
    i := j

    if i == comm.Idx {
      continue
    }

    go func(wp *sync.WaitGroup) {
      err := comm.Connect(i, (*param.Addrs)[i], fmt.Sprintf("%d\n", param.SlaveId))
      if err != nil {
          go comm.Remote(i).Raise(err)
      }

      wp.Done()
    }(&wg)
  }

  wg.Wait()

  if param.Init {
    comm.Remote(0).SendHandshake()
    <- comm.Remote(0).GetHandshake()
  }

  comm.Interface().SetResetHandler(func(i int) {
    comm.RequestReset(i)
    comm.Remote(i).Reset(io.ReadWriteCloser(nil))
  })

  comm.Start()

  return &comm, nil
}

type BasicSlaveComm struct {
  Mutex sync.Mutex
  SlaveIds []int
  SlaveId int
  Ctx context.Context
  Inter Interface
  Id string
  N int
  Idx int
  CommHost ExtHost
  Base protocol.ID
  Remotes *[]Remote
  Standard standardFunctionsCloser
}

func (c *BasicSlaveComm)Start() {

  fmt.Println("[SlaveComm] Starting", c.Idx) //--------------------------

  defer func() {
    if err := recover(); err != nil {
      c.Raise(err.(error))
    }
  }()

  c.Interface().SetErrorHandler(func(err error) {
    c.Raise(err)
  })

  c.Interface().SetCloseHandler(func() {
    if c.Idx == 0 {
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

func (c *BasicSlaveComm)Interface() Interface {
  return c.Inter
}

func (c *BasicSlaveComm)RequestReset(i int) {

  fmt.Println("[SlaveComm] Reset Requested") //--------------------------

  c.Remote(0).RequestReset(i, c.SlaveIds[i])
}

func (c *BasicSlaveComm)SetErrorHandler(handler func(error)) {
  c.Standard.SetErrorHandler(handler)
}

func (c *BasicSlaveComm)SetCloseHandler(handler func()) {
  c.Standard.SetCloseHandler(handler)
}

func (c *BasicSlaveComm)Raise(err error) {
  hErr := NewHeadedError(err, true, SlaveCommHeader)
  c.Standard.Raise(hErr)
}

func (c *BasicSlaveComm)Check() bool {
  return c.Standard.Check()
}

func (c *BasicSlaveComm)Remote(idx int) Remote {
  return (*c.Remotes)[idx]
}

func (c *BasicSlaveComm)Host() ExtHost {
  return c.CommHost
}

func (c *BasicSlaveComm)Close() error {
  defer func() {
    if err := recover(); err != nil {
      c.Raise(err.(error))
    }
  }()

  if c.Check() {

    fmt.Println("[SlaveComm] Closing ", c.Idx) //--------------------------

    c.Standard.Close()

    go c.Interface().Close()

    for j := 0; j < c.N; j++ {
      i := j

      if i == c.Idx {
        continue
      }

      if i != 0 && c.Idx != 0 {
        proto := protocol.ID(fmt.Sprintf("%d/%s/%s", i, c.Id, string(c.Base)))
        c.CommHost.RemoveStreamHandler(proto)
      }

      go func() {
        if c.Idx == 0 {
          c.Remote(i).CloseRemote()
        }
        c.Remote(i).Close()
      }()

    }
  }

  return nil
}

func (c *BasicSlaveComm)Connect(i int, addr peer.ID, msgs ...string) error {

  if c.Idx == 0 { //--------------------------
    fmt.Printf("[MasterComm] connecting to %d with address: %q\n", i, addr) //--------------------------
  } //--------------------------

  defer func() {
    if err := recover(); err != nil {
      c.Raise(err.(error))
    }
  }()

  pid := c.Base
  if c.Idx != 0 {
    pid = protocol.ID(fmt.Sprintf("%d/%s/%s", c.Idx, c.Id, string(c.Base)))
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

  for _, msg := range msgs {
    writer := bufio.NewWriter(rwc)

    _, err = writer.WriteString(msg)
    if err != nil {
      return err
    }

    err = writer.Flush()
    if err != nil {
      return err
    }
  }

  c.Remote(i).Reset(rwc)
  return nil
}
