package core

import (
  "net"
  "strings"
  "errors"
  "context"
  "time"
	"fmt"
  "math/rand"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/event"
  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/peerstore"
  "github.com/libp2p/go-libp2p-core/peer"
  "github.com/libp2p/go-libp2p-core/network"
  "github.com/libp2p/go-libp2p-core/connmgr"
  "github.com/libp2p/go-libp2p-peerstore/pstoremem"

  maddr "github.com/multiformats/go-multiaddr"
)

func ListIpAdresses() ([]maddr.Multiaddr, error) {
  returnAddr := []maddr.Multiaddr{}
	addr, err := maddr.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
  if err != nil {
  	return returnAddr, err
  }

	returnAddr = append(returnAddr, addr)

	addrs, err := net.InterfaceAddrs()
  if err != nil {
    return returnAddr, err
  }

  for _, a := range addrs {
    if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
      block := strings.Split(a.String(), "/")
      if len(block) > 2 {
        return returnAddr, errors.New("Ip adress with too many slash")
      }

      if ipnet.IP.To4() != nil {
        addr, err := maddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/0", block[0]))
        if err != nil {
          return returnAddr, err
        }
        returnAddr = append(returnAddr, addr)
      } else {
        addr, err := maddr.NewMultiaddr(fmt.Sprintf("/ip6/%s/tcp/0", block[0]))
        if err != nil {
          return returnAddr, err
        }
        returnAddr = append(returnAddr, addr)
      }
    }
  }

	return returnAddr, nil
}

func NewHost(ctx context.Context) (ExtHost, error) {
  var nilHost *BasicExtHost

  listenAddresses, err := ListIpAdresses()
  if err != nil {
    return nilHost, err
  }

  priv, _, err := crypto.GenerateKeyPair(
	   crypto.Ed25519, // Select your key type. Ed25519 are nice short
	    -1,             // Select key length when possible (i.e. RSA).
  )
  if err != nil {
    return nilHost, err
  }

  h, err := libp2p.New(ctx,
  	libp2p.Identity(priv),

  	libp2p.ListenAddrs(
      listenAddresses...
  	),
  )

  return &BasicExtHost {
    Ctx: ctx,
    Host: h,
    EndChan: make(chan bool),
    Ended: false,
    PeerStores: make(map[protocol.ID]peerstore.Peerstore),
  }, err
}

type BasicExtHost struct {
  Ctx context.Context
  Host host.Host
  EndChan chan bool
  Ended bool
  PeerStores map[protocol.ID]peerstore.Peerstore
}

func (h *BasicExtHost) Close() error {
  h.EndChan <- true
  h.Ended = true
  return h.Host.Close()
}

func (h *BasicExtHost)CloseChan() chan bool {
  return h.EndChan
}

func (h *BasicExtHost) Check() bool {
  return !h.Ended
}

func (h *BasicExtHost)Listen(pid protocol.ID, rendezvous string) {
  h.PeerStores[pid] = pstoremem.NewPeerstore()
  h.PeerStores[pid].AddAddrs(h.ID(), h.Addrs(), peerstore.TempAddrTTL)

  go func() {
    for h.Check() {
      peerChan := initMDNS(h.Ctx, h.Host, rendezvous)
      for {
        select {
        case peer := <- peerChan:
          h.PeerStores[pid].AddAddrs(peer.ID, peer.Addrs, peerstore.TempAddrTTL)
          go func(){
            h.Connect(h.Ctx, peer)
          }()
        case <- time.After(ScanDuration):
          continue
        }
      }
    }
  }()
}

func (h *BasicExtHost)PeerstoreProtocol(base protocol.ID) (peerstore.Peerstore, error) {
  pstore, ok := h.PeerStores[base]
  if !ok {
    return pstore, errors.New("no such protocol")
  }

  return pstore, nil
}

func (h *BasicExtHost)NewPeer(base protocol.ID) (peer.ID, error) {
  var nilPeerId peer.ID

  pstore, err := h.PeerstoreProtocol(base)
  if err != nil {
    return nilPeerId, err
  }

  peers := pstore.Peers()
  if len(peers) == 0 {
    return nilPeerId, errors.New("no peers supporting this protocol")
  }

  n := rand.Intn(len(peers))
  return peers[n], nil
}

func (h *BasicExtHost)ID() peer.ID {
  return h.Host.ID()
}

func (h *BasicExtHost)Peerstore() peerstore.Peerstore {
  return h.Host.Peerstore()
}

func (h *BasicExtHost)Addrs() []maddr.Multiaddr {
  return h.Host.Addrs()
}

func (h *BasicExtHost)Network() network.Network {
  return h.Host.Network()
}

func (h *BasicExtHost)Mux() protocol.Switch {
  return h.Host.Mux()
}

func (h *BasicExtHost)Connect(ctx context.Context, pi peer.AddrInfo) error {
  return h.Host.Connect(ctx, pi)
}

func (h *BasicExtHost)SetStreamHandler(pid protocol.ID, handler network.StreamHandler) {
  h.Host.SetStreamHandler(pid, handler)
}

func (h *BasicExtHost)SetStreamHandlerMatch(pid protocol.ID, match func(string) bool, handler network.StreamHandler) {
  h.Host.SetStreamHandlerMatch(pid, match, handler)
}

func (h *BasicExtHost)RemoveStreamHandler(pid protocol.ID) {
  h.Host.RemoveStreamHandler(pid)
}

func (h *BasicExtHost)NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error) {
  return h.Host.NewStream(ctx, p, pids...)
}

func (h *BasicExtHost)ConnManager() connmgr.ConnManager {
  return h.Host.ConnManager()
}

func (h *BasicExtHost)EventBus() event.Bus {
  return h.Host.EventBus()
}
