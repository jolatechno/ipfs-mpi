package core

import (
  "time"
  "fmt"
  "context"
  "sync"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/peer"
)

var (
  ResetCooldown = 2 * time.Second
)

func NewSafeWaitgroupTwice(n int, m int) *safeWaitgroupTwice {
  swg := safeWaitgroupTwice {
    Value: make([]int, n),
    Jumped: make([]bool, n),
  }

  swg.WG1.Add(m)
  swg.WG2.Add(m)

  return &swg
}

type safeWaitgroupTwice struct {
  Jumped []bool
  Value []int
  Mutex sync.Mutex
  WG1 sync.WaitGroup
  WG2 sync.WaitGroup
}

func (wg *safeWaitgroupTwice)DoneFirst(i int) {
  wg.Mutex.Lock()
  defer wg.Mutex.Unlock()
  if wg.Value[i] < 1 {
    wg.Value[i] = 1
    wg.WG1.Done()
  }
}

func (wg *safeWaitgroupTwice)CheckFist(i int) bool {
  wg.Mutex.Lock()
  defer wg.Mutex.Unlock()
  return wg.Value[i] >= 1
}

func (wg *safeWaitgroupTwice)DoneSecond(i int) {
  wg.Mutex.Lock()
  defer wg.Mutex.Unlock()
  if wg.Value[i] < 2 {
    if wg.Value[i] < 1 {
      wg.WG1.Done()
    }

    wg.Value[i] = 2
    wg.WG2.Done()
  }
}

func (wg *safeWaitgroupTwice)DoneAll(i int) {
  wg.Mutex.Lock()
  defer wg.Mutex.Unlock()
  wg.Jumped[i] = true
  if wg.Value[i] < 2 {
    if wg.Value[i] < 1 {
      wg.WG1.Done()
    }

    wg.Value[i] = 2
    wg.WG2.Done()
  }
}

func (wg *safeWaitgroupTwice)CheckSecond(i int) bool {
  wg.Mutex.Lock()
  defer wg.Mutex.Unlock()
  return wg.Value[i] >= 2
}

func (wg *safeWaitgroupTwice)Check(i int) bool {
  wg.Mutex.Lock()
  defer wg.Mutex.Unlock()
  return !wg.Jumped[i]
}

func (wg *safeWaitgroupTwice)WaitFirst() {
  wg.WG1.Wait()
}

func (wg *safeWaitgroupTwice)WaitSecond() {
  wg.WG2.Wait()
}

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
    LastReseted: make([]time.Time, n),
    Addrs: &Addrs,
    Comm: BasicSlaveComm {
      Ctx: ctx,
      Inter: inter,
      Id: id,
      N: n,
      Idx: 0,
      CommHost: host,
      Base: base,
      Remotes: &remotes,
      Standard: NewStandardInterface(),
    },
  }

  comm.Comm.SetErrorHandler(func(err error) {
    comm.Raise(err)
  })

  comm.Comm.SetCloseHandler(func() {
    comm.Close()
  })

  wg := NewSafeWaitgroupTwice(n, n - 1)

  for j := 1; j < n; j++ {
    i := j

    (*comm.Comm.Remotes)[i], err = NewRemote()
    if err != nil {
      return nil, err
    }

    comm.SlaveComm().Remote(i).SetCloseHandler(func() {
      comm.Close()
    })

    comm.SlaveComm().Remote(i).SetErrorHandler(func(err error) {
      wg.DoneAll(i)
    })

    go func() {
      comm.SlaveComm().Connect(i, Addrs[i], fmt.Sprintf("%s\n", &Param {
        Init: true,
        Idx: i,
        N: n,
        Id: id,
        Addrs: &Addrs,
      }))

      <- comm.SlaveComm().Remote(i).GetHandshake()

      wg.DoneFirst(i)
    }()
  }

  wg.WaitFirst()

  for j := 1; j < n; j ++ {
    i := j

    if wg.CheckSecond(i) {
      continue
    }

    comm.SlaveComm().Remote(i).SendHandshake()

    go func() {
      <- comm.SlaveComm().Remote(i).GetHandshake()

      wg.DoneSecond(i)
    }()
  }

  wg.WaitSecond()

  for j := 1; j < n; j++ {
    i := j

    comm.SlaveComm().Remote(i).SetErrorHandler(func(err error) {
      comm.Reset(i)
    })

    if wg.Check(i) {
      comm.SlaveComm().Remote(i).SendHandshake()
    } else {
      comm.Reset(i)
    }
  }

  comm.SlaveComm().Interface().SetResetHandler(func(i int) {
    comm.Reset(i)
  })

  comm.SlaveComm().Start()

  return &comm, nil
}

type BasicMasterComm struct {
  LastReseted []time.Time
  Mutex sync.Mutex
  Addrs *[]peer.ID
  Ctx context.Context
  Comm BasicSlaveComm
}

func (c *BasicMasterComm)Close() error {
  return c.SlaveComm().Close()
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

func (c *BasicMasterComm)Reset(i int) {
  var err error

  c.Mutex.Lock()

  t := time.Now()
  if t.Sub(c.LastReseted[i]) < ResetCooldown {
    return
  }

  c.LastReseted[i] = t

  fmt.Println("[MasterComm] reseting ", i) //--------------------------

  (*c.Addrs)[i], err = c.SlaveComm().Host().NewPeer(c.Comm.Base)
  if err != nil {
    c.Mutex.Unlock()

    c.Raise(err)
  }

  param := &Param {
    Init: false,
    Idx: i,
    N: c.Comm.N,
    Id: c.Comm.Id,
    Addrs: c.Addrs,
  }

  c.Mutex.Unlock()

  c.SlaveComm().Connect(i, (*c.Addrs)[i], fmt.Sprintf("%s\n", param))
}
