package main

import (
  "flag"

  "github.com/jolatechno/ipfs-mpi/core"

  "github.com/ipfs/go-log"
  dht "github.com/libp2p/go-libp2p-kad-dht"

)

func ParseFlag() (core.Config, bool, error) {
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
  debugMpi := flag.Bool("debug-mpi", false, "enable debug-mode on mpi")
  debugInterface := flag.Bool("debug-interface", false, "enable debug-mode on interface")
  debugDht := flag.Bool("debug-dht", false, "enable debug-mode on dht")
  debugMdns := flag.Bool("debug-mdns", false, "enable debug-mode on mdns")

	flag.Parse()

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
      if *debugRemote {
        err := log.SetLogLevel(core.RemoteHeader, "debug")
        if err != nil {
          panic(err)
        }
      }

      if *debugSlaveComm {
        err := log.SetLogLevel(core.SlaveCommHeader, "debug")
        if err != nil {
          panic(err)
        }
      }

      if *debugMasterComm {
        err := log.SetLogLevel(core.MasterCommHeader, "debug")
        if err != nil {
          panic(err)
        }
      }

      if *debugStore {
        err := log.SetLogLevel(core.IpfsHeader, "debug")
        if err != nil {
          panic(err)
        }
      }

      if *debugHost {
        err := log.SetLogLevel(core.HostHeader, "debug")
        if err != nil {
          panic(err)
        }
      }

      if *debugMpi {
        err := log.SetLogLevel(core.MpiHeader, "debug")
        if err != nil {
          panic(err)
        }
      }

      if *debugInterface {
        err := log.SetLogLevel(core.InterfaceHeader, "debug")
        if err != nil {
          panic(err)
        }
      }

      if *debugDht {
        err := log.SetLogLevel("dht", "debug")
        if err != nil {
          panic(err)
        }
        err = log.SetLogLevel("dht", "warn")
        if err != nil {
          panic(err)
        }
      }

      if *debugMdns {
        err := log.SetLogLevel("mdns", "debug")
        if err != nil {
          panic(err)
        }
        err = log.SetLogLevel("mdns", "warn")
        if err != nil {
          panic(err)
        }
      }

    }
  }

  return config, *quiet, nil
}
