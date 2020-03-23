package file

import (
  "io/ioutil"

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

func List() ([]File, error) {
  Files := []File{}

  files, err := ioutil.ReadDir("./interpretors")
  if err != nil {
      return Files, err
  }

  for _, f := range files {
    f_name := f.Name()
    versions, err := ioutil.ReadDir("./interpretors/" + f_name)
    if err != nil {
      continue
    }

    for _, v := range versions {
      version, err := semver.NewVersion(v.Name())
      if err != nil{
        continue
      }

      newFile := File{ name:f_name, version:version }
      Files = append(Files, newFile)
    }
  }
  return Files, nil
}

func Has(f File) bool {
  files, err := ioutil.ReadDir("./interpretors")
  if err != nil {
      return false
  }

  for _, F := range files {
        if F.Name() == f.name {
          versions, err := ioutil.ReadDir("./interpretors/" + F.Name())
          if err != nil {
            return false
          }

          for _, v := range versions {
            version, err := semver.NewVersion(v.Name())
            if err != nil{
              continue
            }

            if version.Major == f.version.Major && version.Minor >= f.version.Minor {
              return true
            }
          }
          return false
        }
    }
    return false
}

func Dowload(f File) error {
  //Download the file
  return nil
}

func Occupied() (int64, err) {
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

func Get() File {
  // Get a random program from ipfs
  return File{}
}
