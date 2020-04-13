package core

import (
  "bufio"
  "fmt"
  "errors"
  "context"
  "strings"
  "strconv"
  "sync"

  "github.com/libp2p/go-libp2p-core/network"
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
  param.Addrs = make([]peer.ID, len(addrs))

  for i, addr := range addrs {
    param.Addrs[i], err = peer.IDB58Decode(addr)
    if err != nil {
      return param, err
    }
  }

  param.Idx = idx
  param.N = n
  param.Id = splitted[3]

  return param, nil
}

type Param struct {
  Init bool
  Idx int
  N int
  Id string
  Addrs []peer.ID
}

func (p *Param)String() string {
  addrs := make([]string, len(p.Addrs))
  for i, addr := range p.Addrs {
    addrs[i] = peer.IDB58Encode(addr)
  }

  initInt := 0
  if p.Init {
    initInt = 1
  }

  joinedAddress := strings.Join(addrs, ";")
  return fmt.Sprintf("%d,%d,%d,%s,%s", initInt, p.Idx, p.N, p.Id, joinedAddress)
}

func NewSlaveComm(ctx context.Context, host ExtHost, zeroRw *bufio.ReadWriter, base protocol.ID, inter Interface, param Param) (SlaveComm, error) {

  fmt.Println("[SlaveComm] New") //--------------------------

  comm := BasicSlaveComm {
    Ctx: ctx,
    Ended: false,
    EndChan: make(chan bool),
    Error: make(chan error),
    Inter: inter,
    Id: param.Id,
    Idx: param.Idx,
    Host: host,
    Addrs: param.Addrs,
    Base: base,
    Pid: protocol.ID(fmt.Sprintf("%d/%s/%s", param.Idx, param.Id, string(base))),
    Remotes: make([]Remote, len(param.Addrs)),
  }

  comm.Remotes[0] = Remote {
    Sent: []string{},
    Stream: zeroRw,
    ResetChan: make(chan bool),
  }

  for i := 1; i < len(param.Addrs); i++ {
    comm.Remotes[i] = Remote {
      Sent: []string{},
      Stream: nil,
      ResetChan: make(chan bool),
    }

    streamHandler, err := comm.Remotes[i].StreamHandler()
    if err != nil {
      return nil, err
    }

    proto := protocol.ID(fmt.Sprintf("%d/%s/%s", i, param.Id, string(base)))
    host.SetStreamHandler(proto, streamHandler)
  }

  fmt.Println("[SlaveComm] New, Done") //--------------------------

  fmt.Fprint(zeroRw, "Done\n")
  zeroRw.Flush()

  str, err := zeroRw.ReadString('\n')
  if err != nil {
    return &comm, err
  }
  if str != "Connect\n"{
    return &comm, errors.New("Responce no understood")
  }

  var wg sync.WaitGroup

  if param.Init {
    wg.Add(len(param.Addrs) - param.Idx - 1)
  } else {
    wg.Add(len(param.Addrs) - 1)
  }

  for i, addr := range comm.Addrs {
    if i > 0 && (i > param.Idx || !param.Init) {
      go func(wp *sync.WaitGroup) {
        comm.Connect(i, addr)
        wp.Done()
      }(&wg)
    }
  }

  fmt.Println("[SlaveComm] New, Connected") //--------------------------

  fmt.Fprint(zeroRw, "Connected\n")
  zeroRw.Flush()

  comm.start()

  return &comm, nil
}

func (c *BasicSlaveComm)start() {

  fmt.Printf("[SlaveComm] starting %d out of %d\n", c.Idx, len(c.Addrs)) //--------------------------

  go func(){
    outChan := c.Inter.Message()
    for c.Check() && c.Inter.Check() {
      msg := <- outChan

      go c.Send(msg.To, msg.Content)
    }
  }()

  go func(){
    requestChan := c.Inter.Request()
    for c.Check() {
      req := <- requestChan

      go func() {
        err := c.Inter.Push(c.Get(req))
        if err != nil {

        }
      }()
    }
  }()

  go func(){
    err := <- c.Inter.ErrorChan()
    if c.Check() {
      c.Error <- err
      c.Close()
    }
  }()

  go func(){
    <- c.Inter.CloseChan()
    if c.Check() {
      c.Close()
    }
  }()

  go func(){
    err := <- c.Host.ErrorChan()
    if c.Check() {
      c.Error <- err
      c.Close()
    }
  }()

  go func(){
    <- c.Host.CloseChan()
    if c.Check() {
      c.Close()
    }
  }()
}

