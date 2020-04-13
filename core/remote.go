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
    Received: -handshakeMessage,
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

  fmt.Println("[Remote] reset 0") //--------------------------

  r.Rw = stream
  r.Offset = r.Received
  for _, msg := range r.Sent {
    fmt.Fprint(r.Rw, msg)
    r.Rw.Flush()
  }

  fmt.Println("[Remote] reset 1") //--------------------------

  go func() {
    for r.Check() {
      for r.Offset > 0 {
        _, err := r.Rw.ReadString('\n')
        if err == nil {
          r.Offset --
        }
      }

      str, err := r.Rw.ReadString('\n')
      if err != nil {
        return
      }

      fmt.Printf("[Remote] Received %q\n", str) //--------------------------

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
