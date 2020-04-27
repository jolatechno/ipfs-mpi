package core

import (
  "sync"
)

var (
  nilEndHandler = func() {}
  nilErrorHandler = func(err error) {}
)

func NewStandardInterface(additionalHandler ...func() error) standardFunctionsCloser {
  return &BasicFunctionsCloser {
    EndHandler: &nilEndHandler,
    ErrorHandler: &nilErrorHandler,
    AdditionalHandler: additionalHandler,
  }
}

type BasicFunctionsCloser struct {
  Mutex sync.Mutex
  Ended bool
  EndHandler *func()
  AdditionalHandler []func() error
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
  b.Mutex.Lock()
  defer func() {
    b.Mutex.Unlock()
    recover()
  }()

  if !b.Ended {
    for _, handler := range b.AdditionalHandler {
      err := handler()
      if err != nil {
        return err
      }
    }

    (*b.EndHandler)()

    b.Ended = true
  }

  return nil
}

func (b *BasicFunctionsCloser)Raise(err error) {
  defer recover()

  if b.Check() && err != nil {
    (*b.ErrorHandler)(err)
  }
}
