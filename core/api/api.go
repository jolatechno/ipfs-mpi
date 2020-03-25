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
  handlers map[string]handler
  resp map[int] chan mpi.Message
}

type handler struct{
  handler func([]mpi.Message) error
  list func() (string, []string)
}

type list struct{
  host string
  peers []string
}

type messages struct {
  Pid int
  messages []mpi.Message
}

func NewApi(port int, ReadTimeout int, WriteTimeout int) (*Api, error){
  a := Api{}
  handle := func(w http.ResponseWriter, r *http.Request){
    file := r.Header.Get("File")
    if file == "" {
      http.Error(w, "no file given", 1)
      return
    }

    handler, ok := a.handlers[file]
    if !ok {
      http.Error(w, "no such file", 1)
      return
    }

    if r.Header.Get("List") != "" {
      host, peer_list := handler.list()

      resp := list{ host:host, peers:peer_list }
      js, err := json.Marshal(resp)
      if err != nil {
        http.Error(w, err.Error(), 1)
        return
      }

      w.Write(js)
    } else {
      expected := r.Header.Get("Expected")
      if expected == "" {
        http.Error(w, "no expectency given", 1)
        return
      }

      var int_expected int
      n, err := fmt.Sscanf(expected, "%d", &int_expected)
      if n != 1 || err != nil {
        http.Error(w, err.Error(), 1)
        return
      }

      var msg messages
      err = json.NewDecoder(r.Body).Decode(&msg)
      if err != nil {
        http.Error(w, "expectency error", 1)
        return
      }

      a.resp[msg.Pid] = make(chan mpi.Message)
      err = handler.handler(msg.messages)
      if err != nil {
        http.Error(w, err.Error(), 1)
        return
      }

      resp := messages{ messages:[]mpi.Message{}}
      for i := 0; i < int_expected ; i++ {
        resp.messages = append(resp.messages, <- a.resp[msg.Pid])
      }

      js, err := json.Marshal(resp)
      if err != nil {
        http.Error(w, err.Error(), 1)
        return
      }

      w.Write(js)
    }
  }

  a.server = &http.Server{
  	Addr:           fmt.Sprintf(":%d", port),
  	Handler:        http.HandlerFunc(handle),
  	ReadTimeout:    time.Duration(ReadTimeout) * time.Second,
  	WriteTimeout:   time.Duration(WriteTimeout) * time.Second,
  	MaxHeaderBytes: 1 << 20,
  }

  go func(){
    panic(a.server.ListenAndServe())
  }()

  return &a, nil
}

func (a *Api)AddHandler(key string, handle func([]mpi.Message) error, list func() (string, []string)) {
  a.handlers[key] = handler{ handler:handle, list:list }
}

func (a *Api)Push(msg mpi.Message) error{
  c, ok := a.resp[msg.Pid]
  if !ok {
    return errors.New("no such pid")
  }
  c <- msg
  return nil
}
