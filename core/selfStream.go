package core

import (
  "errors"
  "bytes"
	"time"

  "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
)

func NewStream(pid protocol.ID) SelfStream {
  return &CloseableBuffer {
    WriteBuffer: bytes.NewBuffer([]byte{}),
    ReadBuffer: bytes.NewBuffer([]byte{}),
    Closed: false,
		Pid: pid,
  }
}

type CloseableBuffer struct {
  WriteBuffer *bytes.Buffer
  ReadBuffer *bytes.Buffer
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
  return nil
}

func (b *CloseableBuffer)Read(p []byte) (n int, err error) {
	if b.Closed {
		return 0, errors.New("Stream closed")
	}
	return b.ReadBuffer.Read(p)
}

func (b *CloseableBuffer) Write(p []byte) (n int, err error) {
	if b.Closed {
		return 0, errors.New("Stream closed")
	}
	return b.WriteBuffer.Write(p)
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
