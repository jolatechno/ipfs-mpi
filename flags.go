package main

import (
  "flag"
  "fmt"

  "github.com/jolatechno/libp2p-mpi-core/v2"

  "github.com/ipfs/go-log"
  dht "github.com/libp2p/go-libp2p-kad-dht"

)

type debugList []string

func (i *debugList)String() string {
	return fmt.Sprint(*i)
}

func (i *debugList)Set(value string) error {
	*i = append(*i, value)
	return nil
}

func ParseFlag() (core.Config, bool, error) {
  var config core.Config
  var debugs debugList

  config.Base = "libp2p-mpi/1.0.0" //set to the libp2p-mpi version

	flag.StringVar(&config.Url, "ipfs-api", "/ip4/127.0.0.1/tcp/5001", "Local ipfs daemon url")
	flag.StringVar(&config.Path, "path", "interpreters/", "path to the interpretor directory")
	flag.StringVar(&config.Ipfs_store, "ipfs-store", "QmbK4D76sLDvRn2qKVbtqn95qBW7CvS1nXa6F2ynyn5xaF/",
		"Unique string to identify the ipfs store you are using")
	flag.Uint64Var(&config.Maxsize, "maxsize", 10000000, "Set the max use space, default to 10MB")
  flag.Var(&config.BootstrapPeers, "peer", "Adds a peer multiaddress to the bootstrap list")

  quiet := flag.Bool("q", false, "start on quiet mode (overwrite all debug-mode)")
  debugAll := flag.Bool("debug-all", false, "enable debug-mode on all interfaces")
  flag.Var(&debugs, "debug", `a list of interface name for which to enable debug-mode,
can describe libp2p-mpi interfaces such as:
  Remote, SlaveComm, MasterComm, Interface, IpfsStore, Mpi, Host...
or go-libp2p or go-ipfs interfaces such as:
  dht, mdns, basichost, peerqueue, swarm2...`)

	flag.Parse()

  fmt.Println(debugs)

  if len(config.BootstrapPeers) == 0 {
    config.BootstrapPeers = dht.DefaultBootstrapPeers
  }

  log.SetAllLoggers(log.LevelError)

  for _, header := range []string{core.RemoteHeader, core.SlaveCommHeader, core.MasterCommHeader, core.IpfsHeader, core.HostHeader, core.MpiHeader, core.InterfaceHeader} {
    err := log.SetLogLevel(header, "info")
    if err != nil {
      panic(err)
    }
  }

  if !*quiet {
    for _, header := range []string{core.RemoteHeader, core.SlaveCommHeader, core.MasterCommHeader, core.IpfsHeader, core.HostHeader, core.MpiHeader, core.InterfaceHeader} {
      err := log.SetLogLevel(header, "warn")
      if err != nil {
        panic(err)
      }
    }

    if *debugAll {
      log.SetAllLoggers(log.LevelDebug)

    } else {
      for _, str := range debugs {
        err := log.SetLogLevel(str, "debug")
        if err != nil {
          panic(err)
        }
        /*err = log.SetLogLevel(str, "warn")
        if err != nil {
          panic(err)
        }*/
      }

    }
  }

  return config, *quiet, nil
}
