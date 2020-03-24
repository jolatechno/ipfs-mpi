package mpi

import (
  "github.com/jolatechno/ipfs-mpi/core/ipfs-interface"

  "github.com/coreos/go-semver/semver"
)

type Message struct {
  from string
  to string
}

type Handler func(Message) ([]Message, error)

func (m *Message)String() string {
  //convert a Message to string
  return ""
}

func FromString(msg string) (Message, error){
  //read a Message from string
  return Message{}, nil
}

func Load(f file.File) (Handler, error) {
  //Loading the file
  return nil, nil
}
