package core

import (
  "errors"
  "fmt"
  "os/exec"
  "bufio"
  "strings"
  "strconv"
)

func NewInterface(file string, n int) (Interface, error) {
  inter := StdInterface {
    Ended: false,
    InChan: make(chan string),
    OutChan: make(chan Message),
    RequestChan: make(chan int),
  }

  cmd := exec.Command("python3", file + "/run.py", "0", fmt.Sprint(n))

  stdin, err := cmd.StdinPipe()
	if err != nil {
		return &inter, err
	}

  stdout, err := cmd.StdoutPipe()
	if err != nil {
		return &inter, err
	}

  err = cmd.Start()
  if err != nil {
    return &inter, err
  }

  reader := bufio.NewReader(stdout)

  go func(){
    for !inter.Ended {
      str, err := reader.ReadString('\n')
      if err != nil {
        inter.Close()
      }

      splitted := strings.Split(str, ",")
      if len(splitted) != 2 {
        inter.Close()
      }

      if splitted[0] == "Req" {
        idx, err := strconv.Atoi(splitted[1])
        if err != nil {
          inter.Close()
        }

        inter.RequestChan <- idx
        fmt.Fprint(stdin, <- inter.InChan)
      } else {
        idx, err := strconv.Atoi(splitted[0])
        if err != nil {
          inter.Close()
        }

        inter.OutChan <- Message{
          To: idx,
          Content: splitted[1],
        }
      }
    }
  }()

  return &inter, errors.New("Not yet implemented")
}

type StdInterface struct {
  Ended bool
  InChan chan string
  OutChan chan Message
  RequestChan chan int
}

func (s *StdInterface)Close() {
  s.Ended = true
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