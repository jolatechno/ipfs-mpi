package core

import (
  "bufio"
  "fmt"
  "errors"
  "context"
  "strings"
  "strconv"
  "time"

  "github.com/libp2p/go-libp2p/p2p/protocol/ping"
  "github.com/libp2p/go-libp2p-core/network"
  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/peer"
)

type Param struct {
  Init bool
  Idx int
  Id string
  Addrs []string
}

func ParamFromString(msg string) (Param, error) {
  param := Param{}
  splitted := strings.Split(msg, ",")
  if len(splitted) != 4 {
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

  param.Idx = idx
  param.Id = splitted[2]
  param.Addrs = strings.Split(splitted[3], ";")

  return param, err
}

func AddrsToString(addrs []peer.ID) []string {
  list := make([]string, len(addrs))
  for i, addr := range addrs {
    list[i] = string(addr)
  }

  return list
}

func NewMasterComm(ctx context.Context, host host.Host, n int, base protocol.ID, id string, newPeer func() peer.ID) (Comm, error) {
  Addrs := make([]peer.ID, n)
  for i, _ := range Addrs {
    Addrs[i] = newPeer()
  }

  comm := Comm{
    Id: id,
    Idx: 0,
    Host: host,
    Addrs: Addrs,
    Base: base,
    Pid: protocol.ID(fmt.Sprintf("%s/%s", id, string(base))),
    Pinger: ping.NewPingService(host),
    Remotes: make([]Remote, n),
  }

  for i, addr := range comm.Addrs {
    if i > 0 {
      comm.Remotes[i] = Remote{
        Sent: []string{},
        Stream: nil,
        ResetChan: make(chan bool),
      }

      comm.Connect(ctx, i, addr)


      streamHandler, err := comm.Remotes[i].StreamHandler()
      if err != nil {

        return comm, err
      }

      host.SetStreamHandler(protocol.ID(fmt.Sprintf("%d/%s", i, string(comm.Pid))), streamHandler)
    }
  }

  go func() {
    for {
      for i, addr := range comm.Addrs {
        select {
        case <- comm.Pinger.Ping(ctx, addr):
          continue
        case <- time.After(time.Second):
          comm.Reset(ctx, i)
          continue
        }
      }
    }
  }()

  return comm, nil
}

func (c *Comm)Connect(ctx context.Context, i int, addr peer.ID) {
  stream, err := c.Host.NewStream(ctx, addr, c.Base)
  if err != nil {
    c.Reset(ctx, i)
    return
  }

  rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
  fmt.Fprintf(rw, "1,%d,%s,%s\n", i, c.Id, strings.Join(AddrsToString(c.Addrs), ";"))

  c.Remotes[i].Reset(rw)
}

func (c *Comm)Reset(ctx context.Context, i int) {
  addr := peer.ID("") //generate a random Peer.ID
  c.Connect(ctx, i, addr)
}

func NewSlaveComm(ctx context.Context, host host.Host, base protocol.ID, param Param) (Comm, error) {
  Addrs := make([]peer.ID, len(param.Addrs))
  for i, addr := range param.Addrs {
    Addrs[i] = peer.ID(addr)
  }

  comm := Comm{
    Id: param.Id,
    Idx: param.Idx,
    Host: host,
    Addrs: Addrs,
    Pid: protocol.ID(fmt.Sprintf("%s/%s", param.Id, string(base))),
    Pinger: ping.NewPingService(host),
    Remotes: make([]Remote, len(param.Addrs)),
  }

  for i, addr := range comm.Addrs {
    if i != param.Idx && (i > param.Idx || !param.Init) {
      proto := protocol.ID(fmt.Sprintf("%d/%s", i, string(comm.Pid)))

      stream, err := host.NewStream(ctx, addr, proto)
      if err != nil {
        comm.Stop()
        return comm, err
      }

      rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

      comm.Remotes[i] = Remote{
        Sent: []string{},
        Stream: rw,
        ResetChan: make(chan bool),
      }

      streamHandler, err := comm.Remotes[i].StreamHandler()
      if err != nil {
        comm.Stop()
        return comm, err
      }

      host.SetStreamHandler(proto, streamHandler)
    }
  }

  return comm, nil
}

type Comm struct {
  Id string
  Idx int
  Host host.Host
  Addrs []peer.ID
  Base protocol.ID
  Pid protocol.ID
  Pinger *ping.PingService
  Remotes []Remote
}

func (c *Comm)Stop() {
  for i := range c.Remotes {
    if i != c.Idx {
      proto := protocol.ID(fmt.Sprintf("%d/%s", i, string(c.Pid)))
      c.Host.RemoveStreamHandler(proto)
    }
  }
}

func (c *Comm)Send(idx int, msg string) {
  c.Remotes[idx].Send(msg)
}

func (c *Comm)Get(idx int) string {
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
