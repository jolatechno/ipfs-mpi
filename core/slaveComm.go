package core

import (
  "bufio"
  "fmt"
  "errors"
  "context"
  "strings"
  "strconv"
  "time"

  "github.com/libp2p/go-libp2p-core/network"
  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/peer"
  "github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

type Param struct {
  Init bool
  Idx int
  N int
  Id string
  Addrs []string
}

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

  param.Idx = idx
  param.N = n
  param.Id = splitted[3]
  param.Addrs = strings.Split(splitted[4], ";")

  return param, err
}

func NewSlaveComm(ctx context.Context, host ExtHost, zeroRw *bufio.ReadWriter, base protocol.ID, inter Interface, param Param) (SlaveComm, error) {
  Addrs := make([]peer.ID, len(param.Addrs))
  for i, addr := range param.Addrs {
    Addrs[i] = peer.ID(addr)
  }

  comm := BasicSlaveComm {
    Ended: false,
    EndChan: make(chan bool),
    Inter: inter,
    Id: param.Id,
    Idx: param.Idx,
    Host: host,
    Addrs: Addrs,
    Pid: protocol.ID(fmt.Sprintf("%s/%s", param.Id, string(base))),
    Remotes: make([]Remote, len(param.Addrs)),
  }

  for i, addr := range comm.Addrs {
    if i != param.Idx {
      proto := protocol.ID(fmt.Sprintf("%d/%s", i, string(comm.Pid)))

      if i == 0 {
        comm.Remotes[i] = Remote {
          Sent: []string{},
          Stream: zeroRw,
          ResetChan: make(chan bool),
        }

      } else if i > param.Idx || !param.Init {
        stream, err := host.NewStream(ctx, addr, proto)
        if err != nil {
          comm.Close()
          return nil, err
        }

        rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

        comm.Remotes[i] = Remote {
          Sent: []string{},
          Stream: rw,
          ResetChan: make(chan bool),
        }
      }

      streamHandler, err := comm.Remotes[i].StreamHandler()
      if err != nil {
        comm.Close()
        return nil, err
      }

      host.SetStreamHandler(proto, streamHandler)
    }
  }

  pinger := ping.NewPingService(host)

  go func() {
    for comm.Check() {
      time.Sleep(ScanDuration)
      select {
      case <- pinger.Ping(ctx, comm.Addrs[0]):
        continue

      case <- time.After(ScanDuration):
        comm.Close()
        return
      }
    }
  }()

  comm.start()

  return &comm, nil
}

func (c *BasicSlaveComm)start() {
  go func(){
    outChan := c.Inter.Message()
    for c.Check() {
      msg := <- outChan
      c.Send(msg.To, msg.Content)
    }
  }()

  go func(){
    requestChan := c.Inter.Request()
    for c.Check() {
      req := <- requestChan
      c.Inter.Push(c.Get(req))
    }
  }()

  go func(){
    <- c.Inter.CloseChan()
    if c.Check() {
      c.Close()
    }
  }()
}

type BasicSlaveComm struct {
  Ended bool
  EndChan chan bool
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

func (c *BasicSlaveComm)Check() bool {
  return !c.Ended
}

func (c *BasicSlaveComm)Send(idx int, msg string) {
  c.Remotes[idx].Send(msg)
}

func (c *BasicSlaveComm)Get(idx int) string {
  return c.Remotes[idx].Get()
}

type Remote struct {
  Sent []string
  Stream *bufio.ReadWriter
  Offset int
  Received int
  ResetChan chan bool
}

func (r *Remote)Send(msg string) {
  r.Sent = append(r.Sent, msg)
  fmt.Fprint(r.Stream, msg)
}

func (r *Remote)Get() string {
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
    return msg

  case <- r.ResetChan:
    return r.Get()
  }
}

func (r *Remote)Reset(stream *bufio.ReadWriter) {
  r.Stream = stream
  r.Offset = r.Received
  for _, msg := range r.Sent {
    fmt.Fprint(r.Stream, msg)
  }
  r.ResetChan <- true
}

func (r *Remote)StreamHandler() (network.StreamHandler, error) {
  return func(stream network.Stream) {
    r.Reset(bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream)))
  }, nil
}
