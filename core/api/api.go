package api

import (
  "bufio"
  "net"
  "errors"
  "strings"
  "fmt"
  "strconv"

  "github.com/jolatechno/ipfs-mpi/core/messagestore"
)

type Key struct {
  File string
  Pid int
}

type Api struct {
  Port int
  Store map[Key] *message.MessageStore
  Handler *message.Handler
}

func NewApi(port int, handler *message.Handler) (*Api, error) {
  l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
  if err != nil {
    return nil, err
  }

  a := Api{
    Port:l.Addr().(*net.TCPAddr).Port,
    Handler:handler,
    Store:make(map[Key] *message.MessageStore),
  }

  go func(){
    for {
      c, err := l.Accept()
      if err != nil {
        continue
      }

      reader := bufio.NewReader(c)
      str, err := reader.ReadString('\n')
      if err != nil {
        continue
      }

      splitted := strings.Split(str, ",")
      if len(splitted) != 2 {
        continue
      }

      pid, err := strconv.Atoi(splitted[0])
      if err != nil {
        continue
      }

      k := Key{ File:splitted[1], Pid:pid }
      a.Store[k] = a.Handler.MessageStore(func(str string) error{
        fmt.Fprintf(c, str)
        return nil
      })

      go func(){
        for {
          msg, err := reader.ReadString('\n')
          if err != nil {
            delete(a.Store, k)
            return
          }

          err = a.Store[k].Manage(msg)
          if err != nil {
            delete(a.Store, k)
            return
          }
        }
      }()
    }
  }()

  return &a, nil
}

func (a *Api)Push(msg message.Message) error{
  f, ok := a.Store[Key{ File:msg.File, Pid:msg.Pid }]
  if !ok {
    return errors.New("no such pid")
  }

  f.Add(msg)
  return nil
}
