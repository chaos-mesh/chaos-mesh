// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package bpm

import (
	"bytes"
	"io"
	"sync"

	"go.uber.org/atomic"
)

type blockingBuffer struct {
	buf io.ReadWriteCloser

	cond   *sync.Cond
	closed *atomic.Bool
}

func NewBlockingBuffer() io.ReadWriteCloser {
	m := sync.Mutex{}
	return &blockingBuffer{
		cond:   sync.NewCond(&m),
		buf:    newInternalBuffer(),
		closed: atomic.NewBool(false),
	}
}

func (br *blockingBuffer) Write(b []byte) (ln int, err error) {
	if br.closed.Load() {
		return 0, nil
	}
	ln, err = br.buf.Write(b)

	br.cond.Broadcast()
	return
}

func (br *blockingBuffer) Read(b []byte) (ln int, err error) {
	if br.closed.Load() {
		return 0, io.EOF
	}
	ln, err = br.buf.Read(b)

	for err == io.EOF {
		br.cond.L.Lock()
		if br.closed.Load() {
			return 0, io.EOF
		}
		br.cond.Wait()
		br.cond.L.Unlock()

		ln, err = br.buf.Read(b)
	}
	return
}

func (br *blockingBuffer) Close() error {
	br.closed.Store(true)

	br.cond.Broadcast()

	br.buf.Close()
	return nil
}

// internalBuffer is a buffer with a lock
type internalBuffer struct {
	buf   bytes.Buffer
	mutex sync.Mutex
}

func newInternalBuffer() io.ReadWriteCloser {
	buffer := &internalBuffer{
		buf:   bytes.Buffer{},
		mutex: sync.Mutex{},
	}

	return buffer
}

func (cb *internalBuffer) Write(buf []byte) (ln int, err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	return cb.buf.Write(buf)
}

func (cb *internalBuffer) Read(buf []byte) (ln int, err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	return cb.buf.Read(buf)
}

func (cb *internalBuffer) Close() error {
	return nil
}
