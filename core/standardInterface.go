package core

import (
  "sync"
)

var (
  nilEndHandler = func() {}
  nilErrorHandler = func(err error) {}
)

func NewStandardInterface() standardFunctionsCloser {
  return &BasicFunctionsCloser {
    EndHandler: &nilEndHandler,
    ErrorHandler: &nilErrorHandler,
  }
}

type BasicFunctionsCloser struct {
  Mutex sync.Mutex
  Ended bool
  EndHandler *func()
  ErrorHandler *func(error)
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

func (b *BasicFunctionsCloser)Close() error {
  defer recover()

  b.Mutex.Lock()
  defer b.Mutex.Unlock()
  if !b.Ended {
    (*b.EndHandler)()

    b.Ended = true
  }

  return nil
}

func (b *BasicFunctionsCloser)Raise(err error) {
  defer recover()

  if b.Check() {
    (*b.ErrorHandler)(err)
  }
}
