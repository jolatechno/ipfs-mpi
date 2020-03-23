package store

import (
  "strings"
  "context"
  "bufio"

  "github.com/libp2p/go-libp2p-discovery"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/network"
  "github.com/libp2p/go-libp2p-core/peer"

  "github.com/jolatecno/mpi-peerstore"
  "github.com/jolatecno/ipfs-mpi/ipfs-interface"

  "github.com/coreos/go-semver/semver"
)

type Entry struct {
  store peerstore.Peerstore
  file file.File
}

func NewEntry(host *host.Host, routingDiscovery *discovery.RoutingDiscovery, f file;File) Entry{
  rdv := version + "/" + version.String()
  p := peerstore.NewPeerstore(host, routingDiscovery, rdv)
  return Entry{ store:p, file:f }
}

func (e *Entry)InitEntry() error{
  return file.Dowload(e.file)
}

func (e *Entry)LoadEntry(ctx context.Context, base protocol.ID) error{
  handler, err := mpi.Load(e.file)

  if err != nil {
    return err
  }

  discoveryHandler = func (p *peerstore.Peerstore, id peer.ID){
		Protocol := protocol.ID(e.file.String() + "//" + base.String())
		stream, err := (*e.store.host).NewStream(ctx, id, Protocol)

		if err != nil {
			return
		}

		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

    w := func(str string) error{
    	_, err := rw.WriteString(fmt.Sprintf("%s\n", str))
    	if err != nil {
    		return err
    	}
    	err = rw.Flush()
    	if err != nil {
    		return err
    	}

    	return nil
    }

		e.store.Add(peer.IDB58Encode(id), &w)
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

        reps, err := handler(msg)
        if err != nil {
    			continue
    		}

        for _, rep := range reps{
          if e.store.Has(rep.to){
            e.store.Write(rep.to, rep.String()) // pass on the responces
            continue
          }

          discoveryHandler(e.store, rep.to)
          e.store.Write(rep.to, rep.String()) // pass on the responces
        }
      }
    }
  }

  e.store.SetHostId()
  e.store.SetStreamHandler(base, StreamHandler)
  e.store.Listen(ctx, discoveryHandler)
  e.store.Annonce(ctx)

  return nil
}