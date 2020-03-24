package mpi

import (
  "github.com/coreos/go-semver/semver"
)

type File struct {
  Name string
  Version *semver.Version
}

type Message struct {
  From string
  To string
}

type Handler func(Message) ([]Message, error)

func (m *Message)String() string {
  //convert a Message to string
  return ""
}

func FromString(msg string) (Message, error) {
  //read a Message from string
  return Message{}, nil
}

func Load(f File) (*Handler, error) {
  //Loading the file
  //TODO

  //for now:
  handler := Handler(handle)

  return &handler, nil
}

func Install(f File) error {
  //Install the file
  return nil
}

//for now:
func handle(Message) ([]Message, error) {
  return []Message{}, nil
}
