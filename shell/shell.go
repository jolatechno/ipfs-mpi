package shell

import (
  "bufio"
  "fmt"
  "net"
  "errors"
  "strings"

  "github.com/jolatechno/ipfs-mpi/core/api"
  "github.com/jolatechno/ipfs-mpi/core/messagestore"
)

type list struct {
  host string
  peers []string
}

type Shell struct {
  Conn net.Conn
  ListChan chan list
  MessageChan chan message.Message
}

func NewShell(port int, pid int) (*Shell, error) {
  messageChan := make(chan message.Message)
  listChan := make(chan list)

  s, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
  if err != nil {
    return nil, err
  }

  fmt.Fprintf(s, "%d\n", pid)
  go func(){
    for {
      msg, err := bufio.NewReader(s).ReadString('\n')
      if err != nil {
        panic(err)
      }

      splitted_msg := strings.Split(msg, ";")
      if len(splitted_msg) != 2 {
        panic(errors.New("Message dosen't have a clearly defined header and content"))
      }

      if splitted_msg[0] == "List" {
        host, peers := api.ListFromString(splitted_msg[1])
        listChan <- list{ host:host, peers:peers }
      } else if splitted_msg[0] == "Msg" {
        m, err := message.FromString(splitted_msg[1])
        if err != nil {
          panic(err)
        }

        messageChan <- *m
      } else {
        panic(errors.New("Header not understood"))
      }

    }
  }()

  return &Shell{ Conn:s, ListChan:listChan, MessageChan:messageChan }, nil
}

func (s *Shell)List(file string) (string, []string){
  fmt.Fprintf(s.Conn, "List,%s\n", file)

  list := <- s.ListChan
  return list.host, list.peers
}

func (s *Shell)Send(msg message.Message) {
  fmt.Fprintf(s.Conn, "%Msg;%s\n", msg.String())
}

func (s *Shell)Reqest(From string) message.Message {
  fmt.Fprintf(s.Conn, "%Req;%s\n", From)

  return <- s.MessageChan
}
