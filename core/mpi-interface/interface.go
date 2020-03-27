package mpi

import (
  b64 "encoding/base64"
  "strconv"
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
  return fmt.Sprintf("%d,%s,%s,%s", m.Pid, m.From, m.To, b64.StdEncoding.EncodeToString(m.Data))
}

func FromString(msg string) (*Message, error) {
  splitted := strings.Split(msg, ",")
  if len(splitted) != 4 {
    return nil, errors.New("message dosen't have the write number of field")
  }

  pid, err := strconv.Atoi(splitted[0])
  if err != nil {
    return nil, err
  }

  Data, err := b64.StdEncoding.DecodeString(splitted[3])
  if err != nil {
    return nil, err
  }

  return &Message{ Pid:pid, From:splitted[1], To:splitted[1], Data:Data }, err
}

func Load(path string, responder func(Message) error) Handler {
  return Handler(func(msg Message) ([]Message, error){
    msgs := []Message{}

    if msg.Pid == -1 {
      out, err := exec.Command("python3", path + "/run.py", msg.String()).Output()
      if err != nil{
        return msgs, nil
      }

      out_str := string(out)
      if out_str[len(out_str) - 1:] == "\n" {
        out_str = out_str[:len(out_str) - 1]
      }

      strs := strings.Split(out_str, ";")

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
