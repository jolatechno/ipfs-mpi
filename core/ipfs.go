package core

import (
  //"io/ioutil"
  "os"
  /*"path/filepath"
  "errors"
  "math/rand"

  "github.com/coreos/go-semver/semver"

  shell "github.com/ipfs/go-ipfs-api"*/
)

const (
  ModePerm os.FileMode = 0777
  max_draw int = 1000
)

//straight from version 1.0.1

/*
type IpfsShell struct {
  Shell *shell.Shell
  Store []string
  path string
  ipfs_store string
}

func (s *IpfsShell)Add(f string) {
  s.Store = append(s.Store, f)
}

func NewShell(url string, path string, ipfs_store string) (*IpfsShell, error) {
  Shell := shell.NewShell(url)

  store := make(map[string][] *semver.Version)

  files, err := ioutil.ReadDir(path)
  if err != nil {
      return nil, err
  }

  for _, f := range files {
    f_name := f.Name()
    versions, err := ioutil.ReadDir(path + f_name)
    if err != nil {
      continue
    }

    store[f_name] = [] *semver.Version{}

    for _, v := range versions {
      version, err := semver.NewVersion(v.Name())
      if err != nil{
        continue
      }

      store[f_name] = append(store[f_name], version)
    }
  }

  return &IpfsShell{ Shell:Shell, Store:store, path:path, ipfs_store:ipfs_store }, nil
}

func (s *IpfsShell)List() []string {
  return s.Store
}

func (s *IpfsShell)Has(f string) bool {
  versions, ok := s.Store[f.Name]
  if !ok {
    return false
  }

  for _, vers := range versions {
    if vers.Major == f.Version.Major && vers.Minor >= f.Version.Minor {
      return true
    }
  }
  return false
}

func (s *IpfsShell)Del(f string) error {
  if !s.Has(f){
    return errors.New("No file to delete")
  }

  err := os.Remove(s.path + f.String())
  if err != nil {
    return nil
  }

  for idx, vers := range s.Store[f.Name] {
    if vers == f.Version {
      s.Store[f.Name] = append(s.Store[f.Name][:idx], s.Store[f.Name][idx+1:]...)
      break
    }
  }

  if len(s.Store[f.Name]) == 0 {
    delete(s.Store, f.Name)
  }

  return nil
}

func (s *IpfsShell)Dowload(f string) error {
  if _, err := os.Stat(s.path + f.Name); os.IsNotExist(err) {
    new_err := os.MkdirAll(s.path + f.Name, ModePerm)
    if new_err != nil{
      return err
    }
  } else if err != nil {
    return err
  }

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

func (s *IpfsShell)Get(maxSize uint64) (string, error) {
  List, err := s.Shell.List(s.ipfs_store)
  if err != nil {
    return nil, err
  }

  for i := 0; i < max_draw ; i++ {
    n := rand.Intn(len(List))
    obj := List[n]

    _, ok := s.Store[obj.Name]
    if ok {
      continue
    }

    list, err := s.Shell.List(s.ipfs_store + obj.Name)
    if len(list) == 0 && err != nil {
      continue
    }

    f := list[len(list) - 1]
    if f.Size > maxSize {
      continue
    }

    vers, err := semver.NewVersion(f.Name)
    if err != nil {
      continue
    }


  }

  return nil, errors.New("exceded max draw")
}
*/
