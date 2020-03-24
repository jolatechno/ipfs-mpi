package main

import (
  "context"
  "fmt"

  "github.com/jolatechno/ipfs-mpi/core/store"

  "github.com/libp2p/go-libp2p"
  "github.com/libp2p/go-libp2p-core/protocol"
  maddr "github.com/multiformats/go-multiaddr"
)

func main(){
  config := ParseFlags()
  ctx := context.Background()

  host, err := libp2p.New(ctx,
		libp2p.ListenAddrs([]maddr.Multiaddr(config.ListenAddresses)...),
	)
	if err != nil {
		panic(err)
	}

  fmt.Println("Our adress is: ", host.ID())

	Store, err := store.NewStore(
    ctx,
    config.url,
    host,
    config.BootstrapPeers,
    protocol.ID(config.ipfs_store + config.ProtocolID),
    config.path,
    config.ipfs_store,
    config.maxsize,
  )

  if err != nil {
		panic(err)
	}

  err = Store.Start(ctx)
  if err != nil {
		panic(err)
	}

  select {}
}
