package core

import (
  "bufio"
  "io"
  "fmt"
  "strconv"
  "errors"
  "strings"
  "sync"
  "time"
  "context"

  "github.com/ipfs/go-log"

  "github.com/jolatechno/go-timeout"
)

var (
  RemoteHeader = "Remote"
  RemoteLogger = log.Logger(RemoteHeader)

  //ResetHandShakeHeader = "ResetHandShake"
  PingRespHeader = "PingResp"
  HandShakeHeader = "HandShake"
  MessageHeader = "Msg"
  CloseHeader = "Close"
  PingHeader = "Ping"
  ResetHeader = "Reset"

  StandardTimeout = 2 * time.Second //Will be increase later
  StandardPingInterval = 200 * time.Millisecond //Will be increase later

  NilStreamError = errors.New("nil stream")
  ErrorInterval = 4 * time.Second

  nilRemoteResetHandler = func(int, int) {}
)

func NewChannelBool() *safeChannelBool {
  return &safeChannelBool {
    Chan: make(chan bool),
  }
}

type safeChannelBool struct {
  Chan chan bool
  Mutex sync.Mutex
  Ended bool
}

func (c *safeChannelBool)Send(t bool) {
  c.Mutex.Lock()
  defer c.Mutex.Unlock()

  if !c.Ended {
    go func() {
      c.Chan <- t
    }()
  }
}

func (c *safeChannelBool)Close() {
  c.Mutex.Lock()
  defer func() {
    c.Mutex.Unlock()
    recover()
  }()

  if !c.Ended {
    c.Ended = true
    for len(c.Chan) > 0 {
      <- c.Chan
    }
  }
}

func NewChannelString() *safeChannelString {
  return &safeChannelString {
    Chan: make(chan string),
  }
}

type safeChannelString struct {
  Chan chan string
  Mutex sync.Mutex
  Ended bool
}

func (c *safeChannelString)Send(str string) {
  c.Mutex.Lock()
  defer func() {
    c.Mutex.Unlock()
    recover()
  }()

  if !c.Ended {
    go func() {
      c.Chan <- str
    }()
  }
}

func (c *safeChannelString)Close() {
  c.Mutex.Lock()
  defer func() {
    c.Mutex.Unlock()
    recover()
  }()

  if !c.Ended {
    c.Ended = true
    for len(c.Chan) > 0 {
      <- c.Chan
    }
  }
}

func NewRemote(ctx context.Context, slaveId int) (Remote, error) {
  remote :=  &BasicRemote {
    Ctx: ctx,
    ResetHandler: &nilRemoteResetHandler,
    PingInterval: StandardPingInterval,
    PingTimeout: StandardTimeout,
    ReadChan: NewChannelString(),
    HandshakeChan: NewChannelBool(),
    SendChan: NewChannelString(),
    Sent: []interface{}{},
    Id: slaveId,
    Received: 0,
  }

  close := func() error {
    remote.Mutex.Lock()
    defer remote.Mutex.Unlock()

    if remote.Rw != io.ReadWriteCloser(nil) {
      remote.Rw.Close()
      remote.Rw = io.ReadWriteCloser(nil)
    }

    return nil
  }

  remote.Standard = NewStandardInterface(close)

  return remote, nil
}

type BasicRemote struct {
  Mutex sync.Mutex

  ReadChan *safeChannelString
  HandshakeChan *safeChannelBool
  SendChan *safeChannelString

  Ctx context.Context
  Id int
  ResetHandler *func(int, int)
  PingInterval time.Duration
  PingTimeout time.Duration
  Sent []interface{}
  Rw io.ReadWriteCloser
  Received int
  Standard standardFunctionsCloser
}

func (r *BasicRemote)send(stream io.ReadWriteCloser, slaveId int, msg ...interface{}) {
  if msg[0] != PingHeader && msg[0] != PingRespHeader { //--------------------------
    RemoteLogger.Debugf("Sent %q", fmt.Sprint(msg...)) //--------------------------
  } //--------------------------

  if stream == io.ReadWriteCloser(nil) {
    return
  }

  _, err := fmt.Fprintln(stream, msg...)
  go r.raiseCheck(err, stream, slaveId)
}

func (r *BasicRemote)sendf(stream io.ReadWriteCloser, slaveId int, formatString string, msg ...interface{}) {
  if msg[0] != PingHeader && msg[0] != PingRespHeader { //--------------------------
    RemoteLogger.Debugf("Sent %q", fmt.Sprintf(formatString, msg...)) //--------------------------
  } //--------------------------

  if stream == io.ReadWriteCloser(nil) {
    return
  }

  _, err := fmt.Fprintf(stream, formatString + "\n", msg...)
  go r.raiseCheck(err, stream, slaveId)
}

func (r *BasicRemote)check(stream io.ReadWriteCloser, slaveId int) bool {
  r.Mutex.Lock()
  defer r.Mutex.Unlock()
  return stream == r.Rw && slaveId == r.Id && r.Check()
}

func (r *BasicRemote)raiseCheck(err error, stream io.ReadWriteCloser, slaveId int) bool {
  if err == nil {
    return true
  }

  if r.check(stream, slaveId) {
    r.Raise(err)
  }

  return false
}

func (r *BasicRemote)SetResetHandler(handler func(int, int)) {
  r.ResetHandler = &handler
}

