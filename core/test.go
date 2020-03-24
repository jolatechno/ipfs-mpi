package main

import (
  "context"

  "github.com/jolatechno/ipfs-mpi/core/ipfs-interface"
  "github.com/jolatechno/ipfs-mpi/core/store"

  "github.com/libp2p/go-libp2p"
  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/coreos/go-semver/semver"
  maddr "github.com/multiformats/go-multiaddr"
)

var (
  url = "/ip4/127.0.0.1/tcp/5001"
  examplesHash = "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"
  BootstrapPeers = []maddr.Multiaddr{}
  Id = protocol.ID("test/0.0.0")
  ListenAddresses = []maddr.Multiaddr{}
)

func main(){
  ctx := context.Background()

  host, err := libp2p.New(ctx,
		libp2p.ListenAddrs([]maddr.Multiaddr(ListenAddresses)...),
	)
	if err != nil {
		panic(err)
	}

	Store, err := store.NewStore(ctx, url, host, BootstrapPeers)
  if err != nil {
		panic(err)
	}

  vers, err := semver.NewVersion("0.0.0")
  if err != nil {
		panic(err)
	}

  Store.Add(file.File{ Name:examplesHash, Version:vers})
  err = Store.Start(ctx, Id)
  if err != nil {
		panic(err)
	}
}
