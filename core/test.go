package main

import (
  "context"

  "github.com/jolatechno/ipfs-mpi/core/ipfs-interface"
  "github.com/jolatechno/ipfs-mpi/core/store"

  "github.com/libp2p/go-libp2p"
  "github.com/coreos/go-semver/semver"
  maddr "github.com/multiformats/go-multiaddr"
)

const (
  url = "/ip4/127.0.0.1/tcp/5001"
  examplesHash = "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"
  BootstrapPeers = []maddr.Multiaddr{}
)

func main(){
  ctx := context.Background()

  host, err := libp2p.New(ctx,
		libp2p.ListenAddrs([]multiaddr.Multiaddr(config.ListenAddresses)...),
	)
	if err != nil {
		panic(err)
	}

	logger.Info("Host created. We are:", host.ID())
	if !config.quiet {
		logger.Info(host.Addrs())
	}

	Store, err := store.NewStore(ctx, url, host, BootstrapPeers)
  if err != nil {
		panic(err)
	}

  Store.Add(file.File{ Name:examplesHash, Version:semver.NewVersion("0.0.0")})
  err = Store.Start()
  if err != nil {
		panic(err)
	}
}
