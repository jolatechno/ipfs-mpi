package store

import (
  "github.com/jolatechno/mpi-peerstore"
  "github.com/jolatechno/mpi"
  mh "github.com/multiformats/go-multihash"
  "github.com/ipfs/go-cid/"
)


//constans

const (
  Msg = "Msg"
  Mh = "Mh"
)


//Handler


type Handler func(mpi.Message) error //mpi-handler

func Load(path string) (Handler, error){
  f, err := nil, nil//load mpi from path
  if err != nil {
    return nil, err
  }
  return f, nil
}


//StoreEntry


type StoreEntry struct {
  peerstore peerstore.Peerstore
  prgm *Handler
  path string
}

func (se *StoreEntry)Load() error{
  f, err := Load((*se).path)
  if err != nil {
    return err
  }

  (*se).prgm = &f
  return nil
}

func NewStoreEntry(addr mh.Multihash) (StoreEntry, error){
  c, err := cid.tryNewCidV0(addr)
  if err != nil {
    return nil, err
  }

  path := c.String()

  p := peerstore.NewPeerstore()

  err = nil //download the mpu from addr to path //TODO
  if err != nil {
    return nil, err
  }

  se := StoreEntry{ peerstore: p, prgm: nil, path:path }
  err = se.Load()
  if err != nil {
    return nil, err
  }

  return s, nil
}


//Store


type Store map[string]entry

func NewStore() Store {}
  return make(Store)
}

func (s *Store)Add(addr mh.Multihash) error{
  se, err := NewStoreEntry(addr)
  if err != nil {
    return err
  }

  (*s)[se.path] = se
  return nil
}

func (s *Store)Del(name string){
  delete(*s, name)
}

func (s *Store)Has(name string) bool{
  _, ok := (*s)[addr]
  return ok
}

func (s *Store)WriteAll(str string) {
  for addr, _ := range *s {
    s.Write(addr, str)
  }
}

func (s *Store)Write(addr string, str string) {
  for addr, _ := range *s {
    s.Write(addr, str)
  }
}

func (s *Store)WriteAllMh(addr mh.Multihash) {
  s.WriteAll(name, Mh + addr.String())
}

func (s *Store)WriteMessage(name string, msg mpi.Message) error {
  str, err := msg.String()
  if err != nil {
    return err
  }

  return s.WriteString(name, Msg + str)
}
