package main

import (
	"flag"
	"net"
	"strings"
	"errors"
	"fmt"

	"github.com/jolatechno/ipfs-mpi/core/store"
	maddr "github.com/multiformats/go-multiaddr"
)

func ListIpAdresses() ([]maddr.Multiaddr, error) {
	returnAddr := []maddr.Multiaddr{}
	addr, err := maddr.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		return returnAddr, err
	}

	returnAddr = append(returnAddr, addr)

	addrs, err := net.InterfaceAddrs()
  if err != nil {
    return returnAddr, err
  }

  for _, a := range addrs {
    if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
      block := strings.Split(a.String(), "/")
      if len(block) > 2 {
        return returnAddr, errors.New("Ip adress with too many slash")
      }

      if ipnet.IP.To4() != nil {
        addr, err := maddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/0", block[0]))
        if err != nil {
          return returnAddr, err
        }
        returnAddr = append(returnAddr, addr)
      } else {
        addr, err := maddr.NewMultiaddr(fmt.Sprintf("/ip6/%s/tcp/0", block[0]))
        if err != nil {
          return returnAddr, err
        }
        returnAddr = append(returnAddr, addr)
      }
    }
  }

	return returnAddr, nil
}

func ParseFlags() (store.Config, error) {
	config := store.Config{}
	addrs, err := ListIpAdresses()
	if err != nil {
		return config, err
	}

	config.ListenAddresses = addrs
	config.ProtocolID = "ipfs-mpi/1.0.0" //set to the ipfs-mpi version

	flag.StringVar(&config.Url, "ipfs-api", "/ip4/127.0.0.1/tcp/5001", "Local ipfs daemon url")
	flag.StringVar(&config.Path, "path", "interpreter/", "path to the interpretor directory")
	flag.StringVar(&config.Ipfs_store, "ipfs-store", "QmeqRpxbbjTWfPbr54WLTuPekUDeAaqmJbFQCEkLb6ABRT/",
		"Unique string to identify the ipfs store you are using")
	flag.Var(&config.BootstrapPeers, "peer", "Adds a peer multiaddress to the bootstrap list")
	flag.Uint64Var(&config.Maxsize, "maxsize", 100000, "Set the max use space")
	flag.IntVar(&config.Api_port, "p", 0, "Set the api port")
	flag.Parse()

	return config, nil
}