type BasicSlaveComm struct {
  Ctx context.Context
  Ended bool
  EndChan chan bool
  Error chan error
  Inter Interface
  Id string
  Idx int
  Host ExtHost
  Addrs []peer.ID
  Base protocol.ID
  Pid protocol.ID
  Remotes []Remote
}

func (c *BasicSlaveComm)Interface() Interface {
  return c.Inter
}

func (c *BasicSlaveComm)Close() error {

  fmt.Printf("[SlaveComm] Closing %d out of %d\n", c.Idx, len(c.Addrs)) //--------------------------

  c.EndChan <- true
  c.Ended = true
  err := c.Inter.Close()
  if err != nil {
    return err
  }

  for i := range c.Remotes {
    if i != c.Idx {
      proto := protocol.ID(fmt.Sprintf("%d/%s", i, string(c.Pid)))
      c.Host.RemoveStreamHandler(proto)
    }
  }
  return nil
}

func (c *BasicSlaveComm)CloseChan() chan bool {
  return c.EndChan
}

func (c *BasicSlaveComm)ErrorChan() chan error {
  return c.Error
}

func (c *BasicSlaveComm)Check() bool {
  return !c.Ended
}

func (c *BasicSlaveComm)Send(idx int, msg string) {
  c.Remotes[idx].Send(msg)
}

func (c *BasicSlaveComm)Get(idx int) string {
  return c.Remotes[idx].Get()
}

func (c *BasicSlaveComm)Connect(i int, addr peer.ID) error {
  rwi, err := timeout.MakeTimeout(func() (interface{}, error) {
    stream, err := c.Host.NewStream(c.Ctx, addr, c.Pid)
    if err != nil {
      return nil, err
    }

    rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
    return rw, nil
  }, WaitDuration)

  if err != nil {
    return err
  }

  rw := rwi.(*bufio.ReadWriter)
  c.Remotes[i].Reset(rw)

  return nil
}

type Remote struct {
  Sent []string
  Stream *bufio.ReadWriter
  Offset int
  Received int
  ResetChan chan bool
}

func (r *Remote)Send(msg string) {

  fmt.Printf("[Remote] Sending %q\n", msg) //--------------------------

  r.Sent = append(r.Sent, msg)

  fmt.Fprint(r.Stream, msg)
  r.Stream.Flush()
}

func (r *Remote)Get() string {

  fmt.Println("[Remote] Requesting") //--------------------------

  readChan := make(chan string)
  go func() {
    for r.Offset > 0 {
      _, err := r.Stream.ReadString('\n')
      if err == nil {
        r.Offset --
      }
    }
    str, err := r.Stream.ReadString('\n')
    if err == nil {
      readChan <- str
    }

    close(readChan)
  }()

  select {
  case msg := <- readChan:

    fmt.Printf("[Remote] Requesting %q\n", msg) //--------------------------

    return msg

  case <- r.ResetChan:
    return r.Get()
  }
}

func (r *Remote)Reset(stream *bufio.ReadWriter) {

  fmt.Println("[Remote] reset 0") //--------------------------

  r.Stream = stream
  r.Offset = r.Received
  for _, msg := range r.Sent {
    fmt.Fprint(r.Stream, msg)
    r.Stream.Flush()
  }

  fmt.Println("[Remote] reset 1") //--------------------------

  r.ResetChan <- true
}

func (r *Remote)StreamHandler() (network.StreamHandler, error) {
  return func(stream network.Stream) {
    r.Reset(bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream)))
  }, nil
}
