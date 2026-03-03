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

package util

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/graph"
)

// ReadCommName returns the command name of process
func ReadCommName(pid int) (string, error) {
	f, err := os.Open(fmt.Sprintf("%s/%d/comm", bpm.DefaultProcPrefix, pid))
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// GetChildProcesses will return all child processes's pid. Include all generations.
// Only return error when `/proc` cannot be read.
func GetChildProcesses(ppid uint32, logger logr.Logger) ([]uint32, error) {
	procs, err := os.ReadDir(bpm.DefaultProcPrefix)
	if err != nil {
		return nil, errors.Wrapf(err, "read /proc")
	}

	type processPair struct {
		Pid  uint32
		Ppid uint32
	}

	pairs := make(chan processPair)
	done := make(chan bool)

	go func() {
		var wg sync.WaitGroup

		for _, proc := range procs {
			_, err := strconv.ParseUint(proc.Name(), 10, 32)
			if err != nil {
				continue
			}

			statusPath := bpm.DefaultProcPrefix + "/" + proc.Name() + "/stat"

			wg.Add(1)
			go func() {
				defer wg.Done()

				reader, err := os.Open(statusPath)
				if err != nil {
					logger.Error(err, "read status file error", "path", statusPath)
					return
				}
				defer reader.Close()

				var (
					pid    uint32
					comm   string
					state  string
					parent uint32
				)
				// according to procfs's man page
				fmt.Fscanf(reader, "%d %s %s %d", &pid, &comm, &state, &parent)

				pairs <- processPair{
					Pid:  pid,
					Ppid: parent,
				}
			}()
		}

		wg.Wait()
		done <- true
	}()

	processGraph := graph.NewGraph()
	for {
		select {
		case pair := <-pairs:
			processGraph.Insert(pair.Ppid, pair.Pid)
		case <-done:
			return processGraph.Flatten(ppid, logger), nil
		}
	}
}

func EncodeOutputToError(output []byte, err error) error {
	return errors.Errorf("error code: %v, msg: %s", err, string(output))
}
