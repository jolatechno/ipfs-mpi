package core

import (
  "sync"
  "fmt"
)

var (
  nilEndHandler = func() {}
  nilErrorHandler = func(err error) {}
)

func NewHeadedError(err error, header string) error {
  if err == nil {
    return nil
  }

  errH, ok := err.(*HeadedError)
  if ok {
    return errH
  }

  return &HeadedError {
    Err: err,
    Header: header,
  }
}

type HeadedError struct {
  Err error
  Header string
}

func (err *HeadedError)Error() string {
  return fmt.Sprintf("[%s] %s", err.Header, err.Err.Error())
}

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
