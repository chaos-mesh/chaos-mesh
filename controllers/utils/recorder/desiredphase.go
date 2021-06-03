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

package recorder

type Deleted struct {
}

func (d Deleted) Type() string {
	return "Normal"
}

func (d Deleted) Reason() string {
	return "Deleted"
}

func (d Deleted) Message() string {
	return "Experiment has been deleted"
}

type TimeUp struct {
}

func (t TimeUp) Type() string {
	return "Normal"
}

func (t TimeUp) Reason() string {
	return "TimeUp"
}

func (t TimeUp) Message() string {
	return "Time up according to the duration"
}

type Paused struct {
}

func (p Paused) Type() string {
	return "Normal"
}

func (p Paused) Reason() string {
	return "Paused"
}

func (p Paused) Message() string {
	return "Experiment has been paused"
}

type Started struct {
}

func (p Started) Type() string {
	return "Normal"
}

func (p Started) Reason() string {
	return "Started"
}

func (p Started) Message() string {
	return "Experiment has started"
}

func init() {
	register(Deleted{}, TimeUp{}, Paused{}, Started{})
}
