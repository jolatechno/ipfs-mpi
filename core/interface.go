package core

import (
  "log"
  "fmt"
  "os/exec"
  "bufio"
  "strings"
  "strconv"
  "errors"
  "io"
)

var (
  nilMessageHandler = func(int, string) {}
  nilRequestHandler = func(int) {}
  nilResetHandler = func(int) {}

  logFormat = "\033[33m%s\033[0m\n"
)

func NewInterface(file string, n int, i int, args ...string) (Interface, error) {
  cmdArgs := append([]string{file + "/run.py", fmt.Sprint(n), fmt.Sprint(i)}, args...)
  inter := StdInterface {
    Idx: i,
    Cmd: exec.Command("python3", cmdArgs...),
    MessageHandler: &nilMessageHandler,
    RequestHandler: &nilRequestHandler,
    ResetHandler: &nilResetHandler,
    Standard: NewStandardInterface(),
  }

  return &inter, nil
}

type StdInterface struct {
  Stdin io.Writer
  MessageHandler *func(int, string)
  RequestHandler *func(int)
  ResetHandler *func(int)
  Idx int
  Cmd *exec.Cmd
  Standard standardFunctionsCloser
}

func (s *StdInterface)Start() {
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

  stderr, err := s.Cmd.StderrPipe()
	if err != nil {
    s.Raise(err)
    return
	}

  err = s.Cmd.Start()
  if err != nil {
    s.Raise(err)
    return
  }

  go func() {
    err := s.Cmd.Wait()
    if err != nil {
      s.Raise(err)
    }

    s.Close()
  }()

  errReader := bufio.NewReader(stderr)
  go func() {
    for {
      strErr, err := errReader.ReadString('\n')
      if err != nil {
        s.Close()
        return
      }
      if strErr != "" {
        s.Raise(errors.New(strErr))
        continue
      }
    }
  }()

  reader := bufio.NewReader(stdout)
  go func(){
    for s.Check() {
      str, err := reader.ReadString('\n')
      if str == "" && err == nil {
        err = errors.New("Received an empty string")
      }
      if err != nil {
        s.Close()
        return
      }

      splitted := strings.Split(str, ",")

      switch splitted[0] {
      default:
        s.Raise(errors.New("Not understood"))

      case "Req":
        if len(splitted) != 2 {
          s.Raise(errors.New("Not enough field"))
        }

        idx, err := strconv.Atoi(splitted[1][:len(splitted[1]) - 1])
        if err != nil {
          s.Raise(err)
          continue
        }

        (*s.RequestHandler)(idx)

      case "Reset":
        if s.Idx != 0 {
          s.Raise(errors.New("Can't reset if isn't the MasterComm"))
        }

        if len(splitted) != 2 {
          s.Raise(errors.New("Not enough field"))
        }

        idx, err := strconv.Atoi(splitted[1][:len(splitted[1]) - 1])
        if err != nil {
          s.Raise(err)
          continue
        }

        (*s.ResetHandler)(idx)

      case "Log":
        if s.Idx != 0 {
          s.Raise(errors.New("Can't log if isn't the MasterComm"))
        }

        if len(splitted) < 2 {
          s.Raise(errors.New("Not enough field"))
          continue
        }

        last_idx := len(splitted) - 1
        last_len := len(splitted[last_idx]) - 1
        splitted[last_idx] = splitted[last_idx][:last_len]

        log.Printf(logFormat, strings.Join(splitted[1:], ","))

      case "Send":
        if len(splitted) < 3 {
          s.Raise(errors.New("Not enough field"))
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

func (s *StdInterface)SetRequestHandler(handler func(int)) {
  s.RequestHandler = &handler
}

func (s *StdInterface)SetResetHandler(handler func(int)) {
  s.ResetHandler = &handler
}

func (s *StdInterface)Push(msg string) error {
  if !s.Check() {
    return errors.New("Interface closed")
  }
  fmt.Fprint(s.Stdin, msg)
  return nil
}
