package core

import (
  "errors"
)

type StdInterface struct {
  InChan chan string
  OutChan chan Message
  RequestChan chan int
}

func NewInterface(file string) (Interface, error) {
  inter := StdInterface{
    InChan: make(chan string),
    OutChan: make(chan Message),
    RequestChan: make(chan int),
  }

  return &inter, errors.New("Not yet implemented")
}

func (s *StdInterface)Message() chan Message {
  return s.OutChan
}

func (s *StdInterface)Request() chan int {
  return s.RequestChan
}

func (s *StdInterface)Push(msg string) error {
  s.InChan <- msg
  return nil
}
