package file

import (
  "github.com/coreos/go-semver/semver"
  "github.com/jolatechno/ipfs-mpi/core/mpi-interface"
)

type File struct {
  name string
  version semver.Version
}

func (f *File)String(){
  return f.name
}

func List() []File {
  //list all downloaded file
  return nil
}

func Has(f File) bool {
  //check if file is downloaded
  return false
}

func Dowload(f File) error {
  //Download the file
  return nil
}

func Free() int {
  //read how much space is left
  return -1
}

func Get() File {
  // Get a random program from ipfs
  return File{}
}
