package core

import (
  "bufio"
  "fmt"
  "errors"
  "strings"
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
)

func NewRemote() (Remote, error) {
  return &BasicRemote {
    PingChan: make(chan bool),
    ReadChan: make(chan string),
    HandshakeChan: make(chan bool),
    Sent: &[]string{},
    Rw: nil,
    Received: 0,
    Standard: NewStandardInterface(),
  }, nil
}

type BasicRemote struct {
  PingChan chan bool
  ReadChan chan string
  HandshakeChan chan bool
  Sent *[]string
  Rw *bufio.ReadWriter
  Received int
  HandshakeMessage int
  ReceivedHandshakeMessage int
  Standard standardFunctionsCloser
}

func (r *BasicRemote)Ping(timeoutDuration time.Duration) bool {
  err := timeout.MakeSimpleTimeout(func () error {
    fmt.Fprint(r.Rw, "Ping\n")
    <- r.PingChan
    return nil
  }, timeoutDuration)

  if err != nil {
    return false
  }
  return true
}

func (r *BasicRemote)CloseRemote() {
  fmt.Fprint(r.Rw, "Close\n")
  r.Rw.Flush()
}

func (r *BasicRemote)Send(msg string) {
  *r.Sent = append(*r.Sent, msg)

  if r.Rw != nil {
    fmt.Fprintf(r.Rw, "%s,%s", MessageHeader, msg)
    r.Rw.Flush()
  }
}

func (r *BasicRemote)SendHandshake() {
  if r.Rw != nil {
    fmt.Fprintf(r.Rw, HandShakeHeader)
    r.Rw.Flush()
  }
}


func (r *BasicRemote)Get() string {
  return <- r.ReadChan
}

func (r *BasicRemote)GetHandshake() chan bool {
  return r.HandshakeChan
}

func (r *BasicRemote)Reset(stream *bufio.ReadWriter) {
  r.Rw = stream
  offset := r.Received

  go func() {
    for _, msg := range *r.Sent {
      fmt.Fprintf(stream, "%s,%s", MessageHeader, msg)
      stream.Flush()
    }
  }()

  go func() {
    for r.Check() && r.Rw == stream {
      str, err := stream.ReadString('\n')
      if err != nil {
        r.Raise(err)
        return
      }

      if str == "HandShake\n" {
        r.HandshakeChan <- true
        continue

      } else if str == PingHeader {
        fmt.Fprint(stream, PingRespHeader)
        stream.Flush()
        continue

      } else if str == PingRespHeader {
        r.PingChan <- true
        continue

      } else if str == CloseHeader {

        r.Close()
        continue

      }

      splitted := strings.Split(str, ",")
      if splitted[0] == MessageHeader {
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
        r.ReadChan <- msg

      } else {
        r.Raise(errors.New("command not understood"))
        continue

      }
    }
  }()
  if !r.Check() {
    close(r.PingChan)
    close(r.ReadChan)
    close(r.HandshakeChan)
  }
}

func (r *BasicRemote)StreamHandler() (network.StreamHandler, error) {
  return func(stream network.Stream) {
    r.Reset(bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream)))
  }, nil
}

func (r *BasicRemote)Check() bool {
  return r.Standard.Check()
}

func (r *BasicRemote)Stream() *bufio.ReadWriter {
  return r.Rw
}


func (r *BasicRemote)Close() error {
  if r.Check() {

    r.Standard.Close()
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
