package mpi

import (
  "os/exec"

  "github.com/jolatechno/ipfs-mpi/core/messagestore"
)

type Key struct {
  Origin string
  File string
  Pid int
}

type DaemonStore struct {
  Store map[Key] *message.MessageStore
  Handler *message.Handler
  Path string
}

func NewDaemonStore(path string, handler *message.Handler) DaemonStore {
  return DaemonStore{
    Store: make(map[Key] *message.MessageStore),
    Handler:handler,
    Path:path,
  }
}

func (d *DaemonStore)Push(msg message.Message) error {
  k := Key{ Origin:msg.Origin, Pid:msg.Pid }
  if _, ok := d.Store[k]; !ok {
    if err := d.Load(k); err != nil {
      return err
    }
  }

  d.Store[k].Add(msg)
  return nil
}

func (d *DaemonStore)Load(k Key) error {
  d.Store[k] = d.Handler.MessageStore(func(string) error{
    //Prgm
    return nil
  })

  /*
  cmd := exec.Command("python3", d.Path + "/run.py")
  stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

  stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

  stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

  if err = cmd.Start(); err != nil {
    return err
  }*/

  go func(){

  }()

  return nil
}

func Install(path string) error {
  _, err := exec.Command("python3", path + "/init.py").Output()
	return err
}
