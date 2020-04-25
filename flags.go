package main

import (
  "flag"

  "github.com/jolatechno/ipfs-mpi/core"

  dht "github.com/libp2p/go-libp2p-kad-dht"

)

func ParseFlag() (core.Config, bool, map[string]bool, error) {
  config := core.Config{}

  config.Base = "libp2p-mpi/1.0.0" //set to the libp2p-mpi version

	flag.StringVar(&config.Url, "ipfs-api", "/ip4/127.0.0.1/tcp/5001", "Local ipfs daemon url")
	flag.StringVar(&config.Path, "path", "interpreters/", "path to the interpretor directory")
	flag.StringVar(&config.Ipfs_store, "ipfs-store", "QmbK4D76sLDvRn2qKVbtqn95qBW7CvS1nXa6F2ynyn5xaF/",
		"Unique string to identify the ipfs store you are using")
	flag.Uint64Var(&config.Maxsize, "maxsize", 10000000, "Set the max use space, default to 10MB")
  flag.Var(&config.BootstrapPeers, "peer", "Adds a peer multiaddress to the bootstrap list")

  quiet := flag.Bool("q", false, "start on quiet mode (overwrite all debug-mode)")
  debugAll := flag.Bool("debug-all", false, "enable debug-mode on all interfaces")

  debugRemote := flag.Bool("debug-remote", false, "enable debug-mode on remote")
  debugSlaveComm := flag.Bool("debug-slave", false, "enable debug-mode on SlaveComm")
  debugMasterComm := flag.Bool("debug-master", false, "enable debug-mode on MasterComm")
  debugStore := flag.Bool("debug-store", false, "enable debug-mode on ipfs-store (not used for now)")
  debugHost := flag.Bool("debug-host", false, "enable debug-mode on host (not used for now)")

	flag.Parse()

  if len(config.BootstrapPeers) == 0 {
    config.BootstrapPeers = dht.DefaultBootstrapPeers
  }

  debugs := make(map[string] bool)
  if *quiet {
  } else if *debugAll {
    debugs[core.RemoteHeader] = true
    debugs[core.SlaveCommHeader] = true
    debugs[core.MasterCommHeader] = true
    debugs[core.IpfsHeader] = true
    debugs[core.HostHeader] = true
  } else {
    debugs[core.RemoteHeader] = *debugRemote
    debugs[core.SlaveCommHeader] = *debugSlaveComm
    debugs[core.MasterCommHeader] = *debugMasterComm
    debugs[core.IpfsHeader] = *debugStore
    debugs[core.HostHeader] = *debugHost
  }

  return config, *quiet, debugs, nil
}
