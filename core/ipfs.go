package core

import (
  "io/ioutil"
  "os"
  "path/filepath"
  "errors"
  "math/rand"
  "os/exec"
  "bytes"

  shell "github.com/ipfs/go-ipfs-api"
)

const (
  IpfsHeader = "IpfsStore"

  ModePerm os.FileMode = 0777

  InstalledHeader = "installed/"
  FailedHeader = "failed/"
)

type object struct {
  Name string
  Size uint64
}

func createEmptyFile(path string) {
  defer recover()

  file, _ := os.Create(path)
  file.Close()
}

func shellHas(List []*shell.LsLink, f string) bool {
  for _, obj := range List {
    if obj.Name == f {
      return true
    }
  }

  return false
}

func occupied(path string) (size uint64, err error) {
  err = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
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

func removeFromList(list []string, element string) []string {
  for i, name := range list {
    if name == element {
      return append(list[:i], list[i + 1:]...)
    }
  }

  return list
}

func has(list []string, element string) bool {
  for _, name := range list {
    if name == element {
      return true
    }
  }
  return false
}

//straight from version 1.0.1

func NewStore(url string, path string, ipfs_store string) (Store, error) {
  store := &IpfsShell {
    Path:path,
    Accessible: []object{},
    Store: []string{},
    Failed: []string{},
    IpfsStore:ipfs_store,
    Standard: NewStandardInterface(IpfsHeader),
  }

  defer func() {
    if err := recover(); err != nil {
      store.Raise(err.(error))
    }
  }()

  store.Shell = shell.NewShell(url)

  if _, err := os.Stat(path + InstalledHeader); os.IsNotExist(err) {
    new_err := os.MkdirAll(path + InstalledHeader, ModePerm)
    if new_err != nil{
      return nil, err
    }
  } else if err != nil {
    return nil, err
  }

  if _, err := os.Stat(path + FailedHeader); os.IsNotExist(err) {
    new_err := os.MkdirAll(path + FailedHeader, ModePerm)
    if new_err != nil{
      return nil, err
    }
  } else if err != nil {
    return nil, err
  }

  list_installed, err := ioutil.ReadDir(path + InstalledHeader)
  if err != nil {
      return nil, err
  }

  list_failed, err := ioutil.ReadDir(path + FailedHeader)
  if err != nil {
      return nil, err
  }

  List, err := store.Shell.List(ipfs_store)
  if err != nil {
    return nil, err
  }

  for _, file := range list_installed {
    f := file.Name()

    if shellHas(List, f) {
      store.Store = append(store.Store, f)
    } else {
      err := os.Remove(path + InstalledHeader + f)
      if err != nil {
        return nil, err
      }
    }

  }

  for _, file := range list_failed {
    f := file.Name()

    if shellHas(List, f) {
      store.Failed = append(store.Failed, f)
    } else {
      err := os.Remove(path + FailedHeader + f)
      if err != nil {
        return nil, err
      }
    }

  }

  for _, obj := range List {
    if !has(store.Store, obj.Name) {
      store.Accessible = append(store.Accessible, object {
        Name: obj.Name,
        Size: obj.Size,
      })
    }
  }

  return store, nil
}


type IpfsShell struct {
  Shell *shell.Shell
  Store []string
  Accessible []object
  Failed[] string
  Path string
  IpfsStore string
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
  for i, obj := range s.Accessible {
    if obj.Name == f {
      s.Accessible = append(s.Accessible[:i], s.Accessible[i + 1:]...)
      break
    }
  }
  if has(s.Failed, f) {
    s.Failed = removeFromList(s.Failed, f)
    os.Remove(s.Path + FailedHeader + f)
  }
}

func (s *IpfsShell)List() []string {
  return s.Store
}

func (s *IpfsShell)Has(f string) bool {
  return has(s.Store, f)
}

func (s *IpfsShell)Del(f string, failed bool) error {
  defer func() {
    if err := recover(); err != nil {
      s.Raise(err.(error))
    }
  }()

  if !s.Has(f){
    return errors.New("No file to delete")
  }

  err := os.Remove(s.Path + InstalledHeader + f)
  if err != nil {
    return NewHeadedError(err, IpfsHeader)
  }

  s.Store = removeFromList(s.Store, f)

  if failed {
    for i, obj := range s.Accessible {
      if obj.Name == f {
        s.Accessible = append(s.Accessible[:i], s.Accessible[i + 1:]...)
        break
      }
    }

    s.Failed = append(s.Failed, f)
    createEmptyFile(s.Path + FailedHeader + f)

  } else {
    List, err := s.Shell.List(s.IpfsStore)
    if err != nil {
      return  NewHeadedError(err, IpfsHeader)
    }

    if shellHas(List, f) {
      size, err := occupied(s.Path + InstalledHeader + f)
      if err != nil {
        return NewHeadedError(err, IpfsHeader)
      }

      s.Accessible = append(s.Accessible, object {
        Name: f,
        Size: size,
      })
    }

  }

  return nil
}

func (s *IpfsShell)Dowload(f string) error {
  defer func() {
    if err := recover(); err != nil {
      s.Del(f, true)
      s.Raise(err.(error))
    }
  }()

  err := s.Shell.Get(s.IpfsStore + f, s.Path + InstalledHeader + f)
  if err != nil {
    s.Del(f, true)
    return NewHeadedError(err, IpfsHeader)
  }

  cmd := exec.Command("python3", s.Path + InstalledHeader + f + "/init.py")

  var errorBuffer bytes.Buffer
  cmd.Stderr = &errorBuffer

  err = cmd.Run()

  strError := errorBuffer.String()
  if strError != "" {
    if strError[len(strError) - 1] == '\n' {
      strError = strError[:len(strError) - 1]
    }
    s.Del(f, true)
    return NewHeadedError(errors.New(strError), IpfsHeader)
  }

  if err != nil {
    s.Del(f, true)
    return NewHeadedError(err, IpfsHeader)
  }

  s.Add(f)
  return nil
}

func (s *IpfsShell)Occupied() (uint64, error) {
  return occupied(s.Path)
}

func (s *IpfsShell)Get(maxSize uint64) (string, error) {
  Choices := []string{}
  for _, obj := range s.Accessible {
    if obj.Size <= maxSize {
      Choices = append(Choices, obj.Name)
    }
  }

  if len(Choices) == 0 {
    return "", errors.New("No file with a small enough size")
  }

  n := rand.Intn(len(Choices))

  return Choices[n], nil
}
