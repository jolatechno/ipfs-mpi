package core

import (
  "io/ioutil"
  "os"
  "path/filepath"
  "errors"
  "math/rand"
  "os/exec"

  shell "github.com/ipfs/go-ipfs-api"
)

const (
  IpfsHeader = "IpfsStore"

  ModePerm os.FileMode = 0777
  max_draw int = 1000
)

//straight from version 1.0.1

func NewStore(url string, path string, ipfs_store string) (Store, error) {
  store := &IpfsShell {
    Path:path,
    Ipfs_store:ipfs_store,
    Standard: NewStandardInterface(IpfsHeader),
  }

  defer func() {
    if err := recover(); err != nil {
      store.Raise(err.(error))
    }
  }()

  store.Shell = shell.NewShell(url)

  if _, err := os.Stat(path); os.IsNotExist(err) {
    new_err := os.MkdirAll(path, ModePerm)
    if new_err != nil{
      return nil, err
    }
  } else if err != nil {
    return nil, err
  }

  list, err := ioutil.ReadDir(path)
  if err != nil {
      return nil, err
  }

  store.Store = make([]string, len(list))
  for i, file := range list {
    store.Store[i] = file.Name()
  }

  return store, nil
}


type IpfsShell struct {
  Shell *shell.Shell
  Store []string
  Path string
  Ipfs_store string
  Standard standardFunctionsCloser
}

func (s *IpfsShell)Close() error {
  return s.Standard.Close()
}

func (s *IpfsShell)Check() bool {
  return s.Standard.Check()
}

func (s *IpfsShell)SetErrorHandler(handler func(error)) {
  s.Standard.SetErrorHandler(handler)
}

func (s *IpfsShell)SetCloseHandler(handler func()) {
  s.Standard.SetCloseHandler(handler)
}

func (s *IpfsShell)Raise(err error) {
  s.Standard.Raise(err)
}

func (s *IpfsShell)Add(f string) {
  s.Store = append(s.Store, f)
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
  defer func() {
    if err := recover(); err != nil {
      s.Raise(err.(error))
    }
  }()

  if !s.Has(f){
    return errors.New("No file to delete")
  }

  err := os.Remove(s.Path + f)
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
  defer func() {
    if err := recover(); err != nil {
      s.Raise(err.(error))
    }
  }()

  err := s.Shell.Get(s.Ipfs_store + f, s.Path + f)
  if err != nil {
    return err
  }

  err = exec.Command("python3", s.Path + f + "/init.py").Start()
  if err != nil {
    return err
  }

  s.Add(f)
  return nil
}

func (s *IpfsShell)Occupied() (uint64, error) {
  defer func() {
    if err := recover(); err != nil {
      s.Raise(err.(error))
    }
  }()

  var size uint64
  err := filepath.Walk(s.Path, func(_ string, info os.FileInfo, err error) error {
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
  List, err := s.Shell.List(s.Ipfs_store)
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
