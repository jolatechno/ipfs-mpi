package main

import (
  "errors"
  "context"
  "strings"
  "strconv"
  "fmt"

  "github.com/ipfs/go-log"
  "github.com/jolatechno/libp2p-mpi-core/v2"

  "github.com/carmark/pseudo-terminal-go/terminal"
)

const (
  prompt = "libp2p-mpi>"
)

var (
  MainHeader = "libp2p-mpi"
  MainLogger = log.Logger(MainHeader)
)


func main(){
  log.SetupLogging()

  ctx := context.Background()

  config, quiet, err := ParseFlag()
  if err != nil {
    core.MpiLogger.Panic(err)
  }

  fmt.Println("\nStarting host...")

  host, err := core.NewHost(ctx, config.BootstrapPeers...)
  if err != nil {
    panic(err)
  }

  fmt.Println("Host started")

  if !quiet {
    fmt.Println("Our adress is ", host.ID())
    for _, addr := range host.Addrs() {
      fmt.Println("swarm listening on ", addr)
    }
  }

  fmt.Println("\nStarting store...")

  store, err := core.NewStore(config.Url, config.Path, config.Ipfs_store)
  if err != nil {
    panic(err)
  }

  if !quiet {
    fmt.Println("Connected to store ", config.Ipfs_store)
    for _, file := range store.List() {
      fmt.Printf(" found %q\n", file)
    }
  }

  fmt.Println("Store started")
  fmt.Println("\nStarting libp2p-mpi daemon...")

  mpi, err := core.NewMpi(ctx, config, host, store)
  if err != nil {
    panic(err)
  }

  fmt.Println("Daemon started\n")

  mpi.SetInitFunctions(
    core.NewSlaveComm,
    core.NewMasterSlaveComm,
    core.NewMasterComm,
    core.NewInterface,
    core.NewRemote,
    core.NewNewLogger(quiet),
  )

  mpi.SetErrorHandler(func(err error) {
    panic(err)
  })

  term, err := terminal.NewWithStdInOut()
	if err != nil {
    mpi.Close()
    return
	}
	defer term.ReleaseFromStdInOut() // defer this

  for mpi.Check() {
    cmd, err:= term.ReadLine()
    if err != nil {
      mpi.Close()
      MainLogger.Panic(err)
      return
    }

    splitted := strings.Split(cmd, " ")
    if len(splitted) == 0 {
      continue
    }

    switch splitted[0] {
    default:
      MainLogger.Error(core.CommandNotUnderstood)

    case "list":
      list := mpi.Store().List()
      for _, f := range list {
        fmt.Println(" ", f)
      }

    case "start":
      if len(splitted) < 3 {
        MainLogger.Error("No size given")
        continue
      }

      n, err := strconv.Atoi(splitted[2])
      if n <= 0 && err == nil {
        err = errors.New("Size not understood")
      }
      if err != nil {
        MainLogger.Error(err)
        continue
      }

      go func() {
        err = mpi.Start(splitted[1], n, splitted[3:]...)
        if err != nil {
          MainLogger.Error(err)
        }
      }()

    case "add":
      if len(splitted) < 2 {
        MainLogger.Error("No file given")
        continue
      }

      for _, f := range splitted[1:] {
        go mpi.Add(f)
      }

    case "del":
      if len(splitted) < 2 {
        MainLogger.Error("No file given")
        continue
      }

      for _, f := range splitted[1:] {
        go mpi.Del(f)
      }

    case "exit":
      mpi.Close()
      return
    }
  }
}
