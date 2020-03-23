package file

import (
  "io/ioutil"
  "os"
  "path/filepath"

  "github.com/coreos/go-semver/semver"
  "github.com/jolatechno/ipfs-mpi/core/mpi-interface"

  shell "github.com/ipfs/go-ipfs-api"
)

const (
  base_path = "./interpretors/"
  ModePerm os.FileMode = 0777
)

type File struct {
  name string
  version semver.Version
}

func (f *File)String(){
  return f.name + "/" + f.version.String
}

type IpfsShell struct {
  shell *shell.Shell
  store map[string][]version
}

func (s *IpfsShell)Add(f File) {
  _, ok := s.store[f.name]

  if !ok {
    s.store[f.name] = []version
  }

  s.store[f.name] = append(s.store[f.name], f.version)
}

func NewShell(url string) (IpfsShell, error) {
  Shell := shell.NewShell(url)

  store := make(map[string][]semver.version)

  files, err := ioutil.ReadDir(base_path)
  if err != nil {
      return Files, err
  }

  for _, f := range files {
    f_name := f.Name()
    versions, err := ioutil.ReadDir(base_path + f_name)
    if err != nil {
      continue
    }

    s.store[f.name] = []version

    for _, v := range versions {
      version, err := semver.NewVersion(v.Name())
      if err != nil{
        continue
      }

      store[f.name] = append(store[f.name], version)
    }
  }

  return IpfsShell{ shell:Shell, sore:store }
}

func (s *IpfsShell)List() []File {
  list := []File{}

  for name, versions := range s.store {
    for _, vers := range f {
      f := File{ name:name, version:vers }
      list = append(list, new_file)
    }
  }
  return list
}

func (s *IpfsShell)Has(f File) bool {
  versions, ok := s.store[f.name]
  if !ok {
    return false
  }

  for _, vers := range versions {
    if version.Major == f.version.Major && version.Minor >= f.version.Minor {
      return true
    }
  }
  return false
}

func (s *IpfsShell)Dowload(f File) error {
  if _, err := os.Stat(base_path + f.name); os.IsNotExist(err) {
    new_err := os.Mkdir(base_path + f.name, ModePerm)
    if err != nil{
      return err
    }
  } else if err != nil {
    return err
  }

  err := Shell.Get(f.name, base_path + f.name + "/" + f.version.String())
  if err != nil {
    return err
  }

  s.Add(f)
  return nil
}

func (s *IpfsShell)Occupied() (int64, err) {
  var size int64
  err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
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
