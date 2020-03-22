package compute

import (
  "io"
  "strings"

  "github.com/jolatecno/ipfs-mpi-store/src"
  "github.com/jolatechno/peerstore"
)

type addr string
type device int8

const key_length int = 46
const None addr = strings.repeat("0", length)

type offer struct {
  header string
  address addr
}

func ListOffers() ([]offer, error){
  keys, err := store.list()

  if err != nil {
    return nil, err
  }

  for key := range keys {
    //TODO
  }
}
