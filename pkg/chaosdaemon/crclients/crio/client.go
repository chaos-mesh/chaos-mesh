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

package crio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"
)

const (
	InspectContainersEndpoint = "/containers"

	crioProtocolPrefix    = "cri-o://"
	maxUnixSocketPathSize = len(syscall.RawSockaddrUnix{}.Path)
)

// CrioClient can get information from docker
type CrioClient struct {
	client     *http.Client
	socketPath string
}

// FormatContainerID strips protocol prefix from the container ID
func (c CrioClient) FormatContainerID(ctx context.Context, containerID string) (string, error) {
	if len(containerID) < len(crioProtocolPrefix) {
		return "", fmt.Errorf("container id %s is not a crio container id", containerID)
	}
	if containerID[0:len(crioProtocolPrefix)] != crioProtocolPrefix {
		return "", fmt.Errorf("expected %s but got %s", crioProtocolPrefix, containerID[0:len(crioProtocolPrefix)])
	}
	return containerID[len(crioProtocolPrefix):], nil
}

// most of these implementations are copied from https://github.com/cri-o/cri-o/blob/master/internal/client/client.go

// GetPidFromContainerID fetches PID according to container id
func (c CrioClient) GetPidFromContainerID(ctx context.Context, containerID string) (uint32, error) {
	id, err := c.FormatContainerID(ctx, containerID)
	if err != nil {
		return 0, err
	}

	req, err := c.getRequest(ctx, InspectContainersEndpoint+"/"+id)
	if err != nil {
		return 0, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	cInfo := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&cInfo); err != nil {
		return 0, err
	}

	pid := cInfo["pid"]
	if pid, ok := pid.(float64); ok {
		return uint32(pid), nil
	}

	return 0, errors.New("fail to get pid from container info")
}

// ContainerKillByContainerID kills container according to container id
func (c CrioClient) ContainerKillByContainerID(ctx context.Context, containerID string) error {
	pid, err := c.GetPidFromContainerID(ctx, containerID)
	if err != nil {
		return err
	}
	return syscall.Kill(int(pid), syscall.SIGKILL)
}

func New(socketPath string) (*CrioClient, error) {
	tr := new(http.Transport)
	if err := configureUnixTransport(tr, "unix", socketPath); err != nil {
		return nil, err
	}
	c := &http.Client{
		Transport: tr,
	}
	return &CrioClient{
		client:     c,
		socketPath: socketPath,
	}, nil
}

func configureUnixTransport(tr *http.Transport, proto, addr string) error {
	if len(addr) > maxUnixSocketPathSize {
		return fmt.Errorf("unix socket path %q is too long", addr)
	}
	// No need for compression in local communications.
	tr.DisableCompression = true
	tr.DialContext = func(_ context.Context, _, _ string) (net.Conn, error) {
		return net.DialTimeout(proto, addr, 32*time.Second)
	}
	return nil
}

func (c *CrioClient) getRequest(ctx context.Context, path string) (*http.Request, error) {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	// For local communications over a unix socket, it doesn't matter what
	// the host is. We just need a valid and meaningful host name.
	req.Host = "crio"
	req.URL.Host = c.socketPath
	req.URL.Scheme = "http"
	req = req.WithContext(ctx)
	return req, nil
}
