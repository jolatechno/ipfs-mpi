package core

import (
  "bufio"
  "io"
  "fmt"
  "errors"
  "strings"
  "sync"
  "time"

  "github.com/libp2p/go-libp2p-core/network"

  "github.com/jolatechno/go-timeout"
)

var (
  HandShakeHeader = "HandShake\n"
  MessageHeader = "Msg"
  CloseHeader = "Close\n"
  PingHeader = "Ping\n"
  PingRespHeader = "PingResp\n"

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
  c.Mutex.Lock()
  defer c.Mutex.Unlock()
  if !c.Ended {
    c.Chan <- str
  }
}

func (c *safeChannelString)Close() {
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

  if str != PingHeader && str != PingRespHeader { //--------------------------
    fmt.Printf("[Remote] Sending %q\n", str) //--------------------------
  } //--------------------------

  if stream := r.Stream(); stream != io.ReadWriteCloser(nil) {
    if len(referenceStream) == 1 && referenceStream[0] != stream {
      return
    }

    writer := bufio.NewWriter(stream)

    _, err := writer.WriteString(str)
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

func (r *BasicRemote)SetPingInterval(interval time.Duration) {
  r.PingInterval = interval
}

func (r *BasicRemote)SetPingTimeout(timeoutDuration time.Duration) {
  r.PingTimeout = timeoutDuration
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

func (r *BasicRemote)Reset(stream io.ReadWriteCloser) {
  if !r.Check() {
    return
  }

  r.StreamMutex.Lock()
  defer r.StreamMutex.Unlock()

  r.Rw = stream
  if stream == io.ReadWriteCloser(nil) {
    go func() {
      for r.Check() {
        time.Sleep(ErrorInterval)
        if r.Stream() == io.ReadWriteCloser(nil) {
          r.Raise(NilStreamError)
        } else {
          return
        }
      }
    }()

    return
  }

  offset := r.Received
  pingChan := make(chan bool)

  go func() {
    r.WriteMutex.Lock()
    defer r.WriteMutex.Unlock()

    for _, msg := range *r.Sent {
      r.send(fmt.Sprintf("%s,%s", MessageHeader, msg), true)
    }
  }()

  go func() {
    for r.Check() && r.Stream() == stream {
      time.Sleep(r.PingInterval)

      err := timeout.MakeSimpleTimeout(func() error {
        r.send(PingHeader, false, stream)
        <- pingChan
        return nil
      }, r.PingTimeout)

      if err != nil {
        r.Raise(err)
      }
    }
  }()

  go func() {
    for r.Check() &&  r.Stream() == stream {
      str, err := bufio.NewReader(stream).ReadString('\n')
      if err != nil {
        if stream == r.Stream() {
          r.Raise(err)
        }

        return
      }

      if str != PingHeader && str != PingRespHeader { //--------------------------
        fmt.Printf("[Remote] Received %q\n", str) //--------------------------
      } //--------------------------

      splitted := strings.Split(str, ",")

      switch splitted[0] {
      default:
        r.Raise(errors.New("header not understood"))

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

      case MessageHeader:
        if len(splitted) <= 1 {
          r.Raise(errors.New("not enough fields"))
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
    if !r.Check() {
      r.ReadChan.Close()
      r.HandshakeChan.Close()
    }

  }()
}

func (r *BasicRemote)StreamHandler() (network.StreamHandler, error) {
  return func(stream network.Stream) {

    fmt.Println("[Remote] [StreamHandler]") //--------------------------

    r.Reset(stream.(io.ReadWriteCloser))
  }, nil
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
  if r.Check() {
    r.Standard.Close()

    if stream := r.Stream(); stream != io.ReadWriteCloser(nil) {
      stream.Close()
      r.Rw = io.ReadWriteCloser(nil)
    }
  }
  return nil
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
