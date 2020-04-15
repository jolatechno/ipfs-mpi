package core

import (
  "bufio"
  "io"
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
  Rw io.ReadWriteCloser
  Received int
  HandshakeMessage int
  ReceivedHandshakeMessage int
  Standard standardFunctionsCloser
}

func (r *BasicRemote)Ping(timeoutDuration time.Duration) bool {
  if !r.Check() || r.Rw == nil {
    return false
  }

  err := timeout.MakeSimpleTimeout(func () error {
    fmt.Fprint(r.Rw, PingHeader)
    _, ok := <- r.PingChan
    if !ok {
      return errors.New("Channel Closed")
    }
    return nil
  }, timeoutDuration)

  if err != nil {
    return false
  }
  return true
}

func (r *BasicRemote)CloseRemote() {
  if r.Rw != nil {

    fmt.Println("[Remote] CloseRemote") //--------------------------

    writer := bufio.NewWriter(r.Rw)

    fmt.Fprint(writer, CloseHeader)
    writer.Flush()
  }
}

func (r *BasicRemote)Send(msg string) {
  *r.Sent = append(*r.Sent, msg)

  if r.Rw != nil {
    writer := bufio.NewWriter(r.Rw)

    fmt.Fprintf(writer, "%s,%s", MessageHeader, msg)
    writer.Flush()
  }
}

func (r *BasicRemote)SendHandshake() {
  if r.Rw != nil {
    writer := bufio.NewWriter(r.Rw)

    fmt.Fprintf(writer, HandShakeHeader)
    writer.Flush()
  }
}


func (r *BasicRemote)Get() string {
  return <- r.ReadChan
}

func (r *BasicRemote)GetHandshake() chan bool {
  return r.HandshakeChan
}

func (r *BasicRemote)Reset(stream io.ReadWriteCloser) {
  r.Rw = stream
  offset := r.Received

  writer := bufio.NewWriter(stream)
  reader := bufio.NewReader(stream)

  go func() {
    for _, msg := range *r.Sent {
      fmt.Fprintf(writer, "%s,%s", MessageHeader, msg)
      writer.Flush()
    }
  }()

  go func() {
    for r.Check() && r.Rw == stream {
      str, err := reader.ReadString('\n')
      if err != nil {
        r.Raise(err)
        return
      }

      fmt.Printf("[Remote] Received %q\n", str) //--------------------------

      if str == HandShakeHeader {
        r.HandshakeChan <- true
        continue

      } else if str == PingHeader {
        fmt.Fprint(writer, PingRespHeader)
        writer.Flush()
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
    r.Reset(stream.(io.ReadWriteCloser))
  }, nil
}

func (r *BasicRemote)Check() bool {
  return r.Standard.Check()
}

func (r *BasicRemote)Stream() io.ReadWriteCloser {
  return r.Rw
}


func (r *BasicRemote)Close() error {
  if r.Check() {

    fmt.Println("[Remote] Closing") //--------------------------

    if r.Rw != nil {
      r.Rw.Close()
      r.Rw = nil
    }

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
