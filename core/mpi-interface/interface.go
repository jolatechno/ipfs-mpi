package mpi

import (
  "github.com/coreos/go-semver/semver"
)

type Message struct {
  //TODO
}

func (m *Message)String() string {
  //convert a Message to string
  return ""
}

func FromString(msg string) (Message, error){
  //read a Message from string
  return Message{}, nil
}

func Load(addr string, version semver.Version) (func(Message) error, error) {
  //Loading the file
  return nil, nil
}
