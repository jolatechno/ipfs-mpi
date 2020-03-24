package main

import (
  "context"

  "github.com/jolatechno/ipfs-mpi/core/store"

  "github.com/libp2p/go-libp2p"
  "github.com/libp2p/go-libp2p-core/protocol"
  maddr "github.com/multiformats/go-multiaddr"
)

var (
  url = "/ip4/127.0.0.1/tcp/5001"
  examplesHash = "QmddRNU2VWkpm8FaK2S4QcXCHD3x5kUSiDpKLa1MttRUso/"
  BootstrapPeers = []maddr.Multiaddr{}
  Id = protocol.ID("test/0.0.0")
  ListenAddresses = []maddr.Multiaddr{}
  path = "interpretors/"
  maxsize uint64 = 60000000
)

func main(){
  ctx := context.Background()

  host, err := libp2p.New(ctx,
		libp2p.ListenAddrs([]maddr.Multiaddr(ListenAddresses)...),
	)
	if err != nil {
		panic(err)
	}

	Store, err := store.NewStore(ctx, url, host, BootstrapPeers, Id, path, examplesHash, maxsize)
  if err != nil {
		panic(err)
	}

  err = Store.Get(ctx)
  if err != nil {
		panic(err)
	}

  err = Store.Start(ctx)
  if err != nil {
		panic(err)
	}

  select {}
}
