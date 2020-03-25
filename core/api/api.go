package api

import (
  "net/http"
  "encoding/json"
  "errors"
  "time"
  "fmt"

  "github.com/jolatechno/ipfs-mpi/core/mpi-interface"
)

type Api struct {
  server *http.Server
  handlers map[string]func(mpi.Message) error
  resp map[int] chan mpi.Message
}

func NewApi(port int, ReadTimeout int, WriteTimeout int) (*Api, error){
  a := Api{}
  handle := func(w http.ResponseWriter, r *http.Request){
    file := r.Header.Get("File")
    if file == "" {
      panic(errors.New("no file given"))
    }

    var msg mpi.Message
    err := json.NewDecoder(r.Body).Decode(&msg)
    if err != nil {
      panic(err)
    }

    handler, ok := a.handlers[file]
    if !ok {
      panic(errors.New("no such file"))
    }

    a.resp[msg.Pid] = make(chan mpi.Message)
    err = handler(msg)
    if err != nil {
      panic(err)
    }

    js, err := json.Marshal(<- a.resp[msg.Pid])
    if err != nil {
      panic(err)
    }

    w.Write(js)
  }

  a.server = &http.Server{
  	Addr:           fmt.Sprintf(":%d", port),
  	Handler:        http.HandlerFunc(handle),
  	ReadTimeout:    time.Duration(ReadTimeout) * time.Second,
  	WriteTimeout:   time.Duration(WriteTimeout) * time.Second,
  	MaxHeaderBytes: 1 << 20,
  }

  err := a.server.ListenAndServe()
  if err != nil {
    return nil, err
  }

  return &a, nil
}

func (a *Api)AddHandler(key string, handle func(mpi.Message) error) {
  a.handlers[key] = handle
}

func (a *Api)Push(msg mpi.Message) error{
  c, ok := a.resp[msg.Pid]
  if !ok {
    return errors.New("no such pid")
  }
  c <- msg
  return nil
}
