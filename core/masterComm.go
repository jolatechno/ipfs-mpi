package core

import (
  "bufio"
  "fmt"
  "context"
  "strings"
  "time"

  "github.com/libp2p/go-libp2p/p2p/protocol/ping"
  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/peer"
)

type BasicMasterComm struct {
  Ctx context.Context
  Pinger *ping.PingService
  Comm BasicSlaveComm
}

func NewMasterComm(ctx context.Context, host ExtHost, n int, base protocol.ID, inter Interface, id string) (MasterComm, error) {
  Addrs := make([]peer.ID, n)
  for i, _ := range Addrs {
    Addrs[i] = host.NewPeer(base)
  }

  comm := BasicMasterComm{
    Ctx:ctx,
    Pinger: ping.NewPingService(host),
    Comm: BasicSlaveComm{
      Ended: false,
      Inter: inter,
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

      comm.Connect(i, addr)


      streamHandler, err := comm.Comm.Remotes[i].StreamHandler()
      if err != nil {
        comm.Close()
        return &comm, err
      }

      host.SetStreamHandler(protocol.ID(fmt.Sprintf("%d/%s", i, string(comm.Comm.Pid))), streamHandler)
    }
  }

  for i := range comm.Comm.Addrs {
    go func() {
      for !comm.Comm.Ended {
        time.Sleep(scanDuration)
        if !comm.Check(i) {
          comm.Reset(i)
        }
      }
    }()
  }

  comm.Comm.start()

  return &comm, nil
}

func (c *BasicMasterComm)Interface() Interface {
  return c.Comm.Interface()
}

func (c *BasicMasterComm)Close() {
  c.Comm.Close()
}

func (c *BasicMasterComm)Send(idx int, msg string) {
  c.Comm.Send(idx, msg)
}

func (c *BasicMasterComm)Get(idx int) string {
  return c.Comm.Get(idx)
}

func (c *BasicMasterComm)Check(idx int) bool {
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

func (c *BasicMasterComm)Connect(i int, addr peer.ID) {
  stream, err := c.Comm.Host.NewStream(c.Ctx, addr, c.Comm.Base)
  if err != nil {
    c.Reset(i)
    return
  }

  rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
  fmt.Fprintf(rw, "1,%d,%s,%s\n", i, c.Comm.Id, strings.Join(AddrsToString(c.Comm.Addrs), ";"))

  c.Comm.Remotes[i].Reset(rw)
}

func (c *BasicMasterComm)Reset(i int) {
  addr := c.Comm.Host.NewPeer(c.Comm.Base)
  c.Connect(i, addr)
}

func AddrsToString(addrs []peer.ID) []string {
  list := make([]string, len(addrs))
  for i, addr := range addrs {
    list[i] = string(addr)
  }

  return list
}
