package main

import (
  "fmt"

  "github.com/jolatechno/ipfs-mpi/core"
)

func start(i int) {
  inter, err := core.NewInterface("../../interpreter/echo", 2, i)
  if err != nil {
    panic(err)
  }

  go func(){
    for {
      fmt.Println("Request from:", <- inter.Request())
      inter.Push("test\n")
    }
  }()

  go func(){
    for {
      fmt.Println("Sending message:", <- inter.Message())
    }
  }()

  fmt.Println("Closed:", <- inter.CloseChan())
  fmt.Println("")
}

func main(){
  start(1)
  start(0)
}
