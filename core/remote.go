package core

import (
  "bufio"
  "fmt"

  "github.com/libp2p/go-libp2p-core/network"
)

func NewRemote(handshakeMessage int) (Remote, error) {
  return &BasicRemote {
    Closed: false,
    EndChan: make(chan bool),
    Error: make(chan error),
    ReadChan: make(chan string),
    Sent: []string{},
    Rw: nil,
    Offset: 0,
    Received: 0,
    HandshakeMessage: handshakeMessage,
    ReceivedHandshakeMessage: 0,
  }, nil
}

type BasicRemote struct {
  Closed bool
  EndChan chan bool
  Error chan error
  ReadChan chan string
  Sent []string
  Rw *bufio.ReadWriter
  Offset int
  Received int
  HandshakeMessage int
  ReceivedHandshakeMessage int
}

func (r *BasicRemote)Send(msg string) {

  fmt.Printf("[Remote] Sending %q\n", msg) //--------------------------

  r.Sent = append(r.Sent, msg)

  fmt.Fprint(r.Rw, msg)
  r.Rw.Flush()
}

func (r *BasicRemote)Get() string {

  fmt.Println("[Remote] Requesting") //--------------------------

  return <- r.ReadChan

}

func (r *BasicRemote)Reset(stream *bufio.ReadWriter) {

  fmt.Println("[Remote] reset") //--------------------------

  r.Rw = stream
  r.Offset = r.Received
  r.ReceivedHandshakeMessage = 0
  for _, msg := range r.Sent {
    fmt.Fprint(r.Rw, msg)
    r.Rw.Flush()
  }

  go func() {
    for r.Check() {
      if r.Rw == nil {
        return
      }

      str, err := r.Rw.ReadString('\n')
      if err != nil {
        r.Error <- err
        return
      }

      if r.ReceivedHandshakeMessage < r.HandshakeMessage {
        r.ReceivedHandshakeMessage++
        r.ReadChan <- str
      }

      if r.Offset > 0 {
        r.Offset --
      }

      fmt.Printf("[Remote] Received %q\n", str) //--------------------------

      r.Received++
      r.ReadChan <- str
    }
  }()
}

func (r *BasicRemote)StreamHandler() (network.StreamHandler, error) {
  return func(stream network.Stream) {
    r.Reset(bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream)))
  }, nil
}

func (r *BasicRemote)Check() bool {
  return !r.Closed
}

func (r *BasicRemote)Stream() *bufio.ReadWriter {
  return r.Rw
}


func (r *BasicRemote)Close() error {
  r.EndChan <- true
  r.Closed = true
  return nil
}

func (r *BasicRemote)CloseChan() chan bool {
  return r.EndChan
}

func (r *BasicRemote)ErrorChan() chan error {
  return r.Error
}
