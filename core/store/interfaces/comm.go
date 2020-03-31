package interfaces

import (
  "errors"
  "time"
  "bufio"
)

type Remote struct {
  Timeout time.Duration
  NewAdrress *func() string
  NewSender *func(string) *bufio.ReadWriter
  NotifyReset *func(string)
  Stream *bufio.ReadWriter
  Sent []string
  InChannel chan string
  Offset int
  Received int
}

func NewRemote(newAdrress *func() string, newSender *func(string) *bufio.ReadWriter, notifyReset *func(string), timeout time.Duration) *Remote {
  return &Remote{
    Timeout:timeout,
    NewAdrress: newAdrress,
    NewSender:newSender,
    NotifyReset:notifyReset,
    Stream:nil,
    Sent:[]string{},
    InChannel:make(chan string),
    Offset:0,
    Received:0,
  }
}

func (r *Remote)Send(msg string) {
  r.Sent = append(r.Sent, msg)
  _, err := r.Stream.WriteString(msg)
  if err != nil {
    r.Reset()
    return
  }

  err = r.Stream.Flush()
  if err != nil {
    r.Reset()
    return
  }
}

func (r *Remote)Get() string {
  readChan := make(chan string)
	errChan := make(chan error)

  for {
    go func() {
  		str, err := r.Stream.ReadString('\n')
  		if err != nil {
  			errChan <- err
  		} else {
  			readChan <- str
  		}
  	}()

    select {
    case res := <- readChan:
      if r.Offset > 0 {
        r.Offset --
      } else {
        close(readChan)
        close(errChan)

        r.Received ++
        return res
      }

    case <- errChan:
      close(readChan)
      close(errChan)

      r.Reset()
      return r.Get()

    case <- time.After(r.Timeout):
      close(readChan)
      close(errChan)

      r.Reset()
      return r.Get()
    }
  }
}

func (r *Remote)Reset() {
  addr := (*r.NewAdrress)()
  r.Offset = r.Received
  r.Stream = (*r.NewSender)(addr)
  (*r.NotifyReset)(addr)

  for _, msg := range r.Sent {
    r.Send(msg)
  }
}

func (r *Remote)Replace(addr string) {
  r.Offset = r.Received
  r.Stream = (*r.NewSender)(addr)

  for _, msg := range r.Sent {
    r.Send(msg)
  }
}

type Comm struct {
  Remotes []*Remote
  Kill *func()
}

func NewComm(n int, kill *func(), newAdrress *func() string, newSender *func(string) *bufio.ReadWriter, encodeNotify *func(int, string) string, timeout time.Duration) Comm {
  addrs := make([]string, n)
  for i := range addrs {
    addrs[i] = ""
  }
  return LoadComm(0, addrs, kill, newAdrress, newSender, encodeNotify, timeout)
}

func LoadComm(idx int, addrs []string, kill *func(), newAdrress *func() string, newSender *func(string) *bufio.ReadWriter, encodeNotify *func(int, string) string, timeout time.Duration) Comm {
  c := Comm{
    Remotes:make([]*Remote, len(addrs)),
    Kill:kill,
  }

  notifyResetIdx := func(i int) *func(string) {
    notify := func(str string) {
      for _, r := range c.Remotes {
        (*r).Send((*encodeNotify)(i, str))
      }
    }

    return &notify
  }

  for i, addr := range addrs {
    if i != idx {
      c.Remotes[i] = NewRemote(newAdrress, newSender, notifyResetIdx(i), timeout)
      if addr == "" {
        (*c.Remotes[i]).Reset()
      } else {
        (*c.Remotes[i]).Replace(addr)
      }
    }
  }

  return c
}

func (c *Comm)Send(i int, msg string) error {
  if len(c.Remotes) <= i {
    return errors.New("Comm index out of range")
  }

  (*c.Remotes[i]).Send(msg)
  return nil
}

func (c *Comm)Get(i int) (string, error) {
  if len(c.Remotes) <= i {
    return "", errors.New("Comm index out of range")
  }

  return (*c.Remotes[i]).Get(), nil
}

func (c *Comm)Replace(i int, addr string) error {
  if len(c.Remotes) <= i {
    return errors.New("Comm index out of range")
  }

  (*c.Remotes[i]).Replace(addr)
  return nil
}
