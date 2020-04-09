package core

import (
  "bufio"
  "fmt"
  "context"
  "time"

  "github.com/libp2p/go-libp2p/p2p/protocol/ping"
  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/peer"
)

type BasicMasterComm struct {
  Ctx context.Context
  Pinger *ping.PingService
  N int
  Comm BasicSlaveComm
}

func NewMasterComm(ctx context.Context, host ExtHost, n int, base protocol.ID, inter Interface, id string) (MasterComm, error) {
  var err error

  Addrs := make([]peer.ID, n)
  for i, _ := range Addrs {
    Addrs[i], err = host.NewPeer(base)
    if err != nil {
      return nil, err
    }
  }

  comm := BasicMasterComm {
    Ctx:ctx,
    Pinger: ping.NewPingService(host),
    N: n,
    Comm: BasicSlaveComm {
      Ended: false,
      EndChan: make(chan bool),
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

      comm.Connect(i, addr, true)

      streamHandler, err := comm.Comm.Remotes[i].StreamHandler()
      if err != nil {
        comm.Close()
        return &comm, err
      }

      proto := protocol.ID(fmt.Sprintf("/%d/%s", i, string(comm.Comm.Pid)))
      host.SetStreamHandler(proto, streamHandler)
    }
  }

  for i := range comm.Comm.Addrs {
    go func() {
      for comm.Check() {
        time.Sleep(WaitDuratio)
        if !comm.CheckPeer(i) {
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

func (c *BasicMasterComm)Close() error {
  return c.Comm.Close()
}

func (c *BasicMasterComm)CloseChan() chan bool {
  return c.Comm.CloseChan()
}

func (c *BasicMasterComm)Check() bool {
  return !c.Comm.Check()
}

func (c *BasicMasterComm)Send(idx int, msg string) {
  c.Comm.Send(idx, msg)
}

func (c *BasicMasterComm)Get(idx int) string {
  return c.Comm.Get(idx)
}

func (c *BasicMasterComm)CheckPeer(idx int) bool {
  select {
  case res := <- c.Pinger.Ping(c.Ctx, c.Comm.Addrs[idx]):
    if res.Error != nil {
      return false
    }
    return true

  case <- time.After(WaitDuratio):
    return false
  }
}

func (c *BasicMasterComm)Connect(i int, addr peer.ID, init bool) {
  fmt.Println("Connect 0, proto : ", c.Comm.Base, ", peer : ", addr) //--------------------------

  stream, err := c.Comm.Host.NewStream(c.Ctx, addr, c.Comm.Base)
  if err != nil {
    fmt.Println("Connect 0, err : ", err) //--------------------------
    c.Reset(i)
    return
  }

  fmt.Println("Connect 1") //--------------------------

  p := Param {
    Init: init,
    Idx: i,
    N: c.N,
    Id: c.Comm.Id,
    Addrs: c.Comm.Addrs,
  }

  fmt.Printf("%s\n", p.String()) //--------------------------

  rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
  fmt.Fprintf(rw, "%s\n", p.String())

  c.Comm.Remotes[i].Reset(rw)
}

func (c *BasicMasterComm)Reset(i int) {
  addr, err := c.Comm.Host.NewPeer(c.Comm.Base)
  if err != nil {
    panic(err) //should never happend here
  }
  c.Connect(i, addr, false)
}
