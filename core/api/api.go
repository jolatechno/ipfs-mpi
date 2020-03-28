package api

import (
  "bufio"
  "net"
  "errors"
  "strings"
  "fmt"
  "strconv"

  "github.com/jolatechno/ipfs-mpi/core/mpi-interface"
)

type handler struct {
  handler *func(mpi.Message) error
  list *func() (string, []string)
}

type Key struct {
  File string
  Pid int
}

type Api struct {
  Port int
  Store *map[Key] *mpi.MessageStore
  handlers *map[string] handler
}

func NewApi(port int) (*Api, error) {
  handlers := make(map[string] handler)
  store := make(map[Key] *mpi.MessageStore)

  l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
  if err != nil {
    return nil, err
  }

  a := Api{
    Port:l.Addr().(*net.TCPAddr).Port,
    handlers:&handlers,
    Store:&store,
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
      m := make(mpi.MessageStore)
      (*a.Store)[k] = &m

      file_handler, ok := (*a.handlers)[splitted[1]]
      if !ok {
        delete(*a.Store, k)
        return
      }
      handle := file_handler.handler

      go func(){
        for {
          msg, err := reader.ReadString('\n')
          if err != nil {
            delete(*a.Store, k)
            return
          }

          var header, content string
          n, err := fmt.Sscanf(msg, "%q;%q\n", &header, &content)
          if err != nil || n != 2 {
            delete(*a.Store, k)
            return
          }

          if header == "List" {
            handler, ok := (*a.handlers)[content]
            if !ok {
              delete(*a.Store, k)
              return
            }

            fmt.Fprintf(c, "\"List\";%q\n", ListToString((*handler.list)()))
            continue

          } else if header == "Req" {
            fmt.Fprintf(c, "\"Msg\";%q\n", (*a.Store)[k].Read(content))
            continue

          } else if header == "Send" {
            m, err := mpi.FromString(content)
            if err != nil {
              delete(*a.Store, k)
              return
            }

            err = (*handle)(*m)
            if err != nil {
              delete(*a.Store, k)
              return
            }

          }
        }
      }()
    }
  }()

  return &a, nil
}

func (a *Api)AddHandler(key string, handle *func(mpi.Message) error, list *func() (string, []string)) {
  (*a.handlers)[key] = handler{ handler:handle, list:list }
}

func (a *Api)Push(msg mpi.Message) error{
  f, ok := (*a.Store)[Key{ File:msg.File, Pid:msg.Pid }]
  if !ok {
    return errors.New("no such pid")
  }

  f.Add(msg)
  return nil
}
