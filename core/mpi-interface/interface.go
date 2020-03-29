package mpi

import (
  "os/exec"
  "io"
  "bufio"
  "fmt"

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
  cmd := exec.Command("python3", d.Path + "/run.py", k.Origin, fmt.Sprint(k.Pid))
  stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

  stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

  if err = cmd.Start(); err != nil {
    return err
  }

  defer stdin.Close()
  defer stdout.Close()

  reader := bufio.NewReader(stdout)

  d.Store[k] = d.Handler.MessageStore(func(str string) error{
    io.WriteString(stdin, str + "\n")
    return nil
  })

  go func(){
    for {
      msg, err := reader.ReadString('\n')
      if err != nil {
        delete(d.Store, k)
        return
      }

      err = d.Store[k].Manage(msg[:len(msg) - 1])
      if err != nil {
        delete(d.Store, k)
        return
      }
    }
  }()

  go func(){
    cmd.Wait()
    delete(d.Store, k)
  }()

  return nil
}

func Install(path string) error {
  _, err := exec.Command("python3", path + "/init.py").Output()
	return err
}
