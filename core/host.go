package core

import (
  "net"
  "strings"
  "errors"
  "context"
	"fmt"
  "time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
  "github.com/libp2p/go-libp2p-core/host"
  libp2ptls "github.com/libp2p/go-libp2p-tls"
  secio "github.com/libp2p/go-libp2p-secio"
  connmgr "github.com/libp2p/go-libp2p-connmgr"

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

func NewHost(ctx context.Context) (host.Host, error) {
  var nilHost host.Host

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

  return libp2p.New(ctx,
  	libp2p.Identity(priv),

  	libp2p.ListenAddrs(
      listenAddresses...
  	),

  	libp2p.Security(libp2ptls.ID, libp2ptls.New),
  	libp2p.Security(secio.ID, secio.New),
  	// Let's prevent our peer from having too many
  	// connections by attaching a connection manager.
  	libp2p.ConnectionManager(connmgr.NewConnManager(
  		100,         // Lowwater
  		400,         // HighWater,
  		time.Minute, // GracePeriod
  	)),
  )
}
