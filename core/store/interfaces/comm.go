package interfaces

import (
  "errors"
  "time"
  "bufio"
)

type Remote struct {
  Timeout time.Duration
  NewAdrress *func() string
  NewSender *func(string) (chan *bufio.ReadWriter, chan error)
  NotifyReset *func(string)
  Stream *bufio.ReadWriter
  Sent []string
  InChannel chan string
  Offset int
  Received int
}

func NewRemote(newAdrress *func() string, newSender *func(string) (chan *bufio.ReadWriter, chan error), notifyReset *func(string), timeout time.Duration) *Remote {
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

  readerChan, errChan := (*r.NewSender)(addr)
  switch {
  case <- errChan || <- time.After(r.Timeout):
    close(readChan)
    close(errChan)

    r.Reset()
    return

  case reader := <- readerChan:
    close(readChan)
    close(errChan)

    r.Stream := reader
    (*r.NotifyReset)(addr)

    for _, msg := range r.Sent {
      r.Send(msg)
    }
    return
  }
}

func (r *Remote)Replace(addr string) {
  readerChan, errChan := (*r.NewSender)(addr)
  switch {
  case <- errChan || <- time.After(r.Timeout):
    close(readChan)
    close(errChan)

    r.Reset()
    return

  case reader := <- readerChan:
    close(readChan)
    close(errChan)

    r.Stream := reader
    (*r.NotifyReset)(addr)

    for _, msg := range r.Sent {
      r.Send(msg)
    }
    return
  }
}

type Comm []*Remote

func NewComm(n int, newAdrress *func() string, newSender *func(string) (chan *bufio.ReadWriter, chan error), encodeNotify *func(int, string) string, timeout time.Duration) Comm {
  addrs := make([]string, n)
  for i := range addrs {
    addrs[i] = ""
  }
  return LoadComm(0, addrs, newAdrress, newSender, encodeNotify, timeout)
}

func LoadComm(idx int, addrs []string, newAdrress *func() string, newSender *func(string) (chan *bufio.ReadWriter, chan error), encodeNotify *func(int, string) string, timeout time.Duration) Comm {
  c := make([]*Remote, len(addrs))
  c[idx] = nil

  notifyResetIdx := func(i int) *func(string) {
    notify := func(str string) {
      for _, r := range c {
        (*r).Send((*encodeNotify)(i, str))
      }
    }

    return &notify
  }

  for i, addr := range addrs {
    if i != idx {
      c[i] = NewRemote(newAdrress, newSender, notifyResetIdx(i), timeout)
      if addr == "" {
        (*c[i]).Reset()
      } else {
        (*c[i]).Replace(addr)
      }
    }
  }

  return c
}

func (c *Comm)Send(i int, msg string) error {
  if len(*c) <= i {
    return errors.New("Comm index out of range")
  }

  (*(*c)[i]).Send(msg)
  return nil
}

func (c *Comm)Get(i int) (string, error) {
  if len(*c) <= i {
    return "", errors.New("Comm index out of range")
  }

  return (*(*c)[i]).Get(), nil
}

func (c *Comm)Replace(i int, addr string) error {
  if len(*c) <= i {
    return errors.New("Comm index out of range")
  }

  (*(*c)[i]).Replace(addr)
  return nil
}
