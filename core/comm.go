package core

import (
  "bufio"
  "fmt"

  "github.com/libp2p/go-libp2p-core/network"
  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/peer"
)

type Comm struct {
  Id string
  Idx int
  Host host.Host
  Addrs []peer.ID
  Base protocol.ID
  Pid protocol.ID
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
