package core

import (
  "fmt"
  "os/exec"
  "bufio"
  "strings"
  "strconv"
  "context"
  "errors"
  "io"
  "bytes"

  "github.com/ipfs/go-log"
)

var (
  InterfaceHeader = "Interface"
  InterfaceLogger = log.Logger(InterfaceHeader)

  HeaderNotUnderstood = errors.New("Header not understood")
  CommandNotUnderstood = errors.New("Command not understood")
  //NotMatserComm = errors.New("Not the MasterComm")
  NotEnoughFields = errors.New("Not enough field")
  EmptyString = errors.New("Received an empty string")

  nilMessageHandler = func(int, string) {}
  nilRequestHandler = func(int) {}
  nilLogger = func(string) {}
  nilResetHandler = func(int) {}

  InterfaceLogHeader = "Log"
  InterfaceSendHeader = "Send"
  InterfaceResetHeader = "Reset"
  InterfaceRequestHeader = "Req"

  LogSubFormat = "%s %d/%d"
)

func NewNewLogger(quiet bool) func(string, int, int) (func(string), error) {
  return func(file string, n int, i int) (func(string), error){
    if i != 0 && quiet {
      return func(string) {}, nil
    }

    var logger log.StandardLogger
    if i == 0 {
      logger = log.Logger(file)
    } else {
      logger = log.Logger(fmt.Sprintf(LogSubFormat, file, i, n))
    }

    return func(str string) {
      logger.Info(str)
    }, nil
  }
}

func NewInterface(ctx context.Context, file string, n int, i int, args ...string) (Interface, error) {
  if checkContextDebug(ctx, InterfaceHeader) { //--------------------------
    InterfaceLogger.Debugf("Starting an interface for file %q, %d/%d with params %s", file, i, n, args) //--------------------------
  } //--------------------------

  cmdArgs := append([]string{file + "/run.py", fmt.Sprint(n), fmt.Sprint(i)}, args...)
  inter := StdInterface {
    Ctx: ctx,
    Idx: i,
    Cmd: exec.CommandContext(ctx, "python3", cmdArgs...),
    MessageHandler: &nilMessageHandler,
    RequestHandler: &nilRequestHandler,
    ResetHandler: &nilResetHandler,
    Logger: &nilLogger,
    Standard: NewStandardInterface(InterfaceHeader),
  }

  return &inter, nil
}

type StdInterface struct {
  Ctx context.Context
  Stdin io.Writer
  MessageHandler *func(int, string)
  RequestHandler *func(int)
  ResetHandler *func(int)
  Logger *func(string)
  Idx int
  Cmd *exec.Cmd
  Standard standardFunctionsCloser
}

func (s *StdInterface)Start() {
  defer func() {
    if err := recover(); err != nil {
      s.Raise(err.(error))
    }
  }()

  var err error

  s.Stdin, err = s.Cmd.StdinPipe()
	if err != nil {
    s.Raise(err)
    return
	}

  stdout, err := s.Cmd.StdoutPipe()
	if err != nil {
    s.Raise(err)
    return
	}

  go func() {
    var errorBuffer bytes.Buffer
    s.Cmd.Stderr = &errorBuffer

    err := s.Cmd.Run()

    strError := errorBuffer.String()
    if strError != "" {
      if strError[len(strError) - 1] == '\n' {
        strError = strError[:len(strError) - 1]
      }
      s.Raise(errors.New(strError))
    }

    if err != nil {
      s.Raise(err)
    }

    s.Close()
  }()

  scanner := bufio.NewScanner(stdout)
  go func(){
    defer func() {
      if err := recover(); err != nil {
        s.Raise(err.(error))
      }
    }()

    for s.Check() && scanner.Scan() {
      splitted := strings.Split(scanner.Text(), ",")
      if !s.Check() { //can happen after a long wait
        break
      }

      if checkContextDebug(s.Ctx, InterfaceHeader) { //--------------------------
        InterfaceLogger.Debugf("Received %q", strings.Join(splitted, ",")) //--------------------------
      } //--------------------------

      switch splitted[0] {
      default:
        s.Raise(HeaderNotUnderstood)

      case InterfaceRequestHeader:
        if len(splitted) < 2 {
          s.Raise(NotEnoughFields)
        }

        idx, err := strconv.Atoi(splitted[1])
        if err != nil {
          s.Raise(err)
          continue
        }

        (*s.RequestHandler)(idx)

      case InterfaceResetHeader:
        if len(splitted) < 2 {
          s.Raise(NotEnoughFields)
        }

        idx, err := strconv.Atoi(splitted[1])
        if err != nil {
          s.Raise(err)
          continue
        }

        (*s.ResetHandler)(idx)

      case InterfaceLogHeader:
        if len(splitted) < 2 {
          s.Raise(NotEnoughFields)
          continue
        }

        (*s.Logger)(strings.Join(splitted[1:], ","))

      case InterfaceSendHeader:
        if len(splitted) < 3 {
          s.Raise(NotEnoughFields)
          continue
        }

        idx, err := strconv.Atoi(splitted[1])
        if err != nil {
          s.Raise(err)
          continue
        }

        (*s.MessageHandler)(idx, strings.Join(splitted[2:], ","))
      }
    }

    //we don't care about error at the end of here
  }()
}

func (s *StdInterface)Close() error {
  return s.Standard.Close()
}

func (s *StdInterface)SetErrorHandler(handler func(error)) {
  s.Standard.SetErrorHandler(handler)
}

func (s *StdInterface)SetCloseHandler(handler func()) {
  s.Standard.SetCloseHandler(handler)
}

func (s *StdInterface)Raise(err error) {
  s.Standard.Raise(err)
}

func (s *StdInterface)Check() bool {
  return s.Standard.Check()
}

func (s *StdInterface)SetMessageHandler(handler func(int, string)) {
  s.MessageHandler = &handler
}

func (s *StdInterface)SetLogger(handler func(string)) {
  s.Logger = &handler
}

func (s *StdInterface)SetRequestHandler(handler func(int)) {
  s.RequestHandler = &handler
}

func (s *StdInterface)SetResetHandler(handler func(int)) {
  s.ResetHandler = &handler
}

func (s *StdInterface)Push(msg string) error {
  defer func() {
    if err := recover(); err != nil {
      s.Raise(err.(error))
    }
  }()

  if !s.Check() {
    return errors.New("Interface closed")
  }
  fmt.Fprintln(s.Stdin, msg)
  return nil
}
