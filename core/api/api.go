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
  Store *map[Key] *message.MessageStore
  Handler *message.Handler
}

func NewApi(port int, handler *message.Handler) (*Api, error) {
  store := make(map[Key] *message.MessageStore)

  l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
  if err != nil {
    return nil, err
  }

  a := Api{
    Port:l.Addr().(*net.TCPAddr).Port,
    Handler:handler,
    Store:&store,
  }

  go func(){
    for {
      c, err := l.Accept()
      if err != nil {
        continue
      }

      fmt.Println("api new 0") //------------------------------------------------------------------------

      reader := bufio.NewReader(c)
      str, err := reader.ReadString('\n')
      if err != nil {
        continue
      }

      fmt.Println("api new 1") //------------------------------------------------------------------------

      splitted := strings.Split(str[:len(str) - 1], ",")
      if len(splitted) != 2 {
        continue
      }

      fmt.Println("api new 2") //------------------------------------------------------------------------

      pid, err := strconv.Atoi(splitted[0])
      if err != nil {
        continue
      }

      fmt.Println("api new 3") //------------------------------------------------------------------------

      k := Key{ File:splitted[1], Pid:pid }
      store[k] = a.Handler.MessageStore(func(str string) error {
        fmt.Fprintf(c, str + "\n")
        return nil
      })

      fmt.Println("api new 4") //------------------------------------------------------------------------

      go func(){
        for {

          fmt.Println("api go 0") //------------------------------------------------------------------------

          msg, err := reader.ReadString('\n')
          if err != nil {
            delete(store, k)
            return
          }

          fmt.Println("api go 1") //------------------------------------------------------------------------

          err = store[k].Manage(msg[:len(msg) - 1])
          if err != nil {
            delete(store, k)
            return
          }

          fmt.Println("api go 2") //------------------------------------------------------------------------
        }
      }()
    }
  }()

  return &a, nil
}

func (a *Api)Push(msg message.Message) error {
  f, ok := (*a.Store)[Key{ File:msg.File, Pid:msg.Pid }]
  if !ok {
    return errors.New("no such pid")
  }

  f.Add(msg)
  return nil
}
