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
    PingChan: NewChannelBool(),
    ReadChan: NewChannelString(),
    HandshakeChan: NewChannelBool(),
    Sent: &[]string{},
    Rw: nil,
    Received: 0,
    Standard: NewStandardInterface(),
  }, nil
}

type BasicRemote struct {
  PingChan *SafeChannelBool
  ReadChan *SafeChannelString
  HandshakeChan *SafeChannelBool
  Sent *[]string
  Rw *bufio.ReadWriter
  Received int
  HandshakeMessage int
  ReceivedHandshakeMessage int
  Standard BasicFunctionsCloser
}

func (r *BasicRemote)Ping(timeoutDuration time.Duration) bool {
  err := timeout.MakeSimpleTimeout(func () error {
    fmt.Fprint(r.Rw, "Ping\n")
    for {
      select {
      case <- r.PingChan.C:
        return nil
      case err, ok := <- r.ErrorChan():
        if !ok {
          break
        } else {
          return err
        }
        continue
      }
    }
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
  return <- r.ReadChan.C
}

func (r *BasicRemote)GetHandshake() chan bool {
  return r.HandshakeChan.C
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
        r.Standard.Push(err)
        return
      }

      if str == "HandShake\n" {
        go func() {
          r.HandshakeChan.Send(true)
        }()
        continue

      } else if str == PingHeader {
        fmt.Fprint(stream, PingRespHeader)
        stream.Flush()
        continue

      } else if str == PingRespHeader {
        go func() {
          r.PingChan.Send(true)
        }()
        continue

      } else if str == CloseHeader {

        r.Close()
        continue

      }

      splitted := strings.Split(str, ",")
      if splitted[0] == MessageHeader {
        if len(splitted) <= 1 {
          r.Standard.Push(errors.New("not enough fields"))
          r.Close()
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

      } else {
        r.Standard.Push(errors.New("command not understood"))
        r.Close()

      }
    }
  }()
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
    go r.PingChan.SafeClose(true)
    go r.HandshakeChan.SafeClose(true)
    go r.ReadChan.SafeClose(false)

    r.Standard.Close()
  }
  return nil
}

func (r *BasicRemote)CloseChan() chan bool {
  return r.Standard.CloseChan()
}

func (r *BasicRemote)ErrorChan() chan error {
  return r.Standard.ErrorChan()
}
