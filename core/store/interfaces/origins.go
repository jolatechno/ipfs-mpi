package interfaces

import (
  "errors"
  "time"
)

type Remote struct {
  Timeout time.Duration
  NewSender *func(func(string)) *func(string) error
  Sender *func(string) error
  Sent []string
  InChannel chan string
  Offset int
  Received int
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
    r.Replace()
  }
}

func (r *Remote)Get() string {
  select {
    case res := <- r.InChannel:
      return res

    case <- time.After(r.Timeout):
      r.Replace()
      return r.Get()
  }
}

func (r *Remote)Replace() {
  r.Offset = r.Received
  r.Sender = (*r.NewSender)(r.Push)

  for _, msg := range r.Sent {
    err := (*r.Sender)(msg)
    if err != nil {
      r.Replace()
    }
  }
}

type Comm []Remote

func (c *Comm)Send(i int, msg string) error {
  if len(*c) <= i {
    return errors.New("Comm index out of range")
  }

  (*c)[i].Send(msg)
  return nil
}

func (c *Comm)Get(i int) (string, error) {
  if len(*c) <= i {
    return "", errors.New("Comm index out of range")
  }

  return (*c)[i].Get(), nil
}
