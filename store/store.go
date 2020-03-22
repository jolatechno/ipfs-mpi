package store

import (
  "github.com/jolatechno/mpi-peerstore"
  "github.com/jolatechno/mpi"
  mh "github.com/multiformats/go-multihash"
  "github.com/ipfs/go-cid/"
)


//Handler


type Handler func(mpi.Message) error //mpi-handler


//StoreEntry


type StoreEntry struct {
  peerstore peerstore.Peerstore
  prgm *Handler
  path string
}

func (se *StoreEntry)Load() error{
  f = func(mpi.Message) error{
    //load mpi from (*S).path
    //TODO
    return nil
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

func (s *Store)Add(name string, addr mh.Multihash) error{
  se, err := NewStoreEntry(addr)
  if err != nil {
    return err
  }

  (*s)[name] = se
  return nil
}

func (s *Store)Del(name string){
  delete(*p, name)
}

func (p *Store)Has(name string) bool{
  _, ok := (*p)[addr]
  return ok
}

func (p *Peerstore)WriteAll(str string) {
  for addr, _ := range *p {
    p.Write(addr, str)
  }
}

func (s *Store)WriteStringAll(str string) {
  for name, _ := range *s {
    s.WriteString(name, str)
  }
}

func (s *Store)WriteString(name string, str string) error {
  p, ok := (*s)[name]
  if ok {
    p.WriteAll(str)
    return nil
  }
  return errors.New("no such peer")
}

func (s *Store)WriteAll(msg mpi.Message) {
  for name, _ := range *s {
    s.Write(name, msg)
  }
}

func (s *Store)Write(name string, msg mpi.Message) error {
  str, err := msg.ToString()
  if err != nil {
    return err
  }

  return s.WriteString(name, str)
}
