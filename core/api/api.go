package api

import (
  "bufio"
  "net"
  "errors"
  "fmt"

  "github.com/jolatechno/ipfs-mpi/core/mpi-interface"
)

type handler struct {
  handler *func(mpi.Message) error
  list *func() (string, []string)
}

type Api struct {
  port int
  handlers *map[string] handler
  resp *map[int] func(mpi.Message)
}

func NewApi(port int) (*Api, error){
  handlers := make(map[string] handler)
  resp := make(map[int] func(mpi.Message))

  a := Api{
    port:port,
    handlers:&handlers,
    resp:&resp,
  }

  l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
  if err != nil {
    return nil, err
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

      var pid int
      n, err := fmt.Sscanf(str, "%d\n", &pid)
      if err != nil || n != 1 {
        continue
      }

      (*a.resp)[pid] = func(msg mpi.Message){
        fmt.Fprint(c, "Msg;%s,\n", msg.String())
      }

      go func(){
        for {
          fmt.Println(0)

          msg, err := reader.ReadString('\n')
          if err != nil {
            delete(*a.resp, pid)
            return
          }

          fmt.Println(1)

          var File, content string
          n, err := fmt.Sscanf(msg, "%s;%s\n", &File, &content)
          if err != nil || n != 1 {
            delete(*a.resp, pid)
            return
          }

          fmt.Println(2)

          handler, ok := (*a.handlers)[File]
          if !ok {
            delete(*a.resp, pid)
            return
          }

          fmt.Println(3)

          if content == "List" {
            fmt.Fprint(c, "List;%s,\n", ListToString((*handler.list)()))

            fmt.Println("4L")

            continue
          } else if content == "Msg" {
            m, err := mpi.FromString(content)
            if err != nil {
              delete(*a.resp, pid)
              return
            }

            fmt.Println("4M")

            err = (*handler.handler)(*m)
            if err != nil {
              delete(*a.resp, pid)
              return
            }

            fmt.Println("5M")
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
  f, ok := (*a.resp)[msg.Pid]
  if !ok {
    return errors.New("no such pid")
  }

  f(msg)
  return nil
}
