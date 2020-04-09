package core

import (
  "errors"
  "bytes"
	"time"

  "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
)

func NewStream(pid protocol.ID) network.Stream {
  return &CloseableBuffer {
    Buffer: *bytes.NewBuffer([]byte{}),
    Closed: false,
		Pid: pid,
  }
}

type CloseableBuffer struct {
  Buffer bytes.Buffer
  Closed bool
	Pid protocol.ID
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
	b.Buffer = *bytes.NewBuffer([]byte{})
	b.Closed = false
	return nil
}

func (b *CloseableBuffer)Read(p []byte) (n int, err error) {
	if b.Closed {
		return 0, errors.New("Stream closed")
	}
	return b.Buffer.Read(p)
}

func (b *CloseableBuffer) Write(p []byte) (n int, err error) {
	if b.Closed {
		return 0, errors.New("Stream closed")
	}
	return b.Buffer.Write(p)
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
