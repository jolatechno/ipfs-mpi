package main

import (
	"flag"
	"strings"

	maddr "github.com/multiformats/go-multiaddr"
)

// A new type we need for writing a custom flag parser
type addrList []maddr.Multiaddr

func (al *addrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

func (al *addrList) Set(value string) error {
	addr, err := maddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}

func StringsToAddrs(addrStrings []string) (maddrs []maddr.Multiaddr, err error) {
	for _, addrString := range addrStrings {
		addr, err := maddr.NewMultiaddr(addrString)
		if err != nil {
			return maddrs, err
		}
		maddrs = append(maddrs, addr)
	}
	return
}

type Config struct {
	url string
	path string
	ipfs_store string
	BootstrapPeers addrList
	ListenAddresses addrList
	ProtocolID string
	maxsize uint64
	api_port int
	WriteTimeout int
	ReadTimeout int
}

func ParseFlags() Config {
	config := Config{}

	config.ProtocolID = "ipfs-mpi/1.0.0" //set to the ipfs-mpi version

	flag.StringVar(&config.url, "ipfs-api", "/ip4/127.0.0.1/tcp/5001", "Local ipfs daemon url")
	flag.StringVar(&config.path, "path", "interpretors/", "path to the interpretor directory")
	flag.StringVar(&config.ipfs_store, "ipfs-store", "QmRfk8DdfrPQUxxThhgRxpPYvoa9qpjwV1veqXaSYgrrWf/",
		"Unique string to identify the ipfs store you are using")
	flag.Var(&config.BootstrapPeers, "peer", "Adds a peer multiaddress to the bootstrap list")
	flag.Var(&config.ListenAddresses, "listen", "Adds a multiaddress to the listen list")
	flag.Uint64Var(&config.maxsize, "maxsize", 100000, "Set the max use space")
	flag.Uint64Var(&config.api_port, "p", 8000, "Set the api port")
	flag.Uint64Var(&config.WriteTimeout, "WriteTimeout", 100, "Set the write timeout")
	flag.Uint64Var(&config.ReadTimeout, "ReadTimeout", 1, "Set the max use space")
	flag.Parse()

	return config
}
