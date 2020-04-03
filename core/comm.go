package core

import (
  "bufio"
  "fmt"
  "context"

  "github.com/libp2p/go-libp2p/p2p/protocol/ping"
  "github.com/libp2p/go-libp2p-core/network"
  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/peer"
)

func NewComm(ctx context.Context, addr []string, idx int, pid protocol.ID, host host.Host, init bool) (*Comm, error) {
  comm := Comm{
    Idx: idx,
    Host: host,
    Pid: pid,
    Pinger: ping.NewPingService(host),
    Remotes: make([]Remote, len(addr)),
  }

  for i := range comm.Remotes {
    if i != idx && (i > idx || !init) {
      proto := protocol.ID(fmt.Sprintf("%d/%s", i, string(pid)))
      stream, err := host.NewStream(ctx, peer.ID(addr[i]), proto)
      if err != nil {
        comm.Stop()
        return nil, err
      }

      comm.Remotes[i] = Remote{
        Sent: []string{},
        Stream: bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream)),
        ResetChan: make(chan bool),
      }

      streamHandler, err := comm.Remotes[i].StreamHandler()
      if err != nil {
        comm.Stop()
        return nil, err
      }

      host.SetStreamHandler(proto, streamHandler)
    }
  }

  return &comm, nil
}

type Comm struct {
  Idx int
  Host host.Host
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

func (c *Comm)Reset(idx int, stream *bufio.ReadWriter) {
  c.Remotes[idx].Reset(stream)
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
