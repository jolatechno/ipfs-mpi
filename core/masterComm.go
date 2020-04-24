package core

import (
  "time"
  "fmt"
  "context"
  "sync"
  "errors"
  "io"

  "github.com/libp2p/go-libp2p-core/protocol"
  //"github.com/libp2p/go-libp2p-core/peer"
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

func NewMasterSlaveComm(ctx context.Context, host ExtHost, base protocol.ID, param Param, inter Interface) (_ SlaveComm, err error) {
  comm := BasicSlaveComm {
      Ctx: ctx,
      Inter: inter,
      Param: param,
      CommHost: host,
      Base: base,
      Remotes: make([]Remote, param.N),
      Standard: NewStandardInterface(MasterCommHeader),
    }

    for i := 1; i < param.N; i++ {
      comm.Remotes[i], err = NewRemote(0)
      if err != nil {
        return nil, err
      }
    }

    return &comm, nil
}

func NewMasterComm(ctx context.Context, slaveComm SlaveComm, param Param) (_ MasterComm, err error) {
  comm := BasicMasterComm {
    Ctx: ctx,
    Param: param,
    Comm: slaveComm,
  }

  slaveComm.SetCloseHandler(func() {
    fmt.Println("[MasterComm] Closing from slaveComm") //--------------------------

    comm.Close()
  })

  slaveComm.SetErrorHandler(func(err error) {
    comm.Raise(err)
  })

  close := func() error { //fmt.Println("[MasterComm] Closing") //--------------------------
    go comm.SlaveComm().Interface().Close()

    for j := 1; j < param.N; j++ {
      i := j

      go func() {
        defer recover()

        go comm.SlaveComm().Remote(i).CloseRemote()
        comm.SlaveComm().Remote(i).Close()
      }()

    }

    return nil
  }

  comm.Standard = NewStandardInterface(MasterCommHeader, close)

  defer func() {
    if err := recover(); err != nil {
      comm.Raise(err.(error))
    }
  }()

  wg := NewSafeWaitgroupTwice(param.N, param.N - 1)

  for j := 1; j < param.N; j++ {
    i := j

    comm.SlaveComm().Remote(i).SetResetHandler(func(i int, slaveId int) {
      comm.Reset(i, slaveId)
    })

    comm.SlaveComm().Remote(i).SetCloseHandler(func() {
      comm.Close()
    })

    comm.SlaveComm().Remote(i).SetErrorHandler(func(err error) {
      go comm.Raise(SetNonPanic(err))
      wg.DoneAll(i)
    })

    go func() {
      err := comm.SlaveComm().Connect(i, param.Addrs[i], &Param {
        Init: true,
        Idx: i,
        N: param.N,
        Id: param.Id,
        SlaveIds: param.SlaveIds,
        Addrs: param.Addrs,
      })
      if err != nil {
        comm.SlaveComm().Remote(i).Raise(err)
        return
      }

      comm.SlaveComm().Remote(i).WaitHandshake()

      wg.DoneFirst(i)
    }()
  }

  wg.WaitFirst()

  for j := 1; j < param.N; j ++ {
    i := j

    if wg.CheckSecond(i) {
      continue
    }

    go comm.SlaveComm().Remote(i).SendHandshake()

    go func() {
      comm.SlaveComm().Remote(i).WaitHandshake()

      wg.DoneSecond(i)
    }()
  }

  wg.WaitSecond()

  for j := 1; j < param.N; j++ {
    i := j

    comm.SlaveComm().Remote(i).SetErrorHandler(func(err error) {
      go comm.Raise(SetNonPanic(err))
      go comm.Raise(SetNonPanic(NewHeadedError(errors.New(fmt.Sprintf("%d hanged-up", i)), MasterCommHeader)))

      if comm.SlaveComm().Remote(i).Stream() != io.ReadWriteCloser(nil) {
        remote := comm.SlaveComm().Remote(i)
        remote.Reset(io.ReadWriteCloser(nil), remote.SlaveId())
      }

      comm.Reset(i, -1)
    })

    if wg.Check(i) {
      go comm.SlaveComm().Remote(i).SendHandshake()
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
  Param Param
  Ctx context.Context
  Comm SlaveComm
  Standard standardFunctionsCloser
}

func (c *BasicMasterComm)Close() error {
  return c.Standard.Close()
}

func (c *BasicMasterComm)SetErrorHandler(handler func(error)) {
  c.Standard.SetErrorHandler(handler)
}

func (c *BasicMasterComm)SetCloseHandler(handler func()) {
  c.Standard.SetCloseHandler(handler)
}

func (c *BasicMasterComm)Raise(err error) {
  c.Standard.Raise(err)
}

func (c *BasicMasterComm)Check() bool {
  return c.Standard.Check()
}

func (c *BasicMasterComm)SlaveComm() SlaveComm {
  return c.Comm
}

func (c *BasicMasterComm)Reset(i int, slaveId int) {
  if i == 0 {
    c.Raise(errors.New("Requesting self remote"))
  }

  c.Mutex.Lock()
  defer func() {
    c.Mutex.Unlock()
    if err := recover(); err != nil {
      c.Raise(err.(error))
    }
  }()

  if slaveId != c.Param.SlaveIds[i] && slaveId != -1 {
    return
  }

  c.SlaveComm().Remote(i).CloseRemote()

  c.Param.SlaveIds[i]++

  go c.Raise(SetNonPanic(NewHeadedError(errors.New(fmt.Sprintf("reseting %d for the %dth time", i, c.Param.SlaveIds[i])), MasterCommHeader)))

  for c.Check() {
    addr, err := c.SlaveComm().Host().NewPeer(c.SlaveComm().Protocol())
    if err != nil {
      panic(err) //will be handlled by the recover
    }

    c.Param.Addrs[i] = addr

    err = c.SlaveComm().Connect(i, addr, &Param {
      Init: false,
      Idx: i,
      N: c.Param.N,
      Id: c.Param.Id,
      SlaveIds: c.Param.SlaveIds,
      Addrs: c.Param.Addrs,
    })
    if err == nil {
      c.SlaveComm().Remote(i).WaitHandshake()
      go c.SlaveComm().Remote(i).SendHandshake()

      break
    }

  }
}
