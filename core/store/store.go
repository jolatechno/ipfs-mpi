package store

import (
  "context"
  "os"
  "errors"

  "github.com/jolatechno/mpi-peerstore/utils"
  "github.com/jolatechno/ipfs-mpi/core/ipfs-interface"
  "github.com/jolatechno/ipfs-mpi/core/mpi-interface"
  "github.com/jolatechno/ipfs-mpi/core/api"
  "github.com/jolatechno/ipfs-mpi/core/messagestore"
  "github.com/jolatechno/mpi-peerstore"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-discovery"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/peer"
)

type Store struct {
  Handler *message.Handler
  DaemonStore *mpi.DaemonStore
  Api *api.Api
  Shell *file.IpfsShell
  Store *map[string] *peerstore.Peerstore

  Host *host.Host
  RoutingDiscovery *discovery.RoutingDiscovery

  Protocol protocol.ID
  Maxsize uint64
  Path string
  Ipfs_store string
}

func NewStore(ctx context.Context, host host.Host, config Config) (*Store, error) {
  store := make(map[string] *peerstore.Peerstore)
  proto := protocol.ID(config.Ipfs_store + "//" + config.ProtocolID)

  routingDiscovery, err := utils.NewKadmeliaDHT(ctx, host, config.BootstrapPeers)
  if err != nil {
    return nil, err
  }

  if _, err := os.Stat(config.Path); os.IsNotExist(err) {
    os.MkdirAll(config.Path, file.ModePerm)
  } else if err != nil {
    return nil, err
  }

  shell, err := file.NewShell(config.Url, config.Path, config.Ipfs_store)
  if err != nil {
    return nil, err
  }

  hostId := peer.IDB58Encode(host.ID())
  list := func(str string) (string, []string, error) {
    s, ok := store[str]
    if !ok {
      return hostId, []string{}, errors.New("no such interpreter")
    }

    peers := s.Store

    keys := make([]string, len(peers))
    i := 0
    for addr := range peers {
      keys[i] = addr
      i++
    }

    return hostId, keys, nil
  }

  send := func(msg message.Message) error {
    s, ok := store[msg.File]
    if !ok {
      return errors.New("no such interpreter")
    }

    if (*s).Has(msg.To){
      (*s).Write(msg.To, msg.String()) // pass on the responces
      return nil
    }

    ID, err := peer.IDB58Decode(msg.To)
    if err != nil {
      return err
    }

    (*s.AddPeer)(ID)
    (*s).Write(msg.To, msg.String()) // pass on the responces

    return nil
  }

  handler := &message.Handler{
    List:&list,
    Send:&send,
  }

  api, err := api.NewApi(config.Api_port, handler)
  if err != nil {
    return nil, err
  }

  dameonStore := mpi.NewDaemonStore(config.Path, handler)

  return &Store{
    Handler: handler,
    DaemonStore: &dameonStore,
    Api: api,
    Shell: shell,
    Store: &store,
    Host: &host,
    RoutingDiscovery: routingDiscovery,
    Protocol: proto,
    Maxsize: config.Maxsize,
    Path: config.Path,
    Ipfs_store: config.Ipfs_store,
  }, nil
}
