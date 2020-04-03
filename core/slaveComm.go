package core

import (
  "bufio"
  "fmt"
  "errors"
  "context"
  "strings"
  "strconv"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/peer"
)

func NewSlaveComm(ctx context.Context, host host.Host, base protocol.ID, param Param) (Comm, error) {
  Addrs := make([]peer.ID, len(param.Addrs))
  for i, addr := range param.Addrs {
    Addrs[i] = peer.ID(addr)
  }

  comm := Comm{
    Id: param.Id,
    Idx: param.Idx,
    Host: host,
    Addrs: Addrs,
    Pid: protocol.ID(fmt.Sprintf("%s/%s", param.Id, string(base))),
    Remotes: make([]Remote, len(param.Addrs)),
  }

  for i, addr := range comm.Addrs {
    if i != param.Idx && (i > param.Idx || !param.Init) {
      proto := protocol.ID(fmt.Sprintf("%d/%s", i, string(comm.Pid)))

      stream, err := host.NewStream(ctx, addr, proto)
      if err != nil {
        comm.Stop()
        return comm, err
      }

      rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

      comm.Remotes[i] = Remote{
        Sent: []string{},
        Stream: rw,
        ResetChan: make(chan bool),
      }

      streamHandler, err := comm.Remotes[i].StreamHandler()
      if err != nil {
        comm.Stop()
        return comm, err
      }

      host.SetStreamHandler(proto, streamHandler)
    }
  }

  return comm, nil
}

type Param struct {
  Init bool
  Idx int
  Id string
  Addrs []string
}

func ParamFromString(msg string) (Param, error) {
  param := Param{}
  splitted := strings.Split(msg, ",")
  if len(splitted) != 4 {
    return param, errors.New("Param dosen't have the right number fields")
  }

  if splitted[0] == "0" {
    param.Init = false
  } else if splitted[0] == "1" {
    param.Init = true
  } else {
    return param, errors.New("bool header not understood")
  }

  idx, err := strconv.Atoi(splitted[1])
  if err != nil {
    return param, err
  }

  param.Idx = idx
  param.Id = splitted[2]
  param.Addrs = strings.Split(splitted[3], ";")

  return param, err
}
