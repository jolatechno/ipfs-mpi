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

func NewRemote(handshakeMessage int) (Remote, error) {
  return &BasicRemote {
    PingChan: make(chan bool),
    ReadChan: make(chan string),
    HandshakeChan: make(chan string),
    Sent: &[]string{},
    Rw: nil,
    Offset: 0,
    Received: 0,
    HandshakeMessage: handshakeMessage,
    ReceivedHandshakeMessage: 0,
    Standard: NewStandardInterface(),
  }, nil
}

type BasicRemote struct {
  PingChan chan bool
  ReadChan chan string
  HandshakeChan chan string
  Sent *[]string
  Rw *bufio.ReadWriter
  Offset int
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

  if r.ReceivedHandshakeMessage >= r.HandshakeMessage { //shouldn't be strictly greater
    *r.Sent = append(*r.Sent, msg)
  }

  if r.Rw != nil {
    fmt.Fprintf(r.Rw, "Msg,%s", msg)
    r.Rw.Flush()
  }
}

func (r *BasicRemote)Get() string {
  return <- r.ReadChan
}

func (r *BasicRemote)GetHandshake() string {
  return <- r.HandshakeChan
}

func (r *BasicRemote)Reset(stream *bufio.ReadWriter) {
  r.Rw = stream
  r.Offset = r.Received
  r.ReceivedHandshakeMessage = 0

  go func() {
    for r.Check() && r.Rw == stream {
      str, err := stream.ReadString('\n')
      if err != nil {
        r.Standard.Push(err)
        return
      }

      splitted := strings.Split(str, ",")
      if splitted[0] == "Msg" {
        if len(splitted) <= 1 {
          r.Standard.Push(errors.New("not enough fields"))
          r.Close()
        }

        msg := strings.Join(splitted[1:], ",")

        fmt.Printf("[Remote] Received %q\n", msg) //--------------------------

        if r.ReceivedHandshakeMessage < r.HandshakeMessage {
          r.ReceivedHandshakeMessage++
          go func() {
            r.HandshakeChan <- msg
          }()

          if r.ReceivedHandshakeMessage == r.HandshakeMessage {
            for _, msg_hist := range *r.Sent {
              fmt.Fprintf(stream, "Msg,%s", msg_hist)
              stream.Flush()
            }
          }

          continue
        }

        if r.Offset > 0 {
          r.Offset --

          continue
        }

        r.Received++
        go func() {
          r.ReadChan <- msg
        }()

      } else if splitted[0] == "Ping\n" {
        fmt.Fprint(r.Rw, "PingResp\n")
        r.Rw.Flush()

      } else if splitted[0] == "PingResp\n" {
        r.PingChan <- true

      } else if splitted[0] == "Close\n" {
        r.Close()

      } else {
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
