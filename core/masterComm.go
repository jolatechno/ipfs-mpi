package core

import (
  "fmt"
  "context"
  "time"
  "sync"
  "bufio"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/peer"
)

func NewMasterComm(ctx context.Context, host ExtHost, n int, base protocol.ID, id string, file string, args ...string) (_ MasterComm, err error) {
  inter, err := NewInterface(file, n, 0, args...)
  if err != nil {
    return nil, err
  }

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

  remotes := make([]Remote, n)
  comm := BasicMasterComm {
    Addrs: &Addrs,
    N: n,
    Comm: BasicSlaveComm {
      Ctx: ctx,
      Inter: inter,
      Id: id,
      Idx: 0,
      CommHost: host,
      Base: protocol.ID(fmt.Sprintf("%s/%s", id, string(base))),
      Pid: base,
      Remotes: &remotes,
      Standard: NewStandardInterface(),
    },
  }

  inter.SetErrorHandler(func(err error) {
    comm.Raise(err)
  })

  inter.SetCloseHandler(func() {
    comm.Close()
  })

  comm.Comm.SetErrorHandler(func(err error) {
    comm.Raise(err)
  })

  comm.Comm.SetCloseHandler(func() {
    if comm.Check() {
      comm.Close()
    }
  })

  var wg sync.WaitGroup

  wg.Add(n - 1)

  state := 0
  reseted := make([]bool, n)

  for j, addr := range *comm.Addrs {
    if j > 0 {
      (*comm.Comm.Remotes)[j], err = NewRemote()
      if err != nil {
        return nil, err
      }

      go func(wp *sync.WaitGroup, i int) {
        comm.Connect(i, addr, true)

        go func() {
          for comm.Check() && state == 0 {
            if !comm.SlaveComm().Remote(i).Ping(WaitDuration) {
              reseted[i] = true
              comm.Reset(i)

              wp.Done()
              return
            }
          }
        }()

        <- comm.SlaveComm().Remote(i).GetHandshake()

        if !reseted[i] {
          wp.Done()
        }
      }(&wg, j)
    }
  }

  wg.Wait()

  fmt.Println("[MasterComm] Handshake 0 ") //--------------------------

  state = 1
  N := 0
  for _, s := range reseted {
    if !s {
      N++
    }
  }

  var wg2 sync.WaitGroup

  wg2.Add(N - 1)

  for j := 1; j < n; j++ {
    if !reseted[j] {
      comm.SlaveComm().Remote(j).SendHandshake()

      go func(wp *sync.WaitGroup, i int) {

        go func() {
          for comm.Check() && state == 1 {
            if !comm.SlaveComm().Remote(i).Ping(WaitDuration) {
              reseted[i] = true
              comm.Reset(i)

              wp.Done()
              return
            }
          }
        }()

        <- comm.SlaveComm().Remote(i).GetHandshake()

        if !reseted[i] {
          wp.Done()
        }
      }(&wg2, j)
    }
  }

  wg2.Wait()

  fmt.Println("[MasterComm] Handshake 1 ") //--------------------------

  state = 2

  for j := 1; j < n; j++ {
    if !reseted[j] {
      comm.SlaveComm().Remote(j).SendHandshake()
    }

    go func(i int) {
      for comm.Check() {
        time.Sleep(WaitDuration)
        if !comm.SlaveComm().Remote(i).Ping(WaitDuration) {
          comm.Reset(i)
        }
      }
    }(j)
  }

  comm.SlaveComm().Start()

  return &comm, nil
}

type BasicMasterComm struct {
  Addrs *[]peer.ID
  Ctx context.Context
  N int
  Comm BasicSlaveComm
}

func (c *BasicMasterComm)Close() error {
  if c.SlaveComm().Check() {
    c.SlaveComm().Close()
  }

  return nil
}

func (c *BasicMasterComm)SetErrorHandler(handler func(error)) {
  c.SlaveComm().SetErrorHandler(handler)
}

func (c *BasicMasterComm)SetCloseHandler(handler func()) {
  c.SlaveComm().SetCloseHandler(handler)
}

func (c *BasicMasterComm)Raise(err error) {
  c.SlaveComm().Raise(err)
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
      Addrs: c.Addrs,
    }

    writer := bufio.NewWriter(c.SlaveComm().Remote(i).Stream())

    fmt.Fprintf(writer, "%s\n", p.String())
    writer.Flush()
  }
}

func (c *BasicMasterComm)Reset(i int) {

  fmt.Println("[MasterComm] reseting ", i) //--------------------------

  addr, err := c.SlaveComm().Host().NewPeer(c.Comm.Pid)
  if err != nil {
    c.Raise(err)
  }

  (*c.Addrs)[i] = addr
  c.Connect(i, addr, false)
}
