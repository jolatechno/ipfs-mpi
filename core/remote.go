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
  Standard BasicFunctionsCloser
}

func (r *BasicRemote)Ping(timeoutDuration time.Duration) bool {
  err := timeout.MakeSimpleTimeout(func () error {
    fmt.Fprint(r.Rw, "Ping\n")
    for {
      select {
      case <- r.PingChan:
        return nil
      case err, ok := <- r.ErrorChan():
        if ok {
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

  fmt.Printf("[Remote] Sending %q\n", msg) //--------------------------

  *r.Sent = append(*r.Sent, msg)

  if r.Rw != nil {
    fmt.Fprintf(r.Rw, "%s,%s", MessageHeader, msg)
    r.Rw.Flush()
  }
}

func (r *BasicRemote)SendHandshake() {

  fmt.Println("[Remote] Sending Handshake") //--------------------------

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
        r.Standard.Push(err)
        return
      }

      if str == "HandShake\n" {

        fmt.Println("[Remote] Received Handshake") //--------------------------

        go func() {
          r.HandshakeChan <- true
        }()
        continue

      } else if str == PingHeader {
        fmt.Fprint(stream, PingRespHeader)
        stream.Flush()
        continue

      } else if str == PingRespHeader {
        go func() {
          r.PingChan <- true
        }()
        continue

      } else if str == CloseHeader {

        fmt.Println("[Remote] Closing requested") //--------------------------

        r.Close()
        continue

      }

      splitted := strings.Split(str, ",")
      if splitted[0] == MessageHeader {
        if len(splitted) <= 1 {

          fmt.Println("[Remote] Closing, not enough fields") //--------------------------

          r.Standard.Push(errors.New("not enough fields"))
          r.Close()
        }

        msg := strings.Join(splitted[1:], ",")

        fmt.Printf("[Remote] Received %q\n", msg) //--------------------------

        if offset > 0 {
          offset --

          continue
        }

        r.Received++
        go func() {
          r.ReadChan <- msg
        }()

      } else {

        fmt.Println("[Remote] Closing, command not understood") //--------------------------

        r.Standard.Push(errors.New("command not understood"))
        r.Close()

      }
    }
  }()
}

func (r *BasicRemote)StreamHandler() (network.StreamHandler, error) {
  return func(stream network.Stream) {

    fmt.Println("[Remote] Streamhandler called ") //--------------------------

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

func (r *BasicRemote)CloseChan() chan bool {
  return r.Standard.CloseChan()
}

func (r *BasicRemote)ErrorChan() chan error {
  return r.Standard.ErrorChan()
}
