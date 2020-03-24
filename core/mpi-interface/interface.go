package mpi

import (
  "strings"
  "errors"
  "os/exec"
)

type Message struct {
  From string
  To string
  Data []byte
}

type Handler func(Message) ([]Message, error)

func (m *Message)String() string {
  return m.From + "," + m.To + "," + string(m.Data)
}

func FromString(msg string) (*Message, error) {
  split := strings.Split(msg, ",")
  if len(split) != 3 {
    return nil, errors.New("Message invalid")
  }
  return &Message{ From:split[0], To:split[1], Data:[]byte(split[2]) }, nil
}

func Load(path string) Handler {
  return Handler(func(msg Message) ([]Message, error){
    msgs := []Message{}

    out, err := exec.Command("python3", path + "run.py", msg.String()).Output()
    if err != nil{
      return msgs, nil
    }

    strs := strings.Split(string(out), ";")
    for _, str := range strs {
      msg, err := FromString(str)
      if err != nil {
        return msgs, err
      }
      msgs = append(msgs, *msg)
    }
    return msgs, nil
  })
}

func Install(path string) error {
  _, err := exec.Command("python3", path + "init.py").Output()
	return err
}
