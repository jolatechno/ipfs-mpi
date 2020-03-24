package main

import (
  file "github.com/jolatechno/ipfs-mpi/core/ipfs-interface"
  "github.com/coreos/go-semver/semver"
)

const (
  url = "/ip4/127.0.0.1/tcp/5001"
  examplesHash = "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"
)

func main(){
  s, err := file.NewShell(url)
  if err != nil {
    panic(err)
  }

  vers, err := semver.NewVersion("0.0.0")
  if err != nil {
    panic(err)
  }

  err = s.Dowload(file.File{ Name:examplesHash, Version:vers})
  if err != nil {
    panic(err)
  }
}
