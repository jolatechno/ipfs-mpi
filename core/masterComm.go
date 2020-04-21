package core

import (
  "time"
  "fmt"
  "context"
  "sync"
  "errors"
  "io"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/peer"
)

var (
  MasterCommHeader = "MasterComm"

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
  defer func() {
    wg.Mutex.Unlock()
    recover()
  }()

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
  defer func() {
    wg.Mutex.Unlock()
    recover()
  }()

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
  defer func() {
    wg.Mutex.Unlock()
    recover()
  }()

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
  defer func() {
    wg.Mutex.Unlock()
    recover()
  }()

  return wg.Value[i] >= 2
}

func (wg *safeWaitgroupTwice)Check(i int) bool {
  wg.Mutex.Lock()
  defer func() {
    wg.Mutex.Unlock()
    recover()
  }()

  return !wg.Jumped[i]
}

func (wg *safeWaitgroupTwice)WaitFirst() {
  defer recover()
  wg.WG1.Wait()
}

func (wg *safeWaitgroupTwice)WaitSecond() {
  defer recover()
  wg.WG2.Wait()
}

func NewMasterComm(ctx context.Context, host ExtHost, n int, base protocol.ID, id string, file string, args ...string) (_ MasterComm, err error) {
  inter, err := NewInterface(ctx, file, n, 0, args...)
  if err != nil {
    return nil, err
  }

  addrs := make([]peer.ID, n)
  lastReseted := make([]time.Time, n)

  for i := 0; i < n; i++ {
    if i == 0 {
      addrs[i] = host.ID()
    } else {
      addrs[i], err = host.NewPeer(base)
      if err != nil {
        return nil, err
      }
    }

    lastReseted[i] = time.Now().Add(-1 * (ResetCooldown + time.Second))
  }

  remotes := make([]Remote, n)
  comm := BasicMasterComm {
    Addrs: &addrs,
    Comm: BasicSlaveComm {
      SlaveIds: make([]int, n),
      Ctx: ctx,
      Inter: inter,
      Id: id,
      N: n,
      Idx: 0,
      CommHost: host,
      Base: base,
      Remotes: &remotes,
      Standard: NewStandardInterface(MasterCommHeader),
    },
  }

  defer func() {
    if err := recover(); err != nil {
      comm.Raise(err.(error))
    }
  }()

  wg := NewSafeWaitgroupTwice(n, n - 1)

  for i := 1; i < n; i++ {
    (*comm.Comm.Remotes)[i], err = NewRemote()
    if err != nil {
      return nil, err
    }
  }

  for j := 1; j < n; j++ {
    i := j

    comm.SlaveComm().Remote(i).SetResetHandler(func(i int, slaveId int) {
      comm.Reset(i, slaveId)
    })

    comm.SlaveComm().Remote(i).SetCloseHandler(func() {
      comm.Close()
    })

    comm.SlaveComm().Remote(i).SetErrorHandler(func(err error) {
      comm.Raise(SetNonPanic(err))
      wg.DoneAll(i)
    })

    go func() {
      err := comm.SlaveComm().Connect(i, addrs[i], fmt.Sprint(&Param {
        Init: true,
        Idx: i,
        N: n,
        Id: id,
        SlaveIds: comm.Comm.SlaveIds,
        Addrs: &addrs,
      }))
      if err != nil {
        comm.SlaveComm().Remote(i).Raise(err)
        return
      }

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
      comm.Raise(SetNonPanic(err))
      comm.SlaveComm().Remote(i).Reset(io.ReadWriteCloser(nil))
      comm.Reset(i, -1)
    })

    if wg.Check(i) {
      comm.SlaveComm().Remote(i).SendHandshake()
    } else {
      comm.Reset(i, -1)
    }
  }

  comm.SlaveComm().Interface().SetResetHandler(func(i int) {
    comm.Reset(i, -1)
  })

  comm.SlaveComm().Start()

  return &comm, nil
}

type BasicMasterComm struct {
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
  return c.SlaveComm().Check()
}

func (c *BasicMasterComm)SlaveComm() SlaveComm {
  return &c.Comm
}

func (c *BasicMasterComm)Reset(i int, slaveId int) {
  c.Mutex.Lock()
  defer func() {
    c.Mutex.Unlock()
    if err := recover(); err != nil {
      c.Raise(err.(error))
    }
  }()

  if slaveId != c.Comm.SlaveIds[i] && slaveId != -1 {
    return
  }

  c.SlaveComm().Remote(i).CloseRemote()

  c.Comm.SlaveIds[i]++

  go c.Raise(SetNonPanic(NewHeadedError(errors.New(fmt.Sprintf("reseting %d for the %dth time", i, c.Comm.SlaveIds[i])), MasterCommHeader)))

  for c.Check() {
    addr, err := c.SlaveComm().Host().NewPeer(c.Comm.Base)
    if err != nil {
      panic(err) //will be handlled by the recover
    }

    (*c.Addrs)[i] = addr

    err = c.SlaveComm().Connect(i, addr, fmt.Sprint(&Param {
      Init: false,
      Idx: i,
      N: c.Comm.N,
      Id: c.Comm.Id,
      SlaveIds: c.Comm.SlaveIds,
      Addrs: c.Addrs,
    }))
    if err == nil {
      break
    }

  }
}
