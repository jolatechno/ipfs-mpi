package compute

import (
  "C"
  "io"
)

type Request C.MPI_Request
type Status C.MPI_Status
type Datatype int8 //MPI type identifier

type message struct {
  data []byte
  count int
  datatype Datatype
  id int8
  status Status
}

func Write(msg message) (io.Reader, error) {
  //TODO
}

func Read(io.Reader) (message, error) {
  //TODO
}

func Exec(key string, stream io.ReadWriter) error {
  //TODO
}
