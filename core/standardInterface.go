package core

import (
  "sync"
)

func NewStandardInterface() BasicFunctionsCloser {
  return BasicFunctionsCloser {
    Ended: false,
    EndChan: [] *SafeChannelBool{},
    Error: [] *SafeChannelError{},
  }
}

type BasicFunctionsCloser struct {
  Mutex sync.Mutex
  Ended bool
  EndChan [] *SafeChannelBool
  Error [] *SafeChannelError
}

func (b *BasicFunctionsCloser)Close() error {
  b.Mutex.Lock()
  defer b.Mutex.Unlock()
  if !b.Ended {
    b.Ended = true

    for i := range b.EndChan {
      go func() {
        b.EndChan[i].Send(true)
        b.EndChan[i].SafeClose(true)
      }()

    }

    for i := range b.Error {
      go b.Error[i].SafeClose(true)
    }
  }

  return nil
}

func (b *BasicFunctionsCloser)Push(err error) {
  if b.Check() && err != nil {
    for i := range b.Error {
      go b.Error[i].Send(err)
    }
  }
}

func (b *BasicFunctionsCloser)Check() bool {
  b.Mutex.Lock()
  defer b.Mutex.Unlock()
  return !b.Ended
}

func (b *BasicFunctionsCloser)CloseChan() chan bool {
  EndChan := NewChannelBool()
  b.EndChan = append(b.EndChan, EndChan)
  return EndChan.C
}

func (b *BasicFunctionsCloser)ErrorChan() chan error {
  Error := NewChannelError()
  b.Error = append(b.Error, Error)
  return Error.C
}
