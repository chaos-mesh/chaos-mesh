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
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

var ErrNotPodPID = errors.New("pid is not PodPID")

type PodID struct {
	podID  UID
	sysPID SysPID
}

func (p PodID) ToID() string {
	return p.podID
}

type ChaosOnPOD interface {
	Injectable
	Recoverable
}

// PodHandler implements injecting & recovering on a kubernetes POD.
type PodHandler struct {
	Main   ChaosOnPOD
	Logger logr.Logger
}

func NewPodHandler(logger logr.Logger, main ChaosOnPOD) PodHandler {
	return PodHandler{
		Main:   main,
		Logger: logr.New(logger.GetSink()),
	}
}

func (p *PodHandler) Inject(pid PID) error {
	podPID, ok := pid.(PodID)
	if !ok {
		return ErrNotPodPID
	}

	err := p.Main.Inject(podPID.sysPID)
	return err
}

func (p *PodHandler) Recover(pid PID) error {
	podPID, ok := pid.(PodID)
	if !ok {
		return ErrNotPodPID
	}

	err := p.Main.Recover(podPID.sysPID)
	return err
}
