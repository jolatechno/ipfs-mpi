package core

import (
  "sync"
  "fmt"
)

var (
  nilEndHandler = func() {}
  nilErrorHandler = func(err error) {}

  ErrorFormat = "\033[1m[%s]\033[0m \033[31m%s\033[0m"
  AlertFormat = "\033[1m[%s]\033[0m \033[33m%s\033[0m"
)

func NewHeadedError(err error, header string) error {
  if err == nil {
    return nil
  }

  errH, ok := err.(*HeadedError)
  if ok {
    if errH.Header == "" {
      errH.Header = header
    }

    return errH
  }

  return &HeadedError {
    Panic: true,
    Err: err,
    Header: header,
  }
}

func SetNonPanic(err error) error {
  if err == nil {
    return nil
  }

  errH, ok := err.(*HeadedError)
  if ok {
    return &HeadedError {
      Panic: false,
      Err: errH.Err,
      Header: errH.Header,
    }
  }

  return &HeadedError {
    Panic: false,
    Err: err,
  }
}

func IsPanic(err error) bool {
  if err == nil {
    return false
  }

  errH, ok := err.(*HeadedError)
  if ok {
    return errH.Panic
  }

  return true
}

type HeadedError struct {
  Panic bool
  Err error
  Header string
}

func (err *HeadedError)Error() string {
  if err.Panic {
    return fmt.Sprintf(ErrorFormat, err.Header, err.Err.Error())
  }
  return fmt.Sprintf(AlertFormat, err.Header, err.Err.Error())
}

func NewStandardInterface(header string) standardFunctionsCloser {
  return &BasicFunctionsCloser {
    Header: header,
    EndHandler: &nilEndHandler,
    ErrorHandler: &nilErrorHandler,
  }
}

type BasicFunctionsCloser struct {
  Header string
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
  b.Mutex.Lock()
  defer func() {
    b.Mutex.Unlock()
    recover()
  }()

  if !b.Ended {
    (*b.EndHandler)()

    b.Ended = true
  }

  return nil
}

func (b *BasicFunctionsCloser)Raise(err error) {
  defer recover()

  if b.Check() {
    errH := NewHeadedError(err, b.Header)
    (*b.ErrorHandler)(errH)
  }
}
