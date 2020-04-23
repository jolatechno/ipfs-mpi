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

  "github.com/libp2p/go-libp2p-core/network"
)

var (
  RemoteHeader = "Remote"

  //ResetHandShakeHeader = "ResetHandShake"
  HandShakeHeader = "HandShake"
  MessageHeader = "Msg"
  CloseHeader = "Close"
  PingHeader = "Ping"
  ResetHeader = "Reset"

  StandardTimeout = 2 * time.Second //Will be increase later
  StandardPingInterval = 500 * time.Millisecond //Will be increase later

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

func NewRemote(slaveId int) (Remote, error) {
  remote :=  &BasicRemote {
    ResetHandler: &nilRemoteResetHandler,
    PingInterval: StandardPingInterval,
    PingTimeout: StandardTimeout,
    ReadChan: NewChannelString(),
    HandshakeChan: NewChannelBool(),
    SendChan: NewChannelString(),
    Sent: &[]string{},
    Id: slaveId,
    Received: 0,
  }

  close := func() error {
    if stream := remote.Stream(); stream != io.ReadWriteCloser(nil) {
      stream.Close()
      remote.Rw = io.ReadWriteCloser(nil)
    }

    return nil
  }

  remote.Standard = NewStandardInterface(RemoteHeader, close)

  return remote, nil
}

type BasicRemote struct {
  WriteMutex sync.Mutex
  StreamMutex sync.Mutex

  ReadChan *safeChannelString
  HandshakeChan *safeChannelBool
  SendChan *safeChannelString

  Id int
  ResetHandler *func(int, int)
  PingInterval time.Duration
  PingTimeout time.Duration
  Sent *[]string
  Rw io.ReadWriteCloser
  Received int
  HandshakeMessage int
  ReceivedHandshakeMessage int
  Standard standardFunctionsCloser
}

func (r *BasicRemote)check(stream io.ReadWriteCloser, slaveId int) bool {
  r.StreamMutex.Lock()
  defer r.StreamMutex.Unlock()
  return stream == r.Rw && slaveId == r.Id && r.Check()
}

func (r *BasicRemote)raiseCheck(err error, stream io.ReadWriteCloser, slaveId int) bool {
  if r.check(stream, slaveId) {
    r.Raise(err)
  }
  return err == nil
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
  if stream == io.ReadWriteCloser(nil) {
    return
  }

  go r.SendChan.Send(fmt.Sprintf("%s,%d,%d", ResetHeader, i, slaveId))
}

func (r *BasicRemote)CloseRemote() {
  go r.SendChan.Send(CloseHeader)
}

func (r *BasicRemote)Send(msg string) {
  r.WriteMutex.Lock()
  defer r.WriteMutex.Unlock()
  *r.Sent = append(*r.Sent, msg)

  go r.SendChan.Send(MessageHeader + "," + msg)
}

func (r *BasicRemote)SendHandshake() {
  go r.SendChan.Send(HandShakeHeader)
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
  r.StreamMutex.Lock()
  defer r.StreamMutex.Unlock()
  return r.Rw
}

func (r *BasicRemote)SlaveId() int {
  r.StreamMutex.Lock()
  defer r.StreamMutex.Unlock()
  return r.Id
}

func (r *BasicRemote)Close() error {
  return r.Standard.Close()
}

func (r *BasicRemote)Reset(stream io.ReadWriteCloser, slaveId int, msgs ...interface{}) {
  if !r.Check() || (slaveId < r.SlaveId() && stream != io.ReadWriteCloser(nil)) {
    return
  }

  r.StreamMutex.Lock()
  r.WriteMutex.Lock()
  defer func() {
    r.StreamMutex.Unlock()
    r.WriteMutex.Unlock()
    if err := recover(); err != nil {
      r.raiseCheck(err.(error), stream, slaveId)
    }
  }()

  r.Rw = stream
  if stream == io.ReadWriteCloser(nil) {
    return
  }

  fmt.Println("[Remote] Reset") //--------------------------

  r.SendChan.Close()

  sendChan := NewChannelString()
  r.SendChan = sendChan

  go func() {
    for r.check(stream, slaveId) {
      msg, ok := <- sendChan.Chan
      if !ok {
        return
      }

      if msg != PingHeader && msg != HandShakeHeader /*&& msg != PingRespHeader*/ && msg != CloseHeader { //--------------------------
        fmt.Printf("[Remote] Sent %q\n", msg) //--------------------------
      } //--------------------------

      if _, err := fmt.Fprintln(stream, msg); err != nil {
        r.raiseCheck(err, stream, slaveId)
      }
    }
  }()

  for _, msg := range msgs {
    go sendChan.Send(fmt.Sprint(msg))
  }

  received := ResetReader(r.Received, *r.Sent, func(msg string) {
    go sendChan.Send(MessageHeader + "," + msg)
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
      go sendChan.Send(PingHeader)
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
      stream.(network.Stream).SetReadDeadline(time.Now().Add(r.PingTimeout))

      splitted := strings.Split(scanner.Text(), ",")

      str := strings.Join(splitted, ",") //--------------------------
      if str != PingHeader && str != HandShakeHeader /*&& str != PingRespHeader*/ && str != CloseHeader { //--------------------------
        fmt.Printf("[Remote] Received %q\n", str) //--------------------------
      } //--------------------------

      switch splitted[0] {
      default:
        r.Raise(HeaderNotUnderstood)

      case PingHeader:
        continue

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
