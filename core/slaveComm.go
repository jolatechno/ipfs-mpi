package core

import (
  "bufio"
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
    list[i], err = peer.IDB58Decode(addr)
    if err != nil {
      return param, err
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
    addrs[i] = peer.IDB58Encode(addr)
  }

  initInt := 0
  if p.Init {
    initInt = 1
  }

  joinedAddress := strings.Join(addrs, ";")
  return fmt.Sprintf("%d,%d,%d,%s,%s", initInt, p.Idx, p.N, p.Id, joinedAddress)
}

func NewSlaveComm(ctx context.Context, host ExtHost, zeroRw *bufio.ReadWriter, base protocol.ID, inter Interface, param Param) (_ SlaveComm, err error) {
  remotes := make([]Remote, len(*param.Addrs))
  comm := BasicSlaveComm {
    Ctx: ctx,
    Inter: inter,
    Id: param.Id,
    Idx: param.Idx,
    CommHost: host,
    Addrs: param.Addrs,
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

  go func(){
    <- comm.Remote(0).CloseChan()
    if comm.Check() {
      comm.Close()
    }
  }()

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

  for i, addr := range *comm.Addrs {
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
  Addrs *[]peer.ID
  Base protocol.ID
  Pid protocol.ID
  Remotes *[]Remote
  Standard BasicFunctionsCloser
}

func (c *BasicSlaveComm)Start() {
  go func() {
    outChan := c.Interface().Message()
    for c.Check() {
      msg, ok := <- outChan
      if ok {
        go c.Remote(msg.To).Send(msg.Content)
      } else {
        break
      }
    }

    if c.Idx == 0 && c.Check() {
      c.Close()
    }
  }()

  go func(){
    requestChan := c.Interface().Request()
    for c.Check() {
      req, ok := <- requestChan
      if ok {
        go c.Interface().Push(c.Remote(req).Get())
      } else {
        break
      }
    }

    if c.Idx == 0 && c.Check() {
      c.Close()
    }
  }()

  go func(){
    err, ok := <- c.Interface().ErrorChan()
    if c.Check() && ok {
      c.Standard.Push(err)
      c.Close()
    }
  }()

  go func(){
    err, ok := <- c.Host().ErrorChan()
    if c.Check() && ok {
      c.Standard.Push(err)
      c.Close()
    }
  }()

  go func(){
    <- c.Host().CloseChan()
    if c.Check() {
      c.Close()
    }
  }()

  c.Interface().Start()
}

func (c *BasicSlaveComm)Interface() Interface {
  return c.Inter
}

func (c *BasicSlaveComm)Close() error {
  if c.Check() {

    //fmt.Printf("[SlaveComm] Closing %d out of %d\n", c.Idx, len(*c.Addrs)) //--------------------------

    c.Standard.Close()

    err := c.Inter.Close()
    if err != nil {
      return err
    }

    for i := range *c.Remotes {
      if i != c.Idx {
        proto := protocol.ID(fmt.Sprintf("%d/%s", i, string(c.Pid)))
        c.CommHost.RemoveStreamHandler(proto)

        if c.Idx == 0 {
          c.Remote(i).CloseRemote()
        }
        c.Remote(i).Close()
      }
    }
  }

  return nil
}

func (c *BasicSlaveComm)CloseChan() chan bool {
  return c.Standard.CloseChan()
}

func (c *BasicSlaveComm)ErrorChan() chan error {
  return c.Standard.ErrorChan()
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

  //fmt.Printf("[SlaveComm] %d connecting to %d out of %d\n", c.Idx, i, len(*c.Addrs)) //--------------------------

  rwi, err := timeout.MakeTimeout(func() (interface{}, error) {
    stream, err := c.CommHost.NewStream(c.Ctx, addr, c.Pid)
    if err != nil {
      return nil, err
    }

    rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
    return rw, nil
  }, WaitDuration)

  if err != nil {
    return err
  }

  rw, ok := rwi.(*bufio.ReadWriter)
  if !ok {
    return errors.New("couldn't convert interface")
  }

  c.Remote(i).Reset(rw)

  return nil
}
