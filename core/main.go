package main

import (
  "context"
  "fmt"

  "github.com/jolatechno/ipfs-mpi/core/store"

  "github.com/libp2p/go-libp2p"
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
    store.Config(config),
  )
  if err != nil {
		panic(err)
	}

  fmt.Printf("Api listening on: /ip4/127.0.0.1/tcp/%d\n", config.Api_port)

  err = Store.Init(ctx)
  if err != nil {
		panic(err)
	}

  fmt.Println("Daemon started")

  select {}
}
