package core

import (
  "errors"
	"time"
  "bytes"

  "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"

  "github.com/jolatechno/go-timeout"
)

var (
  StandardTimeout = time.Hour
  ThrottleDuration = 70 * time.Nanosecond
)

func NewStream(pid protocol.ID) SelfStream {
  return &CloseableBuffer {
    WriteBuffer: bytes.NewBuffer([]byte{}),
    ReadBuffer: bytes.NewBuffer([]byte{}),
    WriteTimeout: StandardTimeout,
    ReadTimeout: StandardTimeout,
    Closed: false,
		Pid: pid,
  }
}

type CloseableBuffer struct {
  WriteBuffer *bytes.Buffer
  ReadBuffer *bytes.Buffer
  WriteTimeout time.Duration
  ReadTimeout time.Duration
  Closed bool
	Pid protocol.ID
}

func (b *CloseableBuffer)Reverse() (SelfStream, error) {
  if b.Closed {
		return nil, errors.New("Stream closed")
	}
  return &CloseableBuffer {
    WriteBuffer: b.ReadBuffer,
    ReadBuffer: b.WriteBuffer,
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
		return errors.New("Stream closed")
	}

  b.WriteBuffer = bytes.NewBuffer([]byte{})
  b.ReadBuffer = bytes.NewBuffer([]byte{})
  b.WriteTimeout = StandardTimeout
  b.ReadTimeout = StandardTimeout
  return nil
}

func (b *CloseableBuffer)Read(p []byte) (int, error) {
  if b.Closed {
    return 0, errors.New("Stream closed")
  }

  done := false
  n, err := timeout.MakeTimeout(func() (interface{}, error) {
    for !done {
      if b.ReadBuffer.Len() > len(p) {
        return b.ReadBuffer.Read(p)
      }
      time.Sleep(ThrottleDuration)
    }

    return nil, timeout.TimeOut
  }, b.ReadTimeout)

  done = true

  if n == nil {
    n = 0
  }

  return n.(int), err
}

func (b *CloseableBuffer) Write(p []byte) (int, error) {
  if b.Closed {
    return 0, errors.New("Stream closed")
  }

  n, err := timeout.MakeTimeout(func() (interface{}, error) {
    return b.WriteBuffer.Write(p)
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
