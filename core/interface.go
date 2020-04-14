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
  inter := StdInterface {
    InChan: make(chan string),
    OutChan: make(chan Message),
    RequestChan: make(chan int),
    Standard: NewStandardInterface(),
  }

  cmdArgs := append([]string{file + "/run.py", fmt.Sprint(n), fmt.Sprint(i)}, args...)
  cmd := exec.Command("python3", cmdArgs...)

  stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

  stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

  stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

  err = cmd.Start()
  if err != nil {
    return nil, err
  }

  go func() {
    err := cmd.Wait()
    if err != nil {
      inter.Standard.Push(err)
    }

    if inter.Check() {
      inter.Close()
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
        if inter.Check() {
          inter.Standard.Push(errors.New(strErr))
          inter.Close()
        }
        return
      }
    }
  }()

  reader := bufio.NewReader(stdout)
  go func(){
    for inter.Check() {
      str, err := reader.ReadString('\n')
      if str == "" && err == nil {
        err = errors.New("Received an empty string")
      }
      if err != nil {
        time.Sleep(SafeWait)
        if inter.Check() {
          inter.Standard.Push(err)
          inter.Close()
        }
        return
      }

      splitted := strings.Split(str, ",")

      if splitted[0] == "Req" {
        if len(splitted) != 2 {
          inter.Standard.Push(errors.New("Not enough field"))
          inter.Close()
          return
        }

        idx, err := strconv.Atoi(splitted[1][:len(splitted[1]) - 1])
        if err != nil {
          inter.Standard.Push(err)
          inter.Close()
          return
        }

        inter.RequestChan <- idx
        go fmt.Fprint(stdin, <- inter.InChan)

      } else if splitted[0] == "Log" && i == 0 {
        if len(splitted) < 2 {
          inter.Standard.Push(errors.New("Not enough field"))
          inter.Close()
          return
        }

        log.Print(strings.Join(splitted[1:], ","))

      } else if splitted[0] == "Send" {
        if len(splitted) < 3 {
          inter.Standard.Push(errors.New("Not enough field"))
          inter.Close()
          return
        }

        idx, err := strconv.Atoi(splitted[1])
        if err != nil {
          inter.Standard.Push(err)
          inter.Close()
          return
        }

        go func() {
          inter.OutChan <- Message {
            To: idx,
            Content: strings.Join(splitted[2:], ","),
          }
        }()
      } else {
        inter.Standard.Push(errors.New("Not understood"))
        inter.Close()
        return
      }
    }
  }()

  return &inter, nil
}

type StdInterface struct {
  InChan chan string
  OutChan chan Message
  RequestChan chan int
  Standard BasicFunctionsCloser
}

func (s *StdInterface)Close() error {
  if s.Check() {
    go func() {
      s.OutChan <- Message {
        To: -1,
      }
    }()

    go func() {
      s.RequestChan <- -1
    }()

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
  return s.OutChan
}

func (s *StdInterface)Request() chan int {
  return s.RequestChan
}

func (s *StdInterface)Push(msg string) error {
  if !s.Check() {
    return errors.New("Interface closed")
  }
  s.InChan <- msg
  return nil
}
