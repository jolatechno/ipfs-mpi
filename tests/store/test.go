package main

import (
  "fmt"

  "github.com/jolatechno/ipfs-mpi/core"
)

func main() {
  store, err := core.NewStore("/ip4/127.0.0.1/tcp/5001", "./interpreter/", "QmQvghovY96ujyrFGgPPRu6UKXuEBLiZrNuQePLCigCWKS/")
  if err != nil {
    panic(err)
  }

  go func() {
    for {
      fmt.Println("\nFiles: ", store.List())

      f, err := store.Get(5000000)
      if err != nil {
        panic(err)
      }

      fmt.Println("Found file ", f)
      fmt.Println("Has file: ", store.Has(f))

      err = store.Dowload(f)
      if err != nil {
        panic(err)
      }

      fmt.Println("Has file: ", store.Has(f))
    }
  }()

  fmt.Println("Store closed: ", <- store.CloseChan())
}
