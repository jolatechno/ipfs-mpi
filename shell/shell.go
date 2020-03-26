package shell

import (
  "bufio"
  "fmt"
  "net"

  "github.com/jolatechno/ipfs-mpi/core/api"
  "github.com/jolatechno/ipfs-mpi/core/mpi-interface"
)

type Shell struct {
  conn net.Conn
}

func NewShell(port int, pid int) (*Shell, chan mpi.Message, error) {
  c := make(chan mpi.Message)

  s, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
  if err != nil {
    return nil, c, err
  }

  fmt.Fprintf(s, "%d\n", pid)
  go func(){
    for {
      msg, err := bufio.NewReader(s).ReadString('\n')
      if err != nil {
        panic(err)
      }

      m, err := mpi.FromString(msg)
      if err != nil {
        panic(err)
      }

      c <- *m
    }
  }()

  return &Shell{ conn:s }, c, nil
}

func (s *Shell)List(file string) (string, []string, error){
  fmt.Fprintf(s.conn, "$s,List\n", file)

  str, err := bufio.NewReader(s.conn).ReadString('\n')
  if err != nil {
    return "", []string{}, nil
  }

  host, peers := api.ListFromString(str)
  return host, peers, nil
}

func (s *Shell)Send(file string, msg mpi.Message) {
  fmt.Fprintf(s.conn, "%s,%s\n", file, msg.String())
}
