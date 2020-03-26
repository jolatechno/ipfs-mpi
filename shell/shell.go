package shell

import (
  "bufio"
  "fmt"
  "net"
  "errors"

  "github.com/jolatechno/ipfs-mpi/core/api"
  "github.com/jolatechno/ipfs-mpi/core/mpi-interface"
)

type list struct {
  host string
  peers []string
}

type Shell struct {
  conn net.Conn
  listChan chan list
}

func NewShell(port int, pid int) (*Shell, chan mpi.Message, error) {
  c := make(chan mpi.Message)
  listChan := make(chan list)

  s, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
  if err != nil {
    return nil, c, err
  }

  fmt.Fprintf(s, "%d\n", pid)
  go func(){
    for {
      msg, err := bufio.NewReader(s).ReadString('\n')
      fmt.Println(msg, err)
      if err != nil {
        panic(err)
      }

      var header, content string
      fmt.Sscanf(msg, "%q;%q\n", &header, &content)

      if header == "List" {
        host, peers := api.ListFromString(content)
        listChan <- list{ host:host, peers:peers }
      } else if header == "Msg" {
        m, err := mpi.FromString(content)
        if err != nil {
          panic(err)
        }

        c <- *m
      } else {
        panic(errors.New("Header not understood"))
      }

    }
  }()

  return &Shell{ conn:s, listChan:listChan }, c, nil
}

func (s *Shell)List(file string) (string, []string){
  fmt.Fprintf(s.conn, "%q;\"List\"\n", file)

  list := <- s.listChan
  return list.host, list.peers
}

func (s *Shell)Send(file string, msg mpi.Message) {
  fmt.Fprintf(s.conn, "%q;%q\n", file, msg.String())
}
