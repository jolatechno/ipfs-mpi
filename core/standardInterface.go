package core

import (
  "sync"
)

func NewStandardInterface() BasicFunctionsCloser {
  return BasicFunctionsCloser {
    Ended: false,
    EndChan: []chan bool{},
    Error: []chan error{},
  }
}

type BasicFunctionsCloser struct {
  Mutex sync.Mutex
  Ended bool
  EndChan []chan bool
  Error []chan error
}

func (b *BasicFunctionsCloser)Close() error {
  b.Mutex.Lock()
  defer b.Mutex.Unlock()
  if !b.Ended {
    b.Ended = true
    for i := range b.EndChan {
      go func() {
        b.EndChan[i] <- true
        close(b.EndChan[i])
      }()
    }

    for i := range b.Error {
      go func() {
        for len(b.Error[i]) > 0 {
          <- b.Error[i]
        }
        close(b.Error[i])
      }()
    }
  }

  return nil
}

func (b *BasicFunctionsCloser)Push(err error) {
  if b.Check() && err != nil {
    for i := range b.Error {
      go func() {
        b.Error[i] <- err
      }()
    }
  }
}

func (b *BasicFunctionsCloser)Check() bool {
  b.Mutex.Lock()
  defer b.Mutex.Unlock()
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
