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

  "github.com/jolatechno/go-timeout"
)

var (
  RemoteHeader = "Remote"

  HandShakeHeader = "HandShake"
  MessageHeader = "Msg"
  CloseHeader = "Close"
  PingHeader = "Ping"
  PingRespHeader = "PingResp"
  ResetHeader = "Reset"

  StandardTimeout = 2 * time.Second
  StandardPingInterval = 2 * time.Second

  NilStreamError = errors.New("nil stream")
  ErrorInterval = 4 * time.Second
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
    c.Chan <- t
  }
}

func (c *safeChannelBool)Close() {
  c.Mutex.Lock()
  defer c.Mutex.Unlock()
  if !c.Ended {
    c.Ended = true

    for len(c.Chan) > 0 {
      <- c.Chan
    }

    close(c.Chan)
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
  defer recover()

  c.Mutex.Lock()
  defer c.Mutex.Unlock()
  if !c.Ended {
    c.Chan <- str
  }
}

func (c *safeChannelString)Close() {
  defer recover()

  c.Mutex.Lock()
  defer c.Mutex.Unlock()
  if !c.Ended {
    c.Ended = true

    for len(c.Chan) > 0 {
      <- c.Chan
    }

    close(c.Chan)
  }
}

func NewRemote() (Remote, error) {
  return &BasicRemote {
    ResetHandler: &nilResetHandler,
    PingInterval: StandardPingInterval,
    PingTimeout: StandardTimeout,
    ReadChan: NewChannelString(),
    HandshakeChan: NewChannelBool(),
    Sent: &[]string{},
    Rw: nil,
    Received: 0,
    Standard: NewStandardInterface(),
  }, nil
}

type BasicRemote struct {
  WriteMutex sync.Mutex
  StreamMutex sync.Mutex

  ResetHandler *func(int)
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

func (r *BasicRemote)send(str string, blocking bool, referenceStream ...io.ReadWriteCloser) {
  defer func() {
    if err := recover(); err != nil {
      r.Raise(err.(error))
    }
  }()

  if str != PingHeader && str != PingRespHeader { //--------------------------
    fmt.Printf("[Remote] Sending %q\n", str) //--------------------------
  } //--------------------------

  if stream := r.Stream(); stream != io.ReadWriteCloser(nil) {
    if len(referenceStream) == 1 && referenceStream[0] != stream {
      return
    }

    writer := bufio.NewWriter(stream)

    _, err := writer.WriteString(str + "\n")
    if err != nil {
      if stream == r.Stream() {
        r.Raise(err)
      }

      return
    }

    flush := func() {
      err := writer.Flush()
      if err != nil {
        if stream == r.Stream() {
          r.Raise(err)
        }

        return
      }
    }

    if blocking {
      flush()
    } else {
      go flush()
    }

  }
}

func (r *BasicRemote)SetResetHandler(handler func(int)) {
  r.ResetHandler = &handler
}

func (r *BasicRemote)SetPingInterval(interval time.Duration) {
  r.PingInterval = interval
}

func (r *BasicRemote)SetPingTimeout(timeoutDuration time.Duration) {
  r.PingTimeout = timeoutDuration
}

func (r *BasicRemote)RequestReset(i int, slaveId int) {
  r.send(fmt.Sprintf("%s,%d,%d", ResetHeader, i, slaveId), false)
}

func (r *BasicRemote)CloseRemote() {
  r.send(CloseHeader, true)
}

func (r *BasicRemote)Send(msg string) {
  r.WriteMutex.Lock()
  *r.Sent = append(*r.Sent, msg)
  r.WriteMutex.Unlock()

  r.send(fmt.Sprintf("%s,%s", MessageHeader, msg), false)
}

func (r *BasicRemote)SendHandshake() {
  r.send(HandShakeHeader, false)
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
  hErr := NewHeadedError(err, true, RemoteHeader)
  r.Standard.Raise(hErr)
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

func (r *BasicRemote)Reset(stream io.ReadWriteCloser) {
  if !r.Check() {
    return
  }

  r.StreamMutex.Lock()
  defer func() {
    r.StreamMutex.Unlock()
    if err := recover(); err != nil {
      r.Raise(err.(error))
    }
  }()

  r.Rw = stream
  if stream == io.ReadWriteCloser(nil) {
    r.Raise(NilStreamError)
    return
  }

  offset := r.Received
  pingChan := make(chan bool)

  go func() {
    defer recover()

    r.WriteMutex.Lock()
    defer r.WriteMutex.Unlock()

    for _, msg := range *r.Sent {
      r.send(fmt.Sprintf("%s,%s", MessageHeader, msg), true)
    }
  }()

  go func() {
    defer func() {
      if err := recover(); err != nil {
        r.Raise(err.(error))
      }
    }()

    for r.Check() && r.Stream() == stream {
      time.Sleep(r.PingInterval)

      err := timeout.MakeSimpleTimeout(func() error {
        r.send(PingHeader, false, stream)
        <- pingChan
        return nil
      }, r.PingTimeout)

      if err != nil {
        if stream == r.Stream() {
          r.Raise(err)
        }
      }
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
      splitted := strings.Split(scanner.Text(), ",")

      str := strings.Join(splitted, ",")//--------------------------
      if str != PingHeader && str != PingRespHeader { //--------------------------
        fmt.Printf("[Remote] Received %q\n", str) //--------------------------
      } //--------------------------

      switch splitted[0] {
      default:
        r.Raise(HeaderNotUnderstood)

      case ResetHeader:
        if len(splitted) <= 1 {
          r.Raise(NotEnoughFields)
          continue
        }

        idx, err := strconv.Atoi(splitted[1])
        if err != nil {
          r.Raise(err)
          continue
        }

        go (*r.ResetHandler)(idx)

      case HandShakeHeader:
        go func() {
          r.HandshakeChan.Send(true)
        }()

      case PingHeader:
        r.send(PingRespHeader, false, stream)

      case PingRespHeader:
        pingChan <- true

      case CloseHeader:
        r.Close()
        break

      case MessageHeader:
        if len(splitted) <= 1 {
          r.Raise(NotEnoughFields)
          continue
        }

        msg := strings.Join(splitted[1:], ",")

        if offset > 0 {
          offset --

          continue
        }

        r.Received++

        go func() {
          r.ReadChan.Send(msg)
        }()
      }
    }

    close(pingChan)

    if err := scanner.Err(); err != nil {
      r.Raise(err)
    }

    if !r.Check() {
      r.ReadChan.Close()
      r.HandshakeChan.Close()
    }

  }()
}
