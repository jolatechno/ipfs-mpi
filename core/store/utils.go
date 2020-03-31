package store

import (
  "context"
  "bufio"
  "fmt"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-discovery"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/network"
  "github.com/libp2p/go-libp2p-core/peer"

  "github.com/jolatechno/mpi-peerstore"
  "github.com/jolatechno/ipfs-mpi/core/ipfs-interface"
  "github.com/jolatechno/ipfs-mpi/core/mpi-interface"
  "github.com/jolatechno/ipfs-mpi/core/api"
)

type Entry struct {
  Store *peerstore.Peerstore
  file file.File
  shell *file.IpfsShell
  api *api.Api
  path string
}

func NewEntry(host *host.Host, routingDiscovery *discovery.RoutingDiscovery, f file.File, shell *file.IpfsShell, api *api.Api, path string) *Entry {
  rdv := f.String()
  p := peerstore.NewPeerstore(host, routingDiscovery, rdv)

  return &Entry{ Store:p, file:f, shell:shell, api:api, path:path }
}

func (e *Entry)InitEntry() error {
  err := e.shell.Dowload(e.file)
  if err != nil {
    return err
  }

  return mpi.Install(e.path + e.file.String())
}

func (e *Entry)LoadEntry(ctx context.Context, base protocol.ID) error {
  handler := mpi.Load(e.path + e.file.String(),
  func(msg mpi.Message) error {
    return (*e.api).Push(msg)
  })

  discoveryHandler := func (p *peerstore.Peerstore, id peer.ID) {
		Protocol := protocol.ID(e.file.String() + "//" + string(base))

		stream, err := (*e.Store.Host).NewStream(ctx, id, Protocol)
		if err != nil {
			return
		}

		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		(*e.Store).Add(peer.IDB58Encode(id), func(str string) error{
    	_, err := rw.WriteString(fmt.Sprintf("%s\n", str))
    	if err != nil {
    		return err
    	}
    	err = rw.Flush()
    	if err != nil {
    		return err
    	}

    	return nil
    })
	}

  messageHandler := func(msg mpi.Message) error {
    if (*e.Store).Has(msg.To){
      (*e.Store).Write(msg.To, msg.String()) // pass on the responces
      return nil
    }

    ID, err := peer.IDB58Decode(msg.To)
    if err != nil {
      return err
    }

    discoveryHandler(e.Store, ID)
    (*e.Store).Write(msg.To, msg.String()) // pass on the responces

    return nil
  }

  hostId := peer.IDB58Encode((*e.Store.Host).ID())
  list := func() (string, []string) {
    peers := (*e.Store).Store

    keys := make([]string, len(peers))
    i := 0
    for addr := range peers {
      keys[i] = addr
      i++
    }

    return hostId, keys
  }

  (*e.api).AddHandler(e.file.String(), &messageHandler, &list)

  StreamHandler := func(stream network.Stream) {
		// Create a buffer stream for non blocking read and write.
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		go func(){
      for {
        str, err := rw.ReadString('\n')
    		if err != nil {
    			continue
    		}

        if str[len(str) - 1:] == "\n" {
          str = str[:len(str) - 1]
        }

        msg, err := mpi.FromString(str)
        if err != nil {
    			continue
    		}

        reps, err := handler(*msg)
        if err != nil {
          continue
        }

        for _, rep := range reps {
          err := messageHandler(rep)
          if err != nil {
            continue
          }
        }
      }
    }()
  }

  err := e.Store.SetStreamHandler(base, StreamHandler)
  if err != nil {
    return err
  }

  e.Store.Listen(ctx, discoveryHandler)
  e.Store.Annonce(ctx)
  return nil
}
