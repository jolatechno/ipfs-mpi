package store

import (
  "strings"
  
  maddr "github.com/multiformats/go-multiaddr"
)

type AddrList []maddr.Multiaddr

func (al *AddrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

type Config struct {
	Url string
	Path string
	Ipfs_store string
	BootstrapPeers AddrList
	ListenAddresses AddrList
	ProtocolID string
	Maxsize uint64
	Api_port int
	WriteTimeout int
	ReadTimeout int
}

func (al *AddrList) Set(value string) error {
	addr, err := maddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}

func StringsToAddrs(addrStrings []string) (maddrs []maddr.Multiaddr, err error) {
	for _, addrString := range addrStrings {
		addr, err := maddr.NewMultiaddr(addrString)
		if err != nil {
			return maddrs, err
		}
		maddrs = append(maddrs, addr)
	}
	return
}
