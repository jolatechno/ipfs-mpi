package core

import (
  "io"
  "fmt"
  "errors"
  "context"
  "strings"
  "strconv"
  "sync"
  "time"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/peer"

  "github.com/jolatechno/go-timeout"
)

func ParamFromString(msg string) (Param, error) {
  param := Param{}
  splitted := strings.Split(msg, ",")
  if len(splitted) != 5 {
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

  n, err := strconv.Atoi(splitted[2])
  if err != nil {
    return param, err
  }

  len_addrs := len(splitted[4]) - 1
  if splitted[4][len_addrs] == '\n' {
    splitted[4] = splitted[4][:len_addrs]
  }

  addrs := strings.Split(splitted[4], ";")
  list := make([]peer.ID, len(addrs))

  for i, addr := range addrs {
    if addr != "" {
      list[i], err = peer.IDB58Decode(addr)
      if err != nil {
        return param, err
      }
    }
  }

  param.Idx = idx
  param.N = n
  param.Id = splitted[3]
  param.Addrs = &list

  return param, nil
}

type Param struct {
  Init bool
  Idx int
  N int
  Id string
  Addrs *[]peer.ID
}

func (p *Param)String() string {
  addrs := make([]string, len(*p.Addrs))
  for i, addr := range *p.Addrs {
    if i != p.Idx && (!p.Init || i > p.Idx) {
      addrs[i] = peer.IDB58Encode(addr)
    }
  }

  initInt := 0
  if p.Init {
    initInt = 1
  }

  joinedAddress := strings.Join(addrs, ";")
  return fmt.Sprintf("%d,%d,%d,%s,%s", initInt, p.Idx, p.N, p.Id, joinedAddress)
}

func NewSlaveComm(ctx context.Context, host ExtHost, zeroRw io.ReadWriteCloser, base protocol.ID, param Param, file string, n int, i int) (_ SlaveComm, err error) {
  inter, err := NewInterface(file, n, i)
  if err != nil {
    return nil, err
  }

  remotes := make([]Remote, len(*param.Addrs))
  comm := BasicSlaveComm {
    Ctx: ctx,
    Inter: inter,
    Id: param.Id,
    Idx: param.Idx,
    CommHost: host,
    Base: base,
    Pid: protocol.ID(fmt.Sprintf("%d/%s/%s", param.Idx, param.Id, string(base))),
    Remotes: &remotes,
    Standard: NewStandardInterface(),
  }

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

  go func(){
    for comm.Check() {
      time.Sleep(WaitDuration)
      if !comm.Remote(0).Ping(WaitDuration) {
        comm.Close()
      }
    }
  }()

  for i := 1; i < len(*param.Addrs); i++ {
    (*comm.Remotes)[i], err = NewRemote()
    if err != nil {
      return nil, err
    }

    streamHandler, err := comm.Remote(i).StreamHandler()
    if err != nil {
      return nil, err
    }

    proto := protocol.ID(fmt.Sprintf("%d/%s/%s", i, param.Id, string(base)))
    host.SetStreamHandler(proto, streamHandler)
  }

  if param.Init {
    comm.Remote(0).SendHandshake()
    <- comm.Remote(0).GetHandshake()
  }

  var wg sync.WaitGroup

  if param.Init {
    wg.Add(len(*param.Addrs) - param.Idx - 1)
  } else {
    wg.Add(len(*param.Addrs) - 1)
  }

  for i, addr := range *param.Addrs {
    if i > 0 && (i > param.Idx || !param.Init) && i != param.Idx {
      go func(wp *sync.WaitGroup) {
        comm.Connect(i, addr)
        wp.Done()
      }(&wg)
    }
  }

  if param.Init {
    comm.Remote(0).SendHandshake()
    <- comm.Remote(0).GetHandshake()
  }

  comm.Start()

  return &comm, nil
}

type BasicSlaveComm struct {
  Ctx context.Context
  Inter Interface
  Id string
  Idx int
  CommHost ExtHost
  Base protocol.ID
  Pid protocol.ID
  Remotes *[]Remote
  Standard standardFunctionsCloser
}

func (c *BasicSlaveComm)Start() {

  fmt.Printf("[SlaveComm] Starting %d\n", c.Idx) //--------------------------

  c.Interface().SetErrorHandler(func(err error) {
    c.Raise(err)
  })

  c.Interface().SetCloseHandler(func() {
    c.Close()
  })

  c.Interface().SetMessageHandler(func(to int, content string) {
    c.Remote(to).Send(content)
  })

  c.Interface().SetRequestHandler(func(i int) {

    fmt.Printf("[SlaveComm] %d requesting from %d\n", c.Idx, i) //--------------------------

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

func (c *BasicSlaveComm)Close() error {
  if c.Check() {

    fmt.Printf("[SlaveComm] Closing %d, 0\n", c.Idx) //--------------------------

    c.Standard.Close()

    go c.Interface().Close()

    for i := range *c.Remotes {
      if i != c.Idx {
        proto := protocol.ID(fmt.Sprintf("%d/%s", i, string(c.Pid)))
        c.CommHost.RemoveStreamHandler(proto)

        go func() {
          if c.Idx == 0 {
            c.Remote(i).CloseRemote()
          }
          c.Remote(i).Close()
        }()

      }
    }
  }

  return nil
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
  return (*c.Remotes)[idx]
}

func (c *BasicSlaveComm)Host() ExtHost {
  return c.CommHost
}

func (c *BasicSlaveComm)Connect(i int, addr peer.ID) error {

  //fmt.Printf("[SlaveComm] %d connecting to %d\n", c.Idx, i) //--------------------------

  rwi, err := timeout.MakeTimeout(func() (interface{}, error) {
    stream, err := c.CommHost.NewStream(c.Ctx, addr, c.Pid)
    if err != nil {
      return nil, err
    }

    return stream, nil
  }, WaitDuration)

  if err != nil {
    return err
  }

  rwc, ok := rwi.(io.ReadWriteCloser)
  if !ok {
    return errors.New("couldn't convert interface")
  }

  c.Remote(i).Reset(rwc)

  return nil
}
