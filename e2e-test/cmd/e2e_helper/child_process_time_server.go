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

package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type childProcessTimeServer interface {
	Start(ctx context.Context) error
	Time() (*time.Time, error)
}

var _ childProcessTimeServer = (*defaultChildProcessTimeServer)(nil)

// the default implementation would create another instance of e2e-helper, but bind with a different port
type defaultChildProcessTimeServer struct {
}

// Start implements childProcessTimeServer
func (s *defaultChildProcessTimeServer) Start(ctx context.Context) error {
	selfPath, err := os.Executable()
	if err != nil {
		return err
	}
	return exec.CommandContext(ctx, selfPath, "--port", "8081").Run()
}

// Time implements childProcessTimeServer
func (s *defaultChildProcessTimeServer) Time() (*time.Time, error) {
	resp, err := http.Get("http://localhost:8081/time")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	time, err := time.Parse(time.RFC3339Nano, string(bytes))
	if err != nil {
		return nil, err
	}
	return &time, nil
}
