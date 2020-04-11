package core

import (
  "log"
  "fmt"
  "os/exec"
  "bufio"
  "strings"
  "strconv"
  "errors"
)

func NewInterface(file string, n int, i int, args ...string) (Interface, error) {
  inter := StdInterface {
    Ended: false,
    EndChan: make(chan bool),
    InChan: make(chan string),
    Error: make(chan error),
    OutChan: make(chan Message),
    RequestChan: make(chan int),
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

  err = cmd.Start()
  if err != nil {
    return nil, err
  }

  reader := bufio.NewReader(stdout)

  go func(){
    for inter.Check() {
      str, err := reader.ReadString('\n')
      if err != nil {
        inter.Error <- err
        inter.Close()
      }

      splitted := strings.Split(str, ",")

      if splitted[0] == "Req" {
        if len(splitted) != 2 {
          inter.Error <- errors.New("Not enough field")
          inter.Close()
        }

        idx, err := strconv.Atoi(splitted[1][:len(splitted[1]) - 1])
        if err != nil {
          inter.Error <- err
          inter.Close()
        }

        inter.RequestChan <- idx
        go fmt.Fprint(stdin, <- inter.InChan)

      } else if splitted[0] == "Log" && i == 0 {
        if len(splitted) < 2 {
          inter.Error <- errors.New("Not enough field")
          inter.Close()
        }

        log.Print(strings.Join(splitted[1:], ","))

      } else if splitted[0] == "Send" {
        if len(splitted) < 3 {
          inter.Error <- errors.New("Not enough field")
          inter.Close()
        }

        idx, err := strconv.Atoi(splitted[1])
        if err != nil {
          inter.Error <- err
          inter.Close()
        }

        go func() {
          inter.OutChan <- Message {
            To: idx,
            Content: strings.Join(splitted[2:], ","),
          }
        }()
      } else {
        inter.Error <- errors.New("Not understood")
        inter.Close()
      }
    }
  }()

  return &inter, nil
}

type StdInterface struct {
  Ended bool
  EndChan chan bool
  InChan chan string
  OutChan chan Message
  Error chan error
  RequestChan chan int
}

func (s *StdInterface)Close() error {
  s.EndChan <- true
  s.Ended = true
  return nil
}

func (s *StdInterface)CloseChan() chan bool {
  return s.EndChan
}

func (s *StdInterface)ErrorChan() chan error {
  return s.Error
}

func (s *StdInterface)Check() bool {
  return !s.Ended
}

func (s *StdInterface)Message() chan Message {
  return s.OutChan
}

func (s *StdInterface)Request() chan int {
  return s.RequestChan
}

func (s *StdInterface)Push(msg string) error {
  if s.Ended {
    return errors.New("Interface closed")
  }
  s.InChan <- msg
  return nil
}
