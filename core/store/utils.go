package store

import (
  "strings"

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
  name strings
  version semver.Version
}

func NewEntry(host *host.Host, routingDiscovery *discovery.RoutingDiscovery, name string, version semver.Version) Entry{
  rdv := version + "/" + version.String()
  p := peerstore.NewPeerstore(host, routingDiscovery, rdv)
  return Entry{store:p, name:rdv, version:version }
}

func (e *Entry)InitEntry() error{
  return file.Dowload(e.name, e.version)
}

func (e *Entry)LoadEntry(base protocol.ID) error{
  handler, err := mpi.Load(e.name, e.version)

  if err != nil {
    return err
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
          e.store.Write(rep.to, rep.String()) // pass on the responces
        }
      }
    }

    p.SetStreamHandler(base, StreamHandler)

    return nil
}
