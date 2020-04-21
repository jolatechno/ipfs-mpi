package main

import (
  "flag"

  "github.com/jolatechno/ipfs-mpi/core"

  dht "github.com/libp2p/go-libp2p-kad-dht"

)

func ParseFlag() (core.Config, bool, error) {
  config := core.Config{}

  config.Base = "libp2p-mpi/1.0.0" //set to the libp2p-mpi version

	flag.StringVar(&config.Url, "ipfs-api", "/ip4/127.0.0.1/tcp/5001", "Local ipfs daemon url")
	flag.StringVar(&config.Path, "path", "interpreters/", "path to the interpretor directory")
	flag.StringVar(&config.Ipfs_store, "ipfs-store", "QmaczSWUzonPp9C2cE52UgNiKpzbtz571pJr6dB65kseVg/",
		"Unique string to identify the ipfs store you are using")
	flag.Uint64Var(&config.Maxsize, "maxsize", 10000000, "Set the max use space, default to 10MB")
  flag.Var(&config.BootstrapPeers, "peer", "Adds a peer multiaddress to the bootstrap list")

  quiet := flag.Bool("q", false, "start on quiet mode")

	flag.Parse()

  if len(config.BootstrapPeers) == 0 {
    config.BootstrapPeers = dht.DefaultBootstrapPeers
  }



  return config, *quiet, nil
}
