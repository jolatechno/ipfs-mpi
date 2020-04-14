package core

import (
  "fmt"
  "context"
  "time"
  "sync"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/peer"
)

type BasicMasterComm struct {
  Ctx context.Context
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
            if !comm.SlaveComm().Remote(i).Ping(WaitDuration) {
              comm.Reset(i)
            }
          }
        }()

        for {
          str := comm.SlaveComm().Remote(i).GetHandshake()
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

  var wg2 sync.WaitGroup

  wg2.Add(n - 1)

  for j := 1; j < n; j++ {
    comm.SlaveComm().Remote(j).Send("Connect\n")

    go func(wp *sync.WaitGroup, i int) {
      str := comm.SlaveComm().Remote(i).GetHandshake()
      if str != "Connected\n" {
        comm.Reset(i)
      }

      wp.Done()
    }(&wg2, j)
  }

  wg2.Wait()

  comm.Comm.start()

  return &comm, nil
}

func (c *BasicMasterComm)Close() error {
  if c.Check() {
    for i := range c.Comm.Remotes {
      c.SlaveComm().Remote(i).CloseRemote()
    }
    c.SlaveComm().Close()
  }

  return nil
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

func (c *BasicMasterComm)Connect(i int, addr peer.ID, init bool) {
  err := c.SlaveComm().Connect(i, addr)

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

    rw := c.SlaveComm().Remote(i).Stream()

    fmt.Fprintf(rw, "%s\n", p.String())
    rw.Flush()
  }
}

func (c *BasicMasterComm)Reset(i int) {

  fmt.Println("[MasterComm] reseting ", i) //--------------------------

  addr, err := c.SlaveComm().Host().NewPeer(c.Comm.Pid)
  if err != nil {
    c.Comm.Standard.Push(err)
    c.Close()
  }

  c.Connect(i, addr, false)
}
