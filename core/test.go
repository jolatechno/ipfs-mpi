package main

import (
  file "github.com/jolatechno/ipfs-mpi/core/ipfs-interface"
  "github.com/coreos/go-semver/semver"
)

const (
  url = "/ip4/127.0.0.1/tcp/4001"
  examplesHash = "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"
)

func main(){
  s, err := file.NewShell(url)
  if err != nil {
    panic(err)
  }

  s.Dowload(file.File{ name:examplesHash, version:semver.Version("0.0.0")})
}
