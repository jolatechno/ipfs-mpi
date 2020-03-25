package main

import (
	"flag"

	"github.com/jolatechno/ipfs-mpi/core/store"
)

func ParseFlags() store.Config {
	config := store.Config{}

	config.ProtocolID = "ipfs-mpi/1.0.0" //set to the ipfs-mpi version

	flag.StringVar(&config.Url, "ipfs-api", "/ip4/127.0.0.1/tcp/5001", "Local ipfs daemon url")
	flag.StringVar(&config.Path, "path", "interpreter/", "path to the interpretor directory")
	flag.StringVar(&config.Ipfs_store, "ipfs-store", "QmRfk8DdfrPQUxxThhgRxpPYvoa9qpjwV1veqXaSYgrrWf/",
		"Unique string to identify the ipfs store you are using")
	flag.Var(&config.BootstrapPeers, "peer", "Adds a peer multiaddress to the bootstrap list")
	flag.Var(&config.ListenAddresses, "listen", "Adds a multiaddress to the listen list")
	flag.Uint64Var(&config.Maxsize, "maxsize", 100000, "Set the max use space")
	flag.IntVar(&config.Api_port, "p", 8000, "Set the api port")
	flag.IntVar(&config.WriteTimeout, "WriteTimeout", 100, "Set the write timeout")
	flag.IntVar(&config.ReadTimeout, "ReadTimeout", 1, "Set the max use space")
	flag.Parse()

	return config
}
