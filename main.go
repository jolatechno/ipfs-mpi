package main

import (
  "errors"
  "context"
  //"bufio"
  "strings"
  "strconv"
  //"os"
  "fmt"
  "log"

  "github.com/jolatechno/ipfs-mpi/core"

  "github.com/carmark/pseudo-terminal-go/terminal"
)

const (
  prompt = "libp2p-mpi>"
)

func main(){
  fmt.Println("Starting libp2p-mpi daemon...")

  ctx := context.Background()

  config, quiet, err := ParseFlag()
  if err != nil {
    panic(err)
  }

  store, err := core.NewMpi(ctx, config)
  if err != nil {
    panic(err)
  }

  store.SetErrorHandler(func(err error) {
    if !quiet {
      log.Println(err.Error())
    }
  })

  fmt.Println("Our adress is ", store.Host().ID())

  for _, addr := range store.Host().Addrs() {
    fmt.Println("swarm listening on ", addr)
  }

  term, err := terminal.NewWithStdInOut()
	if err != nil {
    store.Close()
    return
	}
	defer term.ReleaseFromStdInOut() // defer this

  //scanner := bufio.NewScanner(os.Stdin)
  for store.Check()/* && scanner.Scan()*/ {
    //cmd := scanner.Text()
    cmd, err:= term.ReadLine()
    if err != nil {
      store.Close()
      return
    }

    splitted := strings.Split(cmd, " ")
    if len(splitted) == 0 {
      continue
    }

    switch splitted[0] {
    default:
      store.Raise(core.CommandNotUnderstood)

    case "list":
      list := store.Store().List()
      for _, f := range list {
        fmt.Println(" ", f)
      }

    case "start":
      if len(splitted) < 3 {
        store.Raise(errors.New("No size given"))
        continue
      }

      n, err := strconv.Atoi(splitted[2])
      if n <= 0 && err == nil {
        err = errors.New("Size not understood")
      }
      if err != nil {
        store.Raise(err)
        continue
      }

      go func() {
        err = store.Start(splitted[1], n, splitted[3:]...)
        if err != nil {
          store.Raise(err)
        }
      }()

    case "add":
      if len(splitted) < 2 {
        store.Raise(errors.New("No file given"))
        continue
      }

      for _, f := range splitted[1:] {
        go store.Add(f)
      }

    case "del":
      if len(splitted) < 2 {
        store.Raise(errors.New("No file given"))
        continue
      }

      for _, f := range splitted[1:] {
        go store.Del(f)
      }

    case "exit":
      store.Close()
      return
    }
  }

  /*if err := scanner.Err(); err != nil {
    panic(err)
  }*/
}
