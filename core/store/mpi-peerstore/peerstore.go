package peerstore

import (
  "errors"
  "context"
  "github.com/libp2p/go-libp2p-discovery"
  "github.com/libp2p/go-libp2p-core/peer"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/network"
)

type Peerstore struct{
  Store map[string] func(string) error
  Host *host.Host
  RoutingDiscovery *discovery.RoutingDiscovery
  RendezvousString string
  AddPeer *func(peer.ID)
}

func (p *Peerstore)Add(addr string, rw func(string) error){
  p.Store[addr] = rw
}

func (p *Peerstore)Del(addr string){
  delete(p.Store, addr)
}

func (p *Peerstore)Has(addr string) bool{
  _, ok := p.Store[addr]
  return ok
}

func (p *Peerstore)WriteAll(str string) {
  for addr, _ := range p.Store {
    p.Write(addr, str)
  }
}

func (p *Peerstore)Write(addr string, str string) error {
  rw, ok := p.Store[addr]
  if ok {
    err := rw(str)

    if err != nil {
      p.Del(addr)
      return err
    }
    return nil
  }
  return errors.New("no such peer")
}

func NewPeerstore(host *host.Host, routingDiscovery *discovery.RoutingDiscovery, RendezvousString string) *Peerstore {
  store := make(map[string] func(string) error)
  return &Peerstore{ Store:store, Host:host, RoutingDiscovery:routingDiscovery, RendezvousString:RendezvousString}
}

func (p *Peerstore)Annonce(ctx context.Context){
  discovery.Advertise(ctx, p.RoutingDiscovery, p.RendezvousString)
}

func (p *Peerstore)Listen(ctx context.Context, discoveryHandler func(*Peerstore, peer.ID)){
  addPeer := func (id peer.ID) {
    discoveryHandler(p, id)
  }

  p.AddPeer = &addPeer

  hostId := peer.IDB58Encode((*p.Host).ID())
  go func(){
    for {
  		peerChan, err := p.RoutingDiscovery.FindPeers(ctx, p.RendezvousString)
  		if err != nil {
  			continue
  		}

  		for Peer := range peerChan {
        str_id := peer.IDB58Encode(Peer.ID)
  			if p.Has(str_id) || hostId == str_id {
  				continue
  			}

        discoveryHandler(p, Peer.ID)
  		}
  	}
  }()
}

func (p *Peerstore)SetStreamHandler(base protocol.ID,  handleStream func(network.Stream)) error{
  Protocol := protocol.ID(p.RendezvousString + "//" + string(base))
  checker, err := MultistreamSemverMatcher(Protocol)
  if err != nil {
    return err
  }
  (*p.Host).SetStreamHandlerMatch(Protocol, checker, handleStream)
  return nil
}
