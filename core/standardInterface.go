package core

import (
  "sync"
)

func NewStandardInterface() standardFunctionsCloser {
  return &BasicFunctionsCloser {}
}

type BasicFunctionsCloser struct {
  Mutex sync.Mutex
  Ended bool
  EndHandler *func()
  ErrorHandler *func(error)
}

func (b *BasicFunctionsCloser)Close() error {
  b.Mutex.Lock()
  defer b.Mutex.Unlock()
  if !b.Ended {
    if b.EndHandler != nil {
      (*b.EndHandler)()
    }
    b.Ended = true
  }

  return nil
}

func (b *BasicFunctionsCloser)Raise(err error) {
  if b.ErrorHandler != nil {
    (*b.ErrorHandler)(err)
  }
}

func (b *BasicFunctionsCloser)Check() bool {
  b.Mutex.Lock()
  defer b.Mutex.Unlock()
  return !b.Ended
}

func (b *BasicFunctionsCloser)SetErrorHandler(handler func(error)) {
  b.ErrorHandler = &handler
}

func (b *BasicFunctionsCloser)SetCloseHandler(handler func()) {
  b.EndHandler = &handler
}
