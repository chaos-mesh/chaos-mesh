// Copyright 2022 Chaos Mesh Authors.
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

package tasks

import (
	"sync"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaoserr"
)

// TODO: is not PodID -> NotType[PodPID] after update to go 1.18
var ErrNotPodPID = errors.New("pid is not PodPID")

type PodID string

func (p PodID) ToID() string {
	return string(p)
}

// ChaosOnPOD stand for the inner process injector for container.
type ChaosOnPOD interface {
	Injectable
	Recoverable
}

type PodProcessMap struct {
	m      map[PodID]SysPID
	rwLock sync.RWMutex
}

func NewPodProcessMap() PodProcessMap {
	return PodProcessMap{
		m:      make(map[PodID]SysPID),
		rwLock: sync.RWMutex{},
	}
}

func (p *PodProcessMap) Read(podPID PodID) (SysPID, error) {
	p.rwLock.RLock()
	defer p.rwLock.RUnlock()
	sysPID, ok := p.m[podPID]
	if !ok {
		return SysPID(0), chaoserr.NotFound("SysPID")
	}
	return sysPID, nil
}

func (p *PodProcessMap) Write(podPID PodID, sysPID SysPID) {
	p.rwLock.Lock()
	defer p.rwLock.Unlock()
	p.m[podPID] = sysPID
}

// PodHandler implements injecting & recovering on a kubernetes POD.
type PodHandler struct {
	PodProcessMap *PodProcessMap
	Main          ChaosOnPOD
	Logger        logr.Logger
}

func NewPodHandler(podProcessMap *PodProcessMap, main ChaosOnPOD, logger logr.Logger) PodHandler {
	return PodHandler{
		PodProcessMap: podProcessMap,
		Main:          main,
		Logger:        logr.New(logger.GetSink()),
	}
}

// Inject get the container process PID and Inject it with Main injector.
// Be careful about the error handling here.
func (p *PodHandler) Inject(pid PID) error {
	podPID, ok := pid.(PodID)
	if !ok {
		return ErrNotPodPID
	}
	if p.PodProcessMap == nil {
		return errors.New("PodProcessMap not init")
	}

	sysPID, err := p.PodProcessMap.Read(podPID)
	if err != nil {
		return err
	}

	err = p.Main.Inject(sysPID)
	return err
}

// Recover get the container process PID and Recover it with Main injector.
// Be careful about the error handling here.
func (p *PodHandler) Recover(pid PID) error {
	podPID, ok := pid.(PodID)
	if !ok {
		return ErrNotPodPID
	}
	if p.PodProcessMap == nil {
		return errors.New("PodProcessMap not init")
	}

	sysPID, err := p.PodProcessMap.Read(podPID)
	if err != nil {
		return err
	}

	err = p.Main.Recover(sysPID)
	return err
}
