package file

import (
  "github.com/coreos/go-semver/semver"
  "github.com/jolatechno/ipfs-mpi/core/mpi-interface"
)

type File struct {
  addr string
  version semver.Version
}

func List() []File{
  //list all downloaded file
  return nil
}

func Has(addr string, version semver.Version) bool {
  //check if file is downloaded
  return false
}

func Dowload(addr string, version semver.Version) error {
  //Download the file
  return nil
}
