package file

import (
  "io/ioutil"
  "os"
  "path/filepath"

  "github.com/coreos/go-semver/semver"

  shell "github.com/ipfs/go-ipfs-api"
)

const (
  base_path = "./interpretors/"
  ModePerm os.FileMode = 0777
)

type File struct {
  Name string
  Version *semver.Version
}

func (f *File)String() string{
  return f.Name + "/" + f.Version.String()
}

func (f *File)GetHash() (string, error){
  // get the write version
  //TODO
  return f.Name, nil
}


type IpfsShell struct {
  Shell *shell.Shell
  Store map[string][] *semver.Version
}

func (s *IpfsShell)Add(f File) {
  _, ok := s.Store[f.Name]

  if !ok {
    s.Store[f.Name] = [] *semver.Version{}
  }

  s.Store[f.Name] = append(s.Store[f.Name], f.Version)
}

func NewShell(url string) (*IpfsShell, error) {
  Shell := shell.NewShell(url)

  store := make(map[string][] *semver.Version)

  files, err := ioutil.ReadDir(base_path)
  if err != nil {
      return nil, err
  }

  for _, f := range files {
    f_name := f.Name()
    versions, err := ioutil.ReadDir(base_path + f_name)
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

  return &IpfsShell{ Shell:Shell, Store:store }, nil
}

func (s *IpfsShell)List() []File {
  list := []File{}

  for name, versions := range s.Store {
    for _, vers := range versions {
      f := File{ Name:name, Version:vers }
      list = append(list, f)
    }
  }
  return list
}

func (s *IpfsShell)Has(f File) bool {
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

func (s *IpfsShell)Dowload(f File) error {
  if _, err := os.Stat(base_path + f.Name); os.IsNotExist(err) {
    new_err := os.Mkdir(base_path + f.Name, ModePerm)
    if new_err != nil{
      return err
    }
  } else if err != nil {
    return err
  }

  err := s.Shell.Get(f.GetHash(), base_path + f.Name + "/" + f.Version.String())
  if err != nil {
    return err
  }

  s.Add(f)
  return nil
}

func (s *IpfsShell)Occupied() (int64, error) {
  var size int64
  err := filepath.Walk(base_path, func(_ string, info os.FileInfo, err error) error {
      if err != nil {
          return err
      }
      if !info.IsDir() {
          size += info.Size()
      }
      return err
  })
  return size, err
}

func (s *IpfsShell)Get() File {
  //select a random interpretor file from IPFS
  //TODO
  return File{}
}
