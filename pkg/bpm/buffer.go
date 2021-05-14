// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package bpm

import (
	"bytes"
	"context"
	"io"
	"sync"
)

type blockingBuffer struct {
	buf    bytes.Buffer
	cond   *sync.Cond
	closed bool
}

type concurrentBuffer struct {
	ctx    context.Context
	cancel context.CancelFunc

	buf       bytes.Buffer
	writeChan chan []byte
	readChan  chan *readCtx
}

type readCtx struct {
	ret  chan readRet
	data []byte
}

type readRet struct {
	ln  int
	err error
}

func NewBlockingBuffer() io.ReadWriteCloser {
	m := sync.Mutex{}
	return &blockingBuffer{
		cond:   sync.NewCond(&m),
		buf:    bytes.Buffer{},
		closed: false,
	}
}

func NewConcurrentBuffer() io.ReadWriteCloser {
	ctx, cancel := context.WithCancel(context.Background())

	buffer := &concurrentBuffer{
		ctx:       ctx,
		cancel:    cancel,
		writeChan: make(chan []byte),
		readChan:  make(chan *readCtx),
	}

	go buffer.start()
	return buffer
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

func (br *blockingBuffer) Close() error {
	br.closed = true
	br.cond.Broadcast()
	return nil
}

func (cb *concurrentBuffer) start() {
	for {
		if cb.buf.Len() == 0 {
			select {
			case data := <-cb.writeChan:
				cb.buf.Write(data)
				continue
			case <-cb.ctx.Done():
				return
			}
		}

		select {
		case data := <-cb.writeChan:
			cb.buf.Write(data)
		case <-cb.ctx.Done():
			return
		case ctx := <-cb.readChan:
			ln, err := cb.buf.Read(ctx.data)
			ctx.ret <- readRet{
				ln:  ln,
				err: err,
			}
		}
	}
}

func (cb *concurrentBuffer) Write(b []byte) (ln int, err error) {
	select {
	case <-cb.ctx.Done():
		return 0, nil
	case cb.writeChan <- b:
		return len(b), nil
	}
}

func (cb *concurrentBuffer) Read(b []byte) (ln int, err error) {
	ret := make(chan readRet)

	select {
	case <-cb.ctx.Done():
		return 0, io.EOF
	case cb.readChan <- &readCtx{
		ret:  ret,
		data: b,
	}:
		select {
		case <-cb.ctx.Done():
			return 0, io.EOF
		case result := <-ret:
			return result.ln, result.err
		}
	}
}

func (cb *concurrentBuffer) Close() error {
	cb.cancel()
	return nil
}
