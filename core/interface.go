package core

import (
  "fmt"
  "os/exec"
  "bufio"
  "strings"
  "strconv"
)

func NewInterface(file string, n int, i int, args ...string) (Interface, error) {
  inter := StdInterface {
    Ended: false,
    EndChan: make(chan bool),
    InChan: make(chan string),
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
        inter.Close()
      }

      splitted := strings.Split(str, ",")
      if len(splitted) != 2 {
        inter.Close()
      }

      if splitted[0] == "Req" {
        idx, err := strconv.Atoi(splitted[1][:len(splitted[1]) - 1])
        if err != nil {
          inter.Close()
        }

        inter.RequestChan <- idx
        fmt.Fprint(stdin, <- inter.InChan)
      } else if splitted[0] == "Log" && i == 0 {
        fmt.Print(splitted[1])
      } else {
        idx, err := strconv.Atoi(splitted[0])
        if err != nil {
          inter.Close()
        }

        inter.OutChan <- Message {
          To: idx,
          Content: splitted[1],
        }
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
  s.InChan <- msg
  return nil
}
