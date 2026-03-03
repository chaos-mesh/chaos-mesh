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

	"github.com/chaos-mesh/chaos-mesh/pkg/cerr"
)

var ErrNotPodContainerName = cerr.NotType[PodContainerName]()

var ErrPodProcessMapNotInit = cerr.NotInit[map[PodContainerName]SysPID]().WrapName("PodContainerNameProcessMap").Err()

type PodContainerName string

func (p PodContainerName) ToID() string {
	return string(p)
}

// ChaosOnPOD stand for the inner process injector for container.
type ChaosOnPOD interface {
	Injectable
	Recoverable
}

type PodContainerNameProcessMap struct {
	m      map[PodContainerName]SysPID
	rwLock sync.RWMutex
}

func NewPodProcessMap() PodContainerNameProcessMap {
	return PodContainerNameProcessMap{
		m:      make(map[PodContainerName]SysPID),
		rwLock: sync.RWMutex{},
	}
}

func (p *PodContainerNameProcessMap) Read(PodContainerName PodContainerName) (SysPID, error) {
	p.rwLock.RLock()
	defer p.rwLock.RUnlock()
	sysPID, ok := p.m[PodContainerName]
	if !ok {
		return SysPID(0), ErrNotFoundSysID.WithStack().Err()
	}
	return sysPID, nil
}

func (p *PodContainerNameProcessMap) Write(PodContainerName PodContainerName, sysPID SysPID) {
	p.rwLock.Lock()
	defer p.rwLock.Unlock()
	p.m[PodContainerName] = sysPID
}

func (p *PodContainerNameProcessMap) Delete(podPID PodContainerName) {
	p.rwLock.Lock()
	defer p.rwLock.Unlock()

	delete(p.m, podPID)
}

// PodHandler implements injecting & recovering on a kubernetes POD.
type PodHandler struct {
	PodProcessMap *PodContainerNameProcessMap
	SubProcess    ChaosOnPOD
	Logger        logr.Logger
}

func NewPodHandler(podProcessMap *PodContainerNameProcessMap, sub ChaosOnPOD, logger logr.Logger) PodHandler {
	return PodHandler{
		PodProcessMap: podProcessMap,
		SubProcess:    sub,
		Logger:        logr.New(logger.GetSink()),
	}
}

// Inject get the container process IsID and Inject it with major injector.
// Be careful about the error handling here.
func (p *PodHandler) Inject(id IsID) error {
	podPID, ok := id.(PodContainerName)
	if !ok {
		return ErrNotPodContainerName.WrapInput(id).Err()
	}
	if p.PodProcessMap == nil {
		return ErrPodProcessMapNotInit
	}

	sysPID, err := p.PodProcessMap.Read(podPID)
	if err != nil {
		return err
	}

	err = p.SubProcess.Inject(sysPID)
	return err
}

// Recover get the container process IsID and Recover it with major injector.
// Be careful about the error handling here.
func (p *PodHandler) Recover(id IsID) error {
	podPID, ok := id.(PodContainerName)
	if !ok {
		return ErrNotPodContainerName.WrapInput(id).Err()
	}
	if p.PodProcessMap == nil {
		return ErrPodProcessMapNotInit
	}

	sysPID, err := p.PodProcessMap.Read(podPID)
	if err != nil {
		return err
	}

	err = p.SubProcess.Recover(sysPID)
	return err
}
