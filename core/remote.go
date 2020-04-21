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

func send(stream io.ReadWriteCloser, str string) error {
  defer recover()

  if stream == io.ReadWriteCloser(nil) {
    return nil
  }

  if str != PingHeader && str != HandShakeHeader && str != CloseHeader { //--------------------------
    fmt.Printf("[Remote] Sent %q\n", str) //--------------------------
  } //--------------------------

  writer := bufio.NewWriter(stream)

  _, err := writer.WriteString(str + "\n")
  if err != nil {
    return err
  }

  err = writer.Flush()
  return err
}

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

  c.Ended = true
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

  c.Ended = true
}

func NewRemote() (Remote, error) {
  return &BasicRemote {
    ResetHandler: &nilRemoteResetHandler,
    PingInterval: StandardPingInterval,
    PingTimeout: StandardTimeout,
    ReadChan: NewChannelString(),
    HandshakeChan: NewChannelBool(),
    Sent: &[]string{},
    Rw: nil,
    Received: 0,
    Standard: NewStandardInterface(RemoteHeader),
  }, nil
}

type BasicRemote struct {
  WriteMutex sync.Mutex
  StreamMutex sync.Mutex

  ResetHandler *func(int, int)
  PingInterval time.Duration
  PingTimeout time.Duration
  ReadChan *safeChannelString
  HandshakeChan *safeChannelBool
  Sent *[]string
  Rw io.ReadWriteCloser
  Received int
  HandshakeMessage int
  ReceivedHandshakeMessage int
  Standard standardFunctionsCloser
}

func (r *BasicRemote)raiseCheck(err error, stream io.ReadWriteCloser) bool {
  if r.Stream() == stream {
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
  go r.raiseCheck(send(stream, fmt.Sprintf("%s,%d,%d", ResetHeader, i, slaveId)), stream)
}

func (r *BasicRemote)CloseRemote() {
  stream := r.Stream()
  go r.raiseCheck(send(stream, CloseHeader), stream)
}

func (r *BasicRemote)Send(msg string) {
  r.WriteMutex.Lock()
  *r.Sent = append(*r.Sent, msg)
  r.WriteMutex.Unlock()

  go r.Raise(send(r.Stream(), fmt.Sprintf("%s,%s", MessageHeader, msg)))
}

func (r *BasicRemote)SendHandshake() {
  go r.Raise(send(r.Stream(), HandShakeHeader))
}

func (r *BasicRemote)Get() string {
  return <- r.ReadChan.Chan
}

func (r *BasicRemote)GetHandshake() chan bool {
  return r.HandshakeChan.Chan
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

func (r *BasicRemote)Close() error {
  defer recover()

  if r.Check() {
    r.Standard.Close()

    if stream := r.Stream(); stream != io.ReadWriteCloser(nil) {
      stream.Close()
      r.Rw = io.ReadWriteCloser(nil)
    }
  }
  return nil
}

func (r *BasicRemote)Reset(stream io.ReadWriteCloser, msgs ...string) {
  if !r.Check() {
    return
  }

  r.StreamMutex.Lock()

  r.Rw = stream
  if stream == io.ReadWriteCloser(nil) {
    return
  }

  defer func() {
    r.StreamMutex.Unlock()
    if err := recover(); err != nil {
      r.raiseCheck(err.(error), stream)
    }
  }()

  offset := r.Received
  pingChan := NewChannelBool()

  go func() {
    r.WriteMutex.Lock()
    defer func() {
      r.WriteMutex.Unlock()
      if err := recover(); err != nil {
        r.raiseCheck(err.(error), stream)
      }
    }()

    for _, msg := range msgs {
      if err := send(stream, msg); err != nil {
        panic(err)
      }
    }

    for _, msg := range *r.Sent {
      if err := send(stream, fmt.Sprintf("%s,%s", MessageHeader, msg)); err != nil {
        panic(err)
      }
    }
  }()

  go func() {
    defer func() {
      if err := recover(); err != nil {
        r.raiseCheck(err.(error), stream)
      }
    }()

    for r.Check() && r.Stream() == stream {
      time.Sleep(r.PingInterval)
      go r.raiseCheck(send(stream, PingHeader), stream)
    }
  }()

  go func() {
    defer func() {
      if err := recover(); err != nil {
        r.Raise(err.(error))
      }
    }()

    scanner := bufio.NewScanner(stream)

    for r.Check() &&  r.Stream() == stream && scanner.Scan() {
      stream.(network.Stream).SetReadDeadline(time.Now().Add(r.PingTimeout))

      splitted := strings.Split(scanner.Text(), ",")

      str := strings.Join(splitted, ",")//--------------------------
      if str != PingHeader && str != HandShakeHeader && str != CloseHeader { //--------------------------
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

        msg := strings.Join(splitted[1:], ",")

        if offset > 0 {
          offset --
          continue
        }

        r.Received++
        go r.ReadChan.Send(msg)
      }
    }

    pingChan.Close()

    r.raiseCheck(scanner.Err(), stream)

    if !r.Check() {
      r.ReadChan.Close()
      r.HandshakeChan.Close()
    }

  }()
}
