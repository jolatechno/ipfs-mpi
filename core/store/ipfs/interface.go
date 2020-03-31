package ipfs

import (
  "io/ioutil"
  "os"
  "path/filepath"
  "errors"
  "math/rand"

  shell "github.com/ipfs/go-ipfs-api"
)

const (
  ModePerm os.FileMode = 0777
  max_draw int = 1000
)

type File string

func (f *File)String() string {
  return string(*f)
}

type IpfsShell struct {
  Shell *shell.Shell
  Store []File
  path string
  ipfs_store string
}

func (s *IpfsShell)Add(f File) {
  s.Store = append(s.Store, f)
}

func NewShell(url string, path string, ipfs_store string) (*IpfsShell, error) {
  Shell := shell.NewShell(url)
  store := []File {}

  files, err := ioutil.ReadDir(path)
  if err != nil {
      return nil, err
  }

  for _, f := range files {
    store = append(store, File(f.Name()))
  }

  return &IpfsShell{ Shell:Shell, Store:store, path:path, ipfs_store:ipfs_store }, nil
}

func (s *IpfsShell)List() []File {
  return s.Store
}

func (s *IpfsShell)Has(f File) bool {
  for _, F := range s.Store {
    if f == F {
      return true
    }
  }
  return false
}

func (s *IpfsShell)Del(f File) error {
  err := os.Remove(s.path + f.String())
  if err != nil {
    return nil
  }

  for i, F := range s.Store {
    if f == F {
      s.Store = append(s.Store[:i], s.Store[i + 1:]...)
      return nil
    }
  }

  return errors.New("no file to delete")
}

func (s *IpfsShell)Dowload(f File) error {
  err := s.Shell.Get(s.ipfs_store + f.String(), s.path + f.String())
  if err != nil {
    return err
  }

  s.Add(f)
  return nil
}

func (s *IpfsShell)Occupied() (uint64, error) {
  var size uint64
  err := filepath.Walk(s.path, func(_ string, info os.FileInfo, err error) error {
      if err != nil {
          return err
      }
      if !info.IsDir() {
          size += uint64(info.Size())
      }
      return err
  })
  return size, err
}

func (s *IpfsShell)Get(maxSize uint64) (*File, error) {
  List, err := s.Shell.List(s.ipfs_store)
  if err != nil {
    return nil, err
  }

  for i := 0; i < max_draw ; i++ {
    n := rand.Intn(len(List))
    obj := List[n]

    if s.Has(File(obj.Name)) {
      continue
    }

    if obj.Size > maxSize {
      continue
    }

    f := File(obj.Name)
    return &f, nil
  }

  return nil, errors.New("exceded max draw")
}
