// Copyright 2020 Chaos Mesh Authors.
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
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/buffer"
)

var _ = Describe("concurrent buffer", func() {
	var testLines = map[string]bool{
		"a":     true,
		"ab":    true,
		"abc":   true,
		"abcd":  true,
		"abcde": true,
		"b":     true,
		"bc":    true,
		"bcd":   true,
		"bcde":  true,
	}

	const seperator = "\r\n"
	const repeated = 100
	const workers = 10
	const testTimes = 10

	testSequentially := func() {
		linesChan := makeLinesChain(testLines, repeated)
		buffer := NewConcurrentBuffer()
		Expect(writeBuffer(linesChan, buffer)).To(BeNil())
		result, err := readTimeout(buffer, time.Second)
		Expect(err).To(BeNil())
		check(testLines, result, seperator, repeated)
	}

	testConcurrently := func() {
		linesChan := makeLinesChain(testLines, repeated)
		buffer := NewConcurrentBuffer()
		for i := 0; i < workers; i++ {
			go func() {
				defer GinkgoRecover()
				Expect(writeBuffer(linesChan, buffer)).To(BeNil())
			}()
		}
		result, err := readTimeout(buffer, time.Second)
		Expect(err).To(BeNil())
		check(testLines, result, seperator, repeated)
	}

	multiple := func(fn func()) func() {
		return func() {
			wg := sync.WaitGroup{}
			for i := 0; i < testTimes; i++ {
				wg.Add(1)
				go func() {
					defer GinkgoRecover()
					fn()
					wg.Done()
				}()
			}
			wg.Wait()
		}
	}

	Context("sequential write and read", func() {
		It("normal", testSequentially)
		It("multiple times", multiple(testSequentially))
	})
	Context("concurrent write and read", func() {
		It("normal", testConcurrently)
		It("multiple times", multiple(testConcurrently))
	})
})

func makeLinesChain(lines map[string]bool, repeated int) <-chan string {
	linesChan := make(chan string, len(lines)*repeated)
	for i := 0; i < repeated; i++ {
		for line := range lines {
			l := line
			linesChan <- l
		}
	}
	close(linesChan)
	return linesChan
}

func writeBuffer(lines <-chan string, buffer io.Writer) error {
	line, ok := <-lines
	for ok {
		_, err := buffer.Write([]byte(fmt.Sprintf("%s\r\n", line)))
		if err != nil {
			return err
		}
		line, ok = <-lines
	}
	return nil
}

func check(lines map[string]bool, result []byte, seperator string, repeated int) {
	resultMap := make(map[string]int)
	for _, line := range strings.Split(strings.TrimRight(string(result), "\r\n"), seperator) {
		resultMap[line]++
	}

	for line := range lines {
		Expect(resultMap[line]).To(Equal(repeated))
	}

	for line := range resultMap {
		Expect(lines[line]).To(BeTrue())
	}
}

func readTimeout(reader io.Reader, timeout time.Duration) ([]byte, error) {
	errChan := make(chan error)
	buffer := buffer.Buffer{}
	go func() {
		var err error
		var ln int
		chunk := make([]byte, 2)
		for {
			ln, err = reader.Read(chunk)
			if err != nil {
				break
			}
			_, err = buffer.Write(chunk[:ln])
			if err != nil {
				break
			}
		}
		errChan <- err
	}()

	select {
	case err := <-errChan:
		return nil, err
	case <-time.Tick(timeout):
		return buffer.Bytes(), nil
	}
}
