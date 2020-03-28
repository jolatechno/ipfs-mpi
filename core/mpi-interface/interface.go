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
  Handler *handler
  Path string
}

func NewDaemonStore(path string, send *func(Message) error, list *func(string) (string, []string)) DaemonStore {
  return DaemonStore{
    Store: make(map[Key] *MessageStore),
    Handler:&handler{
      list:list,
      send:send,
    },
    Path:path,
  }
}

func (d *DaemonStore)Push(msg Message) error {
  k := Key{ Origin:msg.Origin, Pid:msg.Pid }
  if _, ok := d.Store[k]; !ok {
    if err := d.load(k); err != nil {
      return err
    }
  }

  d.Store[k].Add(msg)
  return nil
}

func (d *DaemonStore)load(k Key) error {
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
