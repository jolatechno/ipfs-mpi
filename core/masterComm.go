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

type MasterComm struct {
  Pinger *ping.PingService
  Comm Comm
}

func NewMasterComm(ctx context.Context, host host.Host, n int, base protocol.ID, id string) (MasterComm, error) {
  Addrs := make([]peer.ID, n)
  for i, _ := range Addrs {
    Addrs[i] = newPeer(base)
  }

  comm := MasterComm{
    Pinger: ping.NewPingService(host),
    Comm: Comm{
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

        return comm, err
      }

      host.SetStreamHandler(protocol.ID(fmt.Sprintf("%d/%s", i, string(comm.Comm.Pid))), streamHandler)
    }
  }

  go func() {
    for {
      for i, addr := range comm.Comm.Addrs {
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

func (c *MasterComm)Connect(ctx context.Context, i int, addr peer.ID) {
  stream, err := c.Comm.Host.NewStream(ctx, addr, c.Comm.Base)
  if err != nil {
    c.Reset(ctx, i)
    return
  }

  rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
  fmt.Fprintf(rw, "1,%d,%s,%s\n", i, c.Comm.Id, strings.Join(AddrsToString(c.Comm.Addrs), ";"))

  c.Comm.Remotes[i].Reset(rw)
}

func (c *MasterComm)Reset(ctx context.Context, i int) {
  addr := newPeer(c.Comm.Base)
  c.Connect(ctx, i, addr)
}

func AddrsToString(addrs []peer.ID) []string {
  list := make([]string, len(addrs))
  for i, addr := range addrs {
    list[i] = string(addr)
  }

  return list
}