func (r *BasicRemote)SetPingInterval(interval time.Duration) {
  r.PingInterval = interval
}

func (r *BasicRemote)SetPingTimeout(timeoutDuration time.Duration) {
  r.PingTimeout = timeoutDuration
}

func (r *BasicRemote)RequestReset(i int, slaveId int) {
  stream := r.Stream()
  go r.sendf(stream, slaveId, "%s,%d,%d", ResetHeader, i, slaveId)
}

func (r *BasicRemote)CloseRemote() {
  stream := r.Stream()
  slaveId := r.SlaveId()

  r.send(stream, slaveId, CloseHeader)
}

func (r *BasicRemote)Send(msg string) {
  str := MessageHeader + "," + msg

  r.Mutex.Lock()
  defer r.Mutex.Unlock()
  r.Sent = append(r.Sent, str)

  r.send(r.Rw, r.Id, str)
}

func (r *BasicRemote)SendHandshake() {
  stream := r.Stream()
  slaveId := r.SlaveId()

  r.send(stream, slaveId, HandShakeHeader)
}

func (r *BasicRemote)Get() string {
  return <- r.ReadChan.Chan
}

func (r *BasicRemote)WaitHandshake() {
  <- r.HandshakeChan.Chan
}

func (r *BasicRemote)SetErrorHandler(handler func(error)) {
  r.Standard.SetErrorHandler(handler)
}

func (r *BasicRemote)SetCloseHandler(handler func()) {
  r.Standard.SetCloseHandler(handler)
}

func (r *BasicRemote)Raise(err error) {
  r.Standard.Raise(err)
}

func (r *BasicRemote)Check() bool {
  return r.Standard.Check()
}

func (r *BasicRemote)Stream() io.ReadWriteCloser {
  r.Mutex.Lock()
  defer r.Mutex.Unlock()
  return r.Rw
}

func (r *BasicRemote)SlaveId() int {
  r.Mutex.Lock()
  defer r.Mutex.Unlock()
  return r.Id
}

func (r *BasicRemote)Close() error {
  return r.Standard.Close()
}

func (r *BasicRemote)Reset(stream io.ReadWriteCloser, slaveId int, msgs ...interface{}) {
  if !r.Check() || (slaveId < r.SlaveId() && stream != io.ReadWriteCloser(nil)) {
    return
  }

  r.Mutex.Lock()
  defer func() {
    r.Mutex.Unlock()
    if err := recover(); err != nil {
      r.raiseCheck(err.(error), stream, slaveId)
    }
  }()

  if r.Rw != io.ReadWriteCloser(nil) {
    r.Rw.Close()
  }

  r.Rw = stream
  if stream == io.ReadWriteCloser(nil) {
    return
  }

  RemoteLogger.Debugf("Sending %q on reset  ", append(msgs, r.Sent...)) //--------------------------

  received := ResetReader(r.Received, append(msgs, r.Sent...), func(msg interface{}) {
    r.send(stream, slaveId, msg)
  }, func(msg string) {
    r.Received++
    r.ReadChan.Send(msg)
  })

  pingChan := NewChannelBool()

  go func() {
    defer func() {
      if err := recover(); err != nil {
        r.raiseCheck(err.(error), stream, slaveId)
      }
    }()

    for r.check(stream, slaveId) {
      time.Sleep(r.PingInterval)

      go r.send(stream, slaveId, PingHeader)

      err := timeout.MakeSimpleTimeout(func() error {
        <- pingChan.Chan
        return nil
      }, r.PingTimeout)

      if err != nil {
        r.raiseCheck(err, stream, slaveId)
      }
    }
  }()

  go func() {
    defer func() {
      if err := recover(); err != nil {
        r.raiseCheck(err.(error), stream, slaveId)
      }
    }()

    scanner := bufio.NewScanner(stream)

    for r.check(stream, slaveId) && scanner.Scan() {
      splitted := strings.Split(scanner.Text(), ",")

      str := strings.Join(splitted, ",") //--------------------------
      if str != PingHeader && str != PingRespHeader { //--------------------------
        RemoteLogger.Debugf("Received %q", str) //--------------------------
      } //--------------------------

      switch splitted[0] {
      default:
        r.Raise(HeaderNotUnderstood)

      case PingHeader:
        go r.send(stream, slaveId, PingRespHeader)

      case PingRespHeader:
        go pingChan.Send(true)

      case ResetHeader:
        if len(splitted) < 2 {
          r.Raise(NotEnoughFields)
          continue
        }

        idx, err := strconv.Atoi(splitted[1])
        if err != nil {
          r.Raise(err)
          continue
        }

        slaveId, err := strconv.Atoi(splitted[2])
        if err != nil {
          r.Raise(err)
          continue
        }

        go (*r.ResetHandler)(idx, slaveId)

      case HandShakeHeader:
        go r.HandshakeChan.Send(true)

      case CloseHeader:
        r.Close()
        break

      case MessageHeader:
        if len(splitted) < 2 {
          r.Raise(NotEnoughFields)
          continue
        }

        received(strings.Join(splitted[1:], ","))

      }
    }

    pingChan.Close()

    r.raiseCheck(scanner.Err(), stream, slaveId)

    if !r.Check() {
      r.ReadChan.Close()
      r.HandshakeChan.Close()
      r.SendChan.Close()
    }

  }()
}
