package core

import (
  "log"
  "fmt"
  "os/exec"
  "bufio"
  "strings"
  "strconv"
  "errors"
  "time"
)

var (
  SafeWait = 10 * time.Millisecond
)

func NewInterface(file string, n int, i int, args ...string) (Interface, error) {
  cmdArgs := append([]string{file + "/run.py", fmt.Sprint(n), fmt.Sprint(i)}, args...)
  inter := StdInterface {
    Idx: i,
    Cmd: exec.Command("python3", cmdArgs...),
    InChan: NewChannelString(),
    OutChan: NewChannelMessage(),
    RequestChan: NewChannelInt(),
    Standard: NewStandardInterface(),
  }

  return &inter, nil
}

type StdInterface struct {
  Idx int
  Cmd *exec.Cmd
  InChan *SafeChannelString
  OutChan *SafeChannelMessage
  RequestChan *SafeChannelInt
  Standard BasicFunctionsCloser
}

func (s *StdInterface)Start() {
  stdin, err := s.Cmd.StdinPipe()
	if err != nil {
    s.Standard.Push(err)
    s.Close()
    return
	}

  stdout, err := s.Cmd.StdoutPipe()
	if err != nil {
    s.Standard.Push(err)
    s.Close()
    return
	}

  stderr, err := s.Cmd.StderrPipe()
	if err != nil {
    s.Standard.Push(err)
    s.Close()
    return
	}

  err = s.Cmd.Start()
  if err != nil {
    s.Standard.Push(err)
    s.Close()
    return
  }

  go func() {
    err := s.Cmd.Wait()
    if err != nil {
      s.Standard.Push(err)
    }

    if s.Check() {
      s.Close()
    }
  }()

  errReader := bufio.NewReader(stderr)
  go func() {
    for {
      strErr, err := errReader.ReadString('\n')
      if err != nil {
        return
      }
      if strErr != "" {
        time.Sleep(SafeWait)
        if s.Check() {
          s.Standard.Push(errors.New(strErr))
          s.Close()
        }
        return
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
        time.Sleep(SafeWait)
        if s.Check() {
          s.Standard.Push(err)
          s.Close()
        }
        return
      }

      splitted := strings.Split(str, ",")

      if splitted[0] == "Req" {
        if len(splitted) != 2 {
          s.Standard.Push(errors.New("Not enough field"))
          s.Close()
          return
        }

        idx, err := strconv.Atoi(splitted[1][:len(splitted[1]) - 1])
        if err != nil {
          s.Standard.Push(err)
          s.Close()
          return
        }

        s.RequestChan.Send(idx)
        go fmt.Fprint(stdin, <- s.InChan.C)

      } else if splitted[0] == "Log" && s.Idx == 0 {
        if len(splitted) < 2 {
          s.Standard.Push(errors.New("Not enough field"))
          s.Close()
          return
        }

        log.Print(strings.Join(splitted[1:], ","))

      } else if splitted[0] == "Send" {
        if len(splitted) < 3 {
          s.Standard.Push(errors.New("Not enough field"))
          s.Close()
          return
        }

        idx, err := strconv.Atoi(splitted[1])
        if err != nil {
          s.Standard.Push(err)
          s.Close()
          return
        }

        go func() {
          s.OutChan.Send(Message {
            To: idx,
            Content: strings.Join(splitted[2:], ","),
          })
        }()
      } else {
        s.Standard.Push(errors.New("Not understood"))
        s.Close()
        return
      }
    }
  }()
}

func (s *StdInterface)Close() error {
  if s.Check() {
    go s.OutChan.SafeClose(false)
    go s.RequestChan.SafeClose(false)
    go s.InChan.SafeClose(true)

    s.Standard.Close()
  }

  return nil
}

func (s *StdInterface)CloseChan() chan bool {
  return s.Standard.CloseChan()
}

func (s *StdInterface)ErrorChan() chan error {
  return s.Standard.ErrorChan()
}

func (s *StdInterface)Check() bool {
  return s.Standard.Check()
}

func (s *StdInterface)Message() chan Message {
  return s.OutChan.C
}

func (s *StdInterface)Request() chan int {
  return s.RequestChan.C
}

func (s *StdInterface)Push(msg string) error {
  if !s.Check() {
    return errors.New("Interface closed")
  }
  s.InChan.Send(msg)
  return nil
}
