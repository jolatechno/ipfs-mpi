package mpi

import (
  "strings"
  "errors"
  "os/exec"
  "fmt"
)

type Message struct {
  pid int
  From string
  To string
  Data []byte
}

type Handler func(Message) ([]Message, error)

func (m *Message)String() string {
  return fmt.Sprintf("%d,%s,%s,%x", m.pid, m.From, m.To, m.Data)
}

func FromString(msg string) (*Message, error) {
  m := Message{}
  n, err := fmt.Sscanf(msg, "%d,%s,%s,%x", &m.pid, &m.From, &m.To, &m.Data)
  if n != 4 {
    return nil, errors.New("message dosen't have the write number of field")
  }

  return &m, err
}

func Load(path string) Handler {
  return Handler(func(msg Message) ([]Message, error){
    msgs := []Message{}

    if msg.pid == -1 {
      out, err := exec.Command("python3", path + "run.py", msg.String()).Output()
      if err != nil{
        return msgs, nil
      }

      strs := strings.Split(string(out), ";")
      for _, str := range strs {
        m, err := FromString(str)
        if err != nil {
          return msgs, err
        }
        msgs = append(msgs, *m)
      }
      return msgs, nil
    }

    //message is an answer
    return nil, errors.New("returning argument isn't yet implemented !")
  })
}

func Install(path string) error {
  _, err := exec.Command("python3", path + "init.py").Output()
	return err
}
