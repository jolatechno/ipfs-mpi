package mpi

import (
  "os/exec"
)

type Key struct {
  Origin string
  File string
  Pid int
}

type DaemonStore struct {
  Store map[Key] *MessageStore
  Sender func(Message) error
  Path string
}

func NewDaemonStore(path string, Sender func(Message) error) DaemonStore {
  return DaemonStore{
    Store: make(map[Key] *MessageStore),
    Sender:Sender,
    Path:path,
  }
}

func (d *DaemonStore)Push(msg Message) {
  k := Key{ Origin:msg.Origin, Pid:msg.Pid }
  if _, ok := d.Store[k]; !ok {
    d.load(k)
  }

  d.Store[k].Add(msg)
}

func (d *DaemonStore)load(k Key) {
  s := make(MessageStore)
  d.Store[k] = &s
  // Load the interpretor

  go func(){

  }()
}

func Install(path string) error {
  _, err := exec.Command("python3", path + "/init.py").Output()
	return err
}
