package core

import (
  "errors"
	"time"
  "io"

  "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"

  "github.com/jolatechno/go-timeout"
)

var (
  StreamClosed = errors.New("Stream closed")
  StandardTimeout = time.Hour
)

func NewStream(pid protocol.ID) SelfStream {
  readPipe, writePipe := io.Pipe()
  readPipeReversed, writePipeReversed := io.Pipe()
  return &CloseableBuffer {
    WritePipe: writePipe,
    ReadPipe: readPipe,
    WritePipeReversed: writePipeReversed,
    ReadPipeReversed: readPipeReversed,
    WriteTimeout: StandardTimeout,
    ReadTimeout: StandardTimeout,
    Closed: false,
		Pid: pid,
  }
}

type CloseableBuffer struct {
  WritePipe *io.PipeWriter
  ReadPipe *io.PipeReader
  WritePipeReversed *io.PipeWriter
  ReadPipeReversed *io.PipeReader
  WriteTimeout time.Duration
  ReadTimeout time.Duration
  Closed bool
	Pid protocol.ID
}

func (b *CloseableBuffer)Reverse() (SelfStream, error) {
  if b.Closed {
		return nil, StreamClosed
	}
  return &CloseableBuffer {
    WritePipe: b.WritePipeReversed,
    ReadPipe: b.ReadPipeReversed,
    WritePipeReversed: b.WritePipe,
    ReadPipeReversed: b.ReadPipe,
    WriteTimeout: b.ReadTimeout,
    ReadTimeout: b.WriteTimeout,
    Closed: false,
		Pid: b.Pid,
  }, nil
}

func (b *CloseableBuffer)Close() error {
  b.Closed = true
  return nil
}

func (b *CloseableBuffer)SetProtocol(pid protocol.ID) {
  b.Pid = pid
}

func (b *CloseableBuffer)Protocol() protocol.ID {
	return b.Pid
}

func (b *CloseableBuffer)Reset() error {
  if b.Closed {
		return StreamClosed
	}

  b.ReadPipe, b.WritePipe = io.Pipe()
  b.ReadPipeReversed, b.WritePipeReversed = io.Pipe()
  b.WriteTimeout = StandardTimeout
  b.ReadTimeout = StandardTimeout
  return nil
}

func (b *CloseableBuffer)Read(p []byte) (int, error) {
  if b.Closed {
		return 0, StreamClosed
	}

  n, err := timeout.MakeTimeout(func() (interface{}, error) {
    return b.ReadPipe.Read(p)
  }, b.ReadTimeout)

  if n == nil {
    n = 0
  }

  return n.(int), err
}

func (b *CloseableBuffer) Write(p []byte) (int, error) {
  if b.Closed {
		return 0, StreamClosed
	}

  n, err := timeout.MakeTimeout(func() (interface{}, error) {
    return b.WritePipeReversed.Write(p)
  }, b.WriteTimeout)

  if n == nil {
    n = 0
  }

  return n.(int), err
}

func (b *CloseableBuffer)Stat() network.Stat {
	return network.Stat{}
}

func (b *CloseableBuffer)Conn() network.Conn {
	return nil
}

func (b *CloseableBuffer)SetDeadline(time.Time) error {
	return nil
}

func (b *CloseableBuffer)SetReadDeadline(time.Time) error {
  return nil
}

func (b *CloseableBuffer)SetWriteDeadline(time.Time) error {
  return nil
}
