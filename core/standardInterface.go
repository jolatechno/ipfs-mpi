package core

import (

)

func NewStandardInterface() BasicFunctionsCloser {
  return BasicFunctionsCloser {
    Ended: false,
    EndChan: []chan bool{},
    Error: []chan error{},
  }
}

type BasicFunctionsCloser struct {
  Ended bool
  EndChan []chan bool
  Error []chan error
}

func (b *BasicFunctionsCloser)Close() error {
  b.Ended = true

  for _, EndChan := range b.EndChan {
    go func() {
      EndChan <- true
    }()
  }

  return nil
}

func (b *BasicFunctionsCloser)Push(err error) {
  for _, Error := range b.Error {
    go func() {
      Error <- err
    }()
  }
}

func (b *BasicFunctionsCloser)Check() bool {
  return !b.Ended
}

func (b *BasicFunctionsCloser)CloseChan() chan bool {
  EndChan := make(chan bool)
  b.EndChan = append(b.EndChan, EndChan)
  return EndChan
}

func (b *BasicFunctionsCloser)ErrorChan() chan error {
  Error := make(chan error)
  b.Error = append(b.Error, Error)
  return Error
}
