package main

import (
  "fmt"

  "github.com/jolatechno/ipfs-mpi/core"
)
func main(){
  inter, err := core.NewInterface(".", 2, 1)
  if err != nil {
    panic(err)
  }

  go func(){
    for {
      fmt.Println(<- inter.Request())
      inter.Push("test\n")
    }
  }()

  go func(){
    for {
      fmt.Println(<- inter.Message())
    }
  }()


  fmt.Println("Closed : ", <- inter.CloseChan())
}
