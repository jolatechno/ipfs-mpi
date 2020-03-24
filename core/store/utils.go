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
)

type Entry struct {
  store peerstore.Peerstore
  file file.File
  shell *file.IpfsShell
}

func NewEntry(host *host.Host, routingDiscovery *discovery.RoutingDiscovery, f file.File, shell *file.IpfsShell) *Entry {
  rdv := f.String()
  p := peerstore.NewPeerstore(host, routingDiscovery, rdv)

  return &Entry{ store:*p, file:f, shell:shell }
}

func (e *Entry)InitEntry() error{
  err := e.shell.Dowload(e.file)
  if err != nil {
    return err
  }

  return mpi.Install(mpi.File(e.file))
}

func (e *Entry)LoadEntry(ctx context.Context, base protocol.ID) error {
  handler, err := mpi.Load(mpi.File(e.file))
  if err != nil {
    return err
  }

  discoveryHandler := func (p *peerstore.Peerstore, id peer.ID) {
		Protocol := protocol.ID(e.file.String() + "//" + string(base))
		stream, err := (*e.store.Host).NewStream(ctx, id, Protocol)

		if err != nil {
			return
		}

		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		e.store.Add(peer.IDB58Encode(id), func(str string) error{
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

  StreamHandler := func(stream network.Stream) {
		// Create a buffer stream for non blocking read and write.
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		go func(){
      for {
        str, err := rw.ReadString('\n')
    		if err != nil {
    			continue
    		}

        msg, err := mpi.FromString(str)
        if err != nil {
    			continue
    		}

        reps, err := (*handler)(msg)
        if err != nil {
    			continue
    		}

        for _, rep := range reps{
          if e.store.Has(rep.To){
            e.store.Write(rep.To, rep.String()) // pass on the responces
            continue
          }

          ID, err := peer.IDB58Decode(rep.To)
          if err != nil {
            continue
          }
          discoveryHandler(&e.store, ID)
          e.store.Write(rep.To, rep.String()) // pass on the responces
        }
      }
    }()
  }

  fmt.Println("store/utils.go/LoadEntry ~ SetHostId ...")
  e.store.SetHostId()
  fmt.Println("store/utils.go/LoadEntry ~ SetHostId v")

  err = e.store.SetStreamHandler(base, StreamHandler)
  fmt.Println("store/utils.go/LoadEntry ~ SetStreamHandler v")
  if err != nil {
    fmt.Println("store/utils.go/LoadEntry ~ SetStreamHandler x")
    return err
  }

  e.store.Listen(ctx, discoveryHandler)
  fmt.Println("store/utils.go/LoadEntry ~ discoveryHandler v")
  e.store.Annonce(ctx)
  fmt.Println("store/utils.go/LoadEntry ~ discoveryHandler v")
  return nil
}
