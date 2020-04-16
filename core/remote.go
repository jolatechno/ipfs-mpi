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

  StandardTimeout = time.Second
  StandardPingInterval = time.Second
)

func NewRemote() (Remote, error) {
  return &BasicRemote {
    PingInterval: StandardPingInterval,
    PingTimeout: StandardTimeout,
    ReadChan: make(chan string),
    HandshakeChan: make(chan bool),
    Sent: &[]string{},
    Rw: nil,
    Received: 0,
    Standard: NewStandardInterface(),
  }, nil
}

type BasicRemote struct {
  PingInterval time.Duration
  PingTimeout time.Duration
  ReadChan chan string
  HandshakeChan chan bool
  Sent *[]string
  Rw io.ReadWriteCloser
  Received int
  HandshakeMessage int
  ReceivedHandshakeMessage int
  Standard standardFunctionsCloser
}

func (r *BasicRemote)SetPingInterval(interval time.Duration) {
  r.PingInterval = interval
}

func (r *BasicRemote)SetPingTimeout(timeoutDuration time.Duration) {
  r.PingTimeout = timeoutDuration
}

func (r *BasicRemote)CloseRemote() {
  if stream := r.Rw; stream != io.ReadWriteCloser(nil) {

    fmt.Println("[Remote] CloseRemote") //--------------------------

    writer := bufio.NewWriter(stream)

    _, err := writer.WriteString(CloseHeader)
    if err != nil {
      r.Raise(err)
    }

    go func() {
      err = writer.Flush()
      if err != nil {
        r.Raise(err)
        return
      }
    }()
  }
}

func (r *BasicRemote)Send(msg string) {
  *r.Sent = append(*r.Sent, msg)

  fmt.Printf("[Remote] Sending %q\n", msg ) //--------------------------

  if stream := r.Rw; stream != io.ReadWriteCloser(nil) {
    writer := bufio.NewWriter(stream)

    _, err := writer.WriteString(fmt.Sprintf("%s,%s", MessageHeader, msg))
    if err != nil {
      r.Raise(err)
      return
    }

    go func() {
      err = writer.Flush()
      if err != nil {
        r.Raise(err)
        return
      }
    }()
  }
}

func (r *BasicRemote)SendHandshake() {
  if stream := r.Rw; stream != io.ReadWriteCloser(nil) {
    writer := bufio.NewWriter(stream)

    _, err := writer.WriteString(HandShakeHeader)
    if err != nil {
      r.Raise(err)
    }

    go func() {
      err = writer.Flush()
      if err != nil {
        r.Raise(err)
        return
      }
    }()
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
  if stream == io.ReadWriteCloser(nil) {
    return
  }

  offset := r.Received
  writer := bufio.NewWriter(stream)
  pingChan := make(chan bool)

  go func() {
    for _, msg := range *r.Sent {
      _, err := writer.WriteString(fmt.Sprintf("%s,%s", MessageHeader, msg))
      if err != nil {
        r.Raise(err)
        return
      }

      go func() {
        err = writer.Flush()
        if err != nil {
          r.Raise(err)
          return
        }
      }()
    }
  }()

  go func() {
    for r.Check() && r.Rw == stream {
      time.Sleep(r.PingInterval)

      _, err := bufio.NewWriter(stream).WriteString(PingHeader)
      if err != nil {
        r.Raise(err)
      }

      go func() {
        err = writer.Flush()
        if err != nil {
          r.Raise(err)
          return
        }
      }()

      err = timeout.MakeSimpleTimeout(func() error {
        <- pingChan
        return nil
      }, r.PingTimeout)

      if err != nil {
        if r.Check() {
          r.Raise(err)
        }
      }
    }
  }()

  go func() {
    for r.Check() && r.Rw == stream {
      str, err := bufio.NewReader(stream).ReadString('\n')
      if err != nil {
        if r.Check() {
          r.Raise(err)
        }
        return
      }

      if str == HandShakeHeader {
        r.HandshakeChan <- true
        continue

      } else if str == PingHeader {
        _, err := bufio.NewWriter(stream).WriteString(PingRespHeader)
        if err != nil {
          r.Raise(err)
        }

        go func() {
          err = writer.Flush()
          if err != nil {
            r.Raise(err)
            return
          }
        }()
        continue

      } else if str == PingRespHeader {
        pingChan <- true
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
  close(pingChan)
  if !r.Check() {
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
    if r.Rw != io.ReadWriteCloser(nil) {
      r.Rw.Close()
      r.Rw = io.ReadWriteCloser(nil)
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
