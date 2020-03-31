package interfaces

import (
)

type Remote struct {
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
  return <- r.InChannel
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

type EndPoint struct {
  Remotes []Remote

}
