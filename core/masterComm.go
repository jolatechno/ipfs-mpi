package core

import (
  "bufio"
  "fmt"
  "context"
  "strings"
  "time"

  "github.com/libp2p/go-libp2p/p2p/protocol/ping"
  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/peer"
)

type BasicMasterComm struct {
  Ctx context.Context
  Pinger *ping.PingService
  Ended bool
  Comm BasicSlaveComm
}

func NewMasterComm(ctx context.Context, host host.Host, n int, base protocol.ID, id string) (MasterComm, error) {
  Addrs := make([]peer.ID, n)
  for i, _ := range Addrs {
    Addrs[i] = newPeer(base)
  }

  comm := BasicMasterComm{
    Ctx:ctx,
    Pinger: ping.NewPingService(host),
    Ended: false,
    Comm: BasicSlaveComm{
      Id: id,
      Idx: 0,
      Host: host,
      Addrs: Addrs,
      Base: base,
      Pid: protocol.ID(fmt.Sprintf("%s/%s", id, string(base))),
      Remotes: make([]Remote, n),
    },
  }

  for i, addr := range comm.Comm.Addrs {
    if i > 0 {
      comm.Comm.Remotes[i] = Remote{
        Sent: []string{},
        Stream: nil,
        ResetChan: make(chan bool),
      }

      comm.Connect(ctx, i, addr)


      streamHandler, err := comm.Comm.Remotes[i].StreamHandler()
      if err != nil {
        comm.Stop()
        return &comm, err
      }

      host.SetStreamHandler(protocol.ID(fmt.Sprintf("%d/%s", i, string(comm.Comm.Pid))), streamHandler)
    }
  }

  go func() {
    for {
      for i := range comm.Comm.Addrs {
        if comm.Ended {
          return
        }
        if !comm.Present(ctx, i) {
          comm.Reset(ctx, i)
        }
      }
    }
  }()

  return &comm, nil
}

func (c *BasicMasterComm)Stop() {
  c.Ended = true
  c.Comm.Stop()
}

func (c *BasicMasterComm)Present(idx int) bool {
  select {
  case res := <- c.Pinger.Ping(c.Ctx, c.Comm.Addrs[idx]):
    if res.Error != nil {
      return false
    }
    return true

  case <- time.After(time.Second):
    return false
  }
}

func (c *BasicMasterComm)Send(idx int, msg string) {
  c.Comm.Send(idx, msg)
}

func (c *BasicMasterComm)Get(idx int) string {
  return c.Comm.Get(idx)
}

func (c *BasicMasterComm)Connect(i int, addr peer.ID) {
  stream, err := c.Comm.Host.NewStream(c.Ctx, addr, c.Comm.Base)
  if err != nil {
    c.Reset(ctx, i)
    return
  }

  rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
  fmt.Fprintf(rw, "1,%d,%s,%s\n", i, c.Comm.Id, strings.Join(AddrsToString(c.Comm.Addrs), ";"))

  c.Comm.Remotes[i].Reset(rw)
}

func (c *BasicMasterComm)Reset(i int) {
  addr := newPeer(c.Comm.Base)
  c.Connect(c.Ctx, i, addr)
}

func AddrsToString(addrs []peer.ID) []string {
  list := make([]string, len(addrs))
  for i, addr := range addrs {
    list[i] = string(addr)
  }

  return list
}
