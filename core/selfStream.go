package core

import (
  "errors"
	"time"
  "io"
  "sync"

  "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"

  "github.com/jolatechno/go-timeout"
)

var (
  StreamEnded = errors.New("Stream closed")
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
  Ended bool
  Mutex sync.Mutex
	Pid protocol.ID
}

func (b *CloseableBuffer)Reverse() (SelfStream, error) {
  if b.Ended {
		return nil, StreamEnded
	}
  return &CloseableBuffer {
    WritePipe: b.WritePipeReversed,
    ReadPipe: b.ReadPipeReversed,
    WritePipeReversed: b.WritePipe,
    ReadPipeReversed: b.ReadPipe,
    WriteTimeout: b.ReadTimeout,
    ReadTimeout: b.WriteTimeout,
    Ended: false,
		Pid: b.Pid,
  }, nil
}

func (b *CloseableBuffer)check() bool {
  b.Mutex.Lock()
  defer b.Mutex.Unlock()

  return !b.Ended
}

func (b *CloseableBuffer)Close() error {
  b.Mutex.Lock()
  defer b.Mutex.Unlock()

  b.Ended = true
  return nil
}

func (b *CloseableBuffer)SetProtocol(pid protocol.ID) {
  b.Pid = pid
}

func (b *CloseableBuffer)Protocol() protocol.ID {
	return b.Pid
}

func (b *CloseableBuffer)Reset() error {
  if !b.check() {
		return StreamEnded
	}

  b.ReadPipe, b.WritePipe = io.Pipe()
  b.ReadPipeReversed, b.WritePipeReversed = io.Pipe()
  b.WriteTimeout = StandardTimeout
  b.ReadTimeout = StandardTimeout
  return nil
}

func (b *CloseableBuffer)Read(p []byte) (int, error) {
  if !b.check() {
		return 0, StreamEnded
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
  if !b.check() {
		return 0, StreamEnded
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
