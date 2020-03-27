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
  Port int
  handlers *map[string] handler
  resp *map[int] func(mpi.Message)
}

func NewApi(port int) (*Api, error) {
  handlers := make(map[string] handler)
  resp := make(map[int] func(mpi.Message))

  l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
  if err != nil {
    return nil, err
  }

  a := Api{
    Port:l.Addr().(*net.TCPAddr).Port,
    handlers:&handlers,
    resp:&resp,
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

      (*a.resp)[pid] = func(msg mpi.Message) {
        fmt.Fprintf(c, "\"Msg\";%q\n", msg.String())
      }

      go func(){
        for {
          msg, err := reader.ReadString('\n')
          if err != nil {
            delete(*a.resp, pid)
            return
          }

          var File, content string
          n, err := fmt.Sscanf(msg, "%q;%q\n", &File, &content)

          if err != nil || n != 2 {
            delete(*a.resp, pid)
            return
          }

          handler, ok := (*a.handlers)[File]
          if !ok {
            delete(*a.resp, pid)
            return
          }

          if content == "List" {
            fmt.Fprintf(c, "\"List\";%q\n", ListToString((*handler.list)()))
            continue
          } else {
            m, err := mpi.FromString(content)
            if err != nil {
              delete(*a.resp, pid)
              return
            }

            err = (*handler.handler)(*m)
            if err != nil {
              delete(*a.resp, pid)
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
  f, ok := (*a.resp)[msg.Pid]
  if !ok {
    return errors.New("no such pid")
  }

  f(msg)
  return nil
}
