package main

import (
  "context"

  "github.com/jolatechno/ipfs-mpi/core"
)

func main(){
  ctx := context.Background()

  config, err := ParseFlag()
  if err != nil {
    panic(err)
  }

  _, err = core.NewMpi(ctx, config)
  if err != nil {
    panic(err)
  }
}
