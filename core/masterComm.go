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
      Inter: inter,
      Id: id,
      Idx: 0,
      CommHost: host,
      Addrs: Addrs,
      Base: protocol.ID(fmt.Sprintf("%s/%s", id, string(base))),
      Pid: base,
      Remotes: make([]Remote, n),
      Standard: NewStandardInterface(),
    },
  }

  var wg sync.WaitGroup

  wg.Add(n - 1)

  for j, addr := range comm.Comm.Addrs {
    if j > 0 {
      comm.Comm.Remotes[j], err = NewRemote(2)
      if err != nil {
        return nil, err
      }

      go func(wp *sync.WaitGroup, i int) {
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
            err := <- comm.SlaveComm().Remote(i).ErrorChan()

            fmt.Println("[MasterComm] ", i, " error : ", err) //--------------------------

            comm.Reset(i)
          }
        }()

        for {
          str := comm.SlaveComm().Remote(i).Get()
          if str != "Done\n" {
            comm.Reset(i)
          } else {
            break
          }
        }

        wp.Done()
      }(&wg, j)
    }
  }

  wg.Wait()

  fmt.Println("[MasterComm] Done") //--------------------------

  var wg2 sync.WaitGroup

  wg2.Add(n - 1)

  for j := 1; j < n; j++ {
    go func(wp *sync.WaitGroup, i int) {
      str := comm.SlaveComm().Remote(i).Get()
      if str != "Connected\n" {
        comm.Reset(i)
      }

      wp.Done()
    }(&wg2, j)
  }

  wg2.Wait()

  fmt.Printf("[MasterComm] Started") //--------------------------

  comm.Comm.start()

  return &comm, nil
}

func (c *BasicMasterComm)Close() error {

  fmt.Println("[MasterComm] closing ") //--------------------------

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

  addr, err := c.SlaveComm().Host().NewPeer(c.Comm.Pid)
  if err != nil {

    fmt.Println("[MasterComm] NewPeer err : ", err) //--------------------------

    c.Comm.Standard.Push(err)
    c.Close()
  }

  c.Connect(i, addr, false)
}
