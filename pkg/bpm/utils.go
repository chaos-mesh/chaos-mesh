package bpm

import (
	"bytes"
	"io"
	"sync"
)

type blockingBuffer struct {
	buf    bytes.Buffer
	cond   *sync.Cond
	closed bool
}

func NewBlockingBuffer() *blockingBuffer {
	m := sync.Mutex{}
	return &blockingBuffer{
		cond:   sync.NewCond(&m),
		buf:    bytes.Buffer{},
		closed: false,
	}
}

func (br *blockingBuffer) Write(b []byte) (ln int, err error) {
	if br.closed {
		return 0, nil
	}
	ln, err = br.buf.Write(b)
	br.cond.Broadcast()
	return
}

func (br *blockingBuffer) Read(b []byte) (ln int, err error) {
	if br.closed {
		return 0, io.EOF
	}
	ln, err = br.buf.Read(b)
	for err == io.EOF {
		br.cond.L.Lock()
		if br.closed {
			return 0, io.EOF
		}
		br.cond.Wait()
		br.cond.L.Unlock()
		ln, err = br.buf.Read(b)
	}
	return
}

func (br *blockingBuffer) Close() {
	br.closed = true
	br.cond.Broadcast()
}
