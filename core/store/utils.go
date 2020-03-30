package store

import (
  "context"
  "bufio"
  "fmt"

  "github.com/jolatechno/ipfs-mpi/core/ipfs-interface"
  "github.com/jolatechno/ipfs-mpi/core/messagestore"
  "github.com/jolatechno/mpi-peerstore"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/peer"
  "github.com/libp2p/go-libp2p-core/network"
)

func (s *Store)Add(f file.File, ctx context.Context) error {
  s.Shell.Dowload(f)
  err := message.Install(s.Path + f.String())
  if err != nil {
    return err
  }

  return s.Load(f, ctx)
}

func (s *Store)Load(f file.File, ctx context.Context) error {
  p := peerstore.NewPeerstore(s.Host, s.RoutingDiscovery, f.String())

  hostId := peer.IDB58Encode((*s.Host).ID())
  err := p.SetStreamHandler(s.Protocol, func(stream network.Stream) {
		// Create a buffer stream for non blocking read and write.
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		go func(){
      for {

        fmt.Println("Load go 0") //------------------------------------------------------------------------

        str, err := rw.ReadString('\n')
    		if err != nil {
    			continue
    		}

        fmt.Println("Load go 1, msg : ", str) //------------------------------------------------------------------------

        msg, err := message.FromString(str[:len(str) - 1])
        if err != nil {
    			continue
    		}

        fmt.Println("Load go 2") //------------------------------------------------------------------------

        if msg.Origin == hostId {

          fmt.Println("load go 3 push back") //------------------------------------------------------------------------

          (*s.Api).Push(*msg)
        } else {

          fmt.Println("load go 3 exec") //------------------------------------------------------------------------

          (*s.DaemonStore).Push(*msg)
        }
      }
    }()
  })
  if err != nil {
    return err
  }

  Protocol := protocol.ID(f.String() + "//" + string(s.Protocol))

  fmt.Print(Protocol)

  p.Listen(ctx, func (p *peerstore.Peerstore, id peer.ID) {
    stream, err := (*s.Host).NewStream(ctx, id, Protocol)
    if err != nil {
      return
    }

    rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

    p.Add(peer.IDB58Encode(id), func(str string) error{
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
  })
  p.Annonce(ctx)

  (*s.Store)[f.String()] = p

  return nil
}

func (s *Store)Init(ctx context.Context) error {
  files := (*s.Shell).List()

  for _, f := range files {
    err := s.Load(f, ctx)

    if err != nil {
      return err
    }
  }

  go func(){
    for{
      err := s.Get(ctx)
      if err != nil { //No new file to add
        return
      }
    }
  }()

  return nil
}

func (s *Store)Del(f file.File) error {
  return s.Shell.Del(f)
}

func (s *Store)Get(ctx context.Context) error {
  used, err := s.Shell.Occupied()
  if err != nil {
    return err
  }

  f, err := s.Shell.Get(s.Maxsize - used)
  if err != nil {
    return err
  }

  return s.Add(*f, ctx)
}
