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

package timer

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Timer represents a running timer process
type Timer struct {
	Stdin    io.WriteCloser
	TimeChan chan TimeResult
	process  *exec.Cmd
	pid      int
}

// Pid returns the pid of timer
func (timer *Timer) Pid() int {
	return timer.pid
}

// TimeResult represents a get time result with an error
type TimeResult struct {
	Time  *time.Time
	Error error
}

// StartTimer will start a timer process
func StartTimer() (*Timer, error) {
	process := exec.Command("./bin/test/timer")

	stdout, err := process.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stdoutScanner := bufio.NewScanner(stdout)

	output := make(chan TimeResult)
	go func() {
		for stdoutScanner.Scan() {
			line := stdoutScanner.Text()
			sections := strings.Split(line, " ")

			sec, err := strconv.ParseInt(sections[0], 10, 64)
			if err != nil {
				output <- TimeResult{
					Error: err,
				}
			}
			nsec, err := strconv.ParseInt(sections[1], 10, 64)
			if err != nil {
				output <- TimeResult{
					Error: err,
				}
			}

			t := time.Unix(sec, nsec)
			output <- TimeResult{
				Time: &t,
			}
		}
	}()

	stdin, err := process.StdinPipe()
	if err != nil {
		return nil, err
	}

	err = process.Start()
	if err != nil {
		return nil, err
	}

	return &Timer{
		Stdin:    stdin,
		TimeChan: output,
		pid:      process.Process.Pid,
		process:  process,
	}, nil
}

// GetTime will run `time.Now()` in timer
func (timer *Timer) GetTime() (*time.Time, error) {
	_, err := fmt.Fprintf(timer.Stdin, "\n")
	if err != nil {
		return nil, err
	}

	result := <-timer.TimeChan
	if result.Error != nil {
		return nil, result.Error
	}

	return result.Time, nil
}

// Stop stops the process
func (timer *Timer) Stop() error {
	_, err := fmt.Fprintf(timer.Stdin, "STOP\n")

	return err
}
