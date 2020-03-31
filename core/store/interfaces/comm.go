package interfaces

import (
  "errors"
  "time"
)

type Remote struct {
  Timeout time.Duration
  NewAdrress *func() string
  NewSender *func(string, func(string)) *func(string) error
  NotifyReset *func(string)
  Sender *func(string) error
  Sent []string
  InChannel chan string
  Offset int
  Received int
}

func NewRemote(newAdrress *func() string, newSender *func(string, func(string)) *func(string) error, notifyReset *func(string), timeout time.Duration) *Remote {
  return &Remote{
    Timeout:timeout,
    NewAdrress: newAdrress,
    NewSender:newSender,
    NotifyReset:notifyReset,
    Sender:nil,
    Sent:[]string{},
    InChannel:make(chan string),
    Offset:0,
    Received:0,
  }

}

func (r *Remote)Push(msg string) {
  if r.Offset == 0 {
    r.Received++
    go func(){
      r.InChannel <- msg
    }()
  } else {
    r.Offset--
  }
}

func (r *Remote)Send(msg string) {
  r.Sent = append(r.Sent, msg)
  err := (*r.Sender)(msg)
  if err != nil {
    r.Reset()
  }
}

func (r *Remote)Get() string {
  select {
    case res := <- r.InChannel:
      return res

    case <- time.After(r.Timeout):
      r.Reset()
      return r.Get()
  }
}

func (r *Remote)Reset() {
  addr := (*r.NewAdrress)()
  r.Offset = r.Received
  r.Sender = (*r.NewSender)(addr, r.Push)

  for _, msg := range r.Sent {
    err := (*r.Sender)(msg)
    if err != nil {
      r.Reset()
      return
    }
  }

  (*r.NotifyReset)(addr)
}

func (r *Remote)Replace(addr string) {
  r.Offset = r.Received
  r.Sender = (*r.NewSender)(addr, r.Push)

  for _, msg := range r.Sent {
    err := (*r.Sender)(msg)
    if err != nil {
      r.Reset()
      return
    }
  }
}

type Comm struct {
  Remotes []*Remote
  Kill *func()
}

func NewComm(n int, kill *func(), newAdrress *func() string, newSender *func(string, func(string)) *func(string) error, notifyReset *func(int, string), timeout time.Duration) Comm {
  addrs := make([]string, n)
  for i := range addrs {
    addrs[i] = ""
  }
  return LoadComm(0, addrs, kill, newAdrress, newSender, notifyReset, timeout)
}

func LoadComm(idx int, addrs []string, kill *func(), newAdrress *func() string, newSender *func(string, func(string)) *func(string) error, notifyReset *func(int, string), timeout time.Duration) Comm {
  c := Comm{
    Remotes:make([]*Remote, len(addrs)),
    Kill:kill,
  }

  for i, addr := range addrs {
    if i != idx {
      c.Remotes[i] = NewRemote(newAdrress, newSender, notifyResetIdx(i, notifyReset), timeout)
      if addr == "" {
        (*c.Remotes[i]).Reset()
      } else {
        (*c.Remotes[i]).Replace(addr)
      }
    } else {
      newAddressSelf := func() string {
        (*c.Kill)()
        return ""
      }

      notifyResetSelf := func(_ string) {
        (*c.Kill)()
      }

      newSenderSelf := func(_ string, push func(string)) *func(string) error{
        send := func(str string) error {
          push(str)
          return nil
        }

        return &send
      }

      c.Remotes[i] = NewRemote(&newAddressSelf, &newSenderSelf, &notifyResetSelf, timeout)
      (*c.Remotes[i]).Replace("")
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

func notifyResetIdx(i int, notifyReset *func(int, string)) *func(string) {
  notify := func(addr string) {
    (*notifyReset)(i, addr)
  }

  return &notify
}
