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
    host,
    protocol.ID(config.ipfs_store + config.ProtocolID),

  )
  if err != nil {
		panic(err)
	}

  err = Store.StartShell(
    config.url,
    config.path,
    config.ipfs_store,
    config.maxsize,
  )
  if err != nil {
		panic(err)
	}

  fmt.Println("Successfully connected to ipfs-api on:", config.url)
  fmt.Println("Connected to:", config.ipfs_store)

  err = Store.StartApi(
    config.api_port,
    config.ReadTimeout,
    config.WriteTimeout,
  )
  if err != nil {
		panic(err)
	}

  fmt.Println("Api listening on: /127.0.0.1:", config.api_port)

  err = Store.StartDiscovery(
    ctx,
    config.BootstrapPeers,
  )
  if err != nil {
		panic(err)
	}

  fmt.Println("Successfully started peer discovery")

  err = Store.Init(ctx)
  if err != nil {
		panic(err)
	}

  fmt.Println("Daemon started")

  select {}
}
