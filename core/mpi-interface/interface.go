package mpi

import (
  "strings"
  "errors"
  "os/exec"
  "fmt"
)

type Message struct {
  Pid int
  From string
  To string
  Data []byte
}

type Handler func(Message) ([]Message, error)

func (m *Message)String() string {
  return fmt.Sprintf("%d,%x,%x,%x", m.Pid, m.From, m.To, m.Data)
}

func FromString(msg string) (*Message, error) {
  m := Message{}
  n, err := fmt.Sscanf(msg, "%d,%x,%x,%x", &m.Pid, &m.From, &m.To, &m.Data)
  if err != nil {
    return nil, err
  }
  if n != 4 {
    return nil, errors.New("message dosen't have the write number of field")
  }

  return &m, err
}

func Load(path string, responder func(Message) error) Handler {
  return Handler(func(msg Message) ([]Message, error){
    msgs := []Message{}

    if msg.Pid == -1 {

      fmt.Println("mpi handler 0") //----------------------------------------------------------------------------

      out, err := exec.Command("python3", path + "/run.py", msg.String()).Output()
      if err != nil{
        return msgs, nil
      }

      out_str := string(out)
      if out_str[len(out_str) - 2:] == "\n" {
        out_str = out_str[:len(out_str) - 2]
      }

      fmt.Println("mpi handler 1, ", out_str) //----------------------------------------------------------------------------

      strs := strings.Split(out_str, ";")

      fmt.Println("mpi handler 2, ", len(out_str), " , ", out_str) //----------------------------------------------------------------------------

      for _, str := range strs {
        m, err := FromString(str)
        if err != nil {
          return msgs, err
        }
        msgs = append(msgs, *m)
      }

      return msgs, nil
    }

    return msgs, responder(msg)
  })
}

func Install(path string) error {
  _, err := exec.Command("python3", path + "/init.py").Output()
	return err
}
