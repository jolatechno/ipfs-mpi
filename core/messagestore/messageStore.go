package message

import (
  "strings"
  "errors"
  "fmt"
)

type Handler struct {
  Send *func(Message) error
  List *func(string) (string, []string, error)
}

type MessageStore struct {
  Store *map[string] chan []byte
  Handler *Handler
  Sender func(string) error
}

func (h *Handler)MessageStore(Sender func(string) error) *MessageStore {
  store := make(map[string] chan []byte)

  return &MessageStore{
    Store:&store,
    Handler:h,
    Sender:Sender,
  }
}

func (m *MessageStore)Add(msg Message) {
  if _, ok := (*m.Store)[msg.From]; !ok {
    (*m.Store)[msg.From] = make(chan []byte)
  }
  (*m.Store)[msg.From] <- msg.Data
}

func (m *MessageStore)Read(From string) []byte {
  if _, ok := (*m.Store)[From]; !ok {
    (*m.Store)[From] = make(chan []byte)
  }
  return <- (*m.Store)[From]
}

func (m *MessageStore)Manage(msg string) error {
  splitted_msg := strings.Split(msg, ";")
  if len(splitted_msg) != 2 {
    return errors.New("Message dosen't have a clearly defined header and content")
  }

  if splitted_msg[0] == "List" {
    host, peers, err := (*m.Handler.List)(splitted_msg[1])
    if err != nil {
      return err
    }
    
    (*m).Sender(fmt.Sprintf("List;%s\n", ListToString(host, peers)))
    return nil

  } else if splitted_msg[0] == "Req" {
    (*m).Sender(fmt.Sprintf("Msg;%s\n", m.Read(splitted_msg[1])))
    return nil

  } else if splitted_msg[0] == "Send" {
    message, err := FromString(splitted_msg[1])
    if err != nil {
      return err
    }

    err = (*m.Handler.Send)(*message)
    if err != nil {
      return err
    }

    return nil
  }

  return errors.New("Message header not understood")
}
