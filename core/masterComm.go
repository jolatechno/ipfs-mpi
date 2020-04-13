package core

import (
  "fmt"
  "context"
  "time"
  "sync"

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

func NewMasterComm(ctx context.Context, host ExtHost, n int, base protocol.ID, inter Interface, id string) (_ MasterComm, err error) {
  Addrs := make([]peer.ID, n)
  for i, _ := range Addrs {
    if i == 0 {
      Addrs[i] = host.ID()
    } else {
      Addrs[i], err = host.NewPeer(base)
      if err != nil {
        return nil, err
      }
    }
  }

  comm := BasicMasterComm {
    Pinger: ping.NewPingService(host),
    N: n,
    Comm: BasicSlaveComm {
      Ctx: ctx,
      Ended: false,
      EndChan: make(chan bool),
      Error: make(chan error),
      Inter: inter,
      Id: id,
      Idx: 0,
      CommHost: host,
      Addrs: Addrs,
      Base: protocol.ID(fmt.Sprintf("%s/%s", id, string(base))),
      Pid: base,
      Remotes: make([]Remote, n),
    },
  }

  var wg sync.WaitGroup

  wg.Add(n - 1)

  for i, addr := range comm.Comm.Addrs {
    if i > 0 {
      comm.Comm.Remotes[i], err = NewRemote(2)
      if err != nil {
        return nil, err
      }

      go func(wp *sync.WaitGroup) {
        comm.Connect(i, addr, true)

        go func() {
          for comm.Check() {
            time.Sleep(WaitDuration)
            if !comm.CheckPeer(i) {
              comm.Reset(i)
            }
          }
        }()

        go func() {
          for comm.Check() {
            <- comm.SlaveComm().Remote(i).ErrorChan()
            comm.Reset(i)
          }
        }()

        str := comm.SlaveComm().Remote(i).Get()
        if str != "Done\n" {
          comm.Reset(i)
        }

        wp.Done()
      }(&wg)
    }
  }

  wg.Wait()

  fmt.Printf("[MasterComm] Done") //--------------------------

  var wg2 sync.WaitGroup

  wg2.Add(n - 1)

  for i := 1; i < n; i++ {
    go func(wp *sync.WaitGroup) {
      str := comm.SlaveComm().Remote(i).Get()
      if str != "Connected\n" {
        comm.Reset(i)
      }

      wp.Done()
    }(&wg2)
  }

  wg2.Wait()

  fmt.Printf("[MasterComm] Started") //--------------------------

  comm.Comm.start()

  return &comm, nil
}

func (c *BasicMasterComm)Close() error {
  return c.SlaveComm().Close()
}

func (c *BasicMasterComm)CloseChan() chan bool {
  return c.SlaveComm().CloseChan()
}

func (c *BasicMasterComm)ErrorChan() chan error {
  return c.SlaveComm().ErrorChan()
}

func (c *BasicMasterComm)Check() bool {
  return !c.SlaveComm().Check()
}

func (c *BasicMasterComm)SlaveComm() SlaveComm {
  return &c.Comm
}

func (c *BasicMasterComm)CheckPeer(idx int) bool {
  if c.Comm.Addrs[idx] == c.SlaveComm().Host().ID() {
    return true
  }

  select {
  case res := <- c.Pinger.Ping(c.Comm.Ctx, c.Comm.Addrs[idx]):
    if res.Error != nil {
      return false
    }
    return true

  case <- time.After(WaitDuration):
    return false
  }
}

func (c *BasicMasterComm)Connect(i int, addr peer.ID, init bool) {

  fmt.Println("[MasterComm] Connect 0") //--------------------------

  err := c.SlaveComm().Connect(i, addr)

  fmt.Println("[MasterComm] Connect 1") //--------------------------

  if err != nil {
    c.Reset(i)
  } else {
    p := Param {
      Init: init,
      Idx: i,
      N: c.N,
      Id: c.Comm.Id,
      Addrs: c.Comm.Addrs,
    }

    fmt.Println("[MasterComm] Connect 2 : ", p.String()) //--------------------------

    rw := c.SlaveComm().Remote(i).Stream()

    fmt.Fprintf(rw, "%s\n", p.String())
    rw.Flush()
  }
}

func (c *BasicMasterComm)Reset(i int) {

  fmt.Println("[MasterComm] reseting ", i) //--------------------------

  addr, err := c.SlaveComm().Host().NewPeer(c.Comm.Base)
  if err != nil {
    c.SlaveComm().ErrorChan() <- err
    c.Close()
  }

  c.Connect(i, addr, false)
}
