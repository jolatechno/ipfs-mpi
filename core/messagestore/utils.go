package message

import (
  b64 "encoding/base64"
  "strconv"
  "strings"
  "errors"
  "fmt"
  "os/exec"
)

type Message struct {
  Pid int
  File string
  Origin string
  From string
  To string
  Data []byte
}

func (m *Message)String() string {
  return fmt.Sprintf("%d,%s,%s,%s,%s,%s", m.Pid, m.File, m.Origin, m.From, m.To, b64.StdEncoding.EncodeToString(m.Data))
}

func FromString(msg string) (*Message, error) {
  splitted := strings.Split(msg, ",")
  if len(splitted) != 6 {
    return nil, errors.New("message dosen't have the right number of field")
  }

  pid, err := strconv.Atoi(splitted[0])
  if err != nil {
    return nil, err
  }

  Data, err := b64.StdEncoding.DecodeString(splitted[5])
  if err != nil {
    return nil, err
  }

  return &Message{ Pid:pid, File:splitted[1], Origin:splitted[2], From:splitted[3], To:splitted[4], Data:Data }, nil
}

func ListToString(host string, peers []string) string {
  return fmt.Sprintf("%s,%s", host, strings.Join(peers, ","))
}

func ListFromString(str string) (string, []string) {
  splited := strings.Split(str, ",")
  return splited[0], splited[1:]
}

func Install(path string) error {
  _, err := exec.Command("python3", path + "/init.py").Output()
	return err
}
