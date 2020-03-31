package main

import (
  "fmt"
  "os"
  "strconv"

  "github.com/jolatechno/ipfs-mpi/shell"
  "github.com/jolatechno/ipfs-mpi/core/mpi-interface"
)

func main() {
  port, err := strconv.Atoi(os.Args[1])
  if err != nil {
    panic(err)
  }

  pid, err := strconv.Atoi(os.Args[2])
  if err != nil {
    panic(err)
  }

  Shell, c, err := shell.NewShell(port, pid)
  if err != nil {
    panic(err)
  }

  host, peers := Shell.List("echo")
  fmt.Println(host, peers)

  if len(peers) != 0 {
    for _, peer := range peers {
      msg := mpi.Message{ Pid:-1, To:peer, From:host, Data:[]byte(fmt.Sprint(pid))}
      Shell.Send("echo", msg)
      fmt.Println(<- c)
    }
  } else {
    fmt.Println("No peer are connected")
  }
}
