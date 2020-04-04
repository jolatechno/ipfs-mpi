package core

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

//straight from version 1.0.1

type IpfsShell struct {
  Shell *shell.Shell
  Store []string
  path string
  ipfs_store string
}

func (s *IpfsShell)Close() {
  
}

func (s *IpfsShell)Add(f string) {
  s.Store = append(s.Store, f)
}

func NewShell(url string, path string, ipfs_store string) (Store, error) {
  Shell := shell.NewShell(url)
  list, err := ioutil.ReadDir(path)
  if err != nil {
      return nil, err
  }

  store := make([]string, len(list))
  for i, file := range list {
    store[i] = file.Name()
  }

  return &IpfsShell{ Shell:Shell, Store:store, path:path, ipfs_store:ipfs_store }, nil
}

func (s *IpfsShell)List() []string {
  return s.Store
}

func (s *IpfsShell)Has(f string) bool {
  for _, name := range s.Store {
    if name == f {
      return true
    }
  }
  return false
}

func (s *IpfsShell)Del(f string) error {
  if !s.Has(f){
    return errors.New("No file to delete")
  }

  err := os.Remove(s.path + f)
  if err != nil {
    return nil
  }

  for i, name := range s.Store {
    if name == f {
      s.Store = append(s.Store[:i], s.Store[i + 1:]...)
      return nil
    }
  }

  return errors.New("File not in store")
}

func (s *IpfsShell)Dowload(f string) error {
  err := s.Shell.Get(s.ipfs_store + f, s.path + f)
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

func (s *IpfsShell)Get(maxSize uint64) (string, error) {
  List, err := s.Shell.List(s.ipfs_store)
  if err != nil {
    return "", err
  }

  for i := 0; i < max_draw ; i++ {
    n := rand.Intn(len(List))
    obj := List[n]

    if s.Has(obj.Name) {
      continue
    }

    if obj.Size > maxSize {
      continue
    }

    return obj.Name, nil
  }

  return "", errors.New("exceded max draw")
}
